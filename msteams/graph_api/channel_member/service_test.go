package channelmember_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	"github.com/go-resty/resty/v2"
)

const (
	channelID                               = "19:channel/opaque@thread.tacv2"
	teamID                                  = "11111111-1111-4111-8111-111111111111"
	userID                                  = "22222222-2222-4222-8222-222222222222"
	memberID     channelmember.MembershipID = "opaque+/=:#%雪"
	nativeSource                            = "contracts/endpoints/8.1.0/core.json"
)

func nativeCase(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	if c.Name == "" {
		c.Name = "native-case"
	}
	c.Family = "native_channel_members"
	if c.Scope == "" {
		c.Scope = "service"
	}
	if c.Source == "" {
		c.Source = nativeSource
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("native membership contract differs: %+v", result.Differences())
	}
	return result
}

func wire(method, origin, path, body string, status int, response string) contracttest.Exchange {
	var jsonBody json.RawMessage
	if body != "" {
		jsonBody = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: origin, Headers: headers, JSON: jsonBody},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func snapshot(value *channelmember.Member, response *resty.Response, err error) contracttest.Observation {
	var result any
	if value != nil {
		result = map[string]any{"id": value.ID, "user_id": value.UserID, "roles": value.Roles,
			"type": value.Type, "email": value.Email, "display_name": value.DisplayName, "raw": value.RawFields}
	}
	return contracttest.Observation{Value: result, Response: response, Err: err}
}

func resourceJSON(roles string) string {
	return `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"` + string(memberID) + `","userId":"` + userID + `","roles":` + roles + `,"email":"","displayName":null,"future":{"flag":false}}`
}

func expectedSnapshot(roles, raw string) json.RawMessage {
	return json.RawMessage(`{"id":"` + string(memberID) + `","user_id":"` + userID + `","roles":` + roles + `,"type":"#microsoft.graph.aadUserConversationMember","email":"","display_name":null,"raw":` + raw + `}`)
}

func TestNativeCRUDAndOwnerActions(t *testing.T) {
	t.Parallel()
	path := "/beta/teams/" + teamID + "/channels/" + url.PathEscape(channelID) + "/members"
	byID := path + "/" + url.PathEscape(string(memberID))
	origin := auth.GraphAudience
	owner := []channelmember.Role{channelmember.Owner}
	empty := []channelmember.Role{}
	var results []contracttest.Result
	for _, tc := range []struct {
		name, method, path, body, roles string
		status                          int
		invoke                          func(context.Context, *msteams.Client) contracttest.Observation
	}{
		{"create-owner", "POST", path, `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":["owner"],"user@odata.bind":"` + origin + `/beta/users('` + userID + `')"}`, `["owner"]`, 201,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID, Roles: &owner}))
			}},
		{"create-member-empty", "POST", path, `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"` + origin + `/beta/users('` + userID + `')"}`, `[]`, 201,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID, Roles: &empty}))
			}},
		{"create-member-omitted", "POST", path, `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"` + origin + `/beta/users('` + userID + `')"}`, `[]`, 201,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID}))
			}},
		{"get-opaque", "GET", byID, "", `["owner"]`, 200,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.GetByID(ctx, teamID, channelID, memberID))
			}},
		{"update-omitted-retains", "PATCH", byID, `{}`, `["owner"]`, 200,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, memberID, channelmember.UpdateRequest{}))
			}},
		{"update-owner", "PATCH", byID, `{"roles":["owner"]}`, `["owner"]`, 200,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, memberID, channelmember.UpdateRequest{Roles: &owner}))
			}},
		{"promote-owner", "PATCH", byID, `{"roles":["owner"]}`, `["owner"]`, 200,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.PromoteOwner(ctx, teamID, channelID, memberID))
			}},
		{"demote-owner", "PATCH", byID, `{"roles":[]}`, `[]`, 200,
			func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.DemoteOwner(ctx, teamID, channelID, memberID))
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := resourceJSON(tc.roles)
			results = append(results, nativeCase(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{wire(tc.method, origin, tc.path, tc.body, tc.status, raw)},
				Invoke: tc.invoke, Want: contracttest.Expected{Status: tc.status, JSON: expectedSnapshot(tc.roles, raw)}}))
		})
	}
	t.Run("delete-membership-only", func(t *testing.T) {
		results = append(results, nativeCase(t, contracttest.Case{Name: "delete", Steps: []contracttest.Exchange{wire("DELETE", origin, byID, "", 204, "")},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := sdk.GraphAPI.ChannelMembers.DeleteByID(ctx, teamID, channelID, memberID)
				return contracttest.Observation{Response: response, Err: err}
			}, Want: contracttest.Expected{Status: 204}}))
	})
	if revision := os.Getenv("MSTEAMS_CONTRACT_REVISION"); revision != "" {
		root, err := os.OpenRoot(t.ArtifactDir())
		if err != nil {
			t.Fatal("native evidence directory unavailable")
		}
		defer func() { _ = root.Close() }()
		file, err := root.Create("channel-member-service.json")
		if err != nil {
			t.Fatal("native evidence file unavailable")
		}
		writeErr := contracttest.WriteEvidence(file, revision, results)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			t.Fatal("native evidence could not be written")
		}
	}
}

