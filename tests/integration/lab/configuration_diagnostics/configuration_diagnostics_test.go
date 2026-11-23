package configurationdiagnosticsacc

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

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "configuration_diagnostics"
	// Inert probe inputs: a key no cache holds, for a region that does not
	// exist, and a policy type the service does not define.
	probeKey        = "msteams-sdk-probe-nonexistent-key"
	probeRegion     = "msteams-sdk-probe-invalid-region"
	probePolicyType = "MsTeamsSdkUnknownType"
	missingName     = "msteams-sdk-missing"
	fixtureIdentity = "msteams-sdk-fixture-configuration"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

func oracle(t *testing.T, ctx context.Context, script string, env []string, out any) {
	t.Helper()
	command := labkit.OracleCommand(ctx, script)
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
	command.Env = append(command.Env, env...)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell configuration oracle failed; no output retained")
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(out) != nil {
		t.Fatal("pinned PowerShell configuration projection unavailable")
	}
}

type configurationObservation struct {
	Success           bool   `json:"success"`
	ListCount         int    `json:"list_count"`
	Identity          string `json:"identity"`
	UserInitiatedMode string `json:"user_initiated_mode"`
	MissingCode       string `json:"missing_code"`
}

func powerShellConfiguration(t *testing.T, ctx context.Context, name string) configurationObservation {
	t.Helper()
	var observed configurationObservation
	oracle(t, ctx, "read-configuration.ps1", []string{"MSTEAMS_LAB_CONFIGURATION_TYPE=" + probeType, "MSTEAMS_LAB_CONFIGURATION_NAME=" + name}, &observed)
	if !observed.Success {
		t.Fatal("pinned PowerShell configuration read failed")
	}
	return observed
}

func setting(value *cd.Configuration, key string) string {
	var text string
	if value == nil || json.Unmarshal(value.RawFields[key], &text) != nil {
		return ""
	}
	return text
}

