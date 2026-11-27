package extintegration

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	feedback "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
	diagnosticsacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type configurationObservation struct {
	Success         bool   `json:"success"`
	GenericIdentity string `json:"generic_identity"`
	TypedIdentity   string `json:"typed_identity"`
	GenericMode     string `json:"generic_mode"`
	TypedMode       string `json:"typed_mode"`
	GenericLog      bool   `json:"generic_log"`
	TypedLog        bool   `json:"typed_log"`
}

func oracle(t *testing.T, ctx context.Context, script string, extraEnv []string, observed any) {
	t.Helper()
	command := labkit.OracleCommand(ctx, script)
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath", "MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	command.Env = append(command.Env, extraEnv...)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned EXT composition read failed; no output retained")
	}
	if json.Unmarshal(output.Data, observed) != nil {
		t.Fatal("pinned EXT composition projection unavailable")
	}
}

func compareConfiguration(t *testing.T, ctx context.Context, name, mode string, allowLog bool) {
	t.Helper()
	var observed configurationObservation
	oracle(t, ctx, "read-ext-integration.ps1", []string{"MSTEAMS_LAB_CONFIGURATION_NAME=" + name}, &observed)
	if !observed.Success || !labkit.SamePolicyName(observed.GenericIdentity, name) || !labkit.SamePolicyName(observed.TypedIdentity, name) || observed.GenericMode != mode || observed.TypedMode != mode || observed.GenericLog != allowLog || observed.TypedLog != allowLog {
		t.Fatal("PowerShell and Go generic/typed configuration projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go integrated EXT parity: passed")
}

// TestAccGenericTypedConfigurationComposition creates only an owned, unassigned
// feedback policy. Generic and typed APIs must observe each other's writes.
// No tenant Global setting, assignment or licensing state is changed.
func TestAccGenericTypedConfigurationComposition(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("EXT integration requires explicit mutation opt-in and a private journal")
	} else if err != nil {
		t.Fatal(err)
	}
	if !labkit.OracleEnabled(t) {
		t.Fatal("EXT integration requires the pinned PowerShell oracle")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private EXT integration journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		results, err := journal.Recover(ctx, diagnosticsacc.RecoveryHandlers(sdk))
		if err != nil {
			t.Error("EXT integration cleanup incomplete; retained journal requires recovery")
		}
		for _, result := range results {
			if result.Outcome != "cleaned" {
				t.Error("EXT integration resource cleanup unresolved")
			}
		}
		if err := journal.Close(); err != nil {
			t.Error("private EXT integration journal close failed")
		}
		if !t.Failed() {
			t.Log("integrated EXT independent cleanup: passed")
		}
	})
	name, err := journal.FixtureName("config")
	if err != nil {
		t.Fatal("owned configuration name unavailable")
	}
	entry, err := journal.PlanCreate("generic-configuration", name)
	if err != nil {
		t.Fatal("generic configuration create was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	generic, typed := sdk.TeamsAdminAPI.ConfigurationDiagnostics, sdk.TeamsAdminAPI.TeamsFeedbackPolicies
	response, err := generic.CreateConfiguration(ctx, configType, cd.ConfigurationRequest{Identity: name, Settings: map[string]any{"UserInitiatedMode": "Disabled", "ReceiveSurveysMode": "Disabled", "AllowLogCollection": true, "EnableFeatureSuggestions": false}}, nil)
	if err != nil || response == nil || response.StatusCode() != 201 {
		t.Fatal("owned generic configuration create failed; journal retained")
	}
	if err := journal.Applied(entry, name); err != nil {
		t.Fatal("owned configuration identity was not journaled")
	}
	replacements := map[string]string{name: "msteams-sdk-fixture-configuration"}
	labkit.WriteResponseFixture(t, "configuration_diagnostics", "integration_create", response, replacements)
	value, response, err := typed.GetByIdentity(ctx, "Tag:"+name)
	if err != nil || value == nil || value.Identity == nil || !labkit.SamePolicyName(*value.Identity, name) || value.UserInitiatedMode == nil || *value.UserInitiatedMode != "Disabled" || value.AllowLogCollection == nil || !*value.AllowLogCollection || value.EnableFeatureSuggestions == nil || *value.EnableFeatureSuggestions {
		t.Fatal("typed service did not observe the generic create")
	}
	labkit.WriteResponseFixture(t, "teams_feedback_policy", "integration_typed_get", response, replacements)
	compareConfiguration(t, ctx, name, "Disabled", true)
	t.Log("integrated EXT generic-to-typed read: passed")
	if _, _, err := typed.PatchByIdentity(ctx, *value.Identity, feedback.PatchRequest{Settings: feedback.Settings{AllowLogCollection: policy.Set(false)}}); err != nil {
		t.Fatal("owned typed sparse update failed")
	}
	raw, response, err := generic.GetConfiguration(ctx, configType, *value.Identity, nil)
	if err != nil || raw == nil || string(raw.RawFields["AllowLogCollection"]) != "false" || string(raw.RawFields["UserInitiatedMode"]) != `"Disabled"` || string(raw.RawFields["EnableFeatureSuggestions"]) != "false" {
		t.Fatal("generic service did not observe the typed false update with omitted settings preserved")
	}
	labkit.WriteResponseFixture(t, "configuration_diagnostics", "integration_generic_get", response, replacements)
	compareConfiguration(t, ctx, name, "Disabled", false)
	t.Log("integrated EXT typed-to-generic update: passed")
	if _, err := generic.SetConfiguration(ctx, configType, *value.Identity, cd.ConfigurationRequest{Identity: *value.Identity, Settings: map[string]any{"UserInitiatedMode": "Enabled", "ReceiveSurveysMode": "Disabled", "AllowLogCollection": false, "EnableFeatureSuggestions": false}}, nil); err != nil {
		t.Fatal("owned generic replacement failed")
	}
	changed, response, err := typed.GetByIdentity(ctx, *value.Identity)
	if err != nil || changed == nil || changed.UserInitiatedMode == nil || *changed.UserInitiatedMode != "Enabled" || changed.AllowLogCollection == nil || *changed.AllowLogCollection {
		t.Fatal("typed service did not observe the generic replacement")
	}
	labkit.WriteResponseFixture(t, "teams_feedback_policy", "integration_replaced_get", response, replacements)
	compareConfiguration(t, ctx, name, "Enabled", false)
	t.Log("integrated EXT generic replacement/read-back: passed")
}