func TestCreateCloudAndConfiguredModeBindings(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			t.Run(cloud+"/"+string(mode), func(t *testing.T) {
				env, err := auth.ResolveEnvironment(cloud, nil)
				if err != nil {
					t.Fatal(err)
				}
				origin := env.Endpoints.Graph
				raw := resourceJSON(`[]`)
				body := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"` + origin + `/beta/users('` + userID + `')"}`
				nativeCase(t, contracttest.Case{Cloud: cloud, Mode: mode, Steps: []contracttest.Exchange{wire("POST", origin, "/beta/teams/"+teamID+"/channels/"+url.PathEscape(channelID)+"/members", body, 201, raw)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID}))
					},
					Want: contracttest.Expected{Status: 201, JSON: expectedSnapshot(`[]`, raw)}})
			})
		}
	}
}

func TestNativeValidationDoesNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	for index, roles := range [][]channelmember.Role{nil, {"member"}, {channelmember.Guest}, {"future"}, {channelmember.Owner, channelmember.Owner}} {
		t.Run("roles-"+strconv.Itoa(index), func(t *testing.T) {
			for _, action := range []string{"create", "update"} {
				nativeCase(t, contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					if action == "create" {
						return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID, Roles: &roles}))
					}
					return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, memberID, channelmember.UpdateRequest{Roles: &roles}))
				}, Want: contracttest.Expected{Error: client.ErrRequest}})
			}
		})
	}
	for _, id := range []channelmember.MembershipID{"", ".", "..", "private\n", "private\\segment", " "} {
		for _, action := range []string{"get", "update", "delete"} {
			nativeCase(t, contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				switch action {
				case "get":
					return snapshot(sdk.GraphAPI.ChannelMembers.GetByID(ctx, teamID, channelID, id))
				case "update":
					return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, id, channelmember.UpdateRequest{}))
				default:
					response, err := sdk.GraphAPI.ChannelMembers.DeleteByID(ctx, teamID, channelID, id)
					return contracttest.Observation{Response: response, Err: err}
				}
			}, Want: contracttest.Expected{Error: client.ErrRequest}})
		}
	}
	for _, top := range []int{-1, 1000} {
		nativeCase(t, contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.GraphAPI.ChannelMembers.List(ctx, teamID, channelID, channelmember.ListOptions{Top: top})
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Error: client.ErrRequest}})
	}
}

func TestNativeHTTPAndEnvelopeFailures(t *testing.T) {
	t.Parallel()
	path := "/beta/teams/" + teamID + "/channels/" + url.PathEscape(channelID) + "/members/" + url.PathEscape(string(memberID))
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"malformed", `{"id":`, 200, client.ErrDecode},
		{"missing-membership-id", `{"userId":"` + userID + `"}`, 200, client.ErrDecode},
		{"array", `[]`, 200, client.ErrDecode},
		{"null", `null`, 200, client.ErrDecode},
		{"unexpected-empty-update", ``, 204, channelmember.ErrUnexpectedStatus},
		{"forbidden", `{"error":{"code":"Forbidden","message":"private"}}`, 403, auth.ErrForbidden},
		{"unauthenticated", `{"error":{"code":"Unauthorized","message":"private"}}`, 401, auth.ErrUnauthenticated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nativeCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("PATCH", auth.GraphAudience, path, `{}`, tc.status, tc.body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, memberID, channelmember.UpdateRequest{}))
				},
				Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
		})
	}
}

func TestMutationFailureIsNotReplayedOrInferredAbsent(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"POST", "PATCH", "DELETE"} {
		for _, status := range []int{404, 429, 503} {
			t.Run(method+"/"+strconv.Itoa(status), func(t *testing.T) {
				path := "/beta/teams/" + teamID + "/channels/" + url.PathEscape(channelID) + "/members"
				body := ""
				if method != "POST" {
					path += "/" + url.PathEscape(string(memberID))
				}
				if method == "POST" {
					body = `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"user@odata.bind":"` + auth.GraphAudience + `/beta/users('` + userID + `')"}`
				}
				if method == "PATCH" {
					body = `{}`
				}
				nativeCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire(method, auth.GraphAudience, path, body, status, `{"error":{"code":"FixtureError","message":"private"}}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						switch method {
						case "POST":
							return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID}))
						case "PATCH":
							return snapshot(sdk.GraphAPI.ChannelMembers.UpdateByID(ctx, teamID, channelID, memberID, channelmember.UpdateRequest{}))
						default:
							response, err := sdk.GraphAPI.ChannelMembers.DeleteByID(ctx, teamID, channelID, memberID)
							return contracttest.Observation{Response: response, Err: err}
						}
					}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureError"}}}})
			})
		}
	}
}
