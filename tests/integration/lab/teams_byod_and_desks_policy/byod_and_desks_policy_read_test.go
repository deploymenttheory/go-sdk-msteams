package teamsbyodanddeskspolicyacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccBYODAndDesksPolicyReadOnly reads Global, the full list and a
// local Global filter, then compares the public getter with pinned PowerShell.
func TestAccBYODAndDesksPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsBYODAndDesksPolicies
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	global, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("BYOD and desks policy Global read failed")
	}
	if global == nil || global.Identity == nil || !strings.EqualFold(*global.Identity, "Global") {
		t.Fatal("BYOD and desks policy Global identity differs")
	}
	values, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("BYOD and desks policy list request failed")
	}
	if len(values) == 0 || values[0] == nil || values[0].Identity == nil || !strings.EqualFold(*values[0].Identity, "Global") {
		t.Fatal("BYOD and desks policy list did not place Global first")
	}
	if !labkit.SameConfigurationFields(global.RawFields, values[0].RawFields) {
		t.Log(labkit.ResponseComparisonFailure(global.RawFields, values[0].RawFields))
		t.Fatal("BYOD and desks policy Global read/list differ")
	}
	filtered, _, err := service.ListFiltered(ctx, "Global")
	if err != nil || len(filtered) != 1 || filtered[0] == nil || !labkit.SameConfigurationFields(global.RawFields, filtered[0].RawFields) {
		t.Fatal("BYOD and desks policy Global filter selection differs")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("BYOD and desks tenant binding unavailable")
		}
		actual := byodAndDesksPolicyOracleForIdentity(t, ctx, tenant, "Global")
		if !matchesBYODAndDesksPolicyGetterProjection(global, actual.Settings) {
			t.Fatal("Go and PowerShell BYOD and desks policy getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go Global read parity: passed")
	}
	t.Log("BYOD and desks policy Global read/list/filter verified")
}
