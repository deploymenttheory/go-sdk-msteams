package batchdeploymentacc

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccBatchDeploymentStatusReadOnly reads one explicitly selected existing
// orchestration. It never submits a deployment or sends notification email.
func TestAccBatchDeploymentStatusReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_BATCH_STATUS_FILE"))
	if err != nil {
		t.Fatal("explicit private batch-status fixture required")
	}
	defer clear(raw)
	var fixture struct {
		TenantID        string `json:"tenant_id"`
		OrchestrationID string `json:"orchestration_id"`
	}
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || strings.TrimSpace(fixture.OrchestrationID) == "" {
		t.Fatal("invalid batch-status fixture binding")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, _, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, fixture.OrchestrationID)
	if err != nil || value == nil {
		t.Fatal("selected batch status read failed")
	}
	if oracle {
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "batch-status", ID: tenant, OrchestrationID: fixture.OrchestrationID}).BatchStatus
		if actual == nil || *actual != value.Text {
			t.Fatal("PowerShell and Go batch status text differs")
		}
	}
	t.Log("selected raw batch status read observed; text does not establish deployment success")
}
