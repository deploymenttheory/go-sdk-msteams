package teamsmultitenantorganizationconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_multi_tenant_organization_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsMultiTenantOrganizationConfiguration"
const selected = collection + "/configuration/Global"
const set = "Set-CsTeamsMultiTenantOrganizationConfiguration"

func TestUnitUpdateUsesActualPreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	run(t, contracttest.Case{Name: "update", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CopilotFromHomeTenant":"Disabled"}`), 204, string(readResponseFixture(t, "update_204.body"))),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{CopilotFromHomeTenant: policy.Set("Disabled")}})
		if e == nil && (v == nil || v.Before == nil || v.Before.CopilotFromHomeTenant == nil || *v.Before.CopilotFromHomeTenant != "Enabled" || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted) {
			t.Fatal("preparation or mutation acceptance receipt lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}

func TestUnitNullEmptyAndForceRemainExplicit(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		field      policy.Field[string]
	}{{"null", `{"Identity":"Global","CopilotFromHomeTenant":null,"Force":"False"}`, policy.Null[string]()}, {"empty", `{"Identity":"Global","CopilotFromHomeTenant":"","Force":"False"}`, policy.Set("")}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{CopilotFromHomeTenant: tc.field}, Controls: policy.Controls{Force: new(false)}})
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitFilterUsesActualCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"global", "Glo?al", 1}, {"missing", "missing", 0}, {"literal-whitespace", " ", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMultiTenantOrganizationConfiguration", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, "Get-CsTeamsMultiTenantOrganizationConfiguration", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(values) != tc.count {
					t.Fatal("local filter result differs")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
