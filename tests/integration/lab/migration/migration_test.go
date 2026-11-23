package migrationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/migration"
	policyassignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

const (
	fixtureService = "migration"
	// probeIdentity is a reserved fixture-form ID that cannot name a real
	// directory object: a meeting migration started for it has no user and no
	// meetings to act on.
	probeIdentity = "00000000-0000-4000-8000-000000000001"
	// probePartition names partitions of the same reserved, nonexistent tenant
	// in a nonexistent container, so a partition move has no documents to move
	// and touches nothing the lab tenant owns.
	probePartition = probeIdentity + "_msteams-sdk-probe"
	probeName      = "msteams-sdk-probe"
)

func tenantOf(t *testing.T, binding string) (string, map[string]string) {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return tenant, map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

// outcome classifies a read as the oracle does.
func outcome(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, auth.ErrUnauthenticated):
		return "unauthorized"
	case errors.Is(err, auth.ErrForbidden):
		return "forbidden"
	default:
		return "unknown"
	}
}

// denied accepts a successful read or the credential/authorization denial the
// app-only lab identity receives; anything else is a contract change.
func denied(t *testing.T, name string, response *resty.Response, err error) string {
	t.Helper()
	labkit.LogAccessResponse(t, response, err)
	class := outcome(err)
	if class == "unknown" {
		t.Fatalf("%s outcome changed; no payload retained", name)
	}
	return class
}

type oracleObservation struct {
	Success     bool   `json:"success"`
	Status      string `json:"status"`
	StatusCount int    `json:"status_count"`
	Summary     string `json:"summary"`
}

func powerShellMeetingMigration(t *testing.T, ctx context.Context) oracleObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-meeting-migration.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		command.Env = append(command.Env, key+"="+os.Getenv(key))
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell meeting migration read failed; no output retained")
	}
	var observed oracleObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell meeting migration projection unavailable")
	}
	return observed
}

// TestAccMigrationReads reads meeting migration status, summary and history,
// tenant move task status and the OCE provisioning failover status. The
// app-only lab identity is denied most of them; denials are recorded as
// fixtures and successful reads are decoded. Nothing is migrated or moved.
func TestAccMigrationReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	tenant, replacements := tenantOf(t, binding)
	service := sdk.TeamsAdminAPI.Migration
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	statuses, response, err := service.ListMeetingMigrationStatuses(ctx, nil)
	status := denied(t, "meeting migration status", response, err)
	if status == "ok" && statuses == nil {
		t.Fatal("successful status read returned no slice")
	}
	labkit.WriteResponseFixture(t, fixtureService, "status", response, replacements)
	summary, response, err := service.GetMeetingMigrationStatusSummary(ctx, nil)
	summaryOutcome := denied(t, "meeting migration summary", response, err)
	if summaryOutcome == "ok" && summary == nil {
		t.Fatal("successful summary read returned no value")
	}
	labkit.WriteResponseFixture(t, fixtureService, "summary", response, replacements)
	t.Logf("meeting migration status %s, summary %s", status, summaryOutcome)

	_, response, err = service.ListMeetingMigrationTransactionHistory(ctx, probeIdentity, nil)
	t.Logf("meeting migration history for a nonexistent user: %s", denied(t, "meeting migration history", response, err))
	labkit.WriteResponseFixture(t, fixtureService, "history", response, replacements)

	_, response, err = service.ListTenantMoveTaskStatuses(ctx)
	t.Logf("tenant move task status: %s", denied(t, "tenant move task status", response, err))
	labkit.WriteResponseFixture(t, fixtureService, "task_status", response, replacements)

	_, response, err = sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, policyassignments.OCEContext{TenantID: tenant}, policyassignments.RegionNOAM)
	t.Logf("OCE NgtProv failover status: %s", denied(t, "OCE NgtProv failover status", response, err))
	labkit.WriteResponseFixture(t, fixtureService, "ngtprov_failover_status", response, replacements)

	if labkit.OracleEnabled(t) {
		observed := powerShellMeetingMigration(t, ctx)
		if observed.Status != status || observed.Summary != summaryOutcome || (status == "ok" && observed.StatusCount != len(statuses)) {
			t.Fatal("PowerShell and Go meeting migration outcomes differ")
		}
		t.Log("PowerShell 8.1.0 / Go meeting migration status parity: passed")
	}
}

// TestAccMigrationRejectedWrites proves the meeting migration and AVS
// partition move request contracts with requests that cannot act: the
// meeting migration names a reserved ID no user has, and the partition move
// names partitions of a reserved, nonexistent tenant in a nonexistent
// container. Tenant moves (cross-region, service instance) and the OCE
// NgtProv failover take no target the service could reject first, so they are
// never sent.
func TestAccMigrationRejectedWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	_, replacements := tenantOf(t, binding)
	service := sdk.TeamsAdminAPI.Migration
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	value, response, err := service.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: probeIdentity})
	if value != nil || err == nil || response == nil || response.StatusCode() < 400 || response.StatusCode() >= 500 {
		t.Fatal("meeting migration for a nonexistent user was not rejected")
	}
	labkit.LogAccessResponse(t, response, err)
	labkit.WriteResponseFixture(t, fixtureService, "start_meeting_migration", response, replacements)

	workload, container, base, zero := probeName, probeName, probeIdentity, int64(0)
	results, response, err := service.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{
		SourcePartitionKey: probePartition + "-source", TargetPartitionKey: probePartition + "-target",
		Workload: &workload, ContainerName: &container, BasePartitionKey: &base, NumberOfDocuments: &zero})
	if results != nil || err == nil || response == nil || response.StatusCode() < 400 || response.StatusCode() >= 500 {
		t.Fatal("partition move for a nonexistent tenant was not rejected")
	}
	labkit.LogAccessResponse(t, response, err)
	labkit.WriteResponseFixture(t, fixtureService, "move_avs_partition", response, replacements)
	t.Log("meeting migration and AVS partition move rejection: passed")
}
