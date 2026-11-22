package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type applicationMeetingObservation struct {
	Success  bool     `json:"success"`
	Identity string   `json:"identity"`
	AppIDs   []string `json:"app_ids"`
}

func decodeApplicationMeetingOracle(raw []byte) (applicationMeetingObservation, error) {
	var wire struct {
		Success  bool            `json:"success"`
		Identity string          `json:"identity"`
		AppIDs   json.RawMessage `json:"app_ids"`
	}
	if len(raw) > labkit.MaxOracleBytes {
		return applicationMeetingObservation{}, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&wire) != nil || !wire.Success || !strings.EqualFold(wire.Identity, "Global") || len(wire.AppIDs) == 0 {
		return applicationMeetingObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return applicationMeetingObservation{}, labkit.ErrOracle
	}
	var ids []string
	if json.Unmarshal(wire.AppIDs, &ids) != nil {
		return applicationMeetingObservation{}, labkit.ErrOracle
	}
	return applicationMeetingObservation{Success: true, Identity: wire.Identity, AppIDs: ids}, nil
}

func applicationMeetingOracle(t *testing.T, ctx context.Context) applicationMeetingObservation {
	t.Helper()
	cmd := labkit.OracleCommand(ctx, "read-application-meeting.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned configuration read oracle failed; no output retained")
	}
	value, err := decodeApplicationMeetingOracle(output.Data)
	if err != nil {
		t.Fatal("pinned configuration projection unavailable")
	}
	return value
}

// TestAccApplicationMeetingConfigurationReadOnly never updates the tenant-wide
// allow-list. Mutation acceptance needs a separate explicitly authorized restore plan.
func TestAccApplicationMeetingConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
	value, fixtureResponse, err := s.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "application_meeting_configuration", "get_global", fixtureResponse, nil)
	if err != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("application meeting Global read failed")
	}
	listed, listResponse, err := s.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "application_meeting_configuration", "list", listResponse, nil)
	if err != nil || len(listed) != 1 || !labkit.SameConfigurationFields(value.RawFields, listed[0].RawFields) {
		t.Fatal("application meeting filtered read differs")
	}
	if oracle {
		actual := applicationMeetingOracle(t, ctx)
		if !strings.EqualFold(actual.Identity, value.Identity) || !reflect.DeepEqual(actual.AppIDs, value.AppIDs) {
			t.Fatal("PowerShell and Go application meeting fields differ")
		}
	}
	t.Log("selected configuration reads observed; tenant-wide update effects remain separately qualified")
}

func TestUnitApplicationMeetingOracleRejectsIncompleteAndMixedOutput(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"success":false}`, `{"success":true,"identity":"Other"}`, `{"success":true,"identity":"Global"}`, `{"success":true,"identity":"Global","app_ids":[],"unknown":false}`, `{"success":true,"identity":"Global"} {}`, `{"success":true,"identity":"Global","app_ids":42}`} {
		if _, err := decodeApplicationMeetingOracle([]byte(raw)); err == nil {
			t.Fatal("invalid oracle output accepted")
		}
	}
	for _, ids := range []string{"null", "[]", `["a",""]`} {
		if _, err := decodeApplicationMeetingOracle([]byte(`{"success":true,"identity":"Global","app_ids":` + ids + `}`)); err != nil {
			t.Fatal("valid nullable list projection rejected")
		}
	}
}
