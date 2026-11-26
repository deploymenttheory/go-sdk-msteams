package tenantfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

// Authored success against the pinned generated Set/PUT/204 contract and
// explicit request-list normalizer, not a captured tenant mutation. Source
// owners and null/operator/route distinctions are documented in the published
// POL90 audit's generated helper and domain-constructor source notes. AllowedDomains response choice is deliberately
// not converted back into this request's flattened string list.
func TestUnitTenantFederationAuthoredDomainBuilderAndPUTAcceptance(t *testing.T) {
	t.Parallel()
	domain, err := resource.NewDomainPattern("fixture.example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		list []string
		body json.RawMessage
	}{
		{"domain", []string{domain.Domain}, json.RawMessage(`{"Identity":"Global","AllowFederatedUsers":false,"AllowedDomains":["fixture.example.invalid"]}`)},
		{"explicit-empty", []string{}, json.RawMessage(`{"Identity":"Global","AllowFederatedUsers":false,"AllowedDomains":[]}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := resource.NewAllowList(tc.list...)
			input := resource.UpdateRequest{Settings: resource.Settings{AllowFederatedUsers: new(false)}, AllowedDomains: value}
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: "put-" + tc.name, Family: "tenant_federation_configuration", Scope: "service", Source: "msteams/teams_admin_api/tenant_federation_configuration/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(t), wire("PUT", "/Skype.Policy/configurations/TenantFederationSettings/configuration/Global", setCommand, tc.body, 204, "")}, Want: contracttest.Expected{Status: 204, Body: []byte{}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, input)
				if err == nil && (response == nil || response.StatusCode() != 204) {
					t.Fatal("PUT acceptance lost its no-content response")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("authored tenant federation PUT differs: %v %+v", err, r.Differences())
			}
		})
	}
}
