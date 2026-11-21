// Package teamscallingpolicy manages the compiled Teams calling policy resource.
package teamscallingpolicy

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy retains nullable known fields and the complete received snapshot.
// Known JSON fields require native string, Boolean and signed Int64 types.
type Policy struct {
	Identity                                         *string                    `json:"Identity"`
	Description                                      *string                    `json:"Description"`
	AllowPrivateCalling                              *bool                      `json:"AllowPrivateCalling"`
	AllowWebPSTNCalling                              *bool                      `json:"AllowWebPSTNCalling"`
	AllowSIPDevicesCalling                           *bool                      `json:"AllowSIPDevicesCalling"`
	AllowVoicemail                                   *string                    `json:"AllowVoicemail"`
	AllowCallGroups                                  *bool                      `json:"AllowCallGroups"`
	AllowDelegation                                  *bool                      `json:"AllowDelegation"`
	AllowCallForwardingToUser                        *bool                      `json:"AllowCallForwardingToUser"`
	AllowCallForwardingToPhone                       *bool                      `json:"AllowCallForwardingToPhone"`
	PreventTollBypass                                *bool                      `json:"PreventTollBypass"`
	BusyOnBusyEnabledType                            *string                    `json:"BusyOnBusyEnabledType"`
	MusicOnHoldEnabledType                           *string                    `json:"MusicOnHoldEnabledType"`
	AllowCloudRecordingForCalls                      *bool                      `json:"AllowCloudRecordingForCalls"`
	ExplicitRecordingConsent                         *string                    `json:"ExplicitRecordingConsent"`
	PreventComplianceRecording                       *string                    `json:"PreventComplianceRecording"`
	EnableRecordingAndTranscriptionCustomMessage     *bool                      `json:"EnableRecordingAndTranscriptionCustomMessage"`
	RecordingAndTranscriptionCustomMessageIdentifier *string                    `json:"RecordingAndTranscriptionCustomMessageIdentifier"`
	AllowMeetingKnowledgeGeneration                  *bool                      `json:"AllowMeetingKnowledgeGeneration"`
	AllowTranscriptionForCalling                     *bool                      `json:"AllowTranscriptionForCalling"`
	RecordingAndTranscriptionAudioNotification       *string                    `json:"RecordingAndTranscriptionAudioNotification"`
	PopoutForIncomingPSTNCalls                       *string                    `json:"PopoutForIncomingPstnCalls"`
	PopoutAppPathForIncomingPSTNCalls                *string                    `json:"PopoutAppPathForIncomingPstnCalls"`
	LiveCaptionsEnabledTypeForCalling                *string                    `json:"LiveCaptionsEnabledTypeForCalling"`
	AutoAnswerEnabledType                            *string                    `json:"AutoAnswerEnabledType"`
	SpamFilteringEnabledType                         *string                    `json:"SpamFilteringEnabledType"`
	CallRecordingExpirationDays                      *int64                     `json:"CallRecordingExpirationDays"`
	AllowCallRedirect                                *string                    `json:"AllowCallRedirect"`
	InboundPSTNCallRoutingTreatment                  *string                    `json:"InboundPstnCallRoutingTreatment"`
	InboundFederatedCallRoutingTreatment             *string                    `json:"InboundFederatedCallRoutingTreatment"`
	EnableWebPSTNMediaBypass                         *bool                      `json:"EnableWebPstnMediaBypass"`
	EnableSpendLimits                                *bool                      `json:"EnableSpendLimits"`
	CallingSpendUserLimit                            *int64                     `json:"CallingSpendUserLimit"`
	Copilot                                          *string                    `json:"Copilot"`
	ShowTeamsCallsInCallLog                          *bool                      `json:"ShowTeamsCallsInCallLog"`
	RealTimeText                                     *string                    `json:"RealTimeText"`
	AIInterpreter                                    *string                    `json:"AIInterpreter"`
	VoiceSimulationInInterpreter                     *string                    `json:"VoiceSimulationInInterpreter"`
	ReportCall                                       *string                    `json:"ReportCall"`
	VoicePhishingDetection                           *string                    `json:"VoicePhishingDetection"`
	RawFields                                        map[string]json.RawMessage `json:"-"`
	RawJSON                                          json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	if validate(next.RecordingAndTranscriptionCustomMessageIdentifier, next.CallRecordingExpirationDays) != nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "calling policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false/zero.
// String choices other than the custom message identifier are server validated.
type Settings struct {
	Description                                      policy.Field[string] `json:"Description,omitzero"`
	AllowPrivateCalling                              policy.Field[bool]   `json:"AllowPrivateCalling,omitzero"`
	AllowWebPSTNCalling                              policy.Field[bool]   `json:"AllowWebPSTNCalling,omitzero"`
	AllowSIPDevicesCalling                           policy.Field[bool]   `json:"AllowSIPDevicesCalling,omitzero"`
	AllowVoicemail                                   policy.Field[string] `json:"AllowVoicemail,omitzero"`
	AllowCallGroups                                  policy.Field[bool]   `json:"AllowCallGroups,omitzero"`
	AllowDelegation                                  policy.Field[bool]   `json:"AllowDelegation,omitzero"`
	AllowCallForwardingToUser                        policy.Field[bool]   `json:"AllowCallForwardingToUser,omitzero"`
	AllowCallForwardingToPhone                       policy.Field[bool]   `json:"AllowCallForwardingToPhone,omitzero"`
	PreventTollBypass                                policy.Field[bool]   `json:"PreventTollBypass,omitzero"`
	BusyOnBusyEnabledType                            policy.Field[string] `json:"BusyOnBusyEnabledType,omitzero"`
	MusicOnHoldEnabledType                           policy.Field[string] `json:"MusicOnHoldEnabledType,omitzero"`
	AllowCloudRecordingForCalls                      policy.Field[bool]   `json:"AllowCloudRecordingForCalls,omitzero"`
	ExplicitRecordingConsent                         policy.Field[string] `json:"ExplicitRecordingConsent,omitzero"`
	PreventComplianceRecording                       policy.Field[string] `json:"PreventComplianceRecording,omitzero"`
	EnableRecordingAndTranscriptionCustomMessage     policy.Field[bool]   `json:"EnableRecordingAndTranscriptionCustomMessage,omitzero"`
	RecordingAndTranscriptionCustomMessageIdentifier policy.Field[string] `json:"RecordingAndTranscriptionCustomMessageIdentifier,omitzero"`
	AllowMeetingKnowledgeGeneration                  policy.Field[bool]   `json:"AllowMeetingKnowledgeGeneration,omitzero"`
	AllowTranscriptionForCalling                     policy.Field[bool]   `json:"AllowTranscriptionForCalling,omitzero"`
	RecordingAndTranscriptionAudioNotification       policy.Field[string] `json:"RecordingAndTranscriptionAudioNotification,omitzero"`
	PopoutForIncomingPSTNCalls                       policy.Field[string] `json:"PopoutForIncomingPstnCalls,omitzero"`
	PopoutAppPathForIncomingPSTNCalls                policy.Field[string] `json:"PopoutAppPathForIncomingPstnCalls,omitzero"`
	LiveCaptionsEnabledTypeForCalling                policy.Field[string] `json:"LiveCaptionsEnabledTypeForCalling,omitzero"`
	AutoAnswerEnabledType                            policy.Field[string] `json:"AutoAnswerEnabledType,omitzero"`
	SpamFilteringEnabledType                         policy.Field[string] `json:"SpamFilteringEnabledType,omitzero"`
	CallRecordingExpirationDays                      policy.Field[int64]  `json:"CallRecordingExpirationDays,omitzero"`
	AllowCallRedirect                                policy.Field[string] `json:"AllowCallRedirect,omitzero"`
	InboundPSTNCallRoutingTreatment                  policy.Field[string] `json:"InboundPstnCallRoutingTreatment,omitzero"`
	InboundFederatedCallRoutingTreatment             policy.Field[string] `json:"InboundFederatedCallRoutingTreatment,omitzero"`
	EnableWebPSTNMediaBypass                         policy.Field[bool]   `json:"EnableWebPstnMediaBypass,omitzero"`
	EnableSpendLimits                                policy.Field[bool]   `json:"EnableSpendLimits,omitzero"`
	CallingSpendUserLimit                            policy.Field[int64]  `json:"CallingSpendUserLimit,omitzero"`
	Copilot                                          policy.Field[string] `json:"Copilot,omitzero"`
	ShowTeamsCallsInCallLog                          policy.Field[bool]   `json:"ShowTeamsCallsInCallLog,omitzero"`
	RealTimeText                                     policy.Field[string] `json:"RealTimeText,omitzero"`
	AIInterpreter                                    policy.Field[string] `json:"AIInterpreter,omitzero"`
	VoiceSimulationInInterpreter                     policy.Field[string] `json:"VoiceSimulationInInterpreter,omitzero"`
	ReportCall                                       policy.Field[string] `json:"ReportCall,omitzero"`
	VoicePhishingDetection                           policy.Field[string] `json:"VoicePhishingDetection,omitzero"`
}
type CreateRequest struct {
	Settings
	policy.Controls
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "calling policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

var messageID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func validate(identifier *string, expiration *int64) error {
	// The source .NET dollar anchor accepts exactly one terminal LF.
	if identifier != nil && !messageID.MatchString(strings.TrimSuffix(*identifier, "\n")) {
		return ErrInput
	}
	if expiration != nil && (*expiration < 1 || *expiration > 9999) {
		return ErrInput
	}
	return nil
}
func writeFields(input Settings) (map[string]json.RawMessage, error) {
	var identifier *string
	if v, ok := input.RecordingAndTranscriptionCustomMessageIdentifier.Value(); ok {
		identifier = &v
	}
	var expiration *int64
	if v, ok := input.CallRecordingExpirationDays.Value(); ok {
		expiration = &v
	}
	if err := validate(identifier, expiration); err != nil {
		return nil, err
	}
	return policyrp.Fields(input)
}
