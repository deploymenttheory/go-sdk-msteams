package tenantconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_configuration"
)

const route = "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TenantConfiguration"
const before = `{"Identity":"Global","MaxAllowedDomains":4294967295,"MaxBlockedDomains":1,"Future":[]}`
const after = `{"Identity":"Global","MaxAllowedDomains":null,"MaxBlockedDomains":1,"Future":[]}`

// Bodies below are authored model vectors constrained by the pinned setters and
// serializer observation, not successful tenant responses.
func wire(method, path, command, body string, status int, reply string) testutil.Exchange {
	headers := http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": nil}
	var input json.RawMessage
	if body != "" {
		input = json.RawMessage(body)
		headers.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: route + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("conditional TenantConfiguration contract differed: %v %+v", err, result.Differences())
	}
}

func TestUnitConditionalGetIdentityFilterAndNullableProjection(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "get-identity-filter", Steps: []testutil.Exchange{
		wire("GET", "/configuration/Global", "Get-CsTenantConfiguration", "", 200, before),
		wire("GET", "", "Get-CsTenantConfiguration", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsTenantConfiguration", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsTenantConfiguration", "", 200, `[`+before+`]`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TenantConfigurations
		value, _, err := service.GetByIdentity(ctx, "Global")
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if value == nil || value.Identity == nil || *value.Identity != "Global" || value.MaxAllowedDomains == nil || *value.MaxAllowedDomains != ^uint32(0) || value.MaxBlockedDomains == nil || *value.MaxBlockedDomains != 1 || string(value.RawFields["Future"]) != "[]" {
			t.Fatal("conditional typed identity/presence projection changed")
		}
		values, _, err := service.List(ctx)
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if len(values) != 1 {
			t.Fatal("omitted filter did not preserve the collection")
		}
		values, _, err = service.ListFiltered(ctx, " ")
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if len(values) != 0 {
			t.Fatal("whitespace filter was trimmed instead of matched literally")
		}
		values, _, err = service.ListFiltered(ctx, "Glo*")
		if err == nil && (len(values) != 1 || values[0].Identity == nil || *values[0].Identity != "Global") {
			t.Fatal("conditional local filter discarded Global identity")
		}
		return testutil.Observation{Err: err}
	}})
}

func TestUnitConditionalSetPreservesExplicitNullAndPreparationReceipt(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "set-null", Steps: []testutil.Exchange{
		wire("GET", "/configuration/Global", "Set-CsTenantConfiguration", "", 200, before),
		wire("PATCH", "/configuration/Global", "Set-CsTenantConfiguration", `{"Identity":"Global","MaxAllowedDomains":null}`, 200, after),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, response, err := sdk.TeamsAdminAPI.TenantConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: config.Settings{MaxAllowedDomains: policy.Null[uint32]()}})
		if err == nil && (result == nil || result.Before == nil || result.After == nil || result.After.MaxAllowedDomains != nil || string(result.After.RawFields["MaxAllowedDomains"]) != "null" || !result.PatchInvoked || !result.HTTPAccepted || result.ReadResponse == nil || response == nil) {
			t.Fatal("conditional set lost explicit null or receipt stages")
		}
		return testutil.Observation{Err: err}
	}})
}

