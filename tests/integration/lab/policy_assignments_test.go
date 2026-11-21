package lab_test

import (
	"context"
	"strings"
	"testing"
)

// TestLabPolicyAssignmentsReadOnly never submits batches, edits group rank,
// grants policies, refreshes users or invokes OCE reprocessing operations.
// It lists ordinary generated jobs and reads at most five returned job IDs.
func TestLabPolicyAssignmentsReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	sdk, _, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	jobs, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListBatches(ctx, "")
	if err != nil {
		status, discovery := 0, false
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil {
				discovery = strings.Contains(response.Request.URL, "/Teams.Tenant/serviceDiscovery")
			}
		}
		t.Fatalf("batch assignment list failed (HTTP %d, discovery=%t); no payload retained", status, discovery)
	}
	checked := 0
	for _, job := range jobs {
		if checked == 5 {
			break
		}
		if job == nil || job.OperationID == nil || *job.OperationID == "" {
			t.Fatal("assignment list returned no stable operation identity")
		}
		value, _, err := sdk.TeamsAdminAPI.PolicyAssignments.GetBatch(ctx, *job.OperationID)
		if err != nil || value == nil || value.OperationID == nil || *value.OperationID != *job.OperationID {
			t.Fatal("selected batch assignment read did not preserve its identity")
		}
		checked++
	}
	t.Logf("ordinary assignment list and %d selected reads observed; writes, OCE and effective assignment convergence remain separately qualified", checked)
}
