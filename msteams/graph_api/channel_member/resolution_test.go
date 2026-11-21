package channelmember_test

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
)

func TestUnitResolveByUserRequiresCompleteUnambiguousReturnedIdentity(t *testing.T) {
	t.Parallel()
	path := "/beta/teams/" + teamID + "/channels/" + url.PathEscape(channelID) + "/members"
	next := auth.GraphAudience + path + "?$skiptoken=next"
	first := `{"id":"first-opaque","userId":"` + userID + `","tenantId":"11111111-1111-4111-8111-111111111111"}`
	second := `{"id":"second-opaque","userId":"` + userID + `","tenantId":"22222222-2222-4222-8222-222222222222"}`
	tenant := teamID
	for _, tc := range []struct {
		name   string
		body   string
		tenant *string
		limits client.PagingOptions
		second bool
		err    error
		json   json.RawMessage
	}{
		{name: "complete-returned-opaque", body: `{"value":[],"@odata.nextLink":"` + next + `"}`, second: true, json: json.RawMessage(`{"id":"second-opaque"}`)},
		{name: "cross-tenant-ambiguous", body: `{"value":[` + first + `,` + second + `]}`, err: channelmember.ErrAmbiguous},
		{name: "explicit-tenant", body: `{"value":[` + first + `,` + second + `]}`, tenant: &tenant, json: json.RawMessage(`{"id":"first-opaque"}`)},
		{name: "missing", body: `{"value":[]}`, err: channelmember.ErrNotFound},
		{name: "incomplete-match", body: `{"value":[` + first + `],"@odata.nextLink":"` + next + `"}`, limits: client.PagingOptions{MaxPages: 1}, err: client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", auth.GraphAudience, path, "", 200, tc.body)}
			if tc.second {
				steps = append(steps, wire("GET", auth.GraphAudience, path+"?$skiptoken=next", "", 200, `{"value":[`+second+`]}`))
			}
			nativeCase(t, contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.ChannelMembers.ResolveByUser(ctx, teamID, channelID, userID, tc.tenant, tc.limits)
				var result any
				if v != nil {
					result = map[string]string{"id": string(v.ID)}
				}
				return contracttest.Observation{Value: result, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: 200, Error: tc.err, JSON: tc.json}})
		})
	}
}
