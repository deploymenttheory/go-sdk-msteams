package policies_test

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
	acs "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
	tenant "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

// Captured ACS/Teams-external reads are combined with authored successful
// writes and denials to test independent preparation and partial receipts.
func TestUnitPolicyCompositionFederationProfilesPreserveIndependentPreparation(t *testing.T) {
	t.Parallel()
	beforeACS := policyCompositionFixture(t, "acs_federation_configuration", "get_global.json")
	beforeExternal := policyCompositionFixture(t, "teams_external_access_configuration", "get_global.json")
	discovery := policyCompositionFixture(t, "internal/configapi", "service_discovery.json")
	var acsSnapshot, externalSnapshot map[string]json.RawMessage
	if json.Unmarshal(beforeACS, &acsSnapshot) != nil || json.Unmarshal(beforeExternal, &externalSnapshot) != nil {
		t.Fatal("invalid retained federation snapshots")
	}
	steps := []testutil.Exchange{
		policyCompositionWire("GET", "TeamsAcsFederationConfiguration", "Global", "Set-CsTeamsAcsFederationConfiguration", nil, 200, beforeACS),
		policyCompositionWire("PATCH", "TeamsAcsFederationConfiguration", "Global", "Set-CsTeamsAcsFederationConfiguration", json.RawMessage(`{"Identity":"Global","AllowedAcsResources":["fixture-resource"],"HideBannerForAllowedAcsUsers":false}`), 204, nil),
		policyCompositionWire("GET", "TeamsExternalAccessConfiguration", "Global", "Set-CsTeamsExternalAccessConfiguration", nil, 200, beforeExternal),
		policyCompositionWire("PATCH", "TeamsExternalAccessConfiguration", "Global", "Set-CsTeamsExternalAccessConfiguration", json.RawMessage(`{"Identity":"Global","BlockedUsers":[],"BlockExternalUserAccess":null}`), 204, nil),
		{Request: testutil.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: discovery}},
		assignmentCompositionWire("PUT", "/Skype.Policy/configurations/TenantFederationSettings/configuration/Global", "Set-CsTenantFederationConfiguration", json.RawMessage(`{"Identity":"Global","AllowFederatedUsers":false,"AllowedDomains":[]}`), 403, "", true),
		policyCompositionWire("GET", "TeamsAcsFederationConfiguration", "Global", "Get-CsTeamsAcsFederationConfiguration", nil, 200, beforeACS),
	}
	r, err := testutil.Run(t.Context(), testutil.Case{Name: "federation-profile-isolation",

		Steps: steps, Want: testutil.Expected{Status: 200, Body: beforeACS, JSON: beforeACS}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			first, response, err := sdk.TeamsAdminAPI.ACSFederationConfigurations.UpdateByIdentity(ctx, "Global", acs.UpdateRequest{Fields: acs.Fields{HideBannerForAllowedAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.ChangeStrings([]string{"fixture-resource"}, nil)})
			if err != nil || response == nil || response.StatusCode() != 204 || first == nil || !first.HTTPAccepted || !first.PatchInvoked || first.After != nil || first.Before == nil || first.Before.Identity == nil || *first.Before.Identity != "Global" || first.Before.HideBannerForAllowedAcsUsers == nil || *first.Before.HideBannerForAllowedAcsUsers || len(first.Before.AllowedAcsResources) != 0 || !reflect.DeepEqual(first.Before.RawFields, acsSnapshot) {
				t.Fatal("ACS prepared receipt lost false/list/identity or invented a readback")
			}
			second, response, err := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.UpdateByIdentity(ctx, "Global", external.UpdateRequest{Fields: external.Fields{BlockExternalUserAccess: policy.Null[bool]()}, BlockedUsers: policy.ChangeStrings(nil, nil)})
			if err != nil || response == nil || response.StatusCode() != 204 || second == nil || !second.HTTPAccepted || !second.PatchInvoked || second.After != nil || second.Before == nil || second.Before.Identity == nil || *second.Before.Identity != "Global" || second.Before.BlockExternalUserAccess == nil || *second.Before.BlockExternalUserAccess || len(second.Before.BlockedUsers) != 0 || !reflect.DeepEqual(second.Before.RawFields, externalSnapshot) {
				t.Fatal("Teams-external preparation was confused with the equal-identity ACS configuration")
			}
			response, err = sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, tenant.UpdateRequest{Settings: tenant.Settings{AllowFederatedUsers: new(false)}, AllowedDomains: tenant.NewAllowList()})
			if response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) || !first.HTTPAccepted || !second.HTTPAccepted || !reflect.DeepEqual(first.Before.RawFields, acsSnapshot) || !reflect.DeepEqual(second.Before.RawFields, externalSnapshot) {
				t.Fatal("generated tenant-profile denial erased earlier compiled-profile evidence")
			}
			// This separate retained primitive reply only proves continued SDK
			// usability, not that accepted writes persisted or rolled back remotely.
			last, response, err := sdk.TeamsAdminAPI.ACSFederationConfigurations.GetByIdentity(ctx, "Global")
			if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, acsSnapshot)) {
				t.Fatal("one profile's denial poisoned the independent compiled read")
			}
			return testutil.Observation{Value: last, Response: response, Err: err}
		}})
	if err != nil || !r.Passed() {
		t.Fatalf("federation profile composition differs: %v %+v", err, r.Differences())
	}
}
