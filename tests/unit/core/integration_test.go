package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"testing/synctest"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
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

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func coreExchange(method, path, body string, status int, reply string) testutil.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{
		Request: testutil.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: payload},
		Reply:   testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)},
	}
}

func ownerExchanges(status int) []testutil.Exchange {
	base := "/v1.0/groups/" + integratedGroup
	return []testutil.Exchange{
		coreExchange("GET", base+"/members?$select=id", "", 200, `{"value":[]}`),
		coreExchange("POST", base+"/members/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/directoryObjects/`+integratedUser+`"}`, 204, ""),
		coreExchange("GET", base+"/owners?$select=id", "", 200, `{"value":[]}`),
		coreExchange("POST", base+"/owners/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/users/`+integratedUser+`"}`, status, ""),
	}
}

func runIntegratedCore(t *testing.T, c testutil.Case) {
	t.Helper()

	c.Mode = auth.Delegated

	result, err := testutil.Run(t.Context(), c)
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
	runIntegratedCore(t, testutil.Case{Name: "owner-and-native-member", Steps: steps, Want: testutil.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			added, err := sdk.GraphAPI.TeamUserManagement.Add(ctx, integratedGroup, integratedUser, management.ChangeOptions{Role: management.Owner})
			if err != nil {
				return testutil.Observation{Err: err}
			}
			if string(added.UserID) != integratedUser || len(added.Stages) != 5 || added.Stages[2].State != management.Succeeded || added.Stages[4].State != management.Succeeded {
				t.Error("owner composition lost its directory identity or successful writes")
			}
			value, response, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, integratedGroup, teammember.MembershipID(integratedMembership))
			if err != nil {
				return testutil.Observation{Response: response, Err: err}
			}
			if string(value.ID) != integratedMembership || value.UserID == nil || string(*value.UserID) != integratedUser || len(value.Roles) != 1 || value.Roles[0] != teammember.Owner {
				t.Error("native membership identity was substituted with a directory identity")
			}
			response, err = sdk.GraphAPI.TeamMembers.DeleteByID(ctx, integratedGroup, value.ID)
			return testutil.Observation{Response: response, Err: err}
		},
	})
}

