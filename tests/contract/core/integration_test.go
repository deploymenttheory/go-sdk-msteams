package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	relationship "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	teammanagement "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

const (
	integratedGroup      = "11111111-1111-4111-8111-111111111111"
	integratedUser       = "22222222-2222-4222-8222-222222222222"
	integratedMembership = "opaque/member+identity="
)

// These are authored offline protocol vectors, not tenant response captures.
// Their primary-source provenance and exact qualification limits are recorded in
// docs/implementation/core/integrated-qualification.md.
func coreExchange(method, path, body string, status int, reply string) contracttest.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{
		Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: payload},
		Reply:   contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)},
	}
}

func ownerExchanges(status int) []contracttest.Exchange {
	base := "/v1.0/groups/" + integratedGroup
	return []contracttest.Exchange{
		coreExchange("GET", base+"/members?$select=id", "", 200, `{"value":[]}`),
		coreExchange("POST", base+"/members/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/directoryObjects/`+integratedUser+`"}`, 204, ""),
		coreExchange("GET", base+"/owners?$select=id", "", 200, `{"value":[]}`),
		coreExchange("POST", base+"/owners/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/users/`+integratedUser+`"}`, status, ""),
	}
}

func runIntegratedCore(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "integrated_core"
	c.Scope = "service"
	c.Source = "tests/contract/core/integration_test.go"
	c.Mode = auth.Delegated
	if len(c.Mappings) == 0 {
		c.Mappings = []contracttest.Mapping{{Command: "Add-TeamUser", ParameterSet: "__AllParameterSets"}}
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("integrated core protocol differed: setup=%v differences=%+v", err, result.Differences())
	}
}

