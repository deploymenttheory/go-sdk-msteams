package teamsmeetingbroadcastpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var meetingBroadcastPolicyScalars = []string{"Description", "AllowBroadcastScheduling", "AllowBroadcastTranscription", "BroadcastAttendeeVisibilityMode", "BroadcastRecordingMode"}

type meetingBroadcastPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var meetingBroadcastPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-meeting-broadcast-policy$`)

func decodeMeetingBroadcastPolicyOracleForIdentity(raw []byte, tenant, identity string) (meetingBroadcastPolicyObservation, error) {
	var v meetingBroadcastPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !meetingBroadcastPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return meetingBroadcastPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(meetingBroadcastPolicyScalars) {
		return meetingBroadcastPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return meetingBroadcastPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range meetingBroadcastPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return meetingBroadcastPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model struct {
		Description                     *string
		AllowBroadcastScheduling        *bool
		AllowBroadcastTranscription     *bool
		BroadcastAttendeeVisibilityMode *string
		BroadcastRecordingMode          *string
	}
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return meetingBroadcastPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func meetingBroadcastPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) meetingBroadcastPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid meeting broadcast policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-meeting-broadcast-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned meeting broadcast policy read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeMeetingBroadcastPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned meeting broadcast policy projection unavailable")
	}
	return value
}

func matchesMeetingBroadcastPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return labkit.SameProjection(broadcastSettings(v), settings)
}

// broadcastSettings is the public getter projection compared with PowerShell.
func broadcastSettings(v *resource.Policy) map[string]any {
	return map[string]any{"Description": v.Description, "AllowBroadcastScheduling": v.AllowBroadcastScheduling, "AllowBroadcastTranscription": v.AllowBroadcastTranscription, "BroadcastAttendeeVisibilityMode": v.BroadcastAttendeeVisibilityMode, "BroadcastRecordingMode": v.BroadcastRecordingMode}
}