// A channel conversation membership is independently addressed even when its
// user already has a different native team conversation membership. These
// authored vectors use Graph's documented create/role-patch/delete/list shapes.
func TestUnitIntegratedCoreChannelMembershipOwnershipAndRecovery(t *testing.T) {
	t.Parallel()
	const channelID = "19:private/fixture@thread.tacv2"
	const channelMembership = "opaque/channel+membership="
	teamPath := "/v1.0/teams/" + integratedGroup + "/members/" + url.PathEscape(integratedMembership)
	collection := "/beta/teams/" + integratedGroup + "/channels/" + url.PathEscape(channelID) + "/members"
	byID := collection + "/" + url.PathEscape(channelMembership)
	teamBody := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"` + integratedMembership + `","roles":["owner"],"userId":"` + integratedUser + `"}`
	memberBody := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"` + channelMembership + `","roles":[],"userId":"` + integratedUser + `"}`
	ownerBody := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"` + channelMembership + `","roles":["owner"],"userId":"` + integratedUser + `"}`
	createBody := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"` + auth.GraphAudience + `/beta/users('` + integratedUser + `')"}`
	for _, tc := range []struct {
		name                           string
		promotionStatus, cleanupStatus int
	}{
		{name: "successful-channel-role-and-independent-team", promotionStatus: 200, cleanupStatus: 204},
		{name: "denied-promotion-owned-membership-recovered", promotionStatus: 403, cleanupStatus: 204},
		{name: "denied-promotion-denied-recovery-is-not-completion", promotionStatus: 403, cleanupStatus: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			promotionBody := ""
			if tc.promotionStatus == 200 {
				promotionBody = ownerBody
			}
			steps := []testutil.Exchange{
				coreExchange("GET", teamPath, "", 200, teamBody),
				coreExchange("POST", collection, createBody, 201, memberBody),
				coreExchange("PATCH", byID, `{"roles":["owner"]}`, tc.promotionStatus, promotionBody),
				coreExchange("DELETE", byID, "", tc.cleanupStatus, ""),
			}

			want := testutil.Expected{Status: 403, Error: auth.ErrForbidden}
			if tc.cleanupStatus == 204 {
				steps = append(steps,
					coreExchange("GET", collection, "", 200, `{"value":[]}`),
					coreExchange("GET", teamPath, "", 200, teamBody),
				)
				want = testutil.Expected{Status: 200}
			}
			runIntegratedCore(t, testutil.Case{Name: tc.name, Steps: steps, Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					teamMember, response, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, integratedGroup, teammember.MembershipID(integratedMembership))
					if err != nil {
						return testutil.Observation{Response: response, Err: err}
					}
					if string(teamMember.ID) != integratedMembership || teamMember.UserID == nil || string(*teamMember.UserID) != integratedUser {
						t.Error("team membership prerequisite lost its native or directory identity")
					}
					created, response, err := sdk.GraphAPI.ChannelMembers.Create(ctx, integratedGroup, channelID,
						channelmember.CreateRequest{UserID: channelmember.DirectoryUserID(integratedUser)})
					if err != nil {
						return testutil.Observation{Response: response, Err: err}
					}
					if string(created.ID) != channelMembership || string(created.ID) == string(teamMember.ID) || created.UserID == nil || string(*created.UserID) != integratedUser {
						t.Error("channel membership identity was substituted with team or directory identity")
					}
					promoted, response, err := sdk.GraphAPI.ChannelMembers.PromoteOwner(ctx, integratedGroup, channelID, created.ID)
					if tc.promotionStatus == 200 {
						if err != nil {
							return testutil.Observation{Response: response, Err: err}
						}
						if promoted.ID != created.ID || len(promoted.Roles) != 1 || promoted.Roles[0] != channelmember.Owner {
							t.Error("channel promotion lost identity or owner role")
						}
					} else if !errors.Is(err, auth.ErrForbidden) || promoted != nil {
						t.Error("denied channel promotion was reported as successful")
						return testutil.Observation{Response: response, Err: err}
					}
					// The confirmed creation receipt owns only this channel relationship.
					// Explicit recovery must not target the team membership, user or group.
					response, err = sdk.GraphAPI.ChannelMembers.DeleteByID(ctx, integratedGroup, channelID, created.ID)
					if err != nil {
						return testutil.Observation{Response: response, Err: err}
					}
					remaining, err := sdk.GraphAPI.ChannelMembers.ListAll(ctx, integratedGroup, channelID, channelmember.ListOptions{}, client.PagingOptions{})
					if err != nil {
						return testutil.Observation{Response: remaining.Response, Err: err}
					}
					if remaining.Pages != 1 || len(remaining.Items) != 0 {
						t.Error("channel recovery did not establish complete membership absence")
					}
					retained, response, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, integratedGroup, teamMember.ID)
					if err == nil && (retained.ID != teamMember.ID || retained.UserID == nil || string(*retained.UserID) != integratedUser || len(retained.Roles) != 1 || retained.Roles[0] != teammember.Owner) {
						t.Error("channel recovery altered independent team membership")
					}
					return testutil.Observation{Response: response, Err: err}
				},
			})
		})
	}
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
			want := testutil.Expected{Status: 403, Error: auth.ErrForbidden}
			if tc.cleanupStatus == 204 {
				steps = append(steps, coreExchange("GET", base+"?$select=id", "", 200, `{"value":[]}`))
				want = testutil.Expected{Status: 200}
			}
			runIntegratedCore(t, testutil.Case{Name: tc.name, Steps: steps, Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					partial, err := sdk.GraphAPI.TeamUserManagement.Add(ctx, integratedGroup, integratedUser, management.ChangeOptions{Role: management.Owner})
					if !errors.Is(err, auth.ErrForbidden) || string(partial.UserID) != integratedUser || len(partial.Stages) != 5 {
						t.Error("failed owner grant lost its primary failure, identity or stages")
						return testutil.Observation{Err: err}
					}
					if partial.Stages[2].Action != management.AddMemberReference || partial.Stages[2].State != management.Succeeded || partial.Stages[4].State != management.Unconfirmed {
						t.Error("partial write was silently rolled back or reported complete")
					}
					// This fixture's empty preflight and confirmed204 establish ownership
					// of this relationship only. Never delete the user or backing group.
					response, err := sdk.GraphAPI.GroupMembers.RemoveReference(ctx, integratedGroup, partial.UserID)
					if err != nil {
						return testutil.Observation{Response: response, Err: err}
					}
					remaining, response, err := sdk.GraphAPI.GroupMembers.List(ctx, integratedGroup, relationship.ListOptions{Select: []string{"id"}})
					if err == nil && (len(remaining.Items) != 0 || remaining.NextLink != "") {
						t.Error("recovery absence was not established by a complete independent read")
					}
					return testutil.Observation{Response: response, Err: err}
				},
			})
		})
	}
}

