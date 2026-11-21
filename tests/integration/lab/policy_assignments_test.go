package lab_test

import (
	"context"
	"strings"
	"testing"

	policyassignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// TestAccPolicyAssignmentsReadOnly never submits batches, edits group rank,
// grants policies, refreshes users or invokes OCE reprocessing operations.
// It lists ordinary generated jobs and reads at most five returned job IDs.
func TestAccPolicyAssignmentsReadOnly(t *testing.T) {
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

// TestAccPolicyAssignmentsOCEReadOnly uses the authenticating lab tenant as the
// explicit OCE target. It only lists regional status; no privileged PowerShell
// profile, contributor elevation, refresh or reprocessing action runs.
func TestAccPolicyAssignmentsOCEReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	sdk, binding, _ := labSDKClient(t)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labDirectoryID.MatchString(tenant) {
		t.Fatal("canonical OCE lab tenant binding required")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, policyassignments.OCEContext{TenantID: tenant}, nil)
	if err != nil {
		status, attempts := 0, 0
		if response != nil {
			status = response.StatusCode()
		}
		if value != nil {
			attempts = len(value.Attempts)
		}
		t.Fatalf("OCE regional status list unavailable (last HTTP %d, attempts %d); no payload retained", status, attempts)
	}
	if value == nil {
		t.Fatal("OCE regional status list returned no observation")
	}
	t.Logf("OCE status list observed across %d regional attempts; no actions or eligibility changes ran", len(value.Attempts))
}
