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

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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

func wire(method, address, body string, status int, reply string) contracttest.Exchange {
	headers := http.Header{"Accept": {"application/json"}}
	var data json.RawMessage
	if body != "" {
		data = json.RawMessage(body)
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: address, Audience: "https://graph.microsoft.com", Headers: headers, JSON: data},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "channels", "msteams/graph_api/channel/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), requestBoundary(t, c))
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
		step   contracttest.Exchange
		invoke func(context.Context, *msteams.Client) contracttest.Observation
		want   contracttest.Expected
	}{
		{name: "create-standard", step: wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"standard"}`, 201, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture"})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: v.Response, Err: e}
				}
				return contracttest.Observation{Value: v.Channel, Response: v.Response, Err: e}
			}, want: contracttest.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}},
		{name: "create-private-owner-empty-description", step: wire("POST", base+"/channels", `{"displayName":"Fixture","description":"","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('`+userID+`')","roles":["owner"]}]}`, 201, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				owner := userID
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", Description: &empty, MembershipType: channel.Private, OwnerUserID: &owner})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: v.Response, Err: e}
				}
				return contracttest.Observation{Value: v.Channel, Response: v.Response, Err: e}
			}, want: contracttest.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created","membershipType":"private"}`)}},
		{name: "read-escaped-future-fields", step: wire("GET", channelPath, "", 200, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			},
			want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + channelID + `","membershipType":"futureType","isArchived":false}`)}},
		{name: "update-empty-description", step: wire("PATCH", channelPath, `{"description":""}`, 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{Description: &empty})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Response: r, Err: e}
			}, want: contracttest.Expected{Status: 204}},
		{name: "update-empty-object", step: wire("PATCH", channelPath, `{}`, 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Response: r, Err: e}
			}, want: contracttest.Expected{Status: 204}},
		{name: "delete-channel-only", step: wire("DELETE", channelPath, "", 204, ""),
			invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, channelID)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Response: r, Err: e}
			}, want: contracttest.Expected{Status: 204}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{tc.step}, Invoke: tc.invoke, Want: tc.want})
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
			run(t, contracttest.Case{Name: "list-" + tc.name, Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: "__AllParameterSets"}},
				Steps: []contracttest.Exchange{wire("GET", base+"/"+string(tc.collection), "", 200, `{"value":[{"id":"first"}],"@odata.nextLink":"`+next+`"}`), wire("GET", next, "", 200, `{"value":[{"id":"second"}]}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					v, e := sdk.GraphAPI.Channels.ListComplete(ctx, teamID, channel.ListOptions{Collection: tc.collection}, client.PagingOptions{})
					return contracttest.Observation{Value: v.Items, Response: v.Response, Err: e}
				},
				Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"first"},{"id":"second"}]`)}})
		})
	}
	kind := channel.Private
	address := base + "/channels?" + url.Values{"$filter": {"membershipType eq 'private'"}, "$select": {"id,displayName"}}.Encode()
	step := wire("GET", address, "", 200, `{"value":[],"@odata.count":0}`)
	step.Request.Headers.Set("Prefer", "include-unknown-enum-members")
	run(t, contracttest.Case{Name: "filter-select-prefer", Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.Channels.List(ctx, teamID, channel.ListOptions{MembershipType: &kind, ReadOptions: channel.ReadOptions{Select: []string{"id", "displayName"}, IncludeUnknownEnumMembers: true}})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: r, Err: e}
		}
		return contracttest.Observation{Value: v.Items, Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[]`)}})
	for _, tc := range []struct {
		name, next string
		limits     client.PagingOptions
		err        error
	}{
		{"foreign", "https://foreign.invalid/beta/secret", client.PagingOptions{}, client.ErrDestination},
		{"bounded", base + "/channels?$skiptoken=next", client.PagingOptions{MaxPages: 1}, client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "paging-" + tc.name, Steps: []contracttest.Exchange{wire("GET", base+"/channels", "", 200, `{"value":[{"id":"partial"}],"@odata.nextLink":"`+tc.next+`"}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					v, e := sdk.GraphAPI.Channels.ListComplete(ctx, teamID, channel.ListOptions{}, tc.limits)
					return contracttest.Observation{Value: v.Items, Response: v.Response, Err: e}
				}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"partial"}]`), Error: tc.err}})
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
				run(t, contracttest.Case{Name: "failure-" + methodName(method) + "-" + strconv.Itoa(status), Steps: []contracttest.Exchange{wire(method, address, body, status, `{"error":{"code":"FixtureDenied","message":"private message"}}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						switch method {
						case "POST":
							v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture"})
							return contracttest.Observation{Response: v.Response, Err: e}
						case "PATCH":
							r, e := sdk.GraphAPI.Channels.UpdateByID(ctx, teamID, channelID, channel.UpdateRequest{})
							return contracttest.Observation{Response: r, Err: e}
						default:
							r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, channelID)
							return contracttest.Observation{Response: r, Err: e}
						}
					},
					Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureDenied"}}}})
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
			run(t, contracttest.Case{Name: "read-" + tc.name, Steps: []contracttest.Exchange{wire("GET", channelPath, "", tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
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
		invoke func(context.Context, *msteams.Client) contracttest.Observation
		err    error
	}{
		{"empty-team", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, "", channelID, channel.ReadOptions{})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"empty-channel", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, "")
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Response: r, Err: e}
		}, client.ErrRequest},
		{"dot-channel", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.GraphAPI.Channels.DeleteByID(ctx, teamID, "..")
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Response: r, Err: e}
		}, client.ErrRequest},
		{"blank-name", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: " "})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: v.Response, Err: e}
			}
			return contracttest.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"future-type-write", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: "future"})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: v.Response, Err: e}
			}
			return contracttest.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"standard-owner", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			owner := userID
			v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", OwnerUserID: &owner})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: v.Response, Err: e}
			}
			return contracttest.Observation{Response: v.Response, Err: e}
		}, client.ErrRequest},
		{"incoming-filter", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			kind := channel.Private
			v, r, e := sdk.GraphAPI.Channels.List(ctx, teamID, channel.ListOptions{Collection: channel.Incoming, MembershipType: &kind})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"bad-select", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.GraphAPI.Channels.GetByID(ctx, teamID, channelID, channel.ReadOptions{Select: []string{"id,secret"}})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, client.ErrRequest},
		{"nil-context", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before any I/O.
			v, r, e := sdk.GraphAPI.Channels.GetByID(nil, teamID, channelID, channel.ReadOptions{})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, client.ErrNilContext},
		{"cancelled", func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			v, r, e := sdk.GraphAPI.Channels.GetByID(cancelled, teamID, channelID, channel.ReadOptions{})
			if e != nil && isRequestBoundary(ctx) {
				return contracttest.Observation{Response: r, Err: e}
			}
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Scope: "local", Invoke: tc.invoke, Want: contracttest.Expected{Error: tc.err}})
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
		run(t, contracttest.Case{Name: "shared-accepted-beta-operation", Mappings: []contracttest.Mapping{{Command: "New-TeamChannel", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{step, wire("GET", "https://graph.microsoft.com/beta/"+url.PathEscape("teams('"+teamID+"')")+"/"+url.PathEscape("operations('job')"), "", 200, "")},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, e := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Shared})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: v.Response, Err: e}
				}
				if e != nil {
					return contracttest.Observation{Response: v.Response, Err: e}
				}
				if v.Channel != nil || v.Operation == nil || v.Operation.Snapshot().Status.State != client.OperationAccepted {
					t.Error("acceptance was confused with a channel")
				}
				snapshot, e := v.Operation.Wait(ctx, client.WaitOptions{MaxPolls: 1})
				return contracttest.Observation{Value: map[string]string{"state": string(snapshot.Status.State)}, Response: snapshot.Response, Err: e}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"state":"succeeded"}`)}})
		if time.Since(started) >= 31*time.Second {
			t.Fatal("denied channel submission unexpectedly waited to poll")
		}

	})
}

func TestUnitOperationLocationsFailLocally(t *testing.T) {
	t.Parallel()
	for i, location := range []string{"", "//foreign.invalid/beta/teams/x/operations/job", "https://foreign.invalid/beta/teams/" + teamID + "/operations/job", "/v1.0/teams/" + teamID + "/operations/job", "/beta/teams/" + userID + "/operations/job", "/beta/users/" + userID, "/beta/teams/" + teamID + "/operations/job?secret=x", "/beta/teams/" + teamID + "/operations/job#x"} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			run(t, contracttest.Case{Name: "bad-location-" + strconv.Itoa(i), Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, e := sdk.GraphAPI.Channels.ResumeOperation(teamID, location)
				return contracttest.Observation{Err: e}
			}, Want: contracttest.Expected{Error: client.ErrOperation}})
		})
	}
}

func TestUnitOwnerResolutionComposition(t *testing.T) {
	t.Parallel()
	step := wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('`+userID+`')","roles":["owner"]}]}`, 201, "")
	run(t, contracttest.Case{Name: "owner-upn-resolution", Mappings: []contracttest.Mapping{{Command: "New-TeamChannel", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{wire("GET", "https://graph.microsoft.com/v1.0/users/owner@fixture.invalid?$select=id", "", 200, ""), step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.Channels.CreateWithOwner(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Private}, "owner@fixture.invalid", directoryuser.ResolveOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		return contracttest.Observation{Value: v.Channel, Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}})
}
