package region_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	region "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/region"
	"github.com/go-resty/resty/v2"
)

const audience = "https://migrationsvc.teams.microsoft.com"
const scope = audience + "/Region.Read.All"

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "multi_geo_region", "msteams/teams_admin_api/region/service_test.go"
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

func TestUnitRegionUsesExactDelegatedScopeAndDirectEntityRoutes(t *testing.T) {
	t.Parallel()

	for _, cloud := range []string{"commercial", "gcc"} {
		for _, kind := range []string{"users", "groups"} {
			id := "opaque/user ?#%雪"
			run(t, contracttest.Case{Name: cloud + "-" + kind, Cloud: cloud, Mode: auth.Delegated, Mappings: []contracttest.Mapping{{Command: "Get-MultiGeoRegion", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire(kind, id, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
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
			}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		}
	}
}

func TestUnitApplicationRegionAndUnqualifiedCloudsRejectBeforeTokens(t *testing.T) {
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

func TestUnitRegionMalformedAndDeniedReadsRemainTypedFailures(t *testing.T) {
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

// The field spellings and nullable timestamp follow the pinned 8.1.0
// GetMultiGeoRegionResponse. Future/null retention is an SDK guarantee.
func TestUnitRegionSuccessfulUserAndGroupSnapshots(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"users", "groups"} {
		for _, populated := range []bool{true, false} {
			body := `{"Region":null,"LastMovementTimeStamp":null,"PreferredDataLocation":null}`
			if populated {
				body = `{"Region":"EUR","LastMovementTimeStamp":"2026-11-20T12:00:00Z","PreferredDataLocation":"GBR","future":false}`
			}
			step := wire(kind, "opaque/entity", 200, body)
			step.Reply.ProtocolSources = []string{"https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0"}
			run(t, contracttest.Case{Name: kind + "-success", Mode: auth.Delegated, Mappings: []contracttest.Mapping{{Command: "Get-MultiGeoRegion", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var v *region.Location
				var r *resty.Response
				var err error
				if kind == "users" {
					v, r, err = sdk.TeamsAdminAPI.Regions.GetForUser(ctx, "opaque/entity")
				} else {
					v, r, err = sdk.TeamsAdminAPI.Regions.GetForGroup(ctx, "opaque/entity")
				}
				if err == nil {
					if v == nil {
						t.Fatal("native location missing")
					}
					if populated {
						if v.Region == nil || *v.Region != "EUR" || v.LastMovementTimestamp == nil || *v.LastMovementTimestamp != "2026-11-20T12:00:00Z" || v.PreferredDataLocation == nil || *v.PreferredDataLocation != "GBR" || string(v.RawFields["future"]) != "false" {
							t.Fatal("native fields or future data lost")
						}
					} else if v.Region != nil || v.LastMovementTimestamp != nil || v.PreferredDataLocation != nil || string(v.RawFields["Region"]) != "null" || string(v.RawFields["LastMovementTimeStamp"]) != "null" || string(v.RawFields["PreferredDataLocation"]) != "null" {
						t.Fatal("nullable data fabricated or lost")
					}
				}
				return contracttest.Observation{Response: r, Err: err}
			}, Want: contracttest.Expected{Status: 200, Body: []byte(body)}})
		}
	}
}