func TestUnitIntegratedCoreCreationOwnershipControlsCompensation(t *testing.T) {
	t.Parallel()
	for _, owned := range []bool{true, false} {
		name := "supplied-group-retained"
		steps := []testutil.Exchange{}
		if owned {
			name = "created-group-compensated"
			body := `{"displayName":"Offline fixture","mailNickname":"offline-fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false,"visibility":"Private","creationOptions":["ExchangeProvisioningFlags:3552"],"owners@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + integratedUser + `"],"members@odata.bind":["` + auth.GraphAudience + `/v1.0/users/` + integratedUser + `"]}`
			steps = append(steps, coreExchange("POST", "/v1.0/groups", body, 201, `{"id":"`+integratedGroup+`"}`))
		}
		steps = append(steps, coreExchange("GET", "/v1.0/groups/"+integratedGroup, "", 403, ""))
		if owned {
			steps = append(steps, coreExchange("DELETE", "/v1.0/groups/"+integratedGroup, "", 204, ""))
		}
		t.Run(name, func(t *testing.T) {
			runIntegratedCore(t, testutil.Case{Name: name, Steps: steps,
				Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
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
					return testutil.Observation{Response: result.PrimaryResponse, Err: err}
				},
			})
		})
	}
}

func TestUnitIntegratedCoreSharedChannelOperationAndRead(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, header, shareWire string
		shares                  *[]channel.InitialSharedTeam
	}{
		{name: "location-omitted-shares", header: "Location"},
		{name: "content-location-omitted-shares", header: "Content-Location"},
		{name: "content-location-empty-shares", header: "Content-Location", shareWire: `,"sharedWithTeams":[]`, shares: new([]channel.InitialSharedTeam{})},
		{name: "content-location-host-share", header: "Content-Location", shareWire: `,"sharedWithTeams":[{"id":"` + integratedGroup + `"}]`, shares: new([]channel.InitialSharedTeam{{ID: integratedGroup}})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				operationPath := "/beta/teams/" + integratedGroup + "/operations/offline-operation"
				createBody := `{"displayName":"Offline shared","membershipType":"shared","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"` + auth.GraphAudience + `/beta/users('` + integratedUser + `')","roles":["owner"]}]}`
				accepted := coreExchange("POST", "/beta/teams/"+integratedGroup+"/channels", createBody[:len(createBody)-1]+tc.shareWire+"}", 202, "")
				accepted.Reply.Headers.Set(tc.header, "/teams/"+integratedGroup+"/operations/offline-operation")
				status := `{"id":"offline-operation","status":"succeeded","targetResourceId":"offline-channel"}`
				steps := []testutil.Exchange{
					accepted,
					coreExchange("GET", operationPath, "", 200, status),
					coreExchange("GET", operationPath, "", 200, status),
					coreExchange("GET", "/beta/teams/"+integratedGroup+"/channels/offline-channel", "", 200, `{"id":"offline-channel","displayName":"Offline shared","membershipType":"shared"}`),
				}
				runIntegratedCore(t, testutil.Case{Name: tc.name,

					Steps: steps, Want: testutil.Expected{Status: 200},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						owner := integratedUser
						accepted, err := sdk.GraphAPI.Channels.Create(ctx, integratedGroup, channel.CreateRequest{DisplayName: "Offline shared", MembershipType: channel.Shared, OwnerUserID: &owner, SharedWithTeams: tc.shares})
						if err != nil {
							return testutil.Observation{Response: accepted.Response, Err: err}
						}
						if accepted.Channel != nil || accepted.Operation == nil || accepted.Operation.Snapshot().Status.State != client.OperationAccepted {
							t.Error("shared-channel acceptance was reported as completed channel creation")
							return testutil.Observation{Response: accepted.Response, Err: client.ErrOperation}
						}
						completed, err := accepted.Operation.Wait(ctx, client.WaitOptions{MaxPolls: 1})
						if err != nil {
							return testutil.Observation{Response: completed.Response, Err: err}
						}
						if completed.Status.State != client.OperationSucceeded || completed.Handle.ID != "offline-operation" {
							t.Error("shared-channel operation lost its state or operation identity")
						}
						operation, response, err := sdk.GraphAPI.TeamOperations.GetByID(ctx, integratedGroup, completed.Handle.ID)
						if err != nil {
							return testutil.Observation{Response: response, Err: err}
						}
						if operation.TargetResourceID == nil || *operation.TargetResourceID != "offline-channel" {
							t.Error("operation target identity was replaced with the operation identity")
							return testutil.Observation{Response: response, Err: client.ErrOperation}
						}
						value, response, err := sdk.GraphAPI.Channels.GetByID(ctx, integratedGroup, *operation.TargetResourceID, channel.ReadOptions{})
						if err == nil && (value.ID != "offline-channel" || value.MembershipType == nil || *value.MembershipType != channel.Shared) {
							t.Error("completed shared-channel identity or resource type was lost")
						}
						return testutil.Observation{Response: response, Err: err}
					},
				})
			})
		})
	}
}

