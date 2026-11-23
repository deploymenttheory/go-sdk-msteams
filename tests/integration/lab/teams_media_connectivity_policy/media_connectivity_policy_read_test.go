package teamsmediaconnectivitypolicyacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMediaConnectivityPolicyReadOnly reads Global, the full list and a
// local Global filter, then compares the public getter with pinned PowerShell.
func TestAccMediaConnectivityPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMediaConnectivityPolicies
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	global, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("media connectivity policy Global read failed")
	}
	if global == nil || global.Identity == nil || !strings.EqualFold(*global.Identity, "Global") {
		t.Fatal("media connectivity policy Global identity differs")
	}
	values, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("media connectivity policy list request failed")
	}
	if len(values) == 0 || values[0] == nil || values[0].Identity == nil || !strings.EqualFold(*values[0].Identity, "Global") {
		t.Fatal("media connectivity policy list did not place Global first")
	}
	if !labkit.SameConfigurationFields(global.RawFields, values[0].RawFields) {
		t.Log(labkit.ResponseComparisonFailure(global.RawFields, values[0].RawFields))
		t.Fatal("media connectivity policy Global read/list differ")
	}
	filtered, _, err := service.ListFiltered(ctx, "Global")
	if err != nil || len(filtered) != 1 || filtered[0] == nil || !labkit.SameConfigurationFields(global.RawFields, filtered[0].RawFields) {
		t.Fatal("media connectivity policy Global filter selection differs")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("media connectivity tenant binding unavailable")
		}
		actual := mediaConnectivityPolicyOracleForIdentity(t, ctx, tenant, "Global")
		if !matchesMediaConnectivityPolicyGetterProjection(global, actual.Settings) {
			t.Fatal("Go and PowerShell media connectivity policy getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go Global read parity: passed")
	}
	t.Log("media connectivity policy Global read/list/filter verified")
}
