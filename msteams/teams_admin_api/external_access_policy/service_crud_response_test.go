package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitOwnedCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const initial = "SDK lab non-default initial description"
	const updated = "SDK lab non-default updated description"
	const base = "/Skype.Policy/configurations/ExternalAccessPolicy"
	const tagged = base + "/configuration/fixture-policy"
	create := readResponseFixture(t, "create_201.body")
	before := readResponseFixture(t, "create_readback.json")
	patch := readResponseFixture(t, "update_204.body")
	after := readResponseFixture(t, "update_readback.json")
	list := readResponseFixture(t, "owned_list.json")
	deleted := readResponseFixture(t, "delete_204.body")
	missing := readResponseFixture(t, "get_deleted_404.json")
	steps := []contracttest.Exchange{
		wire("POST", base, "New-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"`+initial+`","EnableFederationAccess":false}`), 201, string(create)),
		wire("GET", base+"/configuration/fixture-policy", "Get-CsExternalAccessPolicy", nil, 200, string(before)),
		wire("GET", tagged, "Set-CsExternalAccessPolicy", nil, 200, string(before)),
		wire("PATCH", tagged, "Set-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"`+updated+`"}`), 204, string(patch)),
		wire("GET", tagged, "Get-CsExternalAccessPolicy", nil, 200, string(after)),
		wire("GET", base, "Get-CsExternalAccessPolicy", nil, 200, string(list)),
		wire("DELETE", tagged, "Remove-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"fixture-policy"}`), 204, string(deleted)),
		wire("GET", tagged, "Get-CsExternalAccessPolicy", nil, 404, string(missing)),
	}
	for i := range steps {
		steps[i].Reply.Headers.Set("Content-Type", "application/json")
	}
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(missing, &envelope) != nil {
		t.Fatal("invalid error fixture")
	}
	run(t, contracttest.Case{Name: "owned-crud", Source: "msteams/teams_admin_api/external_access_policy/testdata/create_readback.json", Steps: steps,
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ExternalAccessPolicies
			_, response, err := s.Create(ctx, resource.CreateRequest{Identity: "fixture-policy", Settings: resource.Settings{Fields: resource.Fields{Description: policy.Set(initial), EnableFederationAccess: policy.Set(false)}}})
			if err != nil || response == nil || response.StatusCode() != 201 {
				t.Fatal("create request or acceptance changed")
			}
			value, response, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "fixture-policy") || value.Description == nil || *value.Description != initial || value.EnableFederationAccess == nil || *value.EnableFederationAccess {
				t.Fatal("actual create response was decoded incorrectly")
			}
			changed, response, err := s.UpdateByIdentity(ctx, "fixture-policy", resource.UpdateRequest{Fields: resource.Fields{Description: policy.Set(updated)}})
			if err != nil || response == nil || response.StatusCode() != 204 || changed == nil || !changed.HTTPAccepted || changed.Before == nil || changed.Before.Description == nil || *changed.Before.Description != initial || !changed.PatchInvoked {
				t.Fatal("sparse update lost preparation or HTTP acceptance")
			}
			value, response, err = s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Description == nil || *value.Description != updated || value.EnableFederationAccess == nil || *value.EnableFederationAccess || value.RawFields["ConfigMetadata"] == nil {
				t.Fatal("actual update response lost the changed or omitted setting")
			}
			values, response, err := s.List(ctx)
			if err != nil || response == nil || response.StatusCode() != 200 {
				t.Fatal("actual owned list response was not decoded")
			}
			found := 0
			for _, item := range values {
				if item != nil && item.Identity != nil && *item.Identity == "fixture-policy" {
					found++
				}
			}
			if found != 1 {
				t.Fatal("owned policy missing or duplicated")
			}
			deleted, err := s.DeleteByIdentity(ctx, "fixture-policy", policy.Controls{})
			if err != nil || deleted.StatusCode() != 204 {
				t.Fatal("actual delete response changed")
			}
			_, response, err = s.GetByIdentity(ctx, "fixture-policy")
			var api *client.APIError
			if !errors.As(err, &api) || api.StatusCode != 404 {
				t.Fatal("actual absence response lost its typed error")
			}
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Status: 404, Body: missing, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}})
}

// Reuse retained primitive GET/empty Create201 bytes. This comparison/create
// sequence and blank 403 are authored offline, not a captured compound run.
// BaseTpmNewCmdlet compares Global but continues creation after comparison
// failure; its bound-parameter serializer never clones the Global defaults.
func TestUnitCreateComparisonRetainsSuccessfulIndependentSubmission(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	accepted := readResponseFixture(t, "create_201.body")
	var snapshot map[string]json.RawMessage
	if json.Unmarshal(before, &snapshot) != nil {
		t.Fatal("invalid retained Global snapshot")
	}
	for _, status := range []int{200, 403} {
		reply := before
		if status == 403 {
			reply = nil
		}
		run(t, contracttest.Case{Name: "comparison-accepted", Source: "msteams/teams_admin_api/external_access_policy/service_crud_response_test.go", Mappings: []contracttest.Mapping{{Command: "New-CsExternalAccessPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
			wire("GET", "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/Global", "New-CsExternalAccessPolicy", nil, status, string(reply)),
			wire("POST", "/Skype.Policy/configurations/ExternalAccessPolicy", "New-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"SDK lab non-default initial description","EnableFederationAccess":false}`), 201, string(accepted)),
		}, Want: contracttest.Expected{Body: accepted}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "fixture-policy", Settings: resource.Settings{Fields: resource.Fields{Description: policy.Set("SDK lab non-default initial description"), EnableFederationAccess: policy.Set(false)}}})
			if err != nil || response == nil || response.StatusCode() != 201 || value == nil || !value.CreateInvoked || !value.HTTPAccepted || value.Created != nil || value.ComparisonResponse == nil || value.ComparisonResponse.StatusCode() != status {
				t.Fatal("successful sparse creation lost independent comparison or acceptance stages")
			}
			if status == 200 {
				if value.ComparisonError != nil || value.Comparison == nil || !reflect.DeepEqual(value.Comparison.RawFields, snapshot) {
					t.Fatal("successful comparison lost or mutated its Global snapshot")
				}
			} else if value.Comparison != nil || !errors.Is(value.ComparisonError, auth.ErrForbidden) {
				t.Fatal("comparison denial was swallowed or blocked successful independent creation")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	}
}
