package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type accessOutcome struct {
	Success   bool    `json:"success"`
	Tenant    string  `json:"tenant_id"`
	Command   string  `json:"command"`
	Outcome   string  `json:"outcome"`
	Code      *string `json:"code"`
	ErrorCode *int    `json:"error_code"`
}

func decodeAccessOutcome(raw []byte, tenant, command, outcome string) (accessOutcome, error) {
	var result accessOutcome
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || d.Decode(&result) != nil || !result.Success || !strings.EqualFold(result.Tenant, tenant) || result.Command != command || result.Outcome != outcome {
		return accessOutcome{}, errTeamOracle
	}
	if d.Decode(new(any)) != io.EOF {
		return accessOutcome{}, errTeamOracle
	}
	switch outcome {
	case "migration-status-message":
		if result.Code != nil || result.ErrorCode != nil {
			return accessOutcome{}, errTeamOracle
		}
	case "feature-unavailable":
		if result.Code == nil || *result.Code != "ClientError" || result.ErrorCode == nil || *result.ErrorCode != 40003 {
			return accessOutcome{}, errTeamOracle
		}
	case "forbidden":
		if result.Code == nil || *result.Code != "Forbidden" || result.ErrorCode == nil || *result.ErrorCode != 40301 {
			return accessOutcome{}, errTeamOracle
		}
	default:
		return accessOutcome{}, errTeamOracle
	}
	return result, nil
}

func accessOutcomeOracle(t *testing.T, ctx context.Context, tenant, resource, command, outcome string) {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "resource": resource})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-access-outcome.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned access outcome oracle failed")
	}
	if _, err := decodeAccessOutcome(output.data, tenant, command, outcome); err != nil {
		t.Fatal("PowerShell access outcome changed or was not verified")
	}
}

func logAccessResponse(t *testing.T, response *resty.Response, err error) {
	t.Helper()
	status := 0
	if response != nil {
		status = response.StatusCode()
	}
	var api *client.APIError
	var codes []int
	if errors.As(err, &api) {
		codes = api.ServiceErrorCodes
	}
	t.Logf("SDK access outcome: HTTP %d; PolicyRP codes %v; error %T", status, codes, err)
}

func TestAccAssignmentMigrationCapabilityUnavailable(t *testing.T) {
	cfg := targetingReadOnly(t)
	if !teamOracleEnabled(t) {
		t.Fatal("PowerShell comparison is required for tenant access outcomes")
	}
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.AssignmentMigrations.List(ctx)
	logAccessResponse(t, response, err)
	var api *client.APIError
	if value != nil || response == nil || response.StatusCode() != 400 || !errors.Is(err, auth.ErrTenantCapabilityUnavailable) || !errors.As(err, &api) || !slices.Equal(api.ServiceErrorCodes, []int{40003}) {
		t.Fatal("expected typed tenant capability denial changed; successful migration needs separate qualification")
	}
	writeResponseFixture(t, "assignment_migration", "list", response, nil)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	accessOutcomeOracle(t, ctx, tenant, "migration", "Get-DirectToGroupAssignmentsMigrationStatus", "migration-status-message")
	t.Log("confirmed capability denial: HTTP 400, PolicyRP 40003; PowerShell status message matched")
}

func TestAccAudioConferencingPromptsCapabilityUnavailable(t *testing.T) {
	cfg := targetingReadOnly(t)
	if !teamOracleEnabled(t) {
		t.Fatal("PowerShell comparison is required for tenant access outcomes")
	}
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.GetByIdentity(ctx, "Global")
	logAccessResponse(t, response, err)
	var api *client.APIError
	if value != nil || response == nil || response.StatusCode() != 400 || !errors.Is(err, auth.ErrTenantCapabilityUnavailable) || !errors.As(err, &api) || !slices.Equal(api.ServiceErrorCodes, []int{40003}) {
		t.Fatal("expected typed tenant capability denial changed; successful prompt operations need separate qualification")
	}
	writeResponseFixture(t, "audio_conferencing_prompts", "get_global", response, nil)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	accessOutcomeOracle(t, ctx, tenant, "prompts", "Get-CsTeamsAudioConferencingCustomPromptsConfiguration", "feature-unavailable")
	t.Log("confirmed capability denial: HTTP 400, PolicyRP 40003; PowerShell feature denial matched")
}

func TestAccPrivacyConfigurationAuthorizationDenied(t *testing.T) {
	cfg := targetingReadOnly(t)
	if !teamOracleEnabled(t) {
		t.Fatal("PowerShell comparison is required for tenant access outcomes")
	}
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
	logAccessResponse(t, response, err)
	var api *client.APIError
	if value != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) || errors.Is(err, auth.ErrTenantCapabilityUnavailable) || !errors.As(err, &api) || !slices.Equal(api.ServiceErrorCodes, []int{40301}) {
		t.Fatal("expected privacy authorization denial changed; successful privacy reads need separate qualification")
	}
	writeResponseFixture(t, "privacy_configuration", "get_global", response, nil)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	accessOutcomeOracle(t, ctx, tenant, "privacy", "Get-CsPrivacyConfiguration", "forbidden")
	t.Log("confirmed authorization denial: HTTP 403, PolicyRP 40301; PowerShell forbidden response matched")
}

func TestUnitAccessOutcomeRejectsUnverifiedOrDifferentResults(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	good := `{"success":true,"tenant_id":"` + tenant + `","command":"Get-CsPrivacyConfiguration","outcome":"forbidden","code":"Forbidden","error_code":40301}`
	for _, test := range []struct {
		name, raw string
		valid     bool
	}{
		{"verified", good, true},
		{"failed-setup", strings.Replace(good, `"success":true`, `"success":false`, 1), false},
		{"wrong-tenant", strings.Replace(good, tenant, "00000000-0000-4000-8000-000000000004", 1), false},
		{"wrong-code", strings.Replace(good, "40301", "40003", 1), false},
		{"wrong-outcome", strings.Replace(good, "forbidden", "success", 1), false},
		{"trailing-data", good + ` {}`, false},
		{"unknown-property", strings.Replace(good, `"success":true`, `"extra":true,"success":true`, 1), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeAccessOutcome([]byte(test.raw), tenant, "Get-CsPrivacyConfiguration", "forbidden")
			if (err == nil) != test.valid {
				t.Fatal("unverified access outcome accepted or verified outcome rejected")
			}
		})
	}
}
