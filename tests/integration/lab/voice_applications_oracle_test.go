package lab_test

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

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

var voiceApplicationsPolicyScalars = []string{"AllowAutoAttendantAfterHoursGreetingChange", "AllowAutoAttendantAfterHoursRoutingChange", "AllowAutoAttendantBusinessHoursChange", "AllowAutoAttendantBusinessHoursGreetingChange", "AllowAutoAttendantBusinessHoursRoutingChange", "AllowAutoAttendantHolidayGreetingChange", "AllowAutoAttendantHolidayRoutingChange", "AllowAutoAttendantHolidaysChange", "AllowAutoAttendantLanguageChange", "AllowAutoAttendantTimeZoneChange", "AllowCallQueueAgentOptChange", "AllowCallQueueConferenceModeChange", "AllowCallQueueLanguageChange", "AllowCallQueueMembershipChange", "AllowCallQueueMusicOnHoldChange", "AllowCallQueueNoAgentSharedVoicemailGreetingChange", "AllowCallQueueNoAgentsRoutingChange", "AllowCallQueueOptOutChange", "AllowCallQueueOverflowRoutingChange", "AllowCallQueueOverflowSharedVoicemailGreetingChange", "AllowCallQueuePresenceBasedRoutingChange", "AllowCallQueueRoutingMethodChange", "AllowCallQueueTimeoutRoutingChange", "AllowCallQueueTimeoutSharedVoicemailGreetingChange", "AllowCallQueueWelcomeGreetingChange", "CallQueueAgentMonitorMode", "CallQueueAgentMonitorNotificationMode", "Description", "HistoricalAgentMetricsPermission", "HistoricalAutoAttendantMetricsPermission", "HistoricalCallQueueMetricsPermission", "RealTimeAgentMetricsPermission", "RealTimeAutoAttendantMetricsPermission", "RealTimeCallQueueMetricsPermission"}

type voiceApplicationsPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var voiceApplicationsPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-voice-applications-policy$`)

func decodeVoiceApplicationsPolicyOracleForIdentity(raw []byte, tenant, identity string) (voiceApplicationsPolicyObservation, error) {
	var v voiceApplicationsPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !voiceApplicationsPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return voiceApplicationsPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(voiceApplicationsPolicyScalars) {
		return voiceApplicationsPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return voiceApplicationsPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range voiceApplicationsPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return voiceApplicationsPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return voiceApplicationsPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func voiceApplicationsPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) voiceApplicationsPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid voice-applications policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-voice-applications-policy.ps1")
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
		t.Fatal("pinned voice-applications policy read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeVoiceApplicationsPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned voice-applications policy projection unavailable")
	}
	return value
}

func matchesVoiceApplicationsPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return labkit.SameProjection(voiceApplicationsProjection(v), settings)
}

func voiceApplicationsProjection(v *resource.Policy) map[string]any {
	return map[string]any{"AllowAutoAttendantAfterHoursGreetingChange": v.AllowAutoAttendantAfterHoursGreetingChange, "AllowAutoAttendantAfterHoursRoutingChange": v.AllowAutoAttendantAfterHoursRoutingChange, "AllowAutoAttendantBusinessHoursChange": v.AllowAutoAttendantBusinessHoursChange, "AllowAutoAttendantBusinessHoursGreetingChange": v.AllowAutoAttendantBusinessHoursGreetingChange, "AllowAutoAttendantBusinessHoursRoutingChange": v.AllowAutoAttendantBusinessHoursRoutingChange, "AllowAutoAttendantHolidayGreetingChange": v.AllowAutoAttendantHolidayGreetingChange, "AllowAutoAttendantHolidayRoutingChange": v.AllowAutoAttendantHolidayRoutingChange, "AllowAutoAttendantHolidaysChange": v.AllowAutoAttendantHolidaysChange, "AllowAutoAttendantLanguageChange": v.AllowAutoAttendantLanguageChange, "AllowAutoAttendantTimeZoneChange": v.AllowAutoAttendantTimeZoneChange, "AllowCallQueueAgentOptChange": v.AllowCallQueueAgentOptChange, "AllowCallQueueConferenceModeChange": v.AllowCallQueueConferenceModeChange, "AllowCallQueueLanguageChange": v.AllowCallQueueLanguageChange, "AllowCallQueueMembershipChange": v.AllowCallQueueMembershipChange, "AllowCallQueueMusicOnHoldChange": v.AllowCallQueueMusicOnHoldChange, "AllowCallQueueNoAgentSharedVoicemailGreetingChange": v.AllowCallQueueNoAgentSharedVoicemailGreetingChange, "AllowCallQueueNoAgentsRoutingChange": v.AllowCallQueueNoAgentsRoutingChange, "AllowCallQueueOptOutChange": v.AllowCallQueueOptOutChange, "AllowCallQueueOverflowRoutingChange": v.AllowCallQueueOverflowRoutingChange, "AllowCallQueueOverflowSharedVoicemailGreetingChange": v.AllowCallQueueOverflowSharedVoicemailGreetingChange, "AllowCallQueuePresenceBasedRoutingChange": v.AllowCallQueuePresenceBasedRoutingChange, "AllowCallQueueRoutingMethodChange": v.AllowCallQueueRoutingMethodChange, "AllowCallQueueTimeoutRoutingChange": v.AllowCallQueueTimeoutRoutingChange, "AllowCallQueueTimeoutSharedVoicemailGreetingChange": v.AllowCallQueueTimeoutSharedVoicemailGreetingChange, "AllowCallQueueWelcomeGreetingChange": v.AllowCallQueueWelcomeGreetingChange, "CallQueueAgentMonitorMode": v.CallQueueAgentMonitorMode, "CallQueueAgentMonitorNotificationMode": v.CallQueueAgentMonitorNotificationMode, "Description": v.Description, "HistoricalAgentMetricsPermission": v.HistoricalAgentMetricsPermission, "HistoricalAutoAttendantMetricsPermission": v.HistoricalAutoAttendantMetricsPermission, "HistoricalCallQueueMetricsPermission": v.HistoricalCallQueueMetricsPermission, "RealTimeAgentMetricsPermission": v.RealTimeAgentMetricsPermission, "RealTimeAutoAttendantMetricsPermission": v.RealTimeAutoAttendantMetricsPermission, "RealTimeCallQueueMetricsPermission": v.RealTimeCallQueueMetricsPermission}
}