// TestAccConfigurationDiagnosticsReads lists one configuration type, reads its
// Global instance and probes a missing instance and an unknown type. It
// changes nothing.
func TestAccConfigurationDiagnosticsReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ConfigurationDiagnostics
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	listed, response, err := service.ListConfigurations(ctx, probeType, nil)
	if err != nil || len(listed) == 0 || listed[0] == nil || listed[0].Identity == nil {
		t.Fatal("configuration list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	global, response, err := service.GetConfiguration(ctx, probeType, "Global", nil)
	if err != nil || global.Identity == nil || *global.Identity != "Global" || global.ConfigID == nil {
		t.Fatal("Global configuration read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_global", response, replacements)
	t.Log("configuration list/get: passed")

	if _, response, err = service.GetConfiguration(ctx, probeType, missingName, nil); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing configuration was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, replacements)
	if _, response, err = service.ListConfigurations(ctx, "MsTeamsSdkUnknownType", nil); !labkit.IsHTTP(err, 400) {
		t.Fatal("unknown configuration type was not a 400")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list_unknown_type", response, replacements)
	t.Log("configuration missing/unknown-type rejection: passed")

	if labkit.OracleEnabled(t) {
		observed := powerShellConfiguration(t, ctx, "Global")
		if observed.ListCount != len(listed) || observed.Identity != "Global" || observed.UserInitiatedMode != setting(global, "UserInitiatedMode") || observed.MissingCode != "ClientError" {
			t.Fatal("PowerShell and Go configuration observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
}

// TestAccConfigurationDiagnosticsSupportCommandsDenied sends the cache clear
// and schema-defaults publish commands with inputs that cannot act: a
// nonexistent cache key for a nonexistent region, and an undefined policy
// type. App-only Teams Administrator credentials are refused both.
func TestAccConfigurationDiagnosticsSupportCommandsDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ConfigurationDiagnostics
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	region := probeRegion
	if _, response, err := service.ClearCache(ctx, cd.ClearCacheRequest{KeysToDelete: []string{probeKey}, Region: &region}); !labkit.IsHTTP(err, 403) {
		t.Fatal("inert cache clear was not refused")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "cache_clear", response, replacements)
	}
	t.Log("cache clear denial: passed")
	if _, response, err := service.PublishPolicySchemaDefaults(ctx, probePolicyType); !labkit.IsHTTP(err, 403) {
		t.Fatal("schema-defaults publish for an undefined type was not refused")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "schema_defaults", response, replacements)
	}
	t.Log("schema defaults publish denial: passed")

	if labkit.OracleEnabled(t) {
		var observed struct {
			Success   bool   `json:"success"`
			ErrorCode string `json:"error_code"`
		}
		oracle(t, ctx, "reject-cache-clear.ps1", []string{"MSTEAMS_LAB_CACHE_PROBE_KEY=" + probeKey, "MSTEAMS_LAB_CACHE_PROBE_REGION=" + probeRegion}, &observed)
		if !observed.Success || observed.ErrorCode != "Forbidden" {
			t.Fatal("PowerShell and Go cache clear denials differ")
		}
		t.Log("PowerShell 8.1.0 / Go cache clear denial parity: passed")
	}
}

// TestAccConfigurationLifecycle creates one uniquely named feedback policy
// instance through the generic route, reads it, replaces it, and removes it.
// It never touches a pre-existing instance and never grants the policy.
func TestAccConfigurationLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("configuration lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ConfigurationDiagnostics
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private configuration recovery journal unavailable")
	}
	cleanup := configurationCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"generic-configuration": cleanup}); err != nil {
			t.Error("configuration cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private configuration journal close failed")
		}
	})
	fixture, err := journal.FixtureName("config")
	if err != nil {
		t.Fatal("configuration fixture name unavailable")
	}
	entry, err := journal.PlanCreate("generic-configuration", fixture)
	if err != nil {
		t.Fatal("configuration create was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := tenantReplacements(t, binding)
	replacements[fixture] = fixtureIdentity

	// Disabled differs from the type default, so the read-back proves the body.
	response, err := service.CreateConfiguration(ctx, probeType, cd.ConfigurationRequest{Identity: fixture,
		Settings: map[string]any{"UserInitiatedMode": "Disabled"}}, nil)
	labkit.WriteResponseFixture(t, fixtureService, "create", response, replacements)
	if err != nil || response.StatusCode() != 201 {
		t.Fatal("live configuration create failed; recovery journal retained")
	}
	if err := journal.Applied(entry, fixture); err != nil {
		t.Fatal("created configuration was not journaled")
	}
	value, response, err := service.GetConfiguration(ctx, probeType, fixture, nil)
	if err != nil || value.Identity == nil || !labkit.SamePolicyName(*value.Identity, fixture) || setting(value, "UserInitiatedMode") != "Disabled" {
		t.Fatal("created configuration read-back differs")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_created", response, replacements)
	if labkit.OracleEnabled(t) {
		observed := powerShellConfiguration(t, ctx, fixture)
		if !labkit.SamePolicyName(observed.Identity, fixture) || observed.UserInitiatedMode != "Disabled" || observed.MissingCode != "ClientError" {
			t.Fatal("PowerShell and Go owned configuration observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go generic configuration parity: passed")
	}
	t.Log("configuration create/read-back: passed")

	response, err = service.SetConfiguration(ctx, probeType, fixture, cd.ConfigurationRequest{Identity: fixture,
		Settings: map[string]any{"UserInitiatedMode": "Enabled", "ReceiveSurveysMode": "Disabled"}}, nil)
	labkit.WriteResponseFixture(t, fixtureService, "set", response, replacements)
	if err != nil || response.StatusCode() != 204 {
		t.Fatal("live configuration replace failed")
	}
	value, response, err = service.GetConfiguration(ctx, probeType, fixture, nil)
	if err != nil || setting(value, "UserInitiatedMode") != "Enabled" || setting(value, "ReceiveSurveysMode") != "Disabled" {
		t.Fatal("replaced configuration read-back differs")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_replaced", response, replacements)
	t.Log("configuration set/read-back: passed")

	response, err = service.RemoveConfiguration(ctx, probeType, fixture)
	labkit.WriteResponseFixture(t, fixtureService, "remove", response, replacements)
	if err != nil || response.StatusCode() != 204 {
		t.Fatal("live configuration remove failed")
	}
	if err := cleanup.absent(ctx, fixture); err != nil {
		t.Fatal("removed configuration is still visible")
	}
	t.Log("configuration remove/absence: passed")
}
