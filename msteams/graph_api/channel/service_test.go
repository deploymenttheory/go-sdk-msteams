package channel_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"strconv"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
)

const (
	teamID      = "11111111-1111-4111-8111-111111111111"
	userID      = "22222222-2222-4222-8222-222222222222"
	channelID   = "19:fixture/part@thread.tacv2"
	base        = "https://graph.microsoft.com/beta/teams/" + teamID
	channelPath = base + "/channels/" + "19:fixture%2Fpart@thread.tacv2"
)

func wire(method, address, body string, status int, reply string) testutil.Exchange {
	headers := http.Header{"Accept": {"application/json"}}
	var data json.RawMessage
	if body != "" {
		data = json.RawMessage(body)
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: address, Audience: "https://graph.microsoft.com", Headers: headers, JSON: data},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("channel contract differs: setup=%v differences=%+v", err, result.Differences())
	}
	return result
}

func TestUnitNativeCRUDContracts(t *testing.T) {
	t.Parallel()
	empty := ""

	for _, tc := range []struct {
		name   string
		step   testutil.Exchange
		invoke func(context.Context, *msteams.Client) testutil.Observation
		want   testutil.Expected
	}{
		{name: "create-standard", step: wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"standard"}`, 201, `{"id":"created"}`),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture"})

				return testutil.Observation{Value: v.Channel, Response: v.Response, Err: e}
			}, want: testutil.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}},
		{name: "create-private-owner-empty-description", step: wire("POST", base+"/channels", `{"displayName":"Fixture","description":"","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('`+userID+`')","roles":["owner"]}]}`, 201, `{"id":"created","membershipType":"private"}`),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				owner := userID
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", Description: &empty, MembershipType: channel.Private, OwnerUserID: &owner})

				return testutil.Observation{Value: v.Channel, Response: v.Response, Err: e}
			}, want: testutil.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created","membershipType":"private"}`)}},
		{name: "read-escaped-future-fields", step: wire("GET", channelPath, "", 200, `{"id":"`+channelID+`","membershipType":"futureType","isArchived":false}`),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{})

				return testutil.Observation{Value: v, Response: r, Err: e}
			},
			want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + channelID + `","membershipType":"futureType","isArchived":false}`)}},
		{name: "update-empty-description", step: wire("PATCH", channelPath, `{"description":""}`, 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{Description: &empty})

				return testutil.Observation{Response: r, Err: e}
			}, want: testutil.Expected{Status: 204}},
		{name: "update-empty-object", step: wire("PATCH", channelPath, `{}`, 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{})

				return testutil.Observation{Response: r, Err: e}
			}, want: testutil.Expected{Status: 204}},
		{name: "delete-channel-only", step: wire("DELETE", channelPath, "", 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, channelID)

				return testutil.Observation{Response: r, Err: e}
			}, want: testutil.Expected{Status: 204}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{tc.step}, Invoke: tc.invoke, Want: tc.want})
		})
	}
}

