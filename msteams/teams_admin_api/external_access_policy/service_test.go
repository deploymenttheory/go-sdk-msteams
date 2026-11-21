package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestExternalAccessPinnedProviderDispatch(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	seen := 0
	for _, f := range captured(t) {
		if strings.HasPrefix(f.Name, "acs-") || strings.HasPrefix(f.Name, "external-") {
			continue
		}
		seen++
		command, set := "Get-CsExternalAccessPolicy", "Identity"
		switch f.Capture.Method {
		case "POST":
			if strings.Contains(f.Capture.Url, "groupPolicyAssignments") {
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
			if f.Capture.FailureType != nil {
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				} else {
					c.Want.Error = client.ErrDecode
				}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.ExternalAccessPolicies
				if f.Name == "list" || f.Name == "list-envelope" {
					v, r, e := s.List(ctx)
					if e == nil && (v == nil || len(v) != 0) {
						e = errors.New("empty provider list changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
				var v *external.Policy
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
					value, r, e := s.PatchByIdentity(ctx, scoped, external.PatchRequest{Settings: external.Settings{Fields: external.Fields{Description: policy.Set(""), EnableFederationAccess: policy.Set(false)}}, BodyIdentity: policy.Set("read-model-identity")})
					if e == nil && value != nil {
						e = errors.New("empty update invented a result")
					}
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
					if e == nil && f.Name == "group-accepted-opaque" && (value == nil || string(value.RawFields["OperationId"]) != `"fixture-operation"`) {
						e = errors.New("opaque accepted fields lost")
					}
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				id := scoped
				if f.Name == "read-tagged" {
					id = "tag:Fixture / Policy"
				}
				v, r, rErr := s.GetByIdentity(ctx, id)
				if rErr == nil && f.Name == "read-no-content" && v != nil {
					rErr = errors.New("empty read invented a resource")
				}
				if rErr == nil && f.Name == "read-future-enum" && (v.CommunicationWithExternalOrgs == nil || *v.CommunicationWithExternalOrgs != 7) {
					rErr = errors.New("future enum lost")
				}
				return contracttest.Observation{Response: r, Err: rErr}
			}
			results = append(results, run(t, c))
		}
	}
	if seen != 16 {
		t.Fatal("external provider fixture count differs")
	}
	evidence(t, results)
}

func TestExternalAccessComposedUpdate(t *testing.T) {
	t.Parallel()
	before := `{"Identity":"read-model-identity","AllowedExternalDomains":["a","a","A","b"],"BlockedExternalDomains":[],"Description":"preserved","Future":{"enabled":false}}`
	input := external.UpdateRequest{Fields: external.Fields{EnableFederationAccess: policy.Null[bool]()}, AllowedExternalDomains: policy.ChangeStrings([]string{"b", "a"}, []string{"a", "b"})}
	result := run(t, contracttest.Case{Name: "composed-update", Mappings: []contracttest.Mapping{{Command: "Set-CsExternalAccessPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		wire("GET", "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child", "Set-CsExternalAccessPolicy", nil, 200, before),
		wire("PATCH", "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child", "Set-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"read-model-identity","EnableFederationAccess":null,"AllowedExternalDomains":["a","A","b","a"]}`), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.UpdateByIdentity(ctx, scoped, input)
		if e == nil && (v == nil || !v.PatchInvoked || !v.HTTPAccepted || v.Before == nil || v.After != nil || v.ReadResponse == nil) {
			e = errors.New("update preparation/acceptance lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{result})
}
