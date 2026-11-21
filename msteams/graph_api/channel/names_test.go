package channel_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
)

func TestUnitNameCompositionsUseCompleteUniqueLookup(t *testing.T) {
	t.Parallel()

	next := base + "/channels?$skiptoken=next"
	for _, tc := range []struct{ name, method, command, payload string }{
		{"update", "PATCH", "Set-TeamChannel", `{"description":""}`},
		{"delete", "DELETE", "Remove-TeamChannel", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "name-" + tc.name,
				Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: "__AllParameterSets"}},
				Steps: []contracttest.Exchange{
					wire("GET", base+"/channels", "", 200, `{"value":[{"id":"other","displayName":"Other"}],"@odata.nextLink":"`+next+`"}`),
					wire("GET", next, "", 200, `{"value":[{"id":"`+channelID+`","displayName":"FIXTURE"}]}`),
					wire(tc.method, channelPath, tc.payload, 204, ""),
				}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					if tc.method == "PATCH" {
						empty := ""
						r, e := sdk.GraphAPI.Channels.UpdateByDisplayName(ctx, teamID, "fixture", channel.UpdateRequest{Description: &empty}, client.PagingOptions{})
						return contracttest.Observation{Response: r, Err: e}
					}
					r, e := sdk.GraphAPI.Channels.DeleteByDisplayName(ctx, teamID, "fixture", client.PagingOptions{})
					return contracttest.Observation{Response: r, Err: e}
				}, Want: contracttest.Expected{Status: 204}})
		})
	}
	for _, tc := range []struct {
		name, body string
		limits     client.PagingOptions
		err        error
	}{
		{name: "missing", body: `{"value":[]}`, err: channel.ErrNotFound},
		{name: "ambiguous", body: `{"value":[{"id":"first","displayName":"fixture"},{"id":"second","displayName":"FIXTURE"}]}`, err: channel.ErrAmbiguous},
		{name: "incomplete", body: `{"value":[{"id":"first","displayName":"fixture"}],"@odata.nextLink":"` + next + `"}`, limits: client.PagingOptions{MaxPages: 1}, err: client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "name-" + tc.name, Steps: []contracttest.Exchange{wire("GET", base+"/channels", "", 200, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := sdk.GraphAPI.Channels.DeleteByDisplayName(ctx, teamID, "fixture", tc.limits)
				return contracttest.Observation{Response: r, Err: e}
			}, Want: contracttest.Expected{Status: 200, Error: tc.err}})
		})
	}
}

func TestUnitExplicitCollectionWrappers(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, path string }{{"all", "allChannels"}, {"incoming", "incomingChannels"}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "wrapper-" + tc.name, Steps: []contracttest.Exchange{wire("GET", base+"/"+tc.path, "", 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var v client.ListResult[channel.Channel]
				var e error
				if tc.name == "all" {
					v, e = sdk.GraphAPI.Channels.ListAll(ctx, teamID, channel.ListOptions{}, client.PagingOptions{})
				} else {
					v, e = sdk.GraphAPI.Channels.ListIncoming(ctx, teamID, channel.ReadOptions{}, client.PagingOptions{})
				}
				return contracttest.Observation{Value: v.Items, Response: v.Response, Err: e}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`null`)}})
		})
	}
}
