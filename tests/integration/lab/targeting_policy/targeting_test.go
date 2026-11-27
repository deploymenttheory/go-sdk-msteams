package targetingpolicyacc

import (
	"context"
	"reflect"
	"strings"
	"testing"

	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func policyObservation(value policy.Policy) labkit.TargetingPolicyObservation {
	return labkit.TargetingPolicyObservation{Identity: value.Identity, CustomTagsMode: value.CustomTagsMode, Description: value.Description, ManageTagsPermissionMode: value.ManageTagsPermissionMode, ShiftBackedTagsMode: value.ShiftBackedTagsMode, SuggestedPresetTags: value.SuggestedPresetTags, TeamOwnersEditWhoCanManageTagsMode: value.TeamOwnersEditWhoCanManageTagsMode}
}

// TestAccTargetingPolicyReadOnly reads Global and a locally filtered list. It
// never changes a policy, assignment, directory grant or hierarchy.
func TestAccTargetingPolicyReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TargetingPolicies.GetByIdentity(ctx, "Global")
	if err != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("selected modern targeting policy read failed (HTTP %d)", status)
	}
	labkit.WriteResponseFixture(t, "targeting_policy", "get_global", response, nil)
	listed, response, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "Global")
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("filtered modern targeting policy read failed (HTTP %d)", status)
	}
	found := 0
	for _, item := range listed {
		if strings.EqualFold(item.Identity, "Global") {
			found++
			if !reflect.DeepEqual(policyObservation(*value), policyObservation(item)) {
				t.Fatal("selected and listed targeting Global policy fields differ")
			}
		}
	}
	if found != 1 {
		t.Fatal("Global policy missing from filtered list")
	}
	labkit.WriteResponseFixture(t, "targeting_policy", "list_global", response, nil)
	t.Log("targeting Global read/list: passed")
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.DirectoryID.MatchString(tenant) {
			t.Fatal("canonical tenant binding required")
		}
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "targeting-policy", ID: tenant}).TargetingPolicy
		if actual == nil || !reflect.DeepEqual(policyObservation(*value), *actual) {
			t.Fatal("PowerShell and Go selected targeting policy fields differ")
		}
		t.Log("PowerShell 8.1.0 / Go targeting policy read parity: passed")
	}
	t.Log("selected modern targeting reads observed; legacy flight eligibility and mutation effects remain separately qualified")
}
