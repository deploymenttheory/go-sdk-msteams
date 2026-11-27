package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitExternalAccessTypedProviderDispatch(t *testing.T) {
	t.Parallel()

	seen := 0
	for _, f := range requestInputs(t) {
		if strings.HasPrefix(f.Name, "acs-") || strings.HasPrefix(f.Name, "external-") {
			continue
		}
		seen++
		command, set := "Get-CsExternalAccessPolicy", "Identity"
		switch f.Request.Method {
		case "POST":
			if strings.Contains(f.Request.Url, "groupPolicyAssignments") {
				command = "Grant-CsExternalAccessPolicy"
				set = "GrantToGroup"
			} else {
				command = "New-CsExternalAccessPolicy"
			}
		case "PATCH":
			command = "Set-CsExternalAccessPolicy"
			if f.Name == "user" {
				command = "Grant-CsExternalAccessPolicy"
				set = "GrantToUser"
			}
			if f.Name == "tenant" {
				command = "Grant-CsExternalAccessPolicy"
				set = "GrantToTenant"
			}
		case "DELETE":
			command = "Remove-CsExternalAccessPolicy"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{exchange(f)}}
			// Retained rejected-wire vectors do not exercise result decoding.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.ExternalAccessPolicies
				if f.Name == "list" || f.Name == "list-envelope" {
					_, r, e := s.List(ctx)

					return contracttest.Observation{Response: r, Err: e}
				}
				var rErr error
				if f.Name == "delete" || f.Name == "delete-force-false" || f.Name == "delete-denied" {
					controls := policy.Controls{}
					if f.Name == "delete-force-false" {
						no := false
						controls.Force = &no
					}
					r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Name == "create" {
					value, r, e := s.Create(ctx, external.CreateRequest{Identity: scoped, Settings: external.Settings{Fields: external.Fields{Description: policy.Set(""), EnableFederationAccess: policy.Set(false)}}})
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				if f.Name == "update" {
					_, r, e := s.PatchByIdentity(ctx, scoped, external.PatchRequest{Settings: external.Settings{Fields: external.Fields{Description: policy.Set(""), EnableFederationAccess: policy.Set(false)}}, BodyIdentity: policy.Set("read-model-identity")})

					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Name == "user" {
					value, r, e := s.AssignUser(ctx, "user/fixture", nil)
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				if f.Name == "tenant" {
					empty := ""
					value, r, e := s.AssignTenantDefault(ctx, &empty, false)
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				if f.Name == "group" || f.Name == "group-accepted-opaque" {
					value, r, e := s.AssignGroup(ctx, "group/fixture", nil, -1)

					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				id := scoped
				if f.Name == "read-tagged" {
					id = "tag:Fixture / Policy"
				}
				_, r, rErr := s.GetByIdentity(ctx, id)

				return contracttest.Observation{Response: r, Err: rErr}
			}
			run(t, c)
		}
	}
	if seen != 16 {
		t.Fatal("external provider fixture count differs")
	}
}

func TestUnitExternalAccessComposedUpdate(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	path := "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/Global"
	run(t, contracttest.Case{Name: "prepared-update-rejected", Steps: []contracttest.Exchange{
		wire("GET", path, "Set-CsExternalAccessPolicy", nil, 200, string(before)),
		wire("PATCH", path, "Set-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"Global","EnableFederationAccess":null,"AllowedExternalDomains":["b","a"]}`), 400, "")},
		Want: contracttest.Expected{JSON: before, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.UpdateByIdentity(ctx, "Global", external.UpdateRequest{Fields: external.Fields{EnableFederationAccess: policy.Null[bool]()}, AllowedExternalDomains: policy.ChangeStrings([]string{"b", "a"}, []string{"a", "b"})})
			if value == nil || !value.PatchInvoked || value.HTTPAccepted || value.Before == nil || value.After != nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 {
				t.Fatal("prepared rejected update lost stage outcomes")
			}
			return contracttest.Observation{Value: value.Before.RawFields, Response: response, Err: err}
		},
	})
}