func TestUnitIntegratedCoreSharedChannelContentLocationTrustBoundary(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, location, contentLocation string }{
		{name: "foreign-content-location", contentLocation: "https://foreign.example/beta/teams/" + integratedGroup + "/operations/offline-operation"},
		{name: "malformed-primary-is-not-replaced", location: "malformed", contentLocation: "/teams/" + integratedGroup + "/operations/offline-operation"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accepted := coreExchange("POST", "/beta/teams/"+integratedGroup+"/channels", `{"displayName":"Offline shared","membershipType":"shared"}`, 202, "")
			if tc.location != "" {
				accepted.Reply.Headers.Set("Location", tc.location)
			}
			accepted.Reply.Headers.Set("Content-Location", tc.contentLocation)
			runIntegratedCore(t, testutil.Case{Name: tc.name,
				Steps: []testutil.Exchange{accepted}, Want: testutil.Expected{Status: 202, Error: client.ErrOperation},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, err := sdk.GraphAPI.Channels.Create(ctx, integratedGroup, channel.CreateRequest{DisplayName: "Offline shared", MembershipType: channel.Shared})
					if value.Operation != nil || value.Channel != nil {
						t.Error("invalid operation location produced a usable handle")
					}
					return testutil.Observation{Response: value.Response, Err: err}
				},
			})
		})
	}
}

func TestUnitIntegratedCoreSharedChannelCancellationRetainsOperation(t *testing.T) {
	t.Parallel()
	accepted := coreExchange("POST", "/beta/teams/"+integratedGroup+"/channels", `{"displayName":"Offline shared","membershipType":"shared"}`, 202, "")
	accepted.Reply.Headers.Set("Content-Location", "/teams/"+integratedGroup+"/operations/offline-operation")
	runIntegratedCore(t, testutil.Case{Name: "cancelled-channel-wait",
		Steps: []testutil.Exchange{accepted}, Want: testutil.Expected{Status: 202, Error: context.Canceled},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, err := sdk.GraphAPI.Channels.Create(ctx, integratedGroup, channel.CreateRequest{DisplayName: "Offline shared", MembershipType: channel.Shared})
			if err != nil {
				return testutil.Observation{Response: value.Response, Err: err}
			}
			waiting, cancel := context.WithCancel(ctx)
			cancel()
			snapshot, err := value.Operation.Wait(waiting, client.WaitOptions{MaxPolls: 1})
			if snapshot.Handle.ID != "offline-operation" || snapshot.Status.State != client.OperationAccepted {
				t.Error("cancelled local wait discarded accepted remote identity")
			}
			return testutil.Observation{Response: snapshot.Response, Err: err}
		},
	})
}

func TestUnitIntegratedCoreInitialSharesValidateBeforeHTTP(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		kind   channel.MembershipType
		shares *[]channel.InitialSharedTeam
	}{
		{name: "invalid-team-id", kind: channel.Shared, shares: new([]channel.InitialSharedTeam{{ID: "not-a-team-id"}})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testutil.Case{Name: tc.name,
				Want: testutil.Expected{Error: client.ErrRequest},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, err := sdk.GraphAPI.Channels.Create(ctx, integratedGroup, channel.CreateRequest{DisplayName: "Offline shared", MembershipType: tc.kind, SharedWithTeams: tc.shares})
					return testutil.Observation{Response: value.Response, Err: err}
				},
			}
			result, err := testutil.Run(t.Context(), c)
			if err != nil || !result.Passed() {
				t.Fatalf("invalid initial sharing reached HTTP: %v %+v", err, result.Differences())
			}
		})
	}
}

// The documented successful shape is an array on shared creation. Other caller
// shapes remain server-validated; these denial vectors claim no acceptance.
func TestUnitIntegratedCoreInitialShareShapesRemainServerValidated(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		kind   channel.MembershipType
		shares *[]channel.InitialSharedTeam
		wire   string
	}{
		{name: "explicit-null", kind: channel.Shared, shares: new([]channel.InitialSharedTeam(nil)), wire: `{"displayName":"Offline shared","membershipType":"shared","sharedWithTeams":null}`},
		{name: "standard-with-share-input", kind: channel.Standard, shares: new([]channel.InitialSharedTeam{{ID: integratedGroup}}), wire: `{"displayName":"Offline shared","membershipType":"standard","sharedWithTeams":[{"id":"` + integratedGroup + `"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runIntegratedCore(t, testutil.Case{Name: tc.name,
				Steps: []testutil.Exchange{coreExchange("POST", "/beta/teams/"+integratedGroup+"/channels", tc.wire, 403, "")}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, err := sdk.GraphAPI.Channels.Create(ctx, integratedGroup, channel.CreateRequest{DisplayName: "Offline shared", MembershipType: tc.kind, SharedWithTeams: tc.shares})
					return testutil.Observation{Response: value.Response, Err: err}
				},
			})
		})
	}
}
