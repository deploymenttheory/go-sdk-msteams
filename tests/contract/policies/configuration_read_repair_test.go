package policies_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

// Inputs are authored from actual own pinned DTO read projections. These are
// successful SDK HTTP branches, not captured tenant responses or write effects.
func TestUnitConfigurationSixOwnScalarReadsKeepTypedAndRawViewsSeparate(t *testing.T) {
	t.Parallel()
	type read func(context.Context, *msteams.Client) (any, *resty.Response, error)
	cases := []struct {
		name, resource, body string
		expected             map[string]string
		read                 read
	}{
		{"validation", "OnlineVoicemailValidationConfiguration", `{"Identity":37,"AudioFileValidationEnabled":1,"AudioFileValidationUri":null,"Future":false}`, map[string]string{"Identity": `"37"`, "AudioFileValidationEnabled": "true"}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.GetByIdentity(ctx, "Global")
		}},
		{"upgrade", "TeamsUpgradeConfiguration", `{"Identity":false,"DownloadTeams":"True","BlockLegacyAuthorization":0,"SfBMeetingJoinUx":37,"Future":false}`, map[string]string{"Identity": `"false"`, "DownloadTeams": "true", "BlockLegacyAuthorization": "false", "SfBMeetingJoinUx": `"37"`}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.TeamsUpgradeConfigurations.GetByIdentity(ctx, "Global")
		}},
		{"sip", "TeamsSipDevicesConfiguration", `{"Identity":1,"BulkSignIn":false,"Future":false}`, map[string]string{"Identity": `"1"`, "BulkSignIn": `"false"`}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.TeamsSipDevicesConfigurations.GetByIdentity(ctx, "Global")
		}},
		{"mto", "TeamsMultiTenantOrganizationConfiguration", `{"Identity":37,"CopilotFromHomeTenant":1e+03,"Future":false}`, map[string]string{"Identity": `"37"`, "CopilotFromHomeTenant": `"1e+03"`}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.GetByIdentity(ctx, "Global")
		}},
		{"tenant-abuse", "TeamsTenantAbuseConfiguration", `{"Identity":37,"LowSeatLimit":2.5,"SkipValidation":false,"BlockedTenants":[null,1,false],"Future":false}`, map[string]string{"Identity": `"37"`, "LowSeatLimit": "2", "SkipValidation": `"false"`, "BlockedTenants": `[null,"1","false"]`}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.TeamsTenantAbuseConfigurations.GetByIdentity(ctx, "Global")
		}},
		{"messaging", "TeamsMessagingConfiguration", `{"Identity":37,"CustomEmojis":"False","Storyline":2147483648,"FileTypeBlockList":[null,1,false],"Future":false}`, map[string]string{"Identity": `"37"`, "CustomEmojis": "false", "Storyline": "-2147483648", "FileTypeBlockList": `[null,"1","false"]`}, func(ctx context.Context, s *msteams.Client) (any, *resty.Response, error) {
			return s.TeamsAdminAPI.TeamsMessagingConfigurations.GetByIdentity(ctx, "Global")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := contracttest.Run(t.Context(), contracttest.Case{Name: tc.name, Family: "policies", Scope: "service", Source: "tests/contract/policies/configuration_read_repair_test.go", Steps: []contracttest.Exchange{pol91ConfigurationWire(tc.resource, "GET", "Get-Cs"+tc.resource, nil, 200, []byte(tc.body))}, Want: contracttest.Expected{Status: 200, Body: []byte(tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := tc.read(ctx, sdk)
				if e != nil || v == nil || r == nil || string(r.Body()) != tc.body {
					t.Fatal("own scalar read lost value or raw receipt")
				}
				value := reflect.ValueOf(v).Elem()
				for name, want := range tc.expected {
					raw, e := json.Marshal(value.FieldByName(name).Interface())
					if e != nil || string(raw) != want {
						t.Fatalf("typed %s differs from actual source", name)
					}
				}
				fields := value.FieldByName("RawFields").Interface().(map[string]json.RawMessage)
				var original map[string]json.RawMessage
				if json.Unmarshal([]byte(tc.body), &original) != nil || !reflect.DeepEqual(fields, original) {
					t.Fatal("coercion changed original raw tokens or future field")
				}
				return contracttest.Observation{Response: r}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("own scalar root read differs: %v %+v", err, result.Differences())
			}
		})
	}
}