func TestUnitCollectionsAndPaging(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, command string
		collection    channel.Collection
	}{
		{"hosted", "Get-TeamChannel", channel.Hosted}, {"all", "Get-TeamAllChannel", channel.All}, {"incoming", "Get-TeamIncomingChannel", channel.Incoming},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next := base + "/" + string(tc.collection) + "?$skiptoken=opaque%2Bvalue"
			run(t, testutil.Case{Name: "list-" + tc.name,
				Steps: []testutil.Exchange{wire("GET", base+"/"+string(tc.collection), "", 200, `{"value":[{"id":"first"}],"@odata.nextLink":"`+next+`"}`), wire("GET", next, "", 200, `{"value":[{"id":"second"}]}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					v, e := sdk.GraphAPI.Channels.ListComplete(ctx, teamID, channel.ListOptions{Collection: tc.collection}, client.PagingOptions{})
					return testutil.Observation{Value: v.Items, Response: v.Response, Err: e}
				},
				Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"first"},{"id":"second"}]`)}})
		})
	}
	kind := channel.Private
	address := base + "/channels?" + url.Values{"$filter": {"membershipType eq 'private'"}, "$select": {"id,displayName"}}.Encode()
	step := wire("GET", address, "", 200, `{"value":[],"@odata.count":0}`)
	step.Request.Headers.Set("Prefer", "include-unknown-enum-members")
	run(t, testutil.Case{Name: "filter-select-prefer", Steps: []testutil.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.GraphAPI.Channels.List(ctx, teamID, channel.ListOptions{MembershipType: &kind, ReadOptions: channel.ReadOptions{Select: []string{"id", "displayName"}, IncludeUnknownEnumMembers: true}})

		return testutil.Observation{Value: v.Items, Response: r, Err: e}
	}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`[]`)}})
	for _, tc := range []struct {
		name, next string
		limits     client.PagingOptions
		err        error
	}{
		{"foreign", "https://foreign.invalid/beta/secret", client.PagingOptions{}, client.ErrDestination},
		{"bounded", base + "/channels?$skiptoken=next", client.PagingOptions{MaxPages: 1}, client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "paging-" + tc.name, Steps: []testutil.Exchange{wire("GET", base+"/channels", "", 200, `{"value":[{"id":"partial"}],"@odata.nextLink":"`+tc.next+`"}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					v, e := sdk.GraphAPI.Channels.ListComplete(ctx, teamID, channel.ListOptions{}, tc.limits)
					return testutil.Observation{Value: v.Items, Response: v.Response, Err: e}
				}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"partial"}]`), Error: tc.err}})
		})
	}
}

