package lab_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
)

func policyObservation(value policy.Policy) labkit.TargetingPolicyObservation {
	return labkit.TargetingPolicyObservation{Identity: value.Identity, CustomTagsMode: value.CustomTagsMode, Description: value.Description, ManageTagsPermissionMode: value.ManageTagsPermissionMode, ShiftBackedTagsMode: value.ShiftBackedTagsMode, SuggestedPresetTags: value.SuggestedPresetTags, TeamOwnersEditWhoCanManageTagsMode: value.TeamOwnersEditWhoCanManageTagsMode}
}
func hierarchyObservation(value hierarchy.Status) labkit.HierarchyStatusObservation {
	return labkit.HierarchyStatusObservation{ID: value.ID, Status: value.Status, LastKnownStatus: value.LastKnownStatus, ErrorMessage: value.ErrorMessage, LastUpdatedAt: value.LastUpdatedAt, LastModifiedBy: value.LastModifiedBy, Filename: value.Filename}
}

// TestAccTargetingPolicyReadOnly reads Global and a locally filtered list. It
// never changes a policy, assignment, directory grant or hierarchy.
func TestAccTargetingPolicyReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, _, err := sdk.TeamsAdminAPI.TargetingPolicies.GetByIdentity(ctx, "Global")
	if err != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("selected modern targeting policy read failed")
	}
	listed, _, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "Global")
	if err != nil {
		t.Fatal("filtered modern targeting policy read failed")
	}
	found := false
	for _, item := range listed {
		found = found || strings.EqualFold(item.Identity, "Global")
	}
	if !found {
		t.Fatal("Global policy missing from filtered list")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.DirectoryID.MatchString(tenant) {
			t.Fatal("canonical tenant binding required")
		}
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "targeting-policy", ID: tenant}).TargetingPolicy
		if actual == nil || !reflect.DeepEqual(policyObservation(*value), *actual) {
			t.Fatal("PowerShell and Go selected targeting policy fields differ")
		}
	}
	t.Log("selected modern targeting reads observed; legacy flight eligibility and mutation effects remain separately qualified")
}

// TestAccTargetingHierarchyStatusReadOnly reads only the explicitly selected
// version/request. It does not create/import/delete a singleton or poll it.
func TestAccTargetingHierarchyStatusReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_HIERARCHY_STATUS_FILE"))
	if err != nil {
		t.Fatal("explicit private hierarchy-status fixture is required")
	}
	defer clear(raw)
	var fixture struct {
		TenantID  string            `json:"tenant_id"`
		Version   hierarchy.Version `json:"api_version"`
		RequestID string            `json:"request_id"`
	}
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || fixture.Version != hierarchy.V1 && fixture.Version != hierarchy.V2 {
		t.Fatal("invalid hierarchy-status fixture binding/version")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, _, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, fixture.RequestID, fixture.Version)
	if err != nil || value == nil || value.Status == nil {
		t.Fatal("selected hierarchy status read failed")
	}
	if oracle {
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "hierarchy-status", ID: tenant, APIVersion: string(fixture.Version), RequestID: fixture.RequestID}).HierarchyStatus
		if actual == nil || !reflect.DeepEqual(hierarchyObservation(*value), *actual) {
			t.Fatal("PowerShell and Go selected hierarchy status fields differ")
		}
	}
	t.Log("selected hierarchy status observed; import, delete, role enforcement and cloud deployment remain separately qualified")
}
