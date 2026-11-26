package teamsguestmeetingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	guestmeeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// guestMeetingSettings are the eight public getters of the pinned compiled type.
var guestMeetingSettings = []string{"AllowIPVideo", "ScreenSharingMode", "AllowMultipleScreenshare", "AllowMeetNow", "LiveCaptionsEnabledType", "AllowTranscription", "AllowParticipantGiveRequestControl", "AllowExternalParticipantGiveRequestControl"}

func guestMeetingProjection(v *guestmeeting.Configuration) map[string]any {
	return map[string]any{
		"AllowIPVideo":                               v.AllowIPVideo,
		"ScreenSharingMode":                          v.ScreenSharingMode,
		"AllowMultipleScreenshare":                   v.AllowMultipleScreenshare,
		"AllowMeetNow":                               v.AllowMeetNow,
		"LiveCaptionsEnabledType":                    v.LiveCaptionsEnabledType,
		"AllowTranscription":                         v.AllowTranscription,
		"AllowParticipantGiveRequestControl":         v.AllowParticipantGiveRequestControl,
		"AllowExternalParticipantGiveRequestControl": v.AllowExternalParticipantGiveRequestControl,
	}
}

// guestMeetingOracle runs the pinned 8.1.0 Global getter in a dedicated
// read-only PowerShell process and returns its eight public settings.
func guestMeetingOracle(t *testing.T, ctx context.Context, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": "guest-meeting"})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-guest-meeting-configuration.ps1")
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
		t.Fatal("pinned guest meeting configuration read process failed")
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
		t.Fatal("pinned guest meeting configuration projection unavailable")
	}
	if len(result.Settings) != len(guestMeetingSettings) {
		t.Fatal("incomplete guest meeting configuration getter projection")
	}
	for _, name := range guestMeetingSettings {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("incomplete guest meeting configuration getter projection")
		}
	}
	return result.Settings
}
