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
			if f.Request.FailureType != nil {
				c.Want.Error = client.ErrDecode
			}
			if len(c.Steps) != 0 {
				c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
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
	run(t, contracttest.Case{Name: "prepared-update-rejected", Steps: []contracttest.Exchange{
		wire("GET", path, "Set-CsTeamsExternalAccessConfiguration", nil, 200, string(before)),
		wire("PATCH", path, "Set-CsTeamsExternalAccessConfiguration", json.RawMessage(`{"Identity":"Global","BlockedUsers":[],"BlockExternalUserAccess":null}`), 400, "")},
		Want: contracttest.Expected{JSON: before, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Fields: config.Fields{BlockExternalUserAccess: policy.Null[bool]()}, BlockedUsers: policy.ChangeStrings(nil, nil)})
			if value == nil || !value.PatchInvoked || value.HTTPAccepted || value.Before == nil || value.After != nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 {
				t.Fatal("prepared rejected update lost stage outcomes")
			}
			return contracttest.Observation{Value: value.Before.RawFields, Response: response, Err: err}
		},
	})
}
