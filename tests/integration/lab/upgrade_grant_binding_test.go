package lab_test

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

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_policy"
)

func TestAccUpgradeGrantMigrationBinding(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	if !labkit.OracleEnabled(t) {
		t.Fatal("pinned upgrade binding comparison requires its oracle")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("invalid oracle tenant binding")
	}
	input, e := json.Marshal(map[string]string{"tenant_id": tenant})
	if e != nil {
		t.Fatal("invalid oracle input")
	}
	defer clear(input)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	cmd := labkit.OracleCommand(ctx, "read-upgrade-grant-binding.ps1")
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
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned upgrade binding oracle failed")
	}
	var observed struct {
		Success bool `json:"success"`
		Results []struct {
			Bind          string `json:"bind"`
			Success       bool   `json:"success"`
			ErrorID       string `json:"error_id"`
			ExceptionType string `json:"exception_type"`
		} `json:"results"`
	}
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success || len(observed.Results) != 2 {
		t.Fatal("invalid pinned binding observation")
	}
	for index, flag := range []bool{false, true} {
		want := "false"
		if flag {
			want = "true"
		}
		item := observed.Results[index]
		if item.Bind != want || item.Success || !strings.HasPrefix(item.ErrorID, "NamedParameterNotFound,") || item.ExceptionType != "System.Management.Automation.ParameterBindingException" {
			t.Fatal("pinned upgrade migration binding outcome changed")
		}
		for _, scope := range []resource.Assignment{{User: &resource.UserAssignment{Identity: "msteams-sdk-binding-only@invalid.example"}}, {Tenant: &resource.TenantAssignment{}}, {Group: &resource.GroupGrant{ID: "msteams-sdk-binding-only"}}} {
			scope.MigrateMeetingsToTeams = &flag
			_, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.Assign(ctx, scope)
			if !errors.Is(e, resource.ErrMigrationParameterBinding) || r != nil {
				t.Fatal("Go did not preserve the pinned pre-HTTP binding failure")
			}
		}
	}
	t.Log("PowerShell 8.1.0 / Go upgrade migration binding parity: passed")
}
