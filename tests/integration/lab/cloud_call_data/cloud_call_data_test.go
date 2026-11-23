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

func powerShellOutcome(t *testing.T, ctx context.Context) string {
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell cloud call data read failed; no output retained")
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
// call data connection. App-only Teams Administrator credentials are denied;
// a successful read (which returns a secret token) needs separate
// qualification. It never creates a connection.
func TestAccCloudCallDataConnectionAuthorizationDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
	labkit.LogAccessResponse(t, response, err)
	if value != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("expected cloud call data authorization denial changed; successful reads need separate qualification")
	}
	labkit.WriteResponseFixture(t, "cloud_call_data", "get", response, nil)
	if labkit.OracleEnabled(t) {
		if powerShellOutcome(t, ctx) != "forbidden" {
			t.Fatal("PowerShell and Go cloud call data outcomes differ")
		}
		t.Log("PowerShell 8.1.0 / Go cloud call data parity: passed")
	}
	t.Log("confirmed authorization denial: HTTP 403 Forbidden")
}