func TestUnitIntegratedCoreOwnerAndNativeMembershipIdentity(t *testing.T) {
	t.Parallel()
	steps := ownerExchanges(204)
	path := "/v1.0/teams/" + integratedGroup + "/members/opaque%2Fmember+identity="
	member := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"` + integratedMembership + `","roles":["owner"],"userId":"` + integratedUser + `"}`
	steps = append(steps,
		coreExchange("GET", path, "", 200, member),
		coreExchange("DELETE", path, "", 204, ""),
	)
	runIntegratedCore(t, contracttest.Case{Name: "owner-and-native-member", Steps: steps, Want: contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			added, err := sdk.GraphAPI.TeamUserManagement.Add(ctx, integratedGroup, integratedUser, management.ChangeOptions{Role: management.Owner})
			if err != nil {
				return contracttest.Observation{Err: err}
			}
			if string(added.UserID) != integratedUser || len(added.Stages) != 5 || added.Stages[2].State != management.Succeeded || added.Stages[4].State != management.Succeeded {
				t.Error("owner composition lost its directory identity or successful writes")
			}
			value, response, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, integratedGroup, teammember.MembershipID(integratedMembership))
			if err != nil {
				return contracttest.Observation{Response: response, Err: err}
			}
			if string(value.ID) != integratedMembership || value.UserID == nil || string(*value.UserID) != integratedUser || len(value.Roles) != 1 || value.Roles[0] != teammember.Owner {
				t.Error("native membership identity was substituted with a directory identity")
			}
			response, err = sdk.GraphAPI.TeamMembers.DeleteByID(ctx, integratedGroup, value.ID)
			return contracttest.Observation{Response: response, Err: err}
		},
	})
}

func TestUnitIntegratedCorePartialOwnerFailureAndExplicitRecovery(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name          string
		cleanupStatus int
	}{
		{name: "confirmed-recovery", cleanupStatus: 204},
		{name: "denied-recovery-is-not-completion", cleanupStatus: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := ownerExchanges(403)
			base := "/v1.0/groups/" + integratedGroup + "/members"
			steps = append(steps, coreExchange("DELETE", base+"/"+integratedUser+"/$ref", "", tc.cleanupStatus, ""))
			want := contracttest.Expected{Status: 403, Error: auth.ErrForbidden}
			if tc.cleanupStatus == 204 {
				steps = append(steps, coreExchange("GET", base+"?$select=id", "", 200, `{"value":[]}`))
				want = contracttest.Expected{Status: 200}
			}
			runIntegratedCore(t, contracttest.Case{Name: tc.name, Steps: steps, Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					partial, err := sdk.GraphAPI.TeamUserManagement.Add(ctx, integratedGroup, integratedUser, management.ChangeOptions{Role: management.Owner})
					if !errors.Is(err, auth.ErrForbidden) || string(partial.UserID) != integratedUser || len(partial.Stages) != 5 {
						t.Error("failed owner grant lost its primary failure, identity or stages")
						return contracttest.Observation{Err: err}
					}
					if partial.Stages[2].Action != management.AddMemberReference || partial.Stages[2].State != management.Succeeded || partial.Stages[4].State != management.Unconfirmed {
						t.Error("partial write was silently rolled back or reported complete")
					}
					// This fixture's empty preflight and confirmed204 establish ownership
					// of this relationship only. Never delete the user or backing group.
					response, err := sdk.GraphAPI.GroupMembers.RemoveReference(ctx, integratedGroup, partial.UserID)
					if err != nil {
						return contracttest.Observation{Response: response, Err: err}
					}
					remaining, response, err := sdk.GraphAPI.GroupMembers.List(ctx, integratedGroup, relationship.ListOptions{Select: []string{"id"}})
					if err == nil && (len(remaining.Items) != 0 || remaining.NextLink != "") {
						t.Error("recovery absence was not established by a complete independent read")
					}
					return contracttest.Observation{Response: response, Err: err}
				},
			})
		})
	}
}

func TestUnitIntegratedCoreCreationOwnershipControlsCompensation(t *testing.T) {
	t.Parallel()
	for _, owned := range []bool{true, false} {
		name, set := "supplied-group-retained", "MigrateGroup"
		steps := []contracttest.Exchange{}
		if owned {
			name, set = "created-group-compensated", "CreateTeam"
			body := `{"displayName":"Offline fixture","mailNickname":"offline-fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false,"visibility":"Private","creationOptions":["ExchangeProvisioningFlags:3552"],"owners@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + integratedUser + `"],"members@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + integratedUser + `"]}`
			steps = append(steps, coreExchange("POST", "/v1.0/groups", body, 201, `{"id":"`+integratedGroup+`"}`))
		}
		steps = append(steps, coreExchange("GET", "/v1.0/groups/"+integratedGroup, "", 403, ""))
		if owned {
			steps = append(steps, coreExchange("DELETE", "/v1.0/groups/"+integratedGroup, "", 204, ""))
		}
		t.Run(name, func(t *testing.T) {
			runIntegratedCore(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "New-Team", ParameterSet: set}}, Steps: steps,
				Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					var result teammanagement.CreationResult
					var err error
					if owned {
						alias := "offline-fixture"
						result, err = sdk.GraphAPI.TeamManagement.Create(ctx, teammanagement.CreateRequest{DisplayName: "Offline fixture", MailNickname: &alias, Owner: integratedUser})
					} else {
						result, err = sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, integratedGroup, teammanagement.CreationOptions{})
					}
					if result.ID != integratedGroup || result.CreatedGroupOwned != owned || result.Value != nil || !errors.Is(result.PrimaryError, auth.ErrForbidden) {
						t.Error("creation failure lost resource ownership, primary failure or partial identity")
					}
					if owned {
						if result.CreatedGroup == nil || result.CreatedGroup.ID != integratedGroup || result.Cleanup.Disposition != teammanagement.CleanupDeleted || result.Cleanup.Error != nil || result.Cleanup.Response == nil || result.Cleanup.Response.StatusCode() != 204 {
							t.Error("confirmed owned-group compensation was not preserved")
						}
					} else if result.Cleanup.Disposition != teammanagement.CleanupSuppliedGroup || result.Cleanup.Response != nil {
						t.Error("caller-supplied group was selected for cleanup")
					}
					return contracttest.Observation{Response: result.PrimaryResponse, Err: err}
				},
			})
		})
	}
}
