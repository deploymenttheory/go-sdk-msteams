package teamscallingpolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/teams_calling_policy/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsCallingPolicy"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"

type inputVectors struct {
	Name, Operation, Input, Response string
	Status                           int
	ForceMode                        int     `json:"force_mode"`
	PolicyName                       *string `json:"policy_name"`
	Request                          struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	v := struct {
		Cases []inputVectors `json:"cases"`
	}{
		Cases: []inputVectors{
			{
				Name:      "read",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "list-empty",
				Operation:  "list",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "create-all",
				Operation: "create",
				Input:     "{\"Description\":\"Future-Description\",\"AllowPrivateCalling\":false,\"AllowWebPSTNCalling\":false,\"AllowSIPDevicesCalling\":false,\"AllowVoicemail\":\"Future-AllowVoicemail\",\"AllowCallGroups\":false,\"AllowDelegation\":false,\"AllowCallForwardingToUser\":false,\"AllowCallForwardingToPhone\":false,\"PreventTollBypass\":false,\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\",\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\",\"AllowCloudRecordingForCalls\":false,\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\",\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\",\"EnableRecordingAndTranscriptionCustomMessage\":false,\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789abcdef\",\"AllowMeetingKnowledgeGeneration\":false,\"AllowTranscriptionForCalling\":false,\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\",\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\",\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\",\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\",\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\",\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\",\"CallRecordingExpirationDays\":9999,\"AllowCallRedirect\":\"Future-AllowCallRedirect\",\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\",\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\",\"EnableWebPstnMediaBypass\":false,\"EnableSpendLimits\":false,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":\"Future-Copilot\",\"ShowTeamsCallsInCallLog\":false,\"RealTimeText\":\"Future-RealTimeText\",\"AIInterpreter\":\"Future-AIInterpreter\",\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\",\"ReportCall\":\"Future-ReportCall\",\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"Future-Description\",\n  \"AllowPrivateCalling\": false,\n  \"AllowWebPSTNCalling\": false,\n  \"AllowSIPDevicesCalling\": false,\n  \"AllowVoicemail\": \"Future-AllowVoicemail\",\n  \"AllowCallGroups\": false,\n  \"AllowDelegation\": false,\n  \"AllowCallForwardingToUser\": false,\n  \"AllowCallForwardingToPhone\": false,\n  \"PreventTollBypass\": false,\n  \"BusyOnBusyEnabledType\": \"Future-BusyOnBusyEnabledType\",\n  \"MusicOnHoldEnabledType\": \"Future-MusicOnHoldEnabledType\",\n  \"AllowCloudRecordingForCalls\": false,\n  \"ExplicitRecordingConsent\": \"Future-ExplicitRecordingConsent\",\n  \"PreventComplianceRecording\": \"Future-PreventComplianceRecording\",\n  \"EnableRecordingAndTranscriptionCustomMessage\": false,\n  \"RecordingAndTranscriptionCustomMessageIdentifier\": \"01234567-89AB-CDEF-0123-456789abcdef\",\n  \"AllowMeetingKnowledgeGeneration\": false,\n  \"AllowTranscriptionForCalling\": false,\n  \"RecordingAndTranscriptionAudioNotification\": \"Future-RecordingAndTranscriptionAudioNotification\",\n  \"PopoutForIncomingPstnCalls\": \"Future-PopoutForIncomingPstnCalls\",\n  \"PopoutAppPathForIncomingPstnCalls\": \"Future-PopoutAppPathForIncomingPstnCalls\",\n  \"LiveCaptionsEnabledTypeForCalling\": \"Future-LiveCaptionsEnabledTypeForCalling\",\n  \"AutoAnswerEnabledType\": \"Future-AutoAnswerEnabledType\",\n  \"SpamFilteringEnabledType\": \"Future-SpamFilteringEnabledType\",\n  \"CallRecordingExpirationDays\": 9999,\n  \"AllowCallRedirect\": \"Future-AllowCallRedirect\",\n  \"InboundPstnCallRoutingTreatment\": \"Future-InboundPstnCallRoutingTreatment\",\n  \"InboundFederatedCallRoutingTreatment\": \"Future-InboundFederatedCallRoutingTreatment\",\n  \"EnableWebPstnMediaBypass\": false,\n  \"EnableSpendLimits\": false,\n  \"CallingSpendUserLimit\": 9007199254740993,\n  \"Copilot\": \"Future-Copilot\",\n  \"ShowTeamsCallsInCallLog\": false,\n  \"RealTimeText\": \"Future-RealTimeText\",\n  \"AIInterpreter\": \"Future-AIInterpreter\",\n  \"VoiceSimulationInInterpreter\": \"Future-VoiceSimulationInInterpreter\",\n  \"ReportCall\": \"Future-ReportCall\",\n  \"VoicePhishingDetection\": \"Future-VoicePhishingDetection\"\n}",

					CmdletName: "New-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "create-empty",
				Operation:  "create",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "New-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"Description\":\"Future-Description\",\"AllowPrivateCalling\":false,\"AllowWebPSTNCalling\":false,\"AllowSIPDevicesCalling\":false,\"AllowVoicemail\":\"Future-AllowVoicemail\",\"AllowCallGroups\":false,\"AllowDelegation\":false,\"AllowCallForwardingToUser\":false,\"AllowCallForwardingToPhone\":false,\"PreventTollBypass\":false,\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\",\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\",\"AllowCloudRecordingForCalls\":false,\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\",\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\",\"EnableRecordingAndTranscriptionCustomMessage\":false,\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789abcdef\",\"AllowMeetingKnowledgeGeneration\":false,\"AllowTranscriptionForCalling\":false,\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\",\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\",\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\",\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\",\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\",\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\",\"CallRecordingExpirationDays\":9999,\"AllowCallRedirect\":\"Future-AllowCallRedirect\",\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\",\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\",\"EnableWebPstnMediaBypass\":false,\"EnableSpendLimits\":false,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":\"Future-Copilot\",\"ShowTeamsCallsInCallLog\":false,\"RealTimeText\":\"Future-RealTimeText\",\"AIInterpreter\":\"Future-AIInterpreter\",\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\",\"ReportCall\":\"Future-ReportCall\",\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"Future-Description\",\n  \"AllowPrivateCalling\": false,\n  \"AllowWebPSTNCalling\": false,\n  \"AllowSIPDevicesCalling\": false,\n  \"AllowVoicemail\": \"Future-AllowVoicemail\",\n  \"AllowCallGroups\": false,\n  \"AllowDelegation\": false,\n  \"AllowCallForwardingToUser\": false,\n  \"AllowCallForwardingToPhone\": false,\n  \"PreventTollBypass\": false,\n  \"BusyOnBusyEnabledType\": \"Future-BusyOnBusyEnabledType\",\n  \"MusicOnHoldEnabledType\": \"Future-MusicOnHoldEnabledType\",\n  \"AllowCloudRecordingForCalls\": false,\n  \"ExplicitRecordingConsent\": \"Future-ExplicitRecordingConsent\",\n  \"PreventComplianceRecording\": \"Future-PreventComplianceRecording\",\n  \"EnableRecordingAndTranscriptionCustomMessage\": false,\n  \"RecordingAndTranscriptionCustomMessageIdentifier\": \"01234567-89AB-CDEF-0123-456789abcdef\",\n  \"AllowMeetingKnowledgeGeneration\": false,\n  \"AllowTranscriptionForCalling\": false,\n  \"RecordingAndTranscriptionAudioNotification\": \"Future-RecordingAndTranscriptionAudioNotification\",\n  \"PopoutForIncomingPstnCalls\": \"Future-PopoutForIncomingPstnCalls\",\n  \"PopoutAppPathForIncomingPstnCalls\": \"Future-PopoutAppPathForIncomingPstnCalls\",\n  \"LiveCaptionsEnabledTypeForCalling\": \"Future-LiveCaptionsEnabledTypeForCalling\",\n  \"AutoAnswerEnabledType\": \"Future-AutoAnswerEnabledType\",\n  \"SpamFilteringEnabledType\": \"Future-SpamFilteringEnabledType\",\n  \"CallRecordingExpirationDays\": 9999,\n  \"AllowCallRedirect\": \"Future-AllowCallRedirect\",\n  \"InboundPstnCallRoutingTreatment\": \"Future-InboundPstnCallRoutingTreatment\",\n  \"InboundFederatedCallRoutingTreatment\": \"Future-InboundFederatedCallRoutingTreatment\",\n  \"EnableWebPstnMediaBypass\": false,\n  \"EnableSpendLimits\": false,\n  \"CallingSpendUserLimit\": 9007199254740993,\n  \"Copilot\": \"Future-Copilot\",\n  \"ShowTeamsCallsInCallLog\": false,\n  \"RealTimeText\": \"Future-RealTimeText\",\n  \"AIInterpreter\": \"Future-AIInterpreter\",\n  \"VoiceSimulationInInterpreter\": \"Future-VoiceSimulationInInterpreter\",\n  \"ReportCall\": \"Future-ReportCall\",\n  \"VoicePhishingDetection\": \"Future-VoicePhishingDetection\"\n}",

					CmdletName: "Set-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"Description\":null,\"AllowPrivateCalling\":null,\"CallingSpendUserLimit\":null,\"CallRecordingExpirationDays\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": null,\n  \"AllowPrivateCalling\": null,\n  \"RecordingAndTranscriptionCustomMessageIdentifier\": null,\n  \"CallRecordingExpirationDays\": null,\n  \"CallingSpendUserLimit\": null\n}",

					CmdletName: "Set-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "patch-empty",
				Operation:  "patch",
				Input:      "{\"Description\":\"\"}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\"\n}",

					CmdletName: "Set-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"Description\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"Description\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "delete-force-false",
				Operation: "delete",
				Input:     "{}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "read-empty",
				Operation:  "read",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "read-null",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "grant-user-null",
				Operation: "user",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsCallingPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "grant-tenant-null",
				Operation: "tenant",
				Input:     "{}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsCallingPolicy",
					Body:   "{\n  \"PolicyName\": null,\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
			{
				Name:      "grant-group-null",
				Operation: "group",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsCallingPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "grant-user-empty",
				Operation:  "user",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsCallingPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "grant-tenant-empty",
				Operation:  "tenant",
				Input:      "{}",
				Status:     400,
				ForceMode:  1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsCallingPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": true\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
			{
				Name:       "grant-group-empty",
				Operation:  "group",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsCallingPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsCallingPolicy",
				},
			},
		},
	}

	return v.Cases
}
func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_calling_policy"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("calling policy contract differs: %v %+v", e, r.Differences())
	}
	return r
}

const setCommand = "Set-CsTeamsCallingPolicy"
