package tenantmigrationconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tenantmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_migration_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func tenantMigrationProjection(v *tenantmigration.Configuration) map[string]any {
	return map[string]any{"MeetingMigrationEnabled": v.MeetingMigrationEnabled}
}

// tenantMigrationComplete reports whether the public getter setting was read.
func tenantMigrationComplete(v *tenantmigration.Configuration) bool {
	return v != nil && v.MeetingMigrationEnabled != nil
}

// tenantMigrationOracle runs the pinned 8.1.0 Global getter in a dedicated
// read-only PowerShell process and returns its public getter setting.
func tenantMigrationOracle(t *testing.T, ctx context.Context, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": "tenant-migration"})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-tenant-migration-configuration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned tenant migration configuration read process failed")
	}
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		t.Fatal("pinned tenant migration configuration projection unavailable")
	}
	if len(result.Settings) != len(tenantMigrationProjection(&tenantmigration.Configuration{})) {
		t.Fatal("incomplete tenant migration configuration getter projection")
	}
	for name := range tenantMigrationProjection(&tenantmigration.Configuration{}) {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("incomplete tenant migration configuration getter projection")
		}
	}
	return result.Settings
}
