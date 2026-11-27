package teamscallingpolicy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

type modelCase struct {
	Name, Input string
	Observation struct {
		Failure *string
		Value   json.RawMessage
	}
}

func modelCases(t *testing.T) []modelCase {
	t.Helper()

	fixture := struct{ Models []modelCase }{
		Models: []modelCase{
			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "all-settings",
				Input: "{\"Description\":\"Future-Description\",\"AllowPrivateCalling\":false,\"AllowWebPSTNCalling\":false,\"AllowSIPDevicesCalling\":false,\"AllowVoicemail\":\"Future-AllowVoicemail\",\"AllowCallGroups\":false,\"AllowDelegation\":false,\"AllowCallForwardingToUser\":false,\"AllowCallForwardingToPhone\":false,\"PreventTollBypass\":false,\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\",\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\",\"AllowCloudRecordingForCalls\":false,\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\",\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\",\"EnableRecordingAndTranscriptionCustomMessage\":false,\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789abcdef\",\"AllowMeetingKnowledgeGeneration\":false,\"AllowTranscriptionForCalling\":false,\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\",\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\",\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\",\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\",\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\",\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\",\"CallRecordingExpirationDays\":9999,\"AllowCallRedirect\":\"Future-AllowCallRedirect\",\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\",\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\",\"EnableWebPstnMediaBypass\":false,\"EnableSpendLimits\":false,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":\"Future-Copilot\",\"ShowTeamsCallsInCallLog\":false,\"RealTimeText\":\"Future-RealTimeText\",\"AIInterpreter\":\"Future-AIInterpreter\",\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\",\"ReportCall\":\"Future-ReportCall\",\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":\"Future-AIInterpreter\",\"AllowCallForwardingToPhone\":false,\"AllowCallForwardingToUser\":false,\"AllowCallGroups\":false,\"AllowCallRedirect\":\"Future-AllowCallRedirect\",\"AllowCloudRecordingForCalls\":false,\"AllowDelegation\":false,\"AllowMeetingKnowledgeGeneration\":false,\"AllowPrivateCalling\":false,\"AllowSIPDevicesCalling\":false,\"AllowTranscriptionForCalling\":false,\"AllowVoicemail\":\"Future-AllowVoicemail\",\"AllowWebPSTNCalling\":false,\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\",\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\",\"CallRecordingExpirationDays\":9999,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":\"Future-Copilot\",\"Description\":\"Future-Description\",\"EnableRecordingAndTranscriptionCustomMessage\":false,\"EnableSpendLimits\":false,\"EnableWebPstnMediaBypass\":false,\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\",\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\",\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\",\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\",\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\",\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\",\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\",\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\",\"PreventTollBypass\":false,\"RealTimeText\":\"Future-RealTimeText\",\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\",\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789abcdef\",\"ReportCall\":\"Future-ReportCall\",\"ShowTeamsCallsInCallLog\":false,\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\",\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\",\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\"}"),
				},
			},
			{
				Name:  "Description-null",
				Input: "{\"Description\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "Description-empty",
				Input: "{\"Description\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":\"\",\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "Description-future",
				Input: "{\"Description\":\"Future-Description\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":\"Future-Description\",\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowPrivateCalling-null",
				Input: "{\"AllowPrivateCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowPrivateCalling-false",
				Input: "{\"AllowPrivateCalling\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":false,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowPrivateCalling-true",
				Input: "{\"AllowPrivateCalling\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":true,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowWebPSTNCalling-null",
				Input: "{\"AllowWebPSTNCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowWebPSTNCalling-false",
				Input: "{\"AllowWebPSTNCalling\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":false,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowWebPSTNCalling-true",
				Input: "{\"AllowWebPSTNCalling\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":true,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowSIPDevicesCalling-null",
				Input: "{\"AllowSIPDevicesCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowSIPDevicesCalling-false",
				Input: "{\"AllowSIPDevicesCalling\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":false,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowSIPDevicesCalling-true",
				Input: "{\"AllowSIPDevicesCalling\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":true,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowVoicemail-null",
				Input: "{\"AllowVoicemail\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowVoicemail-empty",
				Input: "{\"AllowVoicemail\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":\"\",\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowVoicemail-future",
				Input: "{\"AllowVoicemail\":\"Future-AllowVoicemail\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":\"Future-AllowVoicemail\",\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallGroups-null",
				Input: "{\"AllowCallGroups\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallGroups-false",
				Input: "{\"AllowCallGroups\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":false,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallGroups-true",
				Input: "{\"AllowCallGroups\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":true,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowDelegation-null",
				Input: "{\"AllowDelegation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowDelegation-false",
				Input: "{\"AllowDelegation\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":false,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowDelegation-true",
				Input: "{\"AllowDelegation\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":true,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToUser-null",
				Input: "{\"AllowCallForwardingToUser\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToUser-false",
				Input: "{\"AllowCallForwardingToUser\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":false,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToUser-true",
				Input: "{\"AllowCallForwardingToUser\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":true,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToPhone-null",
				Input: "{\"AllowCallForwardingToPhone\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToPhone-false",
				Input: "{\"AllowCallForwardingToPhone\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":false,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallForwardingToPhone-true",
				Input: "{\"AllowCallForwardingToPhone\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":true,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventTollBypass-null",
				Input: "{\"PreventTollBypass\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventTollBypass-false",
				Input: "{\"PreventTollBypass\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":false,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventTollBypass-true",
				Input: "{\"PreventTollBypass\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":true,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "BusyOnBusyEnabledType-null",
				Input: "{\"BusyOnBusyEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "BusyOnBusyEnabledType-empty",
				Input: "{\"BusyOnBusyEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":\"\",\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "BusyOnBusyEnabledType-future",
				Input: "{\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":\"Future-BusyOnBusyEnabledType\",\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "MusicOnHoldEnabledType-null",
				Input: "{\"MusicOnHoldEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "MusicOnHoldEnabledType-empty",
				Input: "{\"MusicOnHoldEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":\"\",\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "MusicOnHoldEnabledType-future",
				Input: "{\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":\"Future-MusicOnHoldEnabledType\",\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCloudRecordingForCalls-null",
				Input: "{\"AllowCloudRecordingForCalls\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCloudRecordingForCalls-false",
				Input: "{\"AllowCloudRecordingForCalls\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":false,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCloudRecordingForCalls-true",
				Input: "{\"AllowCloudRecordingForCalls\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":true,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ExplicitRecordingConsent-null",
				Input: "{\"ExplicitRecordingConsent\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ExplicitRecordingConsent-empty",
				Input: "{\"ExplicitRecordingConsent\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":\"\",\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ExplicitRecordingConsent-future",
				Input: "{\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":\"Future-ExplicitRecordingConsent\",\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventComplianceRecording-null",
				Input: "{\"PreventComplianceRecording\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventComplianceRecording-empty",
				Input: "{\"PreventComplianceRecording\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":\"\",\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PreventComplianceRecording-future",
				Input: "{\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":\"Future-PreventComplianceRecording\",\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableRecordingAndTranscriptionCustomMessage-null",
				Input: "{\"EnableRecordingAndTranscriptionCustomMessage\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableRecordingAndTranscriptionCustomMessage-false",
				Input: "{\"EnableRecordingAndTranscriptionCustomMessage\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":false,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableRecordingAndTranscriptionCustomMessage-true",
				Input: "{\"EnableRecordingAndTranscriptionCustomMessage\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":true,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionCustomMessageIdentifier-null",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionCustomMessageIdentifier-empty",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionCustomMessageIdentifier-future",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":\"Future-RecordingAndTranscriptionCustomMessageIdentifier\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "AllowMeetingKnowledgeGeneration-null",
				Input: "{\"AllowMeetingKnowledgeGeneration\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowMeetingKnowledgeGeneration-false",
				Input: "{\"AllowMeetingKnowledgeGeneration\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":false,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowMeetingKnowledgeGeneration-true",
				Input: "{\"AllowMeetingKnowledgeGeneration\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":true,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowTranscriptionForCalling-null",
				Input: "{\"AllowTranscriptionForCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowTranscriptionForCalling-false",
				Input: "{\"AllowTranscriptionForCalling\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":false,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowTranscriptionForCalling-true",
				Input: "{\"AllowTranscriptionForCalling\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":true,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionAudioNotification-null",
				Input: "{\"RecordingAndTranscriptionAudioNotification\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionAudioNotification-empty",
				Input: "{\"RecordingAndTranscriptionAudioNotification\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":\"\",\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RecordingAndTranscriptionAudioNotification-future",
				Input: "{\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":\"Future-RecordingAndTranscriptionAudioNotification\",\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutForIncomingPstnCalls-null",
				Input: "{\"PopoutForIncomingPstnCalls\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutForIncomingPstnCalls-empty",
				Input: "{\"PopoutForIncomingPstnCalls\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":\"\",\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutForIncomingPstnCalls-future",
				Input: "{\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":\"Future-PopoutForIncomingPstnCalls\",\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutAppPathForIncomingPstnCalls-null",
				Input: "{\"PopoutAppPathForIncomingPstnCalls\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutAppPathForIncomingPstnCalls-empty",
				Input: "{\"PopoutAppPathForIncomingPstnCalls\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":\"\",\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "PopoutAppPathForIncomingPstnCalls-future",
				Input: "{\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":\"Future-PopoutAppPathForIncomingPstnCalls\",\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "LiveCaptionsEnabledTypeForCalling-null",
				Input: "{\"LiveCaptionsEnabledTypeForCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "LiveCaptionsEnabledTypeForCalling-empty",
				Input: "{\"LiveCaptionsEnabledTypeForCalling\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":\"\",\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "LiveCaptionsEnabledTypeForCalling-future",
				Input: "{\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":\"Future-LiveCaptionsEnabledTypeForCalling\",\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AutoAnswerEnabledType-null",
				Input: "{\"AutoAnswerEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AutoAnswerEnabledType-empty",
				Input: "{\"AutoAnswerEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":\"\",\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AutoAnswerEnabledType-future",
				Input: "{\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":\"Future-AutoAnswerEnabledType\",\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "SpamFilteringEnabledType-null",
				Input: "{\"SpamFilteringEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "SpamFilteringEnabledType-empty",
				Input: "{\"SpamFilteringEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":\"\",\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "SpamFilteringEnabledType-future",
				Input: "{\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":\"Future-SpamFilteringEnabledType\",\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-null",
				Input: "{\"CallRecordingExpirationDays\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallRedirect-null",
				Input: "{\"AllowCallRedirect\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallRedirect-empty",
				Input: "{\"AllowCallRedirect\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":\"\",\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AllowCallRedirect-future",
				Input: "{\"AllowCallRedirect\":\"Future-AllowCallRedirect\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":\"Future-AllowCallRedirect\",\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundPstnCallRoutingTreatment-null",
				Input: "{\"InboundPstnCallRoutingTreatment\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundPstnCallRoutingTreatment-empty",
				Input: "{\"InboundPstnCallRoutingTreatment\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":\"\",\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundPstnCallRoutingTreatment-future",
				Input: "{\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":\"Future-InboundPstnCallRoutingTreatment\",\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundFederatedCallRoutingTreatment-null",
				Input: "{\"InboundFederatedCallRoutingTreatment\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundFederatedCallRoutingTreatment-empty",
				Input: "{\"InboundFederatedCallRoutingTreatment\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":\"\",\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "InboundFederatedCallRoutingTreatment-future",
				Input: "{\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":\"Future-InboundFederatedCallRoutingTreatment\",\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableWebPstnMediaBypass-null",
				Input: "{\"EnableWebPstnMediaBypass\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableWebPstnMediaBypass-false",
				Input: "{\"EnableWebPstnMediaBypass\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":false,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableWebPstnMediaBypass-true",
				Input: "{\"EnableWebPstnMediaBypass\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":true,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableSpendLimits-null",
				Input: "{\"EnableSpendLimits\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableSpendLimits-false",
				Input: "{\"EnableSpendLimits\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":false,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "EnableSpendLimits-true",
				Input: "{\"EnableSpendLimits\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":true,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-null",
				Input: "{\"CallingSpendUserLimit\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "Copilot-null",
				Input: "{\"Copilot\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "Copilot-empty",
				Input: "{\"Copilot\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":\"\",\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "Copilot-future",
				Input: "{\"Copilot\":\"Future-Copilot\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":\"Future-Copilot\",\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ShowTeamsCallsInCallLog-null",
				Input: "{\"ShowTeamsCallsInCallLog\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ShowTeamsCallsInCallLog-false",
				Input: "{\"ShowTeamsCallsInCallLog\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":false,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ShowTeamsCallsInCallLog-true",
				Input: "{\"ShowTeamsCallsInCallLog\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":true,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RealTimeText-null",
				Input: "{\"RealTimeText\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RealTimeText-empty",
				Input: "{\"RealTimeText\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":\"\",\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "RealTimeText-future",
				Input: "{\"RealTimeText\":\"Future-RealTimeText\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":\"Future-RealTimeText\",\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AIInterpreter-null",
				Input: "{\"AIInterpreter\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AIInterpreter-empty",
				Input: "{\"AIInterpreter\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":\"\",\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "AIInterpreter-future",
				Input: "{\"AIInterpreter\":\"Future-AIInterpreter\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":\"Future-AIInterpreter\",\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "VoiceSimulationInInterpreter-null",
				Input: "{\"VoiceSimulationInInterpreter\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "VoiceSimulationInInterpreter-empty",
				Input: "{\"VoiceSimulationInInterpreter\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":\"\"}"),
				},
			},
			{
				Name:  "VoiceSimulationInInterpreter-future",
				Input: "{\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":\"Future-VoiceSimulationInInterpreter\"}"),
				},
			},
			{
				Name:  "ReportCall-null",
				Input: "{\"ReportCall\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ReportCall-empty",
				Input: "{\"ReportCall\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":\"\",\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "ReportCall-future",
				Input: "{\"ReportCall\":\"Future-ReportCall\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":\"Future-ReportCall\",\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "VoicePhishingDetection-null",
				Input: "{\"VoicePhishingDetection\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "VoicePhishingDetection-empty",
				Input: "{\"VoicePhishingDetection\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":\"\",\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "VoicePhishingDetection-future",
				Input: "{\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":\"Future-VoicePhishingDetection\",\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays--9223372036854775808",
				Input: "{\"CallRecordingExpirationDays\":-9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit--9223372036854775808",
				Input: "{\"CallingSpendUserLimit\":-9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":-9223372036854775808,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays--1",
				Input: "{\"CallRecordingExpirationDays\":-1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit--1",
				Input: "{\"CallingSpendUserLimit\":-1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":-1,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-0",
				Input: "{\"CallRecordingExpirationDays\":0}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-0",
				Input: "{\"CallingSpendUserLimit\":0}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":0,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-1",
				Input: "{\"CallRecordingExpirationDays\":1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":1,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-1",
				Input: "{\"CallingSpendUserLimit\":1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":1,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-9999",
				Input: "{\"CallRecordingExpirationDays\":9999}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":9999,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-9999",
				Input: "{\"CallingSpendUserLimit\":9999}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":9999,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-10000",
				Input: "{\"CallRecordingExpirationDays\":10000}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-10000",
				Input: "{\"CallingSpendUserLimit\":10000}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":10000,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-9007199254740993",
				Input: "{\"CallRecordingExpirationDays\":9007199254740993}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-9007199254740993",
				Input: "{\"CallingSpendUserLimit\":9007199254740993}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-9223372036854775807",
				Input: "{\"CallRecordingExpirationDays\":9223372036854775807}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-9223372036854775807",
				Input: "{\"CallingSpendUserLimit\":9223372036854775807}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":9223372036854775807,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-json-9223372036854775808",
				Input: "{\"CallRecordingExpirationDays\":9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-json-9223372036854775808",
				Input: "{\"CallingSpendUserLimit\":9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-json-1.5",
				Input: "{\"CallRecordingExpirationDays\":1.5}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":2,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-json-1.5",
				Input: "{\"CallingSpendUserLimit\":1.5}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":2,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "CallRecordingExpirationDays-json-\"9007199254740993\"",
				Input: "{\"CallRecordingExpirationDays\":\"9007199254740993\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "CallingSpendUserLimit-json-\"9007199254740993\"",
				Input: "{\"CallingSpendUserLimit\":\"9007199254740993\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":9007199254740993,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "guid-uppercase",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789ABCDEF\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89AB-CDEF-0123-456789ABCDEF\",\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "guid-trailing-lf",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89ab-cdef-0123-456789abcdef\\n\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":\"01234567-89ab-cdef-0123-456789abcdef\\n\",\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "guid-braces",
				Input: "{\"RecordingAndTranscriptionCustomMessageIdentifier\":\"{01234567-89ab-cdef-0123-456789abcdef}\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
			{
				Name:  "long-description",
				Input: "{\"Description\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBy" +
						"pass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":nu" +
						"ll,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabl" +
						"edType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComp" +
						"lianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNo" +
						"tification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeam" +
						"sCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulation" +
						"InInterpreter\":null}"),
				},
			},
			{
				Name:  "case-duplicate",
				Input: "{\"Description\":\"first\",\"description\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":\"second\",\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
			{
				Name:  "duplicate-null",
				Input: "{\"AllowPrivateCalling\":true,\"AllowPrivateCalling\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AIInterpreter\":null,\"AllowCallForwardingToPhone\":null,\"AllowCallForwardingToUser\":null,\"AllowCallGroups\":null,\"AllowCallRedirect\":null,\"AllowCloudRecordingForCalls\":null,\"AllowDelegation\":null,\"AllowMeetingKnowledgeGeneration\":null,\"AllowPrivateCalling\":null,\"AllowSIPDevicesCalling\":null,\"AllowTranscriptionForCalling\":null,\"AllowVoicemail\":null,\"AllowWebPSTNCalling\":null,\"AutoAnswerEnabledType\":null,\"BusyOnBusyEnabledType\":null,\"CallRecordingExpirationDays\":null,\"CallingSpendUserLimit\":null,\"Copilot\":null,\"Description\":null,\"EnableRecordingAndTranscriptionCustomMessage\":null,\"EnableSpendLimits\":null,\"EnableWebPstnMediaBypass\":null,\"ExplicitRecordingConsent\":null,\"Identity\":null,\"InboundFederatedCallRoutingTreatment\":null,\"InboundPstnCallRoutingTreatment\":null,\"LiveCaptionsEnabledTypeForCalling\":null,\"MusicOnHoldEnabledType\":null,\"PopoutAppPathForIncomingPstnCalls\":null,\"PopoutForIncomingPstnCalls\":null,\"PreventComplianceRecording\":null,\"PreventTollBypass\":null,\"RealTimeText\":null,\"RecordingAndTranscriptionAudioNotification\":null,\"RecordingAndTranscriptionCustomMessageIdentifier\":null,\"ReportCall\":null,\"ShowTeamsCallsInCallLog\":null,\"SpamFilteringEnabledType\":null,\"VoicePhishingDetection\":null,\"VoiceSimulationInInterpreter\":null}"),
				},
			},
		},
	}

	return fixture.Models
}
func exactJSON(t *testing.T, raw []byte) any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestUnitCallingPolicyKnownProjectionsMatchTypedModels(t *testing.T) {
	t.Parallel()
	for _, m := range modelCases(t) {
		t.Run(m.Name, func(t *testing.T) {
			if strings.Contains(m.Name, "-json-1.5") || strings.Contains(m.Name, `-json-"9007199254740993"`) {
				return
			}
			var value resource.Policy
			var err error
			if m.Input != "constructor" {
				err = json.Unmarshal([]byte(m.Input), &value)
			}
			if m.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source-invalid model accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source-valid typed model rejected")
			}
			type known resource.Policy
			raw, err := json.Marshal(known(value))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exactJSON(t, raw), exactJSON(t, m.Observation.Value)) {
				t.Fatal("known model projection differs")
			}
		})
	}
}
func TestUnitCallingPolicyPreservesOwnedRawSnapshotAndExactInt64(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","CallingSpendUserLimit":9223372036854775807,"Future":{"n":9007199254740993},"Description":"first","description":null}`)
	saved := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.CallingSpendUserLimit == nil || *value.CallingSpendUserLimit != 9223372036854775807 || value.Description != nil {
		t.Fatal("nullable/exact projection differs")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	got, err := json.Marshal(value)
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatal("snapshot lost bytes or ownership")
	}
}
func TestUnitCallingPolicyRequiresNativeJSONTypes(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"CallingSpendUserLimit":"9007199254740993"}`, `{"CallingSpendUserLimit":1.5}`, `{"CallRecordingExpirationDays":1.5}`, `{"AllowPrivateCalling":"false"}`, `{"Description":{}}`, `null`, `[]`} {
		var value resource.Policy
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("non-native JSON type accepted")
		}
	}
}

func TestUnitCallingPolicyRequestsRetainTypedModelSettings(t *testing.T) {
	t.Parallel()

	for index, m := range modelCases(t) {
		if m.Observation.Failure != nil || m.Input == "constructor" || m.Input == "{}" {
			continue
		}
		var input map[string]json.RawMessage
		var known map[string]json.RawMessage
		if json.Unmarshal([]byte(m.Input), &input) != nil || json.Unmarshal(m.Observation.Value, &known) != nil {
			t.Fatal("invalid model input/projection")
		}
		bound := map[string]json.RawMessage{}
		for key := range input {
			for canonical, value := range known {
				if strings.EqualFold(key, canonical) {
					bound[canonical] = value
					break
				}
			}
		}
		raw, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		var settings resource.Settings
		if json.Unmarshal(raw, &settings) != nil {
			t.Fatal("accepted final settings not representable")
		}
		bound["Identity"] = json.RawMessage(`"Global"`)
		body, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		run(t, testutil.Case{Name: fmt.Sprintf("typed-model-request-%03d", index), Steps: []testutil.Exchange{compiledWire("PATCH", item, setCommand, body, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{}})
			return testutil.Observation{Response: r, Err: e}
		}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.
			TeamsAdmin, StatusCode: 400,
		}},
		})
	}
}
