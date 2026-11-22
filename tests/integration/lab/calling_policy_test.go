package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

var callingPolicyScalars = []string{"Description", "AllowPrivateCalling", "AllowWebPSTNCalling", "AllowSIPDevicesCalling", "AllowVoicemail", "AllowCallGroups", "AllowDelegation", "AllowCallForwardingToUser", "AllowCallForwardingToPhone", "PreventTollBypass", "BusyOnBusyEnabledType", "MusicOnHoldEnabledType", "AllowCloudRecordingForCalls", "ExplicitRecordingConsent", "PreventComplianceRecording", "EnableRecordingAndTranscriptionCustomMessage", "RecordingAndTranscriptionCustomMessageIdentifier", "AllowMeetingKnowledgeGeneration", "AllowTranscriptionForCalling", "RecordingAndTranscriptionAudioNotification", "PopoutForIncomingPstnCalls", "PopoutAppPathForIncomingPstnCalls", "LiveCaptionsEnabledTypeForCalling", "AutoAnswerEnabledType", "SpamFilteringEnabledType", "CallRecordingExpirationDays", "AllowCallRedirect", "InboundPstnCallRoutingTreatment", "InboundFederatedCallRoutingTreatment", "EnableWebPstnMediaBypass", "EnableSpendLimits", "CallingSpendUserLimit", "Copilot", "ShowTeamsCallsInCallLog", "RealTimeText", "AIInterpreter", "VoiceSimulationInInterpreter", "ReportCall", "VoicePhishingDetection"}

type callingPolicyObservation struct {
	FailureStage string                     `json:"failure_stage,omitempty"`
	Success      bool                       `json:"success"`
	Tenant       string                     `json:"tenant_id"`
	Identity     string                     `json:"identity"`
	Settings     map[string]json.RawMessage `json:"settings"`
}

var callingOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-calling-policy$`)

func decodeCallingPolicyOracle(raw []byte, tenant string) (callingPolicyObservation, error) {
	return decodeCallingPolicyOracleForIdentity(raw, tenant, "Global")
}

func decodeCallingPolicyOracleForIdentity(raw []byte, tenant, identity string) (callingPolicyObservation, error) {
	var v callingPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !callingOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return callingPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(callingPolicyScalars) {
		return callingPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return callingPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range callingPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return callingPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return callingPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func callingPolicyOracle(t *testing.T, ctx context.Context, tenant string) callingPolicyObservation {
	t.Helper()
	return callingPolicyOracleForIdentity(t, ctx, tenant, "Global")
}

func callingPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) callingPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid calling policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-calling-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned calling policy read failed; no output retained")
	}
	value, err := decodeCallingPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		var diagnostic struct {
			Success bool   `json:"success"`
			Stage   string `json:"failure_stage"`
		}
		if json.Unmarshal(output.Data, &diagnostic) == nil && !diagnostic.Success {
			stage := diagnostic.Stage
			if slices.Contains([]string{"enablement", "selectors", "baseline", "module", "identity", "authentication", "read", "selected-result"}, stage) || strings.HasPrefix(stage, "property:") && slices.Contains(callingPolicyScalars, strings.TrimPrefix(stage, "property:")) {
				t.Fatalf("pinned calling policy read failed at %s", stage)
			}
		}
		t.Fatal("pinned calling policy projection identity/types differ")
	}
	return value
}

// TestAccCallingPolicyReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccCallingPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("calling policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsCallingPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_calling_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("calling policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsCallingPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_calling_policy", "list", listResponse, nil)
	selected, selectionErr := globalCallingPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(callingProjection(value), callingProjection(selected)) {
		t.Fatal("calling Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := callingPolicyOracle(t, ctx, tenant)
		if !matchesCallingGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned calling getter projections differ")
		}

	}
	t.Log("selected calling policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func TestUnitCallingPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":"","AllowPrivateCalling":false,"AllowWebPSTNCalling":null,"AllowSIPDevicesCalling":null,"AllowVoicemail":null,"AllowCallGroups":null,"AllowDelegation":null,"AllowCallForwardingToUser":null,"AllowCallForwardingToPhone":null,"PreventTollBypass":null,"BusyOnBusyEnabledType":null,"MusicOnHoldEnabledType":null,"AllowCloudRecordingForCalls":null,"ExplicitRecordingConsent":null,"PreventComplianceRecording":null,"EnableRecordingAndTranscriptionCustomMessage":null,"RecordingAndTranscriptionCustomMessageIdentifier":null,"AllowMeetingKnowledgeGeneration":null,"AllowTranscriptionForCalling":null,"RecordingAndTranscriptionAudioNotification":null,"PopoutForIncomingPstnCalls":null,"PopoutAppPathForIncomingPstnCalls":null,"LiveCaptionsEnabledTypeForCalling":null,"AutoAnswerEnabledType":null,"SpamFilteringEnabledType":null,"CallRecordingExpirationDays":null,"AllowCallRedirect":null,"InboundPstnCallRoutingTreatment":null,"InboundFederatedCallRoutingTreatment":null,"EnableWebPstnMediaBypass":null,"EnableSpendLimits":null,"CallingSpendUserLimit":9223372036854775807,"Copilot":null,"ShowTeamsCallsInCallLog":null,"RealTimeText":null,"AIInterpreter":null,"VoiceSimulationInInterpreter":null,"ReportCall":null,"VoicePhishingDetection":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"CallingSpendUserLimit":9223372036854775807`, `"CallingSpendUserLimit":9223372036854775808`, 1), strings.Replace(valid, `"CallingSpendUserLimit":9223372036854775807`, `"CallingSpendUserLimit":1.5`, 1), strings.Replace(valid, `"Description":""`, `"Description":false`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeCallingPolicyOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid calling policy oracle output accepted")
		}
	}
	if _, err := decodeCallingPolicyOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable calling policy projection rejected")
	}
}

func callingProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "AllowPrivateCalling": v.AllowPrivateCalling, "AllowWebPSTNCalling": v.AllowWebPSTNCalling, "AllowSIPDevicesCalling": v.AllowSIPDevicesCalling, "AllowVoicemail": v.AllowVoicemail, "AllowCallGroups": v.AllowCallGroups, "AllowDelegation": v.AllowDelegation, "AllowCallForwardingToUser": v.AllowCallForwardingToUser, "AllowCallForwardingToPhone": v.AllowCallForwardingToPhone, "PreventTollBypass": v.PreventTollBypass, "BusyOnBusyEnabledType": v.BusyOnBusyEnabledType, "MusicOnHoldEnabledType": v.MusicOnHoldEnabledType, "AllowCloudRecordingForCalls": v.AllowCloudRecordingForCalls, "ExplicitRecordingConsent": v.ExplicitRecordingConsent, "PreventComplianceRecording": v.PreventComplianceRecording, "EnableRecordingAndTranscriptionCustomMessage": v.EnableRecordingAndTranscriptionCustomMessage, "RecordingAndTranscriptionCustomMessageIdentifier": v.RecordingAndTranscriptionCustomMessageIdentifier, "AllowMeetingKnowledgeGeneration": v.AllowMeetingKnowledgeGeneration, "AllowTranscriptionForCalling": v.AllowTranscriptionForCalling, "RecordingAndTranscriptionAudioNotification": v.RecordingAndTranscriptionAudioNotification, "PopoutForIncomingPstnCalls": v.PopoutForIncomingPSTNCalls, "PopoutAppPathForIncomingPstnCalls": v.PopoutAppPathForIncomingPSTNCalls, "LiveCaptionsEnabledTypeForCalling": v.LiveCaptionsEnabledTypeForCalling, "AutoAnswerEnabledType": v.AutoAnswerEnabledType, "SpamFilteringEnabledType": v.SpamFilteringEnabledType, "CallRecordingExpirationDays": v.CallRecordingExpirationDays, "AllowCallRedirect": v.AllowCallRedirect, "InboundPstnCallRoutingTreatment": v.InboundPSTNCallRoutingTreatment, "InboundFederatedCallRoutingTreatment": v.InboundFederatedCallRoutingTreatment, "EnableWebPstnMediaBypass": v.EnableWebPSTNMediaBypass, "EnableSpendLimits": v.EnableSpendLimits, "CallingSpendUserLimit": v.CallingSpendUserLimit, "Copilot": v.Copilot, "ShowTeamsCallsInCallLog": v.ShowTeamsCallsInCallLog, "RealTimeText": v.RealTimeText, "AIInterpreter": v.AIInterpreter, "VoiceSimulationInInterpreter": v.VoiceSimulationInInterpreter, "ReportCall": v.ReportCall, "VoicePhishingDetection": v.VoicePhishingDetection})
	return raw
}
func globalCallingPolicy(values []*resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, "Global") {
			if found != nil {
				return nil, labkit.ErrOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
	}
	return found, nil
}
func TestUnitCallingPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalCallingPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalCallingPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesCallingGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	a, b := json.NewDecoder(bytes.NewReader(callingProjection(value))), json.NewDecoder(bytes.NewReader(raw))
	a.UseNumber()
	b.UseNumber()
	return a.Decode(&left) == nil && b.Decode(&right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitCallingOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"description":"fixture"}`, `{"Description":"earlier","description":"fixture"}`, `{"Description":null,"description":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range callingPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["Description"] = json.RawMessage(`"fixture"`)
		if !matchesCallingGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
