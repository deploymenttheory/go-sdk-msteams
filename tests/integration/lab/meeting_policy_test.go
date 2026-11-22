package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_policy"
)

var meetingPolicyScalars = []string{"Description", "AllowChannelMeetingScheduling", "AllowMeetNow", "AllowPrivateMeetNow", "MeetingChatEnabledType", "AllowExternalNonTrustedMeetingChat", "CopyRestriction", "LiveCaptionsEnabledType", "DesignatedPresenterRoleMode", "AllowIPAudio", "AllowIPVideo", "AllowEngagementReport", "AllowTrackingInReport", "IPAudioMode", "IPVideoMode", "AllowAnonymousUsersToDialOut", "AllowAnonymousUsersToStartMeeting", "AllowAnonymousUsersToJoinMeeting", "ConditionalAccessAttendeeVerification", "BlockedAnonymousJoinClientTypes", "AllowedStreamingMediaInput", "ExplicitRecordingConsent", "PreventComplianceRecording", "EnableRecordingAndTranscriptionCustomMessage", "RecordingAndTranscriptionCustomMessageIdentifier", "AllowLocalRecording", "AutoRecording", "ParticipantNameChange", "AllowPrivateMeetingScheduling", "AutoAdmittedUsers", "AllowCloudRecording", "SetRecordingAndTranscriptOwnership", "RecordingAndTranscriptionAudioNotification", "FilterProfanityInTranscript", "AllowRecordingStorageOutsideRegion", "RecordingStorageMode", "AllowMeetingKnowledgeGeneration", "KnowledgeGenerationMeetingTypes", "MeetingKnowledgeExpirationDays", "AllowOutlookAddIn", "AllowPowerPointSharing", "AllowParticipantGiveRequestControl", "AllowExternalParticipantGiveRequestControl", "AllowSharedNotes", "AllowWhiteboard", "AllowTranscription", "AllowNetworkConfigurationSettingsLookup", "MediaBitRateKb", "ScreenSharingMode", "AllowMultipleScreenshare", "VideoFiltersMode", "AllowPSTNUsersToBypassLobby", "AllowOrganizersToOverrideLobbySettings", "PreferredMeetingProviderForIslandsMode", "AllowNDIStreaming", "SpeakerAttributionMode", "EnrollUserOverride", "RoomAttributeUserOverride", "StreamingAttendeeMode", "AttendeeIdentityMasking", "AllowBreakoutRooms", "TeamsCameraFarEndPTZMode", "AllowMeetingReactions", "AllowMeetingRegistration", "WhoCanRegister", "AllowScreenContentDigitization", "AllowCarbonSummary", "RoomPeopleNameUserOverride", "AllowMeetingCoach", "NewMeetingRecordingExpirationDays", "LiveStreamingMode", "MeetingInviteLanguages", "ChannelRecordingDownload", "AllowCartCaptionsScheduling", "AllowTasksFromTranscript", "InfoShownInReportMode", "LiveInterpretationEnabledType", "QnAEngagementMode", "AllowImmersiveView", "AllowAvatarsInGallery", "AllowAnnotations", "AllowDocumentCollaboration", "AllowWatermarkForScreenSharing", "AllowWatermarkForCameraVideo", "AllowWatermarkCustomizationForCameraVideo", "WatermarkForCameraVideoOpacity", "WatermarkForCameraVideoPattern", "AllowWatermarkCustomizationForScreenSharing", "WatermarkForScreenSharingOpacity", "WatermarkForScreenSharingPattern", "WatermarkForAnonymousUsers", "DetectSensitiveContentDuringScreenSharing", "AudibleRecordingNotification", "ConnectToMeetingControls", "Copilot", "TranscriptionWithMeetingAI", "MeetingAI", "AutomaticallyStartCopilot", "VoiceIsolation", "ExternalMeetingJoin", "ContentSharingInExternalMeetings", "AllowedUsersForMeetingDetails", "SmsNotifications", "CaptchaVerificationForMeetingJoin", "UsersCanAdmitFromLobby", "LobbyChat", "BackroomChat", "AnonymousUserAuthenticationMethod", "NoiseSuppressionForDialInParticipants", "RealTimeText", "AIInterpreter", "VoiceSimulationInInterpreter", "ParticipantSlideControl", "PasscodeComplexity", "ExternalBotAccessMode", "SyntheticMediaDetection", "SyntheticMediaDetectionAppId", "DisableAudioAnnouncementsForResourceAccounts", "EnablePreMeetingConsent", "PreMeetingConsentContentIdentifier", "EnableExternalRecordingDetection", "IntelligentRecapDocxFileExpirationDays", "AllowIntelligentRecap"}

type meetingPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var meetingPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-meeting-policy$`)

func decodeMeetingPolicyOracle(raw []byte, tenant string) (meetingPolicyObservation, error) {
	return decodeMeetingPolicyOracleForIdentity(raw, tenant, "Global")
}
func decodeMeetingPolicyOracleForIdentity(raw []byte, tenant, identity string) (meetingPolicyObservation, error) {
	var v meetingPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !meetingPolicyOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return meetingPolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(meetingPolicyScalars) {
		return meetingPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return meetingPolicyObservation{}, errTeamOracle
	}
	for _, name := range meetingPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return meetingPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return meetingPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func meetingPolicyOracle(t *testing.T, ctx context.Context, tenant string) meetingPolicyObservation {
	t.Helper()
	return meetingPolicyOracleForIdentity(t, ctx, tenant, "Global")
}
func meetingPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) meetingPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid meeting policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-meeting-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned meeting policy read failed; no output retained")
	}
	value, err := decodeMeetingPolicyOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned meeting policy projection unavailable")
	}
	return value
}

// TestAccMeetingPolicyReadOnly performs configuration reads only.
// Creates, updates, deletes, grants and effective policy behavior need separate acceptance.
func TestAccMeetingPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsMeetingPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsMeetingPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("meeting policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsMeetingPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_policy", "list", listResponse, nil)
	selected, selectionErr := globalMeetingPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(meetingPolicyProjection(value), meetingPolicyProjection(selected)) {
		t.Fatal("meeting policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := meetingPolicyOracle(t, ctx, tenant)
		if !matchesMeetingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned meeting policy getter projections differ")
		}

	}
	t.Log("selected meeting policy reads verified; persistence and effective policy behavior remain separately qualified")
}

func TestUnitMeetingPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := map[string]any{}
	for _, name := range meetingPolicyScalars {
		settings[name] = nil
	}
	settings["Description"] = ""
	settings["AllowIPVideo"] = false
	settings["AllowIPAudio"] = true
	raw, err := json.Marshal(meetingPolicyObservation{Success: true, Tenant: tenant, Identity: "Global", Settings: func() map[string]json.RawMessage {
		encoded, _ := json.Marshal(settings)
		var value map[string]json.RawMessage
		_ = json.Unmarshal(encoded, &value)
		return value
	}()})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	for _, input := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"Description":""`, `"Description":false`, 1), strings.Replace(valid, `"AllowIPAudio":true`, `"AllowIPAudio":"true"`, 1), strings.Replace(valid, `,"AllowIPAudio":true`, "", 1)} {
		if _, e := decodeMeetingPolicyOracle([]byte(input), tenant); e == nil {
			t.Fatal("invalid meeting policy oracle accepted")
		}
	}
	if _, e := decodeMeetingPolicyOracle(raw, tenant); e != nil {
		t.Fatal("complete nullable meeting policy projection rejected")
	}
}

func meetingPolicyProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "AllowChannelMeetingScheduling": v.AllowChannelMeetingScheduling, "AllowMeetNow": v.AllowMeetNow, "AllowPrivateMeetNow": v.AllowPrivateMeetNow, "MeetingChatEnabledType": v.MeetingChatEnabledType, "AllowExternalNonTrustedMeetingChat": v.AllowExternalNonTrustedMeetingChat, "CopyRestriction": v.CopyRestriction, "LiveCaptionsEnabledType": v.LiveCaptionsEnabledType, "DesignatedPresenterRoleMode": v.DesignatedPresenterRoleMode, "AllowIPAudio": v.AllowIPAudio, "AllowIPVideo": v.AllowIPVideo, "AllowEngagementReport": v.AllowEngagementReport, "AllowTrackingInReport": v.AllowTrackingInReport, "IPAudioMode": v.IPAudioMode, "IPVideoMode": v.IPVideoMode, "AllowAnonymousUsersToDialOut": v.AllowAnonymousUsersToDialOut, "AllowAnonymousUsersToStartMeeting": v.AllowAnonymousUsersToStartMeeting, "AllowAnonymousUsersToJoinMeeting": v.AllowAnonymousUsersToJoinMeeting, "ConditionalAccessAttendeeVerification": v.ConditionalAccessAttendeeVerification, "BlockedAnonymousJoinClientTypes": v.BlockedAnonymousJoinClientTypes, "AllowedStreamingMediaInput": v.AllowedStreamingMediaInput, "ExplicitRecordingConsent": v.ExplicitRecordingConsent, "PreventComplianceRecording": v.PreventComplianceRecording, "EnableRecordingAndTranscriptionCustomMessage": v.EnableRecordingAndTranscriptionCustomMessage, "RecordingAndTranscriptionCustomMessageIdentifier": v.RecordingAndTranscriptionCustomMessageIdentifier, "AllowLocalRecording": v.AllowLocalRecording, "AutoRecording": v.AutoRecording, "ParticipantNameChange": v.ParticipantNameChange, "AllowPrivateMeetingScheduling": v.AllowPrivateMeetingScheduling, "AutoAdmittedUsers": v.AutoAdmittedUsers, "AllowCloudRecording": v.AllowCloudRecording, "SetRecordingAndTranscriptOwnership": v.SetRecordingAndTranscriptOwnership, "RecordingAndTranscriptionAudioNotification": v.RecordingAndTranscriptionAudioNotification, "FilterProfanityInTranscript": v.FilterProfanityInTranscript, "AllowRecordingStorageOutsideRegion": v.AllowRecordingStorageOutsideRegion, "RecordingStorageMode": v.RecordingStorageMode, "AllowMeetingKnowledgeGeneration": v.AllowMeetingKnowledgeGeneration, "KnowledgeGenerationMeetingTypes": v.KnowledgeGenerationMeetingTypes, "MeetingKnowledgeExpirationDays": v.MeetingKnowledgeExpirationDays, "AllowOutlookAddIn": v.AllowOutlookAddIn, "AllowPowerPointSharing": v.AllowPowerPointSharing, "AllowParticipantGiveRequestControl": v.AllowParticipantGiveRequestControl, "AllowExternalParticipantGiveRequestControl": v.AllowExternalParticipantGiveRequestControl, "AllowSharedNotes": v.AllowSharedNotes, "AllowWhiteboard": v.AllowWhiteboard, "AllowTranscription": v.AllowTranscription, "AllowNetworkConfigurationSettingsLookup": v.AllowNetworkConfigurationSettingsLookup, "MediaBitRateKb": v.MediaBitRateKb, "ScreenSharingMode": v.ScreenSharingMode, "AllowMultipleScreenshare": v.AllowMultipleScreenshare, "VideoFiltersMode": v.VideoFiltersMode, "AllowPSTNUsersToBypassLobby": v.AllowPSTNUsersToBypassLobby, "AllowOrganizersToOverrideLobbySettings": v.AllowOrganizersToOverrideLobbySettings, "PreferredMeetingProviderForIslandsMode": v.PreferredMeetingProviderForIslandsMode, "AllowNDIStreaming": v.AllowNDIStreaming, "SpeakerAttributionMode": v.SpeakerAttributionMode, "EnrollUserOverride": v.EnrollUserOverride, "RoomAttributeUserOverride": v.RoomAttributeUserOverride, "StreamingAttendeeMode": v.StreamingAttendeeMode, "AttendeeIdentityMasking": v.AttendeeIdentityMasking, "AllowBreakoutRooms": v.AllowBreakoutRooms, "TeamsCameraFarEndPTZMode": v.TeamsCameraFarEndPTZMode, "AllowMeetingReactions": v.AllowMeetingReactions, "AllowMeetingRegistration": v.AllowMeetingRegistration, "WhoCanRegister": v.WhoCanRegister, "AllowScreenContentDigitization": v.AllowScreenContentDigitization, "AllowCarbonSummary": v.AllowCarbonSummary, "RoomPeopleNameUserOverride": v.RoomPeopleNameUserOverride, "AllowMeetingCoach": v.AllowMeetingCoach, "NewMeetingRecordingExpirationDays": v.NewMeetingRecordingExpirationDays, "LiveStreamingMode": v.LiveStreamingMode, "MeetingInviteLanguages": v.MeetingInviteLanguages, "ChannelRecordingDownload": v.ChannelRecordingDownload, "AllowCartCaptionsScheduling": v.AllowCartCaptionsScheduling, "AllowTasksFromTranscript": v.AllowTasksFromTranscript, "InfoShownInReportMode": v.InfoShownInReportMode, "LiveInterpretationEnabledType": v.LiveInterpretationEnabledType, "QnAEngagementMode": v.QnAEngagementMode, "AllowImmersiveView": v.AllowImmersiveView, "AllowAvatarsInGallery": v.AllowAvatarsInGallery, "AllowAnnotations": v.AllowAnnotations, "AllowDocumentCollaboration": v.AllowDocumentCollaboration, "AllowWatermarkForScreenSharing": v.AllowWatermarkForScreenSharing, "AllowWatermarkForCameraVideo": v.AllowWatermarkForCameraVideo, "AllowWatermarkCustomizationForCameraVideo": v.AllowWatermarkCustomizationForCameraVideo, "WatermarkForCameraVideoOpacity": v.WatermarkForCameraVideoOpacity, "WatermarkForCameraVideoPattern": v.WatermarkForCameraVideoPattern, "AllowWatermarkCustomizationForScreenSharing": v.AllowWatermarkCustomizationForScreenSharing, "WatermarkForScreenSharingOpacity": v.WatermarkForScreenSharingOpacity, "WatermarkForScreenSharingPattern": v.WatermarkForScreenSharingPattern, "WatermarkForAnonymousUsers": v.WatermarkForAnonymousUsers, "DetectSensitiveContentDuringScreenSharing": v.DetectSensitiveContentDuringScreenSharing, "AudibleRecordingNotification": v.AudibleRecordingNotification, "ConnectToMeetingControls": v.ConnectToMeetingControls, "Copilot": v.Copilot, "TranscriptionWithMeetingAI": v.TranscriptionWithMeetingAI, "MeetingAI": v.MeetingAI, "AutomaticallyStartCopilot": v.AutomaticallyStartCopilot, "VoiceIsolation": v.VoiceIsolation, "ExternalMeetingJoin": v.ExternalMeetingJoin, "ContentSharingInExternalMeetings": v.ContentSharingInExternalMeetings, "AllowedUsersForMeetingDetails": v.AllowedUsersForMeetingDetails, "SmsNotifications": v.SmsNotifications, "CaptchaVerificationForMeetingJoin": v.CaptchaVerificationForMeetingJoin, "UsersCanAdmitFromLobby": v.UsersCanAdmitFromLobby, "LobbyChat": v.LobbyChat, "BackroomChat": v.BackroomChat, "AnonymousUserAuthenticationMethod": v.AnonymousUserAuthenticationMethod, "NoiseSuppressionForDialInParticipants": v.NoiseSuppressionForDialInParticipants, "RealTimeText": v.RealTimeText, "AIInterpreter": v.AIInterpreter, "VoiceSimulationInInterpreter": v.VoiceSimulationInInterpreter, "ParticipantSlideControl": v.ParticipantSlideControl, "PasscodeComplexity": v.PasscodeComplexity, "ExternalBotAccessMode": v.ExternalBotAccessMode, "SyntheticMediaDetection": v.SyntheticMediaDetection, "SyntheticMediaDetectionAppId": v.SyntheticMediaDetectionAppId, "DisableAudioAnnouncementsForResourceAccounts": v.DisableAudioAnnouncementsForResourceAccounts, "EnablePreMeetingConsent": v.EnablePreMeetingConsent, "PreMeetingConsentContentIdentifier": v.PreMeetingConsentContentIdentifier, "EnableExternalRecordingDetection": v.EnableExternalRecordingDetection, "IntelligentRecapDocxFileExpirationDays": v.IntelligentRecapDocxFileExpirationDays, "AllowIntelligentRecap": v.AllowIntelligentRecap})
	return raw
}
func globalMeetingPolicy(values []*resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, "Global") {
			if found != nil {
				return nil, errTeamOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, errTeamOracle
	}
	return found, nil
}
func TestUnitMeetingPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalMeetingPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalMeetingPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesMeetingPolicyGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(meetingPolicyProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitMeetingPolicyOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"description":"fixture"}`, `{"Description":"earlier","description":"fixture"}`, `{"Description":null,"description":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range meetingPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["Description"] = json.RawMessage(`"fixture"`)
		if !matchesMeetingPolicyGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}

func TestUnitMeetingPolicyOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-meeting-policy"
	raw, err := os.ReadFile("../../../msteams/teams_admin_api/teams_meeting_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range meetingPolicyScalars {
		projection[field] = settings[field]
	}
	for _, tc := range []struct {
		name, selected, returned string
		valid                    bool
	}{
		{"global", "Global", "global", true},
		{"owned-bare", name, "Tag:" + name, true},
		{"owned-tagged", "Tag:" + name, name, true},
		{"foreign-selection", "other-policy", "other-policy", false},
		{"wrong-owned-policy", name, strings.Replace(name, "aaaa", "bbbb", 1), false},
		{"tagged-global", "Global", "Tag:Global", false},
		{"wrong-kind", strings.Replace(name, "meeting-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeMeetingPolicyOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
