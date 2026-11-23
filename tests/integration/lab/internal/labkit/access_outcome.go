package labkit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type AccessOutcome struct {
	Success   bool    `json:"success"`
	Tenant    string  `json:"tenant_id"`
	Command   string  `json:"command"`
	Outcome   string  `json:"outcome"`
	Code      *string `json:"code"`
	ErrorCode *int    `json:"error_code"`
}

func DecodeAccessOutcome(raw []byte, tenant, command, outcome string) (AccessOutcome, error) {
	var result AccessOutcome
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if len(raw) > MaxOracleBytes || !DirectoryID.MatchString(tenant) || d.Decode(&result) != nil || !result.Success || !strings.EqualFold(result.Tenant, tenant) || result.Command != command || result.Outcome != outcome {
		return AccessOutcome{}, ErrOracle
	}
	if d.Decode(new(any)) != io.EOF {
		return AccessOutcome{}, ErrOracle
	}
	switch outcome {
	case "migration-status-message":
		if result.Code != nil || result.ErrorCode != nil {
			return AccessOutcome{}, ErrOracle
		}
	case "feature-unavailable":
		if result.Code == nil || *result.Code != "ClientError" || result.ErrorCode == nil || *result.ErrorCode != 40003 {
			return AccessOutcome{}, ErrOracle
		}
	case "forbidden":
		if result.Code == nil || *result.Code != "Forbidden" || result.ErrorCode == nil || *result.ErrorCode != 40301 {
			return AccessOutcome{}, ErrOracle
		}
	default:
		return AccessOutcome{}, ErrOracle
	}
	return result, nil
}

func AccessOutcomeOracle(t *testing.T, ctx context.Context, tenant, resource, command, outcome string) {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "resource": resource})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := OracleCommand(ctx, "read-access-outcome.ps1")
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
	var output OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned access outcome oracle failed")
	}
	if _, err := DecodeAccessOutcome(output.Data, tenant, command, outcome); err != nil {
		t.Fatal("PowerShell access outcome changed or was not verified")
	}
}
