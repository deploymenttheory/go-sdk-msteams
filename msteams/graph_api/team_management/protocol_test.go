package teammanagement_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"testing/synctest"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

// Authored protocol examples from the cited primitive contracts. These exercise
// SDK composition and retained receipts; they are not captured tenant evidence.
func protocol(method, path, request string, status int, body, source string) contracttest.Exchange {
	step := denied(method, path, request)
	step.Reply.Status = status
	step.Reply.Body = []byte(body)
	step.Reply.ProtocolSources = []string{source}
	return step
}

const (
	groupGet     = "https://learn.microsoft.com/en-us/graph/api/group-get?view=graph-rest-1.0"
	teamGet      = "https://learn.microsoft.com/en-us/graph/api/team-get?view=graph-rest-1.0"
	groupBody    = `{"id":"` + groupID + `","displayName":"Group name","description":"Group description","mailNickname":"fixture","visibility":"Private","futureGroup":false}`
	settingsBody = `{"memberSettings":{},"guestSettings":{},"messagingSettings":{},"funSettings":{},"discoverySettings":{}}`
)

func teamBody(archived bool) string {
	if archived {
		return `{"id":"` + groupID + `","displayName":"Team name","isArchived":true,"internalId":"native-internal","futureTeam":null}`
	}
	return `{"id":"` + groupID + `","displayName":"Team name","isArchived":false,"internalId":"native-internal","futureTeam":null}`
}
func reads(archived bool) []contracttest.Exchange {
	return []contracttest.Exchange{protocol("GET", "/v1.0/groups/"+groupID, "", 200, groupBody, groupGet), protocol("GET", "/v1.0/teams/"+groupID, "", 200, teamBody(archived), teamGet)}
}
func checkManaged(t *testing.T, v *management.ManagedTeam) {
	t.Helper()
	if v == nil || v.ID != groupID || v.DisplayName == nil || *v.DisplayName != "Group name" || v.InternalID == nil || *v.InternalID != "native-internal" || v.GroupSnapshot == nil || v.TeamSnapshot == nil || string(v.GroupSnapshot.RawFields["futureGroup"]) != "false" || string(v.TeamSnapshot.RawFields["futureTeam"]) != "null" {
		t.Fatal("paired read lost group precedence, native identity or raw snapshots")
	}
}
func TestUnitManagementSuccessfulPairedReadAndPartialUpdate(t *testing.T) {
	t.Parallel()
	runCase(t, "Get-Team", contracttest.Case{Name: "paired-positive", Steps: reads(false), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, err := sdk.GraphAPI.TeamManagement.GetByID(ctx, groupID)
		if err == nil {
			checkManaged(t, v.Value)
			if len(v.Stages) != 2 || v.Stages[0].State != management.Succeeded || v.Stages[1].State != management.Succeeded {
				t.Fatal("paired stage outcomes differ")
			}
		}
		return contracttest.Observation{Response: v.Response, Err: err}
	}, Want: contracttest.Expected{Status: 200}})
	for _, deniedTeam := range []bool{false, true} {
		patchGroup := protocol("PATCH", "/v1.0/groups/"+groupID, `{"description":""}`, 204, "", "https://learn.microsoft.com/en-us/graph/api/group-update?view=graph-rest-1.0")
		patchTeam := protocol("PATCH", "/v1.0/teams/"+groupID, settingsBody, 204, "", "https://learn.microsoft.com/en-us/graph/api/team-update?view=graph-rest-1.0")
		steps := []contracttest.Exchange{patchGroup, patchTeam}
		var wantErr error
		status := 200
		if deniedTeam {
			steps[1].Reply.Status = 403
			steps[1].Reply.Body = nil
			wantErr = auth.ErrForbidden
			status = 403
		} else {
			steps = append(steps, reads(false)...)
		}
		runCase(t, "Set-Team", contracttest.Case{Name: "group-then-team", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			empty := ""
			v, err := sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, management.UpdateRequest{Properties: group.UpdateRequest{Description: &empty}})
			if deniedTeam {
				if v.Value != nil || v.ID != groupID || len(v.Stages) != 2 || v.Stages[0].State != management.Succeeded || v.Stages[1].State != management.Unconfirmed {
					t.Fatal("partial write lost confirmed prefix")
				}
			} else if err == nil {
				checkManaged(t, v.Value)
			}
			return contracttest.Observation{Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Status: status, Error: wantErr}})
	}
}
func TestUnitManagementSuccessfulOwnedAndSuppliedGroupCreation(t *testing.T) {
	t.Parallel()
	for _, owned := range []bool{true, false} {
		set := "MigrateGroup"
		var steps []contracttest.Exchange
		if owned {
			set = "CreateTeam"
			request := `{"displayName":"Fixture","mailNickname":"fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false,"visibility":"Private","creationOptions":["ExchangeProvisioningFlags:3552"],"owners@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + ownerID + `"],"members@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + ownerID + `"]}`
			create := protocol("POST", "/v1.0/groups", request, 201, groupBody, "https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0")
			create.Reply.ProtocolSources = append(create.Reply.ProtocolSources, "https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0")
			steps = append(steps, create)
		}
		steps = append(steps, protocol("GET", "/v1.0/groups/"+groupID, "", 200, groupBody, groupGet), protocol("PUT", "/v1.0/groups/"+groupID+"/team", settingsBody, 201, teamBody(false), "https://learn.microsoft.com/en-us/graph/api/team-put-teams?view=graph-rest-1.0"))
		steps = append(steps, reads(false)...)
		creationCase(t, set, contracttest.Case{Name: "positive-" + map[bool]string{true: "owned", false: "supplied"}[owned], Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var v management.CreationResult
			var err error
			if owned {
				v, err = sdk.GraphAPI.TeamManagement.Create(ctx, creationRequest())
			} else {
				v, err = sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, groupID, management.CreationOptions{})
			}
			if err == nil {
				checkManaged(t, v.Value)
				if v.CreatedGroupOwned != owned || v.Teamified == nil || v.Teamified.ID != groupID {
					t.Fatal("ownership or teamification receipt differs")
				}
				if owned {
					if v.CreatedGroup == nil || v.CreatedGroupResponse == nil || v.OwnerID != ownerID || v.Cleanup.Disposition != management.CleanupNotRequired {
						t.Fatal("owned creation metadata lost")
					}
				} else if v.CreatedGroup != nil || v.Cleanup.Disposition != management.CleanupSuppliedGroup {
					t.Fatal("supplied group ownership inferred")
				}
			}
			return contracttest.Observation{Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200}})
	}
}
func TestUnitManagementArchiveRequiresIndependentReadback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const operationID = "33333333-3333-4333-8333-333333333333"
		for _, requested := range []bool{true, false} {
			for _, confirmed := range []bool{true, false} {
				steps := reads(!requested)
				actionName := "archive"
				if !requested {
					actionName = "unarchive"
				}
				action := protocol("POST", "/v1.0/teams/"+groupID+"/"+actionName, "", 202, "", "https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0")
				action.Reply.Headers = http.Header{"Location": {"/v1.0/teams/" + groupID + "/operations/" + operationID}}
				steps = append(steps, action, protocol("GET", "/v1.0/teams/"+groupID+"/operations/"+operationID, "", 200, `{"id":"`+operationID+`","status":"succeeded"}`, "https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0"))
				after := requested
				if !confirmed {
					after = !requested
				}
				steps = append(steps, reads(after)...)
				var wantErr error
				if !confirmed {
					wantErr = management.ErrArchivedStateUnconfirmed
				}
				runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: "archive-readback", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					v, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, requested, management.ArchiveStateOptions{})
					if v.Confirmed != confirmed || v.Before == nil || v.After == nil || v.Before == v.After || v.Accepted.Operation == nil || v.Snapshot.Handle.ID != operationID || v.Accepted.Response.StatusCode() != 202 {
						t.Fatal("archive conflated acceptance, completion or readback")
					}
					return contracttest.Observation{Response: v.Response, Err: err}
				}, Want: contracttest.Expected{Status: 200, Error: wantErr}})
			}
		}
	})
}
func TestUnitManagementTenantHydrationCorrelatesAndRetainsPartialFailure(t *testing.T) {
	t.Parallel()
	const secondID = "00000000-0000-4000-8000-000000000003"
	q := url.Values{"$filter": {"resourceProvisioningOptions/Any(x:x eq 'Team')"}, "$top": {"999"}, "$count": {"true"}}
	for _, partial := range []bool{false, true} {
		discovery := protocol("GET", "/beta/groups?"+q.Encode(), "", 200, `{"value":[`+groupBody+`,{"id":"`+secondID+`","displayName":"Second"}]}`, "https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-beta")
		discovery.Request.Headers.Set("ConsistencyLevel", "eventual")
		second := `{"id":"2","status":200,"body":{"id":"` + secondID + `","isArchived":false}}`
		if partial {
			second = `{"id":"2","status":403,"body":{"error":{"code":"Forbidden"}}}`
		}
		batch := protocol("POST", "/v1.0/$batch", `{"requests":[{"id":"1","method":"GET","url":"/teams/`+groupID+`"},{"id":"2","method":"GET","url":"/teams/`+secondID+`"}]}`, 200, `{"responses":[`+second+`,{"id":"1","status":200,"body":`+teamBody(false)+`}]}`, "https://learn.microsoft.com/en-us/graph/json-batching")
		batch.Reply.ProtocolSources = append(batch.Reply.ProtocolSources, teamGet)
		var wantErr error
		if partial {
			wantErr = management.ErrListingIncomplete
		}
		listCase(t, contracttest.Case{Name: "hydrate-correlated", Steps: []contracttest.Exchange{discovery, batch}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
			if len(v.Items) > 0 {
				checkManaged(t, v.Items[0])
			}
			expected := 2
			if partial {
				expected = 1
			}
			if len(v.Items) != expected || !v.DiscoveryComplete || v.Complete == partial || len(v.Outcomes) != 2 || v.Items[0].ID != groupID || v.Outcomes[1].ID != secondID || len(v.Outcomes[1].Subresponses) != 1 {
				t.Fatal("batch correlation or partial result lost")
			}
			if partial && (!errors.Is(v.Outcomes[1].Err, auth.ErrForbidden) || v.Outcomes[1].Subresponses[0].Status != 403) {
				t.Fatal("per-item denial lost")
			}
			return contracttest.Observation{Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: wantErr}})
	}
}

