package channelusermanagement_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
)

const (
	teamID       = "11111111-1111-4111-8111-111111111111"
	userID       = "22222222-2222-4222-8222-222222222222"
	tenantID     = "33333333-3333-4333-8333-333333333333"
	channelID    = "19:fixture/part@thread.tacv2"
	memberID     = "opaque/returned+=#雪"
	channelsPath = "/beta/teams/" + teamID + "/channels"
)

func membersPath() string { return channelsPath + "/" + url.PathEscape(channelID) + "/members" }
func memberPath() string  { return membersPath() + "/" + url.PathEscape(memberID) }

func wire(method, path, body string, status int, response string) testutil.Exchange {
	headers := http.Header{"Accept": {"application/json"}}
	var data json.RawMessage
	if body != "" {
		data = json.RawMessage(body)
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: data},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func channelRead(kind string) testutil.Exchange {
	return wire("GET", channelsPath, "", 200, `{"value":[{"id":"`+channelID+`","displayName":"FIXTURE","membershipType":"`+kind+`"}]}`)
}

func nativeMember(roles string) string {
	return `{"id":"` + memberID + `","userId":"` + userID + `","email":"external@fixture.invalid","tenantId":"` + tenantID + `","roles":` + roles + `}`
}

func roster() testutil.Exchange {
	return wire("GET", membersPath(), "", 200, `{"value":[`+nativeMember(`[]`)+`]}`)
}

func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("channel-user contract differs: setup=%v differences=%+v", err, result.Differences())
	}
	return result
}

func changed(value management.ChangeResult, err error) testutil.Observation {
	actions, states := []management.Action{}, []management.State{}
	for _, stage := range value.Stages {
		actions = append(actions, stage.Action)
		states = append(states, stage.State)
	}
	return testutil.Observation{Value: map[string]any{"channel": value.ChannelID, "user": value.UserID, "membership": value.MembershipID, "actions": actions, "states": states}, Response: value.Response, Err: err}
}

func TestUnitAddRemoveAndOwnerCompositions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, command, kind, user, role, method, body, roles string
		add, lookup                                          bool
		tenant                                               *string
	}{
		{name: "add-private-id", command: "Add-TeamChannelUser", kind: "private", user: userID, method: "POST", roles: `[]`, add: true},
		{name: "add-private-upn", command: "Add-TeamChannelUser", kind: "private", user: "member@fixture.invalid", method: "POST", roles: `[]`, add: true, lookup: true},
		{name: "add-external-shared", command: "Add-TeamChannelUser", kind: "shared", user: userID, method: "POST", roles: `[]`, add: true, tenant: stringPointer(tenantID)},
		{name: "promote-private", command: "Add-TeamChannelUser", kind: "private", user: userID, role: "OwNeR", method: "PATCH", roles: `["owner"]`, add: true},
		{name: "demote-private", command: "Remove-TeamChannelUser", kind: "private", user: userID, role: "Owner", method: "PATCH", roles: `[]`},
		{name: "delete-private", command: "Remove-TeamChannelUser", kind: "private", user: userID, method: "DELETE"},
		{name: "promote-shared-email", command: "Add-TeamChannelUser", kind: "shared", user: "EXTERNAL@fixture.invalid", role: "Owner", method: "PATCH", roles: `["owner"]`, add: true},
		{name: "demote-shared-email", command: "Remove-TeamChannelUser", kind: "shared", user: "external@fixture.invalid", role: "owner", method: "PATCH", roles: `[]`},
		{name: "delete-shared-id", command: "Remove-TeamChannelUser", kind: "shared", user: userID, method: "DELETE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{channelRead(tc.kind)}
			actions := []string{"resolve-channel"}
			if tc.method == "POST" || tc.kind == "private" {
				if tc.lookup {
					steps = append(steps, wire("GET", "/v1.0/users/member@fixture.invalid?$select=id", "", 200, `{"id":"`+userID+`"}`))
				}
				actions = append(actions, "resolve-user")
			}
			path, body, status, response := memberPath(), "", 204, ""
			if tc.method == "POST" {
				path, status, response = membersPath(), 201, nativeMember(`[]`)
				body = `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"https://graph.microsoft.com/beta/users('` + userID + `')"`
				if tc.tenant != nil {
					body += `,"tenantId":"` + *tc.tenant + `"`
				}
				body += `}`
				actions = append(actions, "create-member")
			} else {
				steps = append(steps, roster())
				actions = append(actions, "resolve-membership")
				if tc.method == "PATCH" {
					body, status, response = `{"roles":`+tc.roles+`}`, 200, nativeMember(tc.roles)
					action := "demote-owner"
					if tc.add {
						action = "promote-owner"
					}
					actions = append(actions, action)
				} else {
					actions = append(actions, "delete-member")
				}
			}
			steps = append(steps, wire(tc.method, path, body, status, response))
			states := make([]string, len(actions))
			for i := range states {
				states[i] = "succeeded"
			}
			want, err := json.Marshal(map[string]any{"channel": channelID, "user": userID, "membership": memberID, "actions": actions, "states": states})
			if err != nil {
				t.Fatal(err)
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				options := management.ChangeOptions{Role: management.Role(tc.role), TenantID: tc.tenant}
				if tc.add {
					return changed(sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", tc.user, options))
				}
				return changed(sdk.GraphAPI.ChannelUserManagement.Remove(ctx, teamID, "fixture", tc.user, options))
			}, Want: testutil.Expected{Status: status, JSON: want}})
		})
	}
}

