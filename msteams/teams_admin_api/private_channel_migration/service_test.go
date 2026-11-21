package privatechannelmigration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/private_channel_migration"
)

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "private_channel_migration", "msteams/teams_admin_api/private_channel_migration/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("migration %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
	return r
}
func wire(origin, prefix, audience string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + prefix + "/channels/privateChannelMigrationStatus", Audience: audience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-Target-Uri": nil}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func TestUnitMigrationCloudRequestPrefixesAndAudiences(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ cloud, origin, prefix, audience string }{
		{"commercial", "https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com"},
		{"gcc", "https://teams.microsoft.com", "/fabric/gcc/templates/api", "https://templates.teams.microsoft.com"},
		{"gcch", "https://gov.teams.microsoft.us", "/fabric/gov/templates/api", "https://templates.teams.microsoft.us"},
		{"dod", "https://dod.teams.microsoft.us", "/fabric/dod/templates/api", "https://templates.teams.microsoft.us"},
		{"china", "https://teams.microsoftonline.cn", "/fabric/templates/api", "https://templates.teams.microsoftonline.cn"},
	} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			run(t, contracttest.Case{Name: tc.cloud + "-" + string(mode), Cloud: tc.cloud, Mode: mode, Mappings: []contracttest.Mapping{{Command: "Get-TenantPrivateChannelMigrationStatus", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{wire(tc.origin, tc.prefix, tc.audience, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
				var value any
				if v != nil {
					details, decodeErr := v.DecodeDetails()
					if decodeErr != nil {
						t.Error("typed detail decode failed")
					}
					value = map[string]any{"tenant": v.TenantID, "status": v.MigrationStatus, "checked": v.LastChecked, "start": v.MigrationStartTimestamp, "details": details, "future": v.RawFields["future"]}
				}
				return contracttest.Observation{Value: value, Response: r, Err: err}
			}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		}
	}
}

func TestUnitMigrationProfilesDoNotFallbackAndWrongBackendIsLocal(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"france", "germany"} {
		run(t, contracttest.Case{Name: cloud, Cloud: cloud, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Error: client.ErrDestination}})
	}
	run(t, contracttest.Case{Name: "wrong-backend", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := migration.NewService(sdk.TeamsAdminAPI.Scope())
		v, r, err := s.Get(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: client.ErrDestination}})
	run(t, contracttest.Case{Name: "cancelled", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
}

func TestUnitMigrationMalformedDetailsRemainExplicitCallerData(t *testing.T) {
	t.Parallel()
	var value migration.Status
	if err := json.Unmarshal([]byte(`{"MigrationStatus":"RequiresAdminAttention","Details":"not-json","LastChecked":null,"future":false}`), &value); err != nil || value.Details == nil {
		t.Fatal("local status decoding lost details")
	}
	if _, err := value.DecodeDetails(); err == nil {
		t.Fatal("malformed detail text accepted")
	}
	for _, body := range []string{`null`, `[]`, `{"Details":{}}`, `{"LastChecked":17}`} {
		run(t, contracttest.Case{Name: "malformed-status", Steps: []contracttest.Exchange{wire("https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
	run(t, contracttest.Case{Name: "forbidden", Steps: []contracttest.Exchange{wire("https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com", 403, `{"error":{"code":"Denied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
}
