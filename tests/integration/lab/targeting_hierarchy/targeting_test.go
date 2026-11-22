package targetinghierarchyacc

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func hierarchyObservation(value hierarchy.Status) labkit.HierarchyStatusObservation {
	return labkit.HierarchyStatusObservation{ID: value.ID, Status: value.Status, LastKnownStatus: value.LastKnownStatus, ErrorMessage: value.ErrorMessage, LastUpdatedAt: value.LastUpdatedAt, LastModifiedBy: value.LastModifiedBy, Filename: value.Filename}
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