func stringPointer(value string) *string { return &value }

func TestUnitFailuresStopBeforeFurtherChanges(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"create", "promote", "demote", "delete"} {
		for _, status := range []int{403, 429, 503} {
			t.Run(action+"-"+strconv.Itoa(status), func(t *testing.T) {
				steps := []testutil.Exchange{channelRead("private")}
				method, path, body := "POST", membersPath(), `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"https://graph.microsoft.com/beta/users('`+userID+`')"}`
				role := management.Role("")
				if action != "create" {
					steps = append(steps, roster())
					path = memberPath()
					method = "PATCH"
					role = management.Owner
					body = `{"roles":[]}`
					if action == "promote" {
						body = `{"roles":["owner"]}`
					}
					if action == "delete" {
						method, body, role = "DELETE", "", ""
					}
				}
				steps = append(steps, wire(method, path, body, status, `{"error":{"code":"FixtureDenied","message":"private"}}`))
				run(t, testutil.Case{Name: action + "-" + strconv.Itoa(status), Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					options := management.ChangeOptions{Role: role}
					var value management.ChangeResult
					var err error
					if action == "create" || action == "promote" {
						value, err = sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", userID, options)
					} else {
						value, err = sdk.GraphAPI.ChannelUserManagement.Remove(ctx, teamID, "fixture", userID, options)
					}
					membership := memberID
					if action == "create" {
						membership = ""
					}
					if len(value.Stages) < 3 {
						t.Fatal("failed mutation omitted prerequisite stages")
					}
					if value.ChannelID != channelID || string(value.UserID) != userID || string(value.MembershipID) != membership || value.Stages[len(value.Stages)-1].State != management.Unconfirmed {
						t.Error("failed mutation lost resolved identity or unconfirmed final stage")
					}
					for _, stage := range value.Stages[:len(value.Stages)-1] {
						if stage.State != management.Succeeded {
							t.Error("failed mutation rewrote a successful prerequisite")
						}
					}
					return changed(value, err)
				}, Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureDenied"}}}})
			})
		}
	}
	for _, kind := range []string{"private", "shared"} {
		for _, problem := range []string{"missing", "ambiguous", "incomplete"} {
			t.Run(kind+"-"+problem, func(t *testing.T) {
				body := `{"value":[]}`
				expected := error(member.ErrNotFound)
				limits := client.PagingOptions{}
				if problem == "ambiguous" {
					body = `{"value":[` + nativeMember(`[]`) + `,` + nativeMember(`[]`) + `]}`
					expected = member.ErrAmbiguous
				}
				if problem == "incomplete" {
					body = `{"value":[` + nativeMember(`[]`) + `],"@odata.nextLink":"https://graph.microsoft.com` + membersPath() + `?$skiptoken=next"}`
					limits.MaxPages = 1
					expected = client.ErrPageLimit
				}
				run(t, testutil.Case{Name: kind + "-" + problem, Steps: []testutil.Exchange{channelRead(kind), wire("GET", membersPath(), "", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					return changed(sdk.GraphAPI.ChannelUserManagement.Remove(ctx, teamID, "fixture", userID, management.ChangeOptions{Traversal: limits}))
				}, Want: testutil.Expected{Status: 200, Error: expected}})
			})
		}
	}
}

func TestUnitLocalValidationAndStandardGuard(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"add", "remove"} {
		t.Run(action+"-standard", func(t *testing.T) {
			run(t, testutil.Case{Name: action + "-standard", Steps: []testutil.Exchange{channelRead("standard")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				if action == "add" {
					return changed(sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", userID, management.ChangeOptions{}))
				}
				return changed(sdk.GraphAPI.ChannelUserManagement.Remove(ctx, teamID, "fixture", userID, management.ChangeOptions{}))
			}, Want: testutil.Expected{Status: 200, Error: management.ErrStandardChannel}})
		})
	}
	for _, tc := range []struct {
		name, user string
		options    management.ChangeOptions
	}{
		{"empty-user", "", management.ChangeOptions{}}, {"wrong-role", userID, management.ChangeOptions{Role: management.Member}}, {"unknown-role", userID, management.ChangeOptions{Role: "future"}}, {"bad-tenant", userID, management.ChangeOptions{TenantID: stringPointer("")}}, {"external-upn", "member@fixture.invalid", management.ChangeOptions{TenantID: stringPointer(tenantID)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return changed(sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", tc.user, tc.options))
			}, Want: testutil.Expected{Error: client.ErrRequest}})
		})
	}
	run(t, testutil.Case{Name: "cancelled", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		return changed(sdk.GraphAPI.ChannelUserManagement.Add(cancelled, teamID, "fixture", userID, management.ChangeOptions{}))
	}, Want: testutil.Expected{Error: context.Canceled}})
	run(t, testutil.Case{Name: "redaction", Invoke: func(_ context.Context, _ *msteams.Client) testutil.Observation {
		value := management.ChangeResult{ChannelID: channelID, MembershipID: memberID}
		text := fmt.Sprintf("%+v %#v", value, value)
		if strings.Contains(text, channelID) || strings.Contains(text, memberID) {
			t.Error("formatting disclosed native identities")
		}
		return testutil.Observation{}
	}, Want: testutil.Expected{}})
}
