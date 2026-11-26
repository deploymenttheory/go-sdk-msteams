package acsfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitTypedProviderDispatch(t *testing.T) {
	t.Parallel()

	seen := 0
	for _, f := range requestInputs(t) {
		if !strings.HasPrefix(f.Name, "acs-") {
			continue
		}
		seen++
		command := "Get-CsTeamsAcsFederationConfiguration"
		if f.Request.Method == "PATCH" {
			command = "Set-CsTeamsAcsFederationConfiguration"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{exchange(f)}}
			// These retained vectors prove rejected request bytes, not decoding.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.ACSFederationConfigurations
				if strings.HasSuffix(f.Name, "-list") && f.Name != "acs-null-list" {
					_, r, e := s.List(ctx)

					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Request.Method == "PATCH" {
					_, r, e := s.PatchByIdentity(ctx, scoped, config.PatchRequest{Settings: config.Settings{Fields: config.Fields{EnableAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.Strings()}, BodyIdentity: policy.Set("read-model-identity")})

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
	path := "/Skype.Policy/configurations/TeamsAcsFederationConfiguration/configuration/Global"
	run(t, contracttest.Case{Name: "prepared-update-rejected", Steps: []contracttest.Exchange{
		wire("GET", path, "Set-CsTeamsAcsFederationConfiguration", nil, 200, string(before)),
		wire("PATCH", path, "Set-CsTeamsAcsFederationConfiguration", json.RawMessage(`{"Identity":"Global","AllowedAcsResources":["b","a"],"HideBannerForAllowedAcsUsers":false}`), 400, "")},
		Want: contracttest.Expected{JSON: before, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ACSFederationConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Fields: config.Fields{HideBannerForAllowedAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.ChangeStrings([]string{"b", "a"}, []string{"a", "b"})})
			if value == nil || !value.PatchInvoked || value.HTTPAccepted || value.Before == nil || value.After != nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 {
				t.Fatal("prepared rejected update lost stage outcomes")
			}
			return contracttest.Observation{Value: value.Before.RawFields, Response: response, Err: err}
		},
	})
}
