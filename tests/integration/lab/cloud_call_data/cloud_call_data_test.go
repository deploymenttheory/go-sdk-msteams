package cloudcalldataacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func powerShellOutcome(t *testing.T, ctx context.Context, action string) string {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-cloud-call-data.ps1")
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_CLOUD_CALL_DATA_ACTION="+action)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell cloud call data call failed; no output retained")
	}
	var observed struct {
		Success bool   `json:"success"`
		Outcome string `json:"outcome"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell cloud call data projection unavailable")
	}
	return observed.Outcome
}

// TestAccCloudCallDataConnectionAuthorizationDenied reads the tenant's cloud
// returned cloud call data connections. App-only Teams Administrator credentials are denied;
// a successful read (which returns a secret token) needs separate
// qualification. It never creates a connection.
func TestAccCloudCallDataConnectionAuthorizationDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	values, response, err := sdk.TeamsAdminAPI.CloudCallData.List(ctx)
	labkit.LogAccessResponse(t, response, err)
	if values != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("expected cloud call data authorization denial changed; successful reads need separate qualification")
	}
	labkit.WriteResponseFixture(t, "cloud_call_data", "get", response, nil)
	if labkit.OracleEnabled(t) {
		if powerShellOutcome(t, ctx, "get") != "forbidden" {
			t.Fatal("PowerShell and Go cloud call data outcomes differ")
		}
		t.Log("PowerShell 8.1.0 / Go cloud call data parity: passed")
	}
	t.Log("confirmed authorization denial: HTTP 403 Forbidden")
}

// TestAccCloudCallDataConnectionCreateOutcome sends the tenant singleton
// create (New-CsCloudCallDataConnection), approved for the lab even though
// the connection cannot be removed. App-only Teams Administrator credentials
// are denied with HTTP 403 and the tenant is unchanged. The test
// records the authorized denial or token-bearing returned models. Client source does
// not establish server idempotence, ownership or safe replay after ambiguity.
func TestAccCloudCallDataConnectionCreateOutcome(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	result, response, err := sdk.TeamsAdminAPI.CloudCallData.CreateWithResult(ctx)
	labkit.LogAccessResponse(t, response, err)
	switch {
	case response != nil && response.StatusCode() == 403 && result != nil && result.Returned == nil && errors.Is(err, auth.ErrForbidden):
		labkit.WriteResponseFixture(t, "cloud_call_data", "create", response, nil)
		if labkit.OracleEnabled(t) {
			if powerShellOutcome(t, ctx, "new") != "forbidden" {
				t.Fatal("PowerShell and Go cloud call data create outcomes differ")
			}
			t.Log("PowerShell 8.1.0 / Go cloud call data create parity: passed")
		}
		t.Log("confirmed create authorization denial: HTTP 403 Forbidden")
	case err == nil && response != nil && response.StatusCode() == 200 && result != nil:
		tokenBearing := 0
		for _, value := range result.Returned {
			if value != nil && value.Token != nil && *value.Token != "" {
				tokenBearing++
			}
		}
		if tokenBearing == 0 {
			t.Fatal("accepted create returned no token-bearing model; investigate the retained receipt before replay")
		}
		t.Logf("modern create returned %d rows (%d token-bearing); ownership, persistence and idempotence remain unqualified", len(result.Returned), tokenBearing)
	default:
		t.Fatal("cloud call data create outcome changed; neither the observed denial nor a token-bearing connection")
	}
}