func TestUnitConditionalSetRetainsReadDenialAndAcceptedDecodeFailure(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"preparation-denied", "accepted-invalid-body"} {
		t.Run(scenario, func(t *testing.T) {
			var steps []testutil.Exchange
			var want error
			if scenario == "preparation-denied" {
				steps = []testutil.Exchange{wire("GET", "/configuration/Global", "Set-CsTenantConfiguration", "", 403, `{"code":"Forbidden"}`)}
				want = auth.ErrForbidden
			} else {
				steps = []testutil.Exchange{wire("GET", "/configuration/Global", "Set-CsTenantConfiguration", "", 200, before), wire("PATCH", "/configuration/Global", "Set-CsTenantConfiguration", `{"Identity":"Global","MaxAllowedDomains":null}`, 200, `[]`)}
				want = client.ErrDecode
			}
			run(t, testutil.Case{Name: scenario, Steps: steps, Want: testutil.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, response, err := sdk.TeamsAdminAPI.TenantConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: config.Settings{MaxAllowedDomains: policy.Null[uint32]()}})
				if !errors.Is(err, want) || result == nil || result.ReadResponse == nil || response == nil {
					t.Fatal("conditional partial result lost its error or response")
				}
				if scenario == "preparation-denied" && (result.PatchInvoked || result.HTTPAccepted) {
					t.Fatal("denied preparation invoked a write")
				}
				if scenario == "accepted-invalid-body" && (result.Before == nil || !result.PatchInvoked || !result.HTTPAccepted || result.After != nil) {
					t.Fatal("accepted decode failure lost completed stages")
				}
				return testutil.Observation{Err: err}
			}})
		})
	}
}

func TestUnitConditionalSelectedSettingValidationPrecedesCredentials(t *testing.T) {
	t.Parallel()
	for _, settings := range []config.Settings{{MaxAllowedDomains: policy.Set(uint32(0))}, {MaxBlockedDomains: policy.Set(uint32(0))}, {}} {
		run(t, testutil.Case{Name: "invalid-setting", Want: testutil.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, _, err := sdk.TeamsAdminAPI.TenantConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: settings})
			return testutil.Observation{Err: err}
		}})
	}
}

func TestUnitConditionalResponseSnapshotPreservesRawFields(t *testing.T) {
	t.Parallel()
	var value config.Configuration
	if json.Unmarshal([]byte(before), &value) != nil {
		t.Fatal("conditional model failed to decode")
	}
	got, err := json.Marshal(value)
	var actual, expected any
	if err != nil || json.Unmarshal(got, &actual) != nil || json.Unmarshal([]byte(before), &expected) != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatal("conditional model lost complete response snapshot")
	}
	for _, raw := range []string{`{"Identity":"Global","MaxAllowedDomains":0}`, `{"Identity":"Global","MaxBlockedDomains":4294967296}`, `[]`, `null`} {
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("conditional model accepted an invalid known setting")
		}
	}
}

func TestUnitConditionalSparsePatchOmitsUnboundFieldsAndKeepsFalseForce(t *testing.T) {
	t.Parallel()
	flag := false
	run(t, testutil.Case{Name: "sparse-false-force", Steps: []testutil.Exchange{wire("PATCH", "/configuration/Global", "Set-CsTenantConfiguration", `{"Identity":"Global","MaxAllowedDomains":1,"Force":"False"}`, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, _, err := sdk.TeamsAdminAPI.TenantConfigurations.PatchByIdentity(ctx, "Global", config.PatchRequest{Settings: config.Settings{MaxAllowedDomains: policy.Set[uint32](1)}, Controls: policy.Controls{Force: &flag}})
		return testutil.Observation{Err: err}
	}})
}

func TestUnitConditionalExplicitEmptyFilterFailsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "explicit-empty-filter", Want: testutil.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, _, err := sdk.TeamsAdminAPI.TenantConfigurations.ListFiltered(ctx, "")
		return testutil.Observation{Err: err}
	}})
}

func TestUnitConditionalForceAloneIsNotASetting(t *testing.T) {
	t.Parallel()
	flag := false
	for _, patch := range []bool{false, true} {
		run(t, testutil.Case{Name: "force-only", Want: testutil.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			var e error
			if patch {
				_, _, e = sdk.TeamsAdminAPI.TenantConfigurations.PatchByIdentity(ctx, "Global", config.PatchRequest{Controls: policy.Controls{Force: &flag}})
			} else {
				_, _, e = sdk.TeamsAdminAPI.TenantConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Controls: policy.Controls{Force: &flag}})
			}
			return testutil.Observation{Err: e}
		}})
	}
}
