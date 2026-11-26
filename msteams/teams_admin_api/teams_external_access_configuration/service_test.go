package teamsexternalaccessconfiguration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
)

func TestUnitTypedProviderDispatch(t *testing.T) {
	t.Parallel()

	seen := 0
	for _, f := range requestInputs(t) {
		if !strings.HasPrefix(f.Name, "external-") {
			continue
		}
		seen++
		command := "Get-CsTeamsExternalAccessConfiguration"
		if f.Request.Method == "PATCH" {
			command = "Set-CsTeamsExternalAccessConfiguration"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{exchange(f)}}
			// These retained vectors prove rejected request bytes, not decoding.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations
				if strings.HasSuffix(f.Name, "-list") && f.Name != "external-null-list" {
					_, r, e := s.List(ctx)

					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Request.Method == "PATCH" {
					_, r, e := s.PatchByIdentity(ctx, scoped, config.PatchRequest{Settings: config.Settings{Fields: config.Fields{BlockExternalUserAccess: policy.Null[bool]()}, BlockedUsers: policy.Strings()}, BodyIdentity: policy.Set("read-model-identity")})

					return contracttest.Observation{Response: r, Err: e}
				}
				_, r, e := s.GetByIdentity(ctx, scoped)

				return contracttest.Observation{Response: r, Err: e}
			}
			run(t, c)
		}
	}
	if seen != 4 {
		t.Fatal("provider fixture count differs")
	}
}

func TestUnitComposedUpdatePreservesReadIdentityAndSparseFields(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	path := "/Skype.Policy/configurations/TeamsExternalAccessConfiguration/configuration/Global"
	// Retained GET bytes plus authored empty 204 PATCH acceptance. The
	// compiled PolicyRP Set path and these sparse/list bindings are traced in
	// the published POL04 primary-source audit (dc2bb9a); no persistence or
	// nonempty PATCH response is inferred. Keep the original 400 proof too.
	for _, tc := range []struct {
		name   string
		status int
	}{{"rejected", 400}, {"accepted", 204}} {
		t.Run(tc.name, func(t *testing.T) {
			want := contracttest.Expected{JSON: before}
			if tc.status == 400 {
				want.APIError = &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}
			}
			run(t, contracttest.Case{Name: "prepared-update-" + tc.name, Source: "msteams/teams_admin_api/teams_external_access_configuration/service_test.go", Steps: []contracttest.Exchange{
				wire("GET", path, "Set-CsTeamsExternalAccessConfiguration", nil, 200, string(before)),
				wire("PATCH", path, "Set-CsTeamsExternalAccessConfiguration", json.RawMessage(`{"Identity":"Global","BlockedUsers":[],"BlockExternalUserAccess":null}`), tc.status, "")},
				Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					value, response, err := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Fields: config.Fields{BlockExternalUserAccess: policy.Null[bool]()}, BlockedUsers: policy.ChangeStrings(nil, nil)})
					if value == nil || !value.PatchInvoked || value.HTTPAccepted != (tc.status == 204) || value.Before == nil || value.After != nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 || value.Before.Identity == nil || *value.Before.Identity != "Global" || value.Before.BlockExternalUserAccess == nil || *value.Before.BlockExternalUserAccess || len(value.Before.BlockedUsers) != 0 {
						t.Fatal("prepared update lost acceptance or stage outcomes")
					}
					return contracttest.Observation{Value: value.Before.RawFields, Response: response, Err: err}
				},
			})
		})
	}
}