func TestUnitManagementUserDiscoveryAndFailedLaterReadRetainPrefix(t *testing.T) {
	t.Parallel()
	const secondID = "00000000-0000-4000-8000-000000000003"
	for _, partial := range []bool{false, true} {
		joined := protocol("GET", "/v1.0/users/"+ownerID+"/joinedTeams", "", 200, `{"value":[{"id":"`+groupID+`"},{"id":"`+secondID+`"}]}`, "https://learn.microsoft.com/en-us/graph/api/user-list-joinedteams?view=graph-rest-1.0")
		steps := append([]contracttest.Exchange{joined}, reads(false)...)
		secondGroup := protocol("GET", "/v1.0/groups/"+secondID, "", 200, `{"id":"`+secondID+`","displayName":"Second"}`, groupGet)
		steps = append(steps, secondGroup)
		last := protocol("GET", "/v1.0/teams/"+secondID, "", 200, `{"id":"`+secondID+`","isArchived":false}`, teamGet)
		var wantErr error
		if partial {
			last.Reply.Status = 403
			last.Reply.Body = nil
			wantErr = management.ErrListingIncomplete
		}
		steps = append(steps, last)
		listCase(t, contracttest.Case{Name: "joined-paired-read", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			user := ownerID
			v, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{User: &user})
			if len(v.Items) > 0 {
				checkManaged(t, v.Items[0])
			}
			expected := 2
			if partial {
				expected = 1
			}
			if v.UserID != ownerID || len(v.Items) != expected || len(v.Outcomes) != 2 || !v.DiscoveryComplete || v.Complete == partial || v.Items[0].ID != groupID {
				t.Fatal("user discovery lost identity, prefix or completeness")
			}
			if partial && (!errors.Is(v.Outcomes[1].Err, auth.ErrForbidden) || v.Outcomes[1].Value != nil) {
				t.Fatal("later denied paired read fabricated a complete team")
			}
			return contracttest.Observation{Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Status: map[bool]int{true: 403, false: 200}[partial], Error: wantErr}})
	}
}
