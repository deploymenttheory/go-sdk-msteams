package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
)

type targetingPolicyObservation struct {
	Identity                           string  `json:"identity"`
	CustomTagsMode                     *string `json:"custom_tags_mode"`
	Description                        *string `json:"description"`
	ManageTagsPermissionMode           *string `json:"manage_tags_permission_mode"`
	ShiftBackedTagsMode                *string `json:"shift_backed_tags_mode"`
	SuggestedPresetTags                *string `json:"suggested_preset_tags"`
	TeamOwnersEditWhoCanManageTagsMode *string `json:"team_owners_edit_who_can_manage_tags_mode"`
}
type hierarchyStatusObservation struct {
	ID              string  `json:"id"`
	Status          *string `json:"status"`
	LastKnownStatus *string `json:"last_known_status"`
	ErrorMessage    *string `json:"error_message"`
	LastUpdatedAt   *string `json:"last_updated_at"`
	LastModifiedBy  *string `json:"last_modified_by"`
	Filename        *string `json:"filename"`
}

func policyObservation(value policy.Policy) targetingPolicyObservation {
	return targetingPolicyObservation{Identity: value.Identity, CustomTagsMode: value.CustomTagsMode, Description: value.Description, ManageTagsPermissionMode: value.ManageTagsPermissionMode, ShiftBackedTagsMode: value.ShiftBackedTagsMode, SuggestedPresetTags: value.SuggestedPresetTags, TeamOwnersEditWhoCanManageTagsMode: value.TeamOwnersEditWhoCanManageTagsMode}
}
func hierarchyObservation(value hierarchy.Status) hierarchyStatusObservation {
	return hierarchyStatusObservation{ID: value.ID, Status: value.Status, LastKnownStatus: value.LastKnownStatus, ErrorMessage: value.ErrorMessage, LastUpdatedAt: value.LastUpdatedAt, LastModifiedBy: value.LastModifiedBy, Filename: value.Filename}
}
func targetingReadOnly(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("targeting reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid targeting lab enablement")
	}
	return cfg
}

// TestAccTargetingPolicyReadOnly reads Global and a locally filtered list. It
// never changes a policy, assignment, directory grant or hierarchy.
func TestAccTargetingPolicyReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
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
		if !ok || !labDirectoryID.MatchString(tenant) {
			t.Fatal("canonical tenant binding required")
		}
		actual := runTeamOracle(t, ctx, teamOracleRequest{Kind: "targeting-policy", ID: tenant}).TargetingPolicy
		if actual == nil || !reflect.DeepEqual(policyObservation(*value), *actual) {
			t.Fatal("PowerShell and Go selected targeting policy fields differ")
		}
	}
	t.Log("selected modern targeting reads observed; legacy flight eligibility and mutation effects remain separately qualified")
}

// TestAccTargetingHierarchyStatusReadOnly reads only the explicitly selected
// version/request. It does not create/import/delete a singleton or poll it.
func TestAccTargetingHierarchyStatusReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	raw, err := privateLabFile(os.Getenv("MSTEAMS_LAB_HIERARCHY_STATUS_FILE"))
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
		actual := runTeamOracle(t, ctx, teamOracleRequest{Kind: "hierarchy-status", ID: tenant, APIVersion: string(fixture.Version), RequestID: fixture.RequestID}).HierarchyStatus
		if actual == nil || !reflect.DeepEqual(hierarchyObservation(*value), *actual) {
			t.Fatal("PowerShell and Go selected hierarchy status fields differ")
		}
	}
	t.Log("selected hierarchy status observed; import, delete, role enforcement and cloud deployment remain separately qualified")
}
