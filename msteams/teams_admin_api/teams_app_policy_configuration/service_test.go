package teamsapppolicyconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_policy_configuration"
)

const source = "msteams/teams_admin_api/teams_app_policy_configuration/service_test.go"
const route = "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPolicyConfiguration"
const before = `{"Identity":"Global","AppCatalogUri":"https://catalog.example.invalid/","ResourceUri":null,"Future":[]}`
const after = `{"Identity":"Global","AppCatalogUri":null,"ResourceUri":null,"Future":[]}`

// Bodies below are authored model vectors constrained by the pinned setters and
// serializer observation, not successful tenant responses.
func wire(method, path, command, body string, status int, reply string) contracttest.Exchange {
	headers := http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": nil}
	var input json.RawMessage
	if body != "" {
		input = json.RawMessage(body)
		headers.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: route + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Source = source
	c.Family = "teams_app_policy_configuration"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("conditional TeamsAppPolicyConfiguration contract differed: %v %+v", err, result.Differences())
	}
}

func TestUnitConditionalGetIdentityFilterAndNullableProjection(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "get-identity-filter", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/Global", "Get-CsTeamsAppPolicyConfiguration", "", 200, before),
		wire("GET", "", "Get-CsTeamsAppPolicyConfiguration", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsTeamsAppPolicyConfiguration", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsTeamsAppPolicyConfiguration", "", 200, `[`+before+`]`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations
		value, _, err := service.GetByIdentity(ctx, "Global")
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if value == nil || value.Identity == nil || *value.Identity != "Global" || value.AppCatalogURI == nil || *value.AppCatalogURI != "https://catalog.example.invalid/" || value.ResourceURI != nil || string(value.RawFields["Future"]) != "[]" {
			t.Fatal("conditional typed identity/presence projection changed")
		}
		values, _, err := service.List(ctx)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if len(values) != 1 {
			t.Fatal("omitted filter did not preserve the collection")
		}
		values, _, err = service.ListFiltered(ctx, " ")
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if len(values) != 0 {
			t.Fatal("whitespace filter was trimmed instead of matched literally")
		}
		values, _, err = service.ListFiltered(ctx, "Glo*")
		if err == nil && (len(values) != 1 || values[0].Identity == nil || *values[0].Identity != "Global") {
			t.Fatal("conditional local filter discarded Global identity")
		}
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitConditionalSetPreservesExplicitNullAndPreparationReceipt(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "set-null", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", "", 200, before),
		wire("PATCH", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", `{"Identity":"Global","AppCatalogUri":null}`, 200, after),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, response, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: config.Settings{AppCatalogURI: policy.Null[string]()}})
		if err == nil && (result == nil || result.Before == nil || result.After == nil || result.After.AppCatalogURI != nil || string(result.After.RawFields["AppCatalogUri"]) != "null" || !result.PatchInvoked || !result.HTTPAccepted || result.ReadResponse == nil || response == nil) {
			t.Fatal("conditional set lost explicit null or receipt stages")
		}
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitConditionalSetRetainsReadDenialAndAcceptedDecodeFailure(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"preparation-denied", "accepted-invalid-body"} {
		t.Run(scenario, func(t *testing.T) {
			var steps []contracttest.Exchange
			var want error
			if scenario == "preparation-denied" {
				steps = []contracttest.Exchange{wire("GET", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", "", 403, `{"code":"Forbidden"}`)}
				want = auth.ErrForbidden
			} else {
				steps = []contracttest.Exchange{wire("GET", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", "", 200, before), wire("PATCH", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", `{"Identity":"Global","AppCatalogUri":null}`, 200, `[]`)}
				want = client.ErrDecode
			}
			run(t, contracttest.Case{Name: scenario, Steps: steps, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, response, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: config.Settings{AppCatalogURI: policy.Null[string]()}})
				if !errors.Is(err, want) || result == nil || result.ReadResponse == nil || response == nil {
					t.Fatal("conditional partial result lost its error or response")
				}
				if scenario == "preparation-denied" && (result.PatchInvoked || result.HTTPAccepted) {
					t.Fatal("denied preparation invoked a write")
				}
				if scenario == "accepted-invalid-body" && (result.Before == nil || !result.PatchInvoked || !result.HTTPAccepted || result.After != nil) {
					t.Fatal("accepted decode failure lost completed stages")
				}
				return contracttest.Observation{Err: err}
			}})
		})
	}
}

func TestUnitConditionalSelectedSettingValidationPrecedesCredentials(t *testing.T) {
	t.Parallel()
	for _, settings := range []config.Settings{{AppCatalogURI: policy.Set("")}, {ResourceURI: policy.Set("http://resource.example.invalid/")}, {}} {
		run(t, contracttest.Case{Name: "invalid-setting", Scope: "local", Want: contracttest.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: settings})
			return contracttest.Observation{Err: err}
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
	for _, raw := range []string{`{"Identity":"Global","AppCatalogUri":"http://catalog.example.invalid/"}`, `[]`, `null`} {
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("conditional model accepted an invalid known setting")
		}
	}
}

func TestUnitConditionalSparsePatchOmitsUnboundFieldsAndKeepsFalseForce(t *testing.T) {
	t.Parallel()
	flag := false
	run(t, contracttest.Case{Name: "sparse-false-force", Steps: []contracttest.Exchange{wire("PATCH", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", `{"Identity":"Global","AppCatalogUri":"https://catalog.example.invalid/","Force":"False"}`, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.PatchByIdentity(ctx, "Global", config.PatchRequest{Settings: config.Settings{AppCatalogURI: policy.Set[string]("https://catalog.example.invalid/")}, Controls: policy.Controls{Force: &flag}})
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitConditionalURIPreservesPinnedPrefixAndFinalLFBehavior(t *testing.T) {
	t.Parallel()
	selected := "HTTPS://catalog.example.invalid/\n"
	body := `{"Identity":"Global","AppCatalogUri":"HTTPS://catalog.example.invalid/\n"}`
	run(t, contracttest.Case{Name: "uri-final-lf", Steps: []contracttest.Exchange{wire("PATCH", "/configuration/Global", "Set-CsTeamsAppPolicyConfiguration", body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.PatchByIdentity(ctx, "Global", config.PatchRequest{Settings: config.Settings{AppCatalogURI: policy.Set(selected)}})
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitConditionalExplicitEmptyFilterFailsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "explicit-empty-filter", Scope: "local", Want: contracttest.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.ListFiltered(ctx, "")
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitConditionalForceAloneIsNotASetting(t *testing.T) {
	t.Parallel()
	flag := false
	for _, patch := range []bool{false, true} {
		run(t, contracttest.Case{Name: "force-only", Scope: "local", Want: contracttest.Expected{Error: config.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var e error
			if patch {
				_, _, e = sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.PatchByIdentity(ctx, "Global", config.PatchRequest{Controls: policy.Controls{Force: &flag}})
			} else {
				_, _, e = sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Controls: policy.Controls{Force: &flag}})
			}
			return contracttest.Observation{Err: e}
		}})
	}
}
