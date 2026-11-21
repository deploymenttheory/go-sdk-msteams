package region_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

const audience = "https://migrationsvc.teams.microsoft.com"
const scope = audience + "/Region.Read.All"

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "multi_geo_region", "contracts/endpoints/8.1.0/special-service-profiles.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("region %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
	return r
}
func wire(kind, id string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: "https://teams.cloud.microsoft/api/tms/prod/" + kind + "/" + url.PathEscape(id) + "/region", Audience: audience, TokenScope: scope, Headers: http.Header{"Accept": {"application/json"}, "X-MS-Target-Uri": nil}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func TestRegionUsesExactDelegatedScopeAndDirectEntityRoutes(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, cloud := range []string{"commercial", "gcc"} {
		for _, kind := range []string{"users", "groups"} {
			id := "opaque/user ?#%雪"
			results = append(results, run(t, contracttest.Case{Name: cloud + "-" + kind, Cloud: cloud, Mode: auth.Delegated, Mappings: []contracttest.Mapping{{Command: "Get-MultiGeoRegion", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire(kind, id, 200, `{"Region":"FutureRegion","LastMovementTimeStamp":null,"PreferredDataLocation":"","future":{"v":8}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				if kind == "users" {
					v, r, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, id)
					var value any
					if v != nil {
						value = map[string]any{"region": v.Region, "moved": v.LastMovementTimestamp, "preferred": v.PreferredDataLocation, "raw": v.RawFields}
					}
					return contracttest.Observation{Value: value, Response: r, Err: err}
				}
				v, r, err := sdk.TeamsAdminAPI.Regions.GetForGroup(ctx, id)
				var value any
				if v != nil {
					value = map[string]any{"region": v.Region, "moved": v.LastMovementTimestamp, "preferred": v.PreferredDataLocation, "raw": v.RawFields}
				}
				return contracttest.Observation{Value: value, Response: r, Err: err}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"region":"FutureRegion","moved":null,"preferred":"","raw":{"Region":"FutureRegion","LastMovementTimeStamp":null,"PreferredDataLocation":"","future":{"v":8}}}`)}}))
		}
	}
	evidence(t, results)
}

func TestApplicationRegionAndUnqualifiedCloudsRejectBeforeTokens(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "application-local", Scope: "local", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, "fixture")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: client.ErrAuthenticationMode}})
	for _, cloud := range []string{"gcch", "dod", "china", "france", "germany"} {
		run(t, contracttest.Case{Name: cloud, Cloud: cloud, Scope: "local", Mode: auth.Delegated, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.Regions.GetForGroup(ctx, "fixture")
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Error: client.ErrDestination}})
	}
}

func TestRegionMalformedAndDeniedReadsRemainTypedFailures(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`null`, `[]`, `{"Region":17}`, `{"LastMovementTimeStamp":{}}`} {
		run(t, contracttest.Case{Name: "malformed", Mode: auth.Delegated, Steps: []contracttest.Exchange{wire("users", "fixture", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, "fixture")
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
	run(t, contracttest.Case{Name: "forbidden", Mode: auth.Delegated, Steps: []contracttest.Exchange{wire("groups", "fixture", 403, `{"error":{"code":"Denied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.Regions.GetForGroup(ctx, "fixture")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	for _, id := range []string{"", "..", "bad\nidentity"} {
		run(t, contracttest.Case{Name: "invalid-id", Mode: auth.Delegated, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, id)
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Error: client.ErrRequest}})
	}
}
