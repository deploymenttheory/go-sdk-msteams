package sharedchannel_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	share "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel"
)

const (
	hostID    = "11111111-1111-4111-8111-111111111111"
	targetID  = "22222222-2222-4222-8222-222222222222"
	channelID = "19:fixture/opaque@thread.tacv2"
)

func base() string {
	return "/beta/teams/" + hostID + "/channels/" + url.PathEscape(channelID) + "/sharedWithTeams"
}
func wire(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "shared_channel_relationships", "msteams/graph_api/shared_channel/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, e := contracttest.Run(t.Context(), requestBoundary(t, c))
	if e != nil || !r.Passed() {
		t.Fatalf("sharing contract differs: setup=%v differences=%+v", e, r.Differences())
	}
	return r
}

func TestUnitNativeRelationshipReadListAndUnshare(t *testing.T) {
	t.Parallel()

	raw := `{"id":"` + targetID + `","displayName":null,"tenantId":"foreign-tenant","isHostTeam":false,"future":{"v":7}}`
	run(t, contracttest.Case{Name: "read-target", Steps: []contracttest.Exchange{wire("GET", base()+"/"+targetID, 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.ChannelShares.GetByID(ctx, hostID, channelID, targetID)
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: r, Err: e}
		}
		var value any
		if v != nil {
			value = map[string]any{"id": v.ID, "host": v.IsHostTeam, "raw": v.RawFields}
		}
		return contracttest.Observation{Value: value, Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + targetID + `","host":false,"raw":` + raw + `}`)}})

	run(t, contracttest.Case{Name: "complete-relationships", Steps: []contracttest.Exchange{wire("GET", base(), 200, ""), wire("GET", base()+"?$skiptoken=opaque%2Bvalue", 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.ChannelShares.ListAll(ctx, hostID, channelID, client.PagingOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		return contracttest.Observation{Value: map[string]int{"items": len(v.Items), "pages": v.Pages}, Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"items":1,"pages":2}`)}})
	run(t, contracttest.Case{Name: "unshare-only-relationship", Mappings: []contracttest.Mapping{{Command: "Remove-SharedWithTeam", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{wire("DELETE", base()+"/"+targetID, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, e := sdk.GraphAPI.ChannelShares.DeleteByID(ctx, hostID, channelID, targetID)
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: r, Err: e}
		}
		return contracttest.Observation{Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 204}})
	run(t, contracttest.Case{Name: "allowed-members-exact-relationship", Steps: []contracttest.Exchange{wire("GET", base()+"/"+targetID+"/allowedMembers", 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.ChannelShares.ListAllowedMembers(ctx, hostID, channelID, targetID, client.PagingOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		var value any
		if len(v.Items) == 1 {
			value = map[string]any{"id": v.Items[0].ID, "user": v.Items[0].UserID, "roles": v.Items[0].Roles, "raw": v.Items[0].RawFields}
		}
		return contracttest.Observation{Value: value, Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"opaque-allowed","user":null,"roles":["future"],"raw":{"id":"opaque-allowed","userId":null,"tenantId":"foreign","roles":["future"],"future":5}}`)}})
}

func TestUnitUnshareFailureIsNotReplayed(t *testing.T) {
	t.Parallel()

	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "unshare-" + strconv.Itoa(status), Steps: []contracttest.Exchange{wire("DELETE", base()+"/"+targetID, status, `{"error":{"code":"Denied","message":"private"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.ChannelShares.DeleteByID(ctx, hostID, channelID, targetID)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Response: r, Err: e}
			}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"Denied"}}}})
		})
	}
}

func TestUnitSharingPagesRetainPartialStateAndRejectForeignLinks(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, next string
		bounds     client.PagingOptions
		err        error
	}{
		{"foreign", "https://foreign.invalid/private", client.PagingOptions{}, client.ErrDestination}, {"bounded", auth.GraphAudience + base() + "?$skiptoken=next", client.PagingOptions{MaxPages: 1}, client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{wire("GET", base(), 200, `{"value":[{"id":"`+targetID+`"}],"@odata.nextLink":"`+tc.next+`"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, e := sdk.GraphAPI.ChannelShares.ListAll(ctx, hostID, channelID, tc.bounds)
				return contracttest.Observation{Value: len(v.Items), Response: v.Response, Err: e}
			}, Want: contracttest.Expected{Status: 200, Error: tc.err, JSON: json.RawMessage(`1`)}})
		})
	}
	for _, body := range []string{`null`, `{}`, `{"value":null}`, `{"value":[{"tenantId":"missing-id"}]}`} {
		run(t, contracttest.Case{Name: "malformed-page", Steps: []contracttest.Exchange{wire("GET", base(), 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.GraphAPI.ChannelShares.List(ctx, hostID, channelID)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
}

func TestUnitSharingLocalValidationAndUnexpectedStatus(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, host, channel, target string }{{"bad-host", "not-a-uuid", channelID, targetID}, {"empty-channel", hostID, "", targetID}, {"bad-target", hostID, channelID, "bad"}, {"empty-target", hostID, channelID, ""}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.ChannelShares.DeleteByID(ctx, tc.host, tc.channel, tc.target)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Response: r, Err: e}
			}, Want: contracttest.Expected{Error: client.ErrRequest}})
		})
	}
	run(t, contracttest.Case{Name: "read-202", Steps: []contracttest.Exchange{wire("GET", base()+"/"+targetID, 202, `{"id":"`+targetID+`"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.ChannelShares.GetByID(ctx, hostID, channelID, targetID)
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: r, Err: e}
		}
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 202, Error: share.ErrUnexpectedStatus}})
}
