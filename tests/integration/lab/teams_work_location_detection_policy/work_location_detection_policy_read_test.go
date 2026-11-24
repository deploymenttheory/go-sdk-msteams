package teamsworklocationdetectionpolicyacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccWorkLocationDetectionPolicyReadOnly reads Global, the full list and a
// local Global filter, then compares the public getter with pinned PowerShell.
func TestAccWorkLocationDetectionPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsWorkLocationDetectionPolicies
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	global, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("work location detection policy Global read failed")
	}
	if global == nil || global.Identity == nil || !strings.EqualFold(*global.Identity, "Global") {
		t.Fatal("work location detection policy Global identity differs")
	}
	values, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("work location detection policy list request failed")
	}
	if len(values) == 0 || values[0] == nil || values[0].Identity == nil || !strings.EqualFold(*values[0].Identity, "Global") {
		t.Fatal("work location detection policy list did not place Global first")
	}
	if !labkit.SameConfigurationFields(global.RawFields, values[0].RawFields) {
		t.Log(labkit.ResponseComparisonFailure(global.RawFields, values[0].RawFields))
		t.Fatal("work location detection policy Global read/list differ")
	}
	filtered, _, err := service.ListFiltered(ctx, "Global")
	if err != nil || len(filtered) != 1 || filtered[0] == nil || !labkit.SameConfigurationFields(global.RawFields, filtered[0].RawFields) {
		t.Fatal("work location detection policy Global filter selection differs")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("work location detection tenant binding unavailable")
		}
		actual := workLocationDetectionPolicyOracleForIdentity(t, ctx, tenant, "Global")
		if !matchesWorkLocationDetectionPolicyGetterProjection(global, actual.Settings) {
			t.Fatal("Go and PowerShell work location detection policy getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go Global read parity: passed")
	}
	t.Log("work location detection policy Global read/list/filter verified")
}
