package privatechannelmigration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/private_channel_migration"
)

func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("migration %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
	return r
}
func wire(origin, prefix, audience string, status int, body string) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + prefix + "/channels/privateChannelMigrationStatus", Audience: audience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-Target-Uri": nil, "X-MS-CmdletName": {"Get-TenantPrivateChannelMigrationStatus"}}}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
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
			run(t, testutil.Case{Name: tc.cloud + "-" + string(mode), Cloud: tc.cloud, Mode: mode, Steps: []testutil.Exchange{wire(tc.origin, tc.prefix, tc.audience, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
				var value any
				if v != nil {
					details, decodeErr := v.DecodeDetails()
					if decodeErr != nil {
						t.Error("typed detail decode failed")
					}
					value = map[string]any{"tenant": v.TenantID, "status": v.MigrationStatus, "checked": v.LastChecked, "start": v.MigrationStartTimestamp, "details": details, "future": v.RawFields["future"]}
				}
				return testutil.Observation{Value: value, Response: r, Err: err}
			}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden}})
		}
	}
}

func TestUnitMigrationProfilesDoNotFallbackAndWrongBackendIsLocal(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"france", "germany"} {
		run(t, testutil.Case{Name: cloud, Cloud: cloud, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Error: client.ErrDestination}})
	}
	run(t, testutil.Case{Name: "wrong-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := migration.NewService(sdk.TeamsAdminAPI.Scope())
		v, r, err := s.Get(ctx)
		return testutil.Observation{Value: v, Response: r, Err: err}
	}, Want: testutil.Expected{Error: client.ErrDestination}})
	run(t, testutil.Case{Name: "cancelled", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
		return testutil.Observation{Value: v, Response: r, Err: err}
	}, Want: testutil.Expected{Error: context.Canceled}})
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
		run(t, testutil.Case{Name: "malformed-status", Steps: []testutil.Exchange{wire("https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Status: 200, Error: client.ErrDecode}})
	}
	run(t, testutil.Case{Name: "forbidden", Steps: []testutil.Exchange{wire("https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com", 403, `{"error":{"code":"Denied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
		return testutil.Observation{Value: v, Response: r, Err: err}
	}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden}})
}

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func TestUnitMigrationSuccessfulNativeSnapshotAndExplicitDetailDecode(t *testing.T) {
	t.Parallel()
	for _, details := range []string{`{"pending":2}`, `not-json`} {
		body, _ := json.Marshal(map[string]any{"TenantId": "offline-tenant", "MigrationStatus": "RequiresAdminAttention", "Details": details, "LastChecked": "2026-11-20T12:00:00Z", "MigrationStartTimeStamp": nil, "MigrationCompletionTimeStamp": nil, "future": false})
		step := wire("https://teams.microsoft.com", "/fabric/templates/api", "https://templates.teams.microsoft.com", 200, string(body))

		run(t, testutil.Case{Name: "native-success", Steps: []testutil.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
			if err == nil {
				if v == nil || v.TenantID == nil || *v.TenantID != "offline-tenant" || v.MigrationStatus == nil || *v.MigrationStatus != "RequiresAdminAttention" || v.LastChecked == nil || *v.LastChecked != "2026-11-20T12:00:00Z" || v.Details == nil || *v.Details != details || v.MigrationStartTimestamp != nil || v.MigrationCompletionTimestamp != nil || string(v.RawFields["future"]) != "false" {
					t.Fatal("native snapshot lost typed, opaque or nullable data")
				}
				decoded, decodeErr := v.DecodeDetails()
				if details == "not-json" {
					if !errors.Is(decodeErr, client.ErrDecode) || *v.Details != details {
						t.Fatal("caller decode changed successful native result")
					}
				} else if decodeErr != nil || string(decoded["pending"]) != "2" {
					t.Fatal("explicit detail decoding failed")
				}
			}
			return testutil.Observation{Response: r, Err: err}
		}, Want: testutil.Expected{Status: 200, Body: body}})
	}
}