func TestUnitFailuresPreserveStatusAndDoNotReplayWrites(t *testing.T) {
	t.Parallel()

	for _, method := range []string{"POST", "PATCH", "DELETE"} {
		for _, status := range []int{403, 429, 503} {
			t.Run(method+"-"+strconv.Itoa(status), func(t *testing.T) {
				address, body := channelPath, ""
				if method == "POST" {
					address, body = base+"/channels", `{"displayName":"Fixture","membershipType":"standard"}`
				}
				if method == "PATCH" {
					body = `{}`
				}
				run(t, testutil.Case{Name: "failure-" + methodName(method) + "-" + strconv.Itoa(status), Steps: []testutil.Exchange{wire(method, address, body, status, `{"error":{"code":"FixtureDenied","message":"private message"}}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						switch method {
						case "POST":
							v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture"})
							return testutil.Observation{Response: v.Response, Err: e}
						case "PATCH":
							r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{})
							return testutil.Observation{Response: r, Err: e}
						default:
							r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, channelID)
							return testutil.Observation{Response: r, Err: e}
						}
					},
					Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureDenied"}}}})
			})
		}
	}
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"malformed", "{", 200, client.ErrDecode}, {"null", "null", 200, client.ErrDecode}, {"missing-id", "{}", 200, client.ErrDecode}, {"wrong-status", `{"id":"ignored"}`, 202, channel.ErrUnexpectedStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "read-" + tc.name, Steps: []testutil.Exchange{wire("GET", channelPath, "", tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{})

				return testutil.Observation{Value: v, Response: r, Err: e}
			}, Want: testutil.Expected{Status: tc.status, Error: tc.err}})
		})
	}
}

func methodName(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PATCH":
		return "update"
	default:
		return "delete"
	}
}

func TestUnitLocalValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		invoke func(context.Context, *msteams.Client) testutil.Observation
		err    error
	}{
		{"empty-team", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, "", channelID, channel.ReadOptions{})

			return testutil.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"empty-channel", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, "")

			return testutil.Observation{Response: r, Err: e}
		}, client.ErrRequest},
		{"dot-channel", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, "..")

			return testutil.Observation{Response: r, Err: e}
		}, client.ErrRequest},
		{"blank-name", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: " "})

			return testutil.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"future-type-write", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: "future"})

			return testutil.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"standard-owner", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			owner := userID
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", OwnerUserID: &owner})

			return testutil.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"incoming-filter", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			kind := channel.Private
			v, r, e := sdk.GraphAPI.Channels.List(ctx, teamID, channel.ListOptions{Collection: channel.Incoming, MembershipType: &kind})

			return testutil.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"bad-select", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{Select: []string{"id,secret"}})

			return testutil.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"nil-context", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before any I/O.
			v, r, e := sdk.GraphAPI.Channels.GetByID(nil, teamID, channelID, channel.ReadOptions{})

			return testutil.Observation{Value: v, Response: r, Err: e}
		}, client.ErrNilContext},
		{"cancelled", func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			v, r, e := sdk.GraphAPI.Channels.GetByID(cancelled, teamID, channelID, channel.ReadOptions{})

			return testutil.Observation{Value: v, Response: r, Err: e}
		}, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Invoke: tc.invoke, Want: testutil.Expected{Error: tc.err}})
		})
	}
	var service *channel.Service
	if _, _, err := service.GetByID(t.Context(), teamID, channelID, channel.ReadOptions{}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service did not fail closed")
	}
}

func TestUnitSharedCreationAndOperation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		location := "/teams('" + teamID + "')/operations('job')"
		step := wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"shared"}`, 202, "")
		step.Reply.Headers.Set("Location", location)
		started := time.Now()
		run(t, testutil.Case{Name: "shared-accepted-beta-operation", Steps: []testutil.Exchange{step, wire("GET", "https://graph.microsoft.com/beta/"+url.PathEscape("teams('"+teamID+"')")+"/"+url.PathEscape("operations('job')"), "", 200, `{"id":"job","status":"succeeded"}`)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Shared})

				if e != nil {
					return testutil.Observation{Response: v.Response, Err: e}
				}
				if v.Channel != nil || v.Operation == nil || v.Operation.Snapshot().Status.State != client.OperationAccepted {
					t.Error("acceptance was confused with a channel")
				}
				snapshot, e := v.Operation.Wait(ctx, client.WaitOptions{MaxPolls: 1})
				return testutil.Observation{Value: map[string]string{"state": string(snapshot.Status.State)}, Response: snapshot.Response, Err: e}
			}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"state":"succeeded"}`)}})
		if time.Since(started) < 31*time.Second {
			t.Fatal("accepted channel operation polled before the minimum interval")
		}

	})
}

func TestUnitOperationLocationsFailLocally(t *testing.T) {
	t.Parallel()
	for i, location := range []string{"", "//foreign.invalid/beta/teams/x/operations/job", "https://foreign.invalid/beta/teams/" + teamID + "/operations/job", "/v1.0/teams/" + teamID + "/operations/job", "/beta/teams/" + userID + "/operations/job", "/beta/users/" + userID, "/beta/teams/" + teamID + "/operations/job?secret=x", "/beta/teams/" + teamID + "/operations/job#x"} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			run(t, testutil.Case{Name: "bad-location-" + strconv.Itoa(i), Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, e := sdk.GraphAPI.Channels.ResumeOperation(teamID, location)
				return testutil.Observation{Err: e}
			}, Want: testutil.Expected{Error: client.ErrOperation}})
		})
	}
}

func TestUnitOwnerResolutionComposition(t *testing.T) {
	t.Parallel()
	step := wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('`+userID+`')","roles":["owner"]}]}`, 201, `{"id":"created"}`)
	run(t, testutil.Case{Name: "owner-upn-resolution", Steps: []testutil.Exchange{wire("GET", "https://graph.microsoft.com/v1.0/users/owner@fixture.invalid?$select=id", "", 200, `{"id":"`+userID+`"}`), step}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, e := sdk.GraphAPI.Channels.CreateWithOwner(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Private}, "owner@fixture.invalid", directoryuser.ResolveOptions{})

		return testutil.Observation{Value: v.Channel, Response: v.Response, Err: e}
	}, Want: testutil.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}})
}
