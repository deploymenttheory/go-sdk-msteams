// Package teamsmeetingpolicy manages compiled Teams meeting policy.
package teamsmeetingpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// Strings, Booleans and numeric settings retain source nullability.
type Policy struct {
	Identity                                         *string                    `json:"Identity"`
	Description                                      *string                    `json:"Description"`
	AllowChannelMeetingScheduling                    *bool                      `json:"AllowChannelMeetingScheduling"`
	AllowMeetNow                                     *bool                      `json:"AllowMeetNow"`
	AllowPrivateMeetNow                              *bool                      `json:"AllowPrivateMeetNow"`
	MeetingChatEnabledType                           *string                    `json:"MeetingChatEnabledType"`
	AllowExternalNonTrustedMeetingChat               *bool                      `json:"AllowExternalNonTrustedMeetingChat"`
	CopyRestriction                                  *bool                      `json:"CopyRestriction"`
	LiveCaptionsEnabledType                          *string                    `json:"LiveCaptionsEnabledType"`
	DesignatedPresenterRoleMode                      *string                    `json:"DesignatedPresenterRoleMode"`
	AllowIPAudio                                     *bool                      `json:"AllowIPAudio"`
	AllowIPVideo                                     *bool                      `json:"AllowIPVideo"`
	AllowEngagementReport                            *string                    `json:"AllowEngagementReport"`
	AllowTrackingInReport                            *string                    `json:"AllowTrackingInReport"`
	IPAudioMode                                      *string                    `json:"IPAudioMode"`
	IPVideoMode                                      *string                    `json:"IPVideoMode"`
	AllowAnonymousUsersToDialOut                     *bool                      `json:"AllowAnonymousUsersToDialOut"`
	AllowAnonymousUsersToStartMeeting                *bool                      `json:"AllowAnonymousUsersToStartMeeting"`
	AllowAnonymousUsersToJoinMeeting                 *bool                      `json:"AllowAnonymousUsersToJoinMeeting"`
	ConditionalAccessAttendeeVerification            *bool                      `json:"ConditionalAccessAttendeeVerification"`
	BlockedAnonymousJoinClientTypes                  *string                    `json:"BlockedAnonymousJoinClientTypes"`
	AllowedStreamingMediaInput                       *string                    `json:"AllowedStreamingMediaInput"`
	ExplicitRecordingConsent                         *string                    `json:"ExplicitRecordingConsent"`
	PreventComplianceRecording                       *string                    `json:"PreventComplianceRecording"`
	EnableRecordingAndTranscriptionCustomMessage     *bool                      `json:"EnableRecordingAndTranscriptionCustomMessage"`
	RecordingAndTranscriptionCustomMessageIdentifier *string                    `json:"RecordingAndTranscriptionCustomMessageIdentifier"`
	AllowLocalRecording                              *bool                      `json:"AllowLocalRecording"`
	AutoRecording                                    *string                    `json:"AutoRecording"`
	ParticipantNameChange                            *string                    `json:"ParticipantNameChange"`
	AllowPrivateMeetingScheduling                    *bool                      `json:"AllowPrivateMeetingScheduling"`
	AutoAdmittedUsers                                *string                    `json:"AutoAdmittedUsers"`
	AllowCloudRecording                              *bool                      `json:"AllowCloudRecording"`
	SetRecordingAndTranscriptOwnership               *string                    `json:"SetRecordingAndTranscriptOwnership"`
	RecordingAndTranscriptionAudioNotification       *string                    `json:"RecordingAndTranscriptionAudioNotification"`
	FilterProfanityInTranscript                      *string                    `json:"FilterProfanityInTranscript"`
	AllowRecordingStorageOutsideRegion               *bool                      `json:"AllowRecordingStorageOutsideRegion"`
	RecordingStorageMode                             *string                    `json:"RecordingStorageMode"`
	AllowMeetingKnowledgeGeneration                  *bool                      `json:"AllowMeetingKnowledgeGeneration"`
	KnowledgeGenerationMeetingTypes                  *string                    `json:"KnowledgeGenerationMeetingTypes"`
	MeetingKnowledgeExpirationDays                   *int64                     `json:"MeetingKnowledgeExpirationDays"`
	AllowOutlookAddIn                                *bool                      `json:"AllowOutlookAddIn"`
	AllowPowerPointSharing                           *bool                      `json:"AllowPowerPointSharing"`
	AllowParticipantGiveRequestControl               *bool                      `json:"AllowParticipantGiveRequestControl"`
	AllowExternalParticipantGiveRequestControl       *bool                      `json:"AllowExternalParticipantGiveRequestControl"`
	AllowSharedNotes                                 *bool                      `json:"AllowSharedNotes"`
	AllowWhiteboard                                  *bool                      `json:"AllowWhiteboard"`
	AllowTranscription                               *bool                      `json:"AllowTranscription"`
	AllowNetworkConfigurationSettingsLookup          *bool                      `json:"AllowNetworkConfigurationSettingsLookup"`
	MediaBitRateKb                                   *uint32                    `json:"MediaBitRateKb"`
	ScreenSharingMode                                *string                    `json:"ScreenSharingMode"`
	AllowMultipleScreenshare                         *bool                      `json:"AllowMultipleScreenshare"`
	VideoFiltersMode                                 *string                    `json:"VideoFiltersMode"`
	AllowPSTNUsersToBypassLobby                      *bool                      `json:"AllowPSTNUsersToBypassLobby"`
	AllowOrganizersToOverrideLobbySettings           *bool                      `json:"AllowOrganizersToOverrideLobbySettings"`
	PreferredMeetingProviderForIslandsMode           *string                    `json:"PreferredMeetingProviderForIslandsMode"`
	AllowNDIStreaming                                *bool                      `json:"AllowNDIStreaming"`
	SpeakerAttributionMode                           *string                    `json:"SpeakerAttributionMode"`
	EnrollUserOverride                               *string                    `json:"EnrollUserOverride"`
	RoomAttributeUserOverride                        *string                    `json:"RoomAttributeUserOverride"`
	StreamingAttendeeMode                            *string                    `json:"StreamingAttendeeMode"`
	AttendeeIdentityMasking                          *string                    `json:"AttendeeIdentityMasking"`
	AllowBreakoutRooms                               *bool                      `json:"AllowBreakoutRooms"`
	TeamsCameraFarEndPTZMode                         *string                    `json:"TeamsCameraFarEndPTZMode"`
	AllowMeetingReactions                            *bool                      `json:"AllowMeetingReactions"`
	AllowMeetingRegistration                         *bool                      `json:"AllowMeetingRegistration"`
	WhoCanRegister                                   *string                    `json:"WhoCanRegister"`
	AllowScreenContentDigitization                   *string                    `json:"AllowScreenContentDigitization"`
	AllowCarbonSummary                               *bool                      `json:"AllowCarbonSummary"`
	RoomPeopleNameUserOverride                       *string                    `json:"RoomPeopleNameUserOverride"`
	AllowMeetingCoach                                *bool                      `json:"AllowMeetingCoach"`
	NewMeetingRecordingExpirationDays                *int64                     `json:"NewMeetingRecordingExpirationDays"`
	LiveStreamingMode                                *string                    `json:"LiveStreamingMode"`
	MeetingInviteLanguages                           *string                    `json:"MeetingInviteLanguages"`
	ChannelRecordingDownload                         *string                    `json:"ChannelRecordingDownload"`
	AllowCartCaptionsScheduling                      *string                    `json:"AllowCartCaptionsScheduling"`
	AllowTasksFromTranscript                         *string                    `json:"AllowTasksFromTranscript"`
	InfoShownInReportMode                            *string                    `json:"InfoShownInReportMode"`
	LiveInterpretationEnabledType                    *string                    `json:"LiveInterpretationEnabledType"`
	QnAEngagementMode                                *string                    `json:"QnAEngagementMode"`
	AllowImmersiveView                               *bool                      `json:"AllowImmersiveView"`
	AllowAvatarsInGallery                            *bool                      `json:"AllowAvatarsInGallery"`
	AllowAnnotations                                 *bool                      `json:"AllowAnnotations"`
	AllowDocumentCollaboration                       *string                    `json:"AllowDocumentCollaboration"`
	AllowWatermarkForScreenSharing                   *bool                      `json:"AllowWatermarkForScreenSharing"`
	AllowWatermarkForCameraVideo                     *bool                      `json:"AllowWatermarkForCameraVideo"`
	AllowWatermarkCustomizationForCameraVideo        *bool                      `json:"AllowWatermarkCustomizationForCameraVideo"`
	WatermarkForCameraVideoOpacity                   *int64                     `json:"WatermarkForCameraVideoOpacity"`
	WatermarkForCameraVideoPattern                   *string                    `json:"WatermarkForCameraVideoPattern"`
	AllowWatermarkCustomizationForScreenSharing      *bool                      `json:"AllowWatermarkCustomizationForScreenSharing"`
	WatermarkForScreenSharingOpacity                 *int64                     `json:"WatermarkForScreenSharingOpacity"`
	WatermarkForScreenSharingPattern                 *string                    `json:"WatermarkForScreenSharingPattern"`
	WatermarkForAnonymousUsers                       *string                    `json:"WatermarkForAnonymousUsers"`
	DetectSensitiveContentDuringScreenSharing        *bool                      `json:"DetectSensitiveContentDuringScreenSharing"`
	AudibleRecordingNotification                     *string                    `json:"AudibleRecordingNotification"`
	ConnectToMeetingControls                         *string                    `json:"ConnectToMeetingControls"`
	Copilot                                          *string                    `json:"Copilot"`
	TranscriptionWithMeetingAI                       *string                    `json:"TranscriptionWithMeetingAI"`
	MeetingAI                                        *string                    `json:"MeetingAI"`
	AutomaticallyStartCopilot                        *string                    `json:"AutomaticallyStartCopilot"`
	VoiceIsolation                                   *string                    `json:"VoiceIsolation"`
	ExternalMeetingJoin                              *string                    `json:"ExternalMeetingJoin"`
	ContentSharingInExternalMeetings                 *string                    `json:"ContentSharingInExternalMeetings"`
	AllowedUsersForMeetingDetails                    *string                    `json:"AllowedUsersForMeetingDetails"`
	SmsNotifications                                 *string                    `json:"SmsNotifications"`
	CaptchaVerificationForMeetingJoin                *string                    `json:"CaptchaVerificationForMeetingJoin"`
	UsersCanAdmitFromLobby                           *string                    `json:"UsersCanAdmitFromLobby"`
	LobbyChat                                        *string                    `json:"LobbyChat"`
	BackroomChat                                     *string                    `json:"BackroomChat"`
	AnonymousUserAuthenticationMethod                *string                    `json:"AnonymousUserAuthenticationMethod"`
	NoiseSuppressionForDialInParticipants            *string                    `json:"NoiseSuppressionForDialInParticipants"`
	RealTimeText                                     *string                    `json:"RealTimeText"`
	AIInterpreter                                    *string                    `json:"AIInterpreter"`
	VoiceSimulationInInterpreter                     *string                    `json:"VoiceSimulationInInterpreter"`
	ParticipantSlideControl                          *string                    `json:"ParticipantSlideControl"`
	PasscodeComplexity                               *string                    `json:"PasscodeComplexity"`
	ExternalBotAccessMode                            *string                    `json:"ExternalBotAccessMode"`
	SyntheticMediaDetection                          *string                    `json:"SyntheticMediaDetection"`
	SyntheticMediaDetectionAppId                     *string                    `json:"SyntheticMediaDetectionAppId"`
	DisableAudioAnnouncementsForResourceAccounts     *bool                      `json:"DisableAudioAnnouncementsForResourceAccounts"`
	EnablePreMeetingConsent                          *bool                      `json:"EnablePreMeetingConsent"`
	PreMeetingConsentContentIdentifier               *string                    `json:"PreMeetingConsentContentIdentifier"`
	EnableExternalRecordingDetection                 *bool                      `json:"EnableExternalRecordingDetection"`
	IntelligentRecapDocxFileExpirationDays           *int64                     `json:"IntelligentRecapDocxFileExpirationDays"`
	AllowIntelligentRecap                            *bool                      `json:"AllowIntelligentRecap"`
	RawFields                                        map[string]json.RawMessage `json:"-"`
	RawJSON                                          json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
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
func (Policy) String() string     { return "meeting policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// Source strings remain opaque; only evidenced setter constraints are checked.
type Settings struct {
	Description                                      policy.Field[string] `json:"Description,omitzero"`
	AllowChannelMeetingScheduling                    policy.Field[bool]   `json:"AllowChannelMeetingScheduling,omitzero"`
	AllowMeetNow                                     policy.Field[bool]   `json:"AllowMeetNow,omitzero"`
	AllowPrivateMeetNow                              policy.Field[bool]   `json:"AllowPrivateMeetNow,omitzero"`
	MeetingChatEnabledType                           policy.Field[string] `json:"MeetingChatEnabledType,omitzero"`
	AllowExternalNonTrustedMeetingChat               policy.Field[bool]   `json:"AllowExternalNonTrustedMeetingChat,omitzero"`
	CopyRestriction                                  policy.Field[bool]   `json:"CopyRestriction,omitzero"`
	LiveCaptionsEnabledType                          policy.Field[string] `json:"LiveCaptionsEnabledType,omitzero"`
	DesignatedPresenterRoleMode                      policy.Field[string] `json:"DesignatedPresenterRoleMode,omitzero"`
	AllowIPAudio                                     policy.Field[bool]   `json:"AllowIPAudio,omitzero"`
	AllowIPVideo                                     policy.Field[bool]   `json:"AllowIPVideo,omitzero"`
	AllowEngagementReport                            policy.Field[string] `json:"AllowEngagementReport,omitzero"`
	AllowTrackingInReport                            policy.Field[string] `json:"AllowTrackingInReport,omitzero"`
	IPAudioMode                                      policy.Field[string] `json:"IPAudioMode,omitzero"`
	IPVideoMode                                      policy.Field[string] `json:"IPVideoMode,omitzero"`
	AllowAnonymousUsersToDialOut                     policy.Field[bool]   `json:"AllowAnonymousUsersToDialOut,omitzero"`
	AllowAnonymousUsersToStartMeeting                policy.Field[bool]   `json:"AllowAnonymousUsersToStartMeeting,omitzero"`
	AllowAnonymousUsersToJoinMeeting                 policy.Field[bool]   `json:"AllowAnonymousUsersToJoinMeeting,omitzero"`
	ConditionalAccessAttendeeVerification            policy.Field[bool]   `json:"ConditionalAccessAttendeeVerification,omitzero"`
	BlockedAnonymousJoinClientTypes                  policy.Field[string] `json:"BlockedAnonymousJoinClientTypes,omitzero"`
	AllowedStreamingMediaInput                       policy.Field[string] `json:"AllowedStreamingMediaInput,omitzero"`
	ExplicitRecordingConsent                         policy.Field[string] `json:"ExplicitRecordingConsent,omitzero"`
	PreventComplianceRecording                       policy.Field[string] `json:"PreventComplianceRecording,omitzero"`
	EnableRecordingAndTranscriptionCustomMessage     policy.Field[bool]   `json:"EnableRecordingAndTranscriptionCustomMessage,omitzero"`
	RecordingAndTranscriptionCustomMessageIdentifier policy.Field[string] `json:"RecordingAndTranscriptionCustomMessageIdentifier,omitzero"`
	AllowLocalRecording                              policy.Field[bool]   `json:"AllowLocalRecording,omitzero"`
	AutoRecording                                    policy.Field[string] `json:"AutoRecording,omitzero"`
	ParticipantNameChange                            policy.Field[string] `json:"ParticipantNameChange,omitzero"`
	AllowPrivateMeetingScheduling                    policy.Field[bool]   `json:"AllowPrivateMeetingScheduling,omitzero"`
	AutoAdmittedUsers                                policy.Field[string] `json:"AutoAdmittedUsers,omitzero"`
	AllowCloudRecording                              policy.Field[bool]   `json:"AllowCloudRecording,omitzero"`
	SetRecordingAndTranscriptOwnership               policy.Field[string] `json:"SetRecordingAndTranscriptOwnership,omitzero"`
	RecordingAndTranscriptionAudioNotification       policy.Field[string] `json:"RecordingAndTranscriptionAudioNotification,omitzero"`
	FilterProfanityInTranscript                      policy.Field[string] `json:"FilterProfanityInTranscript,omitzero"`
	AllowRecordingStorageOutsideRegion               policy.Field[bool]   `json:"AllowRecordingStorageOutsideRegion,omitzero"`
	RecordingStorageMode                             policy.Field[string] `json:"RecordingStorageMode,omitzero"`
	AllowMeetingKnowledgeGeneration                  policy.Field[bool]   `json:"AllowMeetingKnowledgeGeneration,omitzero"`
	KnowledgeGenerationMeetingTypes                  policy.Field[string] `json:"KnowledgeGenerationMeetingTypes,omitzero"`
	MeetingKnowledgeExpirationDays                   policy.Field[int64]  `json:"MeetingKnowledgeExpirationDays,omitzero"`
	AllowOutlookAddIn                                policy.Field[bool]   `json:"AllowOutlookAddIn,omitzero"`
	AllowPowerPointSharing                           policy.Field[bool]   `json:"AllowPowerPointSharing,omitzero"`
	AllowParticipantGiveRequestControl               policy.Field[bool]   `json:"AllowParticipantGiveRequestControl,omitzero"`
	AllowExternalParticipantGiveRequestControl       policy.Field[bool]   `json:"AllowExternalParticipantGiveRequestControl,omitzero"`
	AllowSharedNotes                                 policy.Field[bool]   `json:"AllowSharedNotes,omitzero"`
	AllowWhiteboard                                  policy.Field[bool]   `json:"AllowWhiteboard,omitzero"`
	AllowTranscription                               policy.Field[bool]   `json:"AllowTranscription,omitzero"`
	AllowNetworkConfigurationSettingsLookup          policy.Field[bool]   `json:"AllowNetworkConfigurationSettingsLookup,omitzero"`
	MediaBitRateKb                                   policy.Field[uint32] `json:"MediaBitRateKb,omitzero"`
	ScreenSharingMode                                policy.Field[string] `json:"ScreenSharingMode,omitzero"`
	AllowMultipleScreenshare                         policy.Field[bool]   `json:"AllowMultipleScreenshare,omitzero"`
	VideoFiltersMode                                 policy.Field[string] `json:"VideoFiltersMode,omitzero"`
	AllowPSTNUsersToBypassLobby                      policy.Field[bool]   `json:"AllowPSTNUsersToBypassLobby,omitzero"`
	AllowOrganizersToOverrideLobbySettings           policy.Field[bool]   `json:"AllowOrganizersToOverrideLobbySettings,omitzero"`
	PreferredMeetingProviderForIslandsMode           policy.Field[string] `json:"PreferredMeetingProviderForIslandsMode,omitzero"`
	AllowNDIStreaming                                policy.Field[bool]   `json:"AllowNDIStreaming,omitzero"`
	SpeakerAttributionMode                           policy.Field[string] `json:"SpeakerAttributionMode,omitzero"`
	EnrollUserOverride                               policy.Field[string] `json:"EnrollUserOverride,omitzero"`
	RoomAttributeUserOverride                        policy.Field[string] `json:"RoomAttributeUserOverride,omitzero"`
	StreamingAttendeeMode                            policy.Field[string] `json:"StreamingAttendeeMode,omitzero"`
	AttendeeIdentityMasking                          policy.Field[string] `json:"AttendeeIdentityMasking,omitzero"`
	AllowBreakoutRooms                               policy.Field[bool]   `json:"AllowBreakoutRooms,omitzero"`
	TeamsCameraFarEndPTZMode                         policy.Field[string] `json:"TeamsCameraFarEndPTZMode,omitzero"`
	AllowMeetingReactions                            policy.Field[bool]   `json:"AllowMeetingReactions,omitzero"`
	AllowMeetingRegistration                         policy.Field[bool]   `json:"AllowMeetingRegistration,omitzero"`
	WhoCanRegister                                   policy.Field[string] `json:"WhoCanRegister,omitzero"`
	AllowScreenContentDigitization                   policy.Field[string] `json:"AllowScreenContentDigitization,omitzero"`
	AllowCarbonSummary                               policy.Field[bool]   `json:"AllowCarbonSummary,omitzero"`
	RoomPeopleNameUserOverride                       policy.Field[string] `json:"RoomPeopleNameUserOverride,omitzero"`
	AllowMeetingCoach                                policy.Field[bool]   `json:"AllowMeetingCoach,omitzero"`
	NewMeetingRecordingExpirationDays                policy.Field[int64]  `json:"NewMeetingRecordingExpirationDays,omitzero"`
	LiveStreamingMode                                policy.Field[string] `json:"LiveStreamingMode,omitzero"`
	MeetingInviteLanguages                           policy.Field[string] `json:"MeetingInviteLanguages,omitzero"`
	ChannelRecordingDownload                         policy.Field[string] `json:"ChannelRecordingDownload,omitzero"`
	AllowCartCaptionsScheduling                      policy.Field[string] `json:"AllowCartCaptionsScheduling,omitzero"`
	AllowTasksFromTranscript                         policy.Field[string] `json:"AllowTasksFromTranscript,omitzero"`
	InfoShownInReportMode                            policy.Field[string] `json:"InfoShownInReportMode,omitzero"`
	LiveInterpretationEnabledType                    policy.Field[string] `json:"LiveInterpretationEnabledType,omitzero"`
	QnAEngagementMode                                policy.Field[string] `json:"QnAEngagementMode,omitzero"`
	AllowImmersiveView                               policy.Field[bool]   `json:"AllowImmersiveView,omitzero"`
	AllowAvatarsInGallery                            policy.Field[bool]   `json:"AllowAvatarsInGallery,omitzero"`
	AllowAnnotations                                 policy.Field[bool]   `json:"AllowAnnotations,omitzero"`
	AllowDocumentCollaboration                       policy.Field[string] `json:"AllowDocumentCollaboration,omitzero"`
	AllowWatermarkForScreenSharing                   policy.Field[bool]   `json:"AllowWatermarkForScreenSharing,omitzero"`
	AllowWatermarkForCameraVideo                     policy.Field[bool]   `json:"AllowWatermarkForCameraVideo,omitzero"`
	AllowWatermarkCustomizationForCameraVideo        policy.Field[bool]   `json:"AllowWatermarkCustomizationForCameraVideo,omitzero"`
	WatermarkForCameraVideoOpacity                   policy.Field[int64]  `json:"WatermarkForCameraVideoOpacity,omitzero"`
	WatermarkForCameraVideoPattern                   policy.Field[string] `json:"WatermarkForCameraVideoPattern,omitzero"`
	AllowWatermarkCustomizationForScreenSharing      policy.Field[bool]   `json:"AllowWatermarkCustomizationForScreenSharing,omitzero"`
	WatermarkForScreenSharingOpacity                 policy.Field[int64]  `json:"WatermarkForScreenSharingOpacity,omitzero"`
	WatermarkForScreenSharingPattern                 policy.Field[string] `json:"WatermarkForScreenSharingPattern,omitzero"`
	WatermarkForAnonymousUsers                       policy.Field[string] `json:"WatermarkForAnonymousUsers,omitzero"`
	DetectSensitiveContentDuringScreenSharing        policy.Field[bool]   `json:"DetectSensitiveContentDuringScreenSharing,omitzero"`
	AudibleRecordingNotification                     policy.Field[string] `json:"AudibleRecordingNotification,omitzero"`
	ConnectToMeetingControls                         policy.Field[string] `json:"ConnectToMeetingControls,omitzero"`
	Copilot                                          policy.Field[string] `json:"Copilot,omitzero"`
	TranscriptionWithMeetingAI                       policy.Field[string] `json:"TranscriptionWithMeetingAI,omitzero"`
	MeetingAI                                        policy.Field[string] `json:"MeetingAI,omitzero"`
	AutomaticallyStartCopilot                        policy.Field[string] `json:"AutomaticallyStartCopilot,omitzero"`
	VoiceIsolation                                   policy.Field[string] `json:"VoiceIsolation,omitzero"`
	ExternalMeetingJoin                              policy.Field[string] `json:"ExternalMeetingJoin,omitzero"`
	ContentSharingInExternalMeetings                 policy.Field[string] `json:"ContentSharingInExternalMeetings,omitzero"`
	AllowedUsersForMeetingDetails                    policy.Field[string] `json:"AllowedUsersForMeetingDetails,omitzero"`
	SmsNotifications                                 policy.Field[string] `json:"SmsNotifications,omitzero"`
	CaptchaVerificationForMeetingJoin                policy.Field[string] `json:"CaptchaVerificationForMeetingJoin,omitzero"`
	UsersCanAdmitFromLobby                           policy.Field[string] `json:"UsersCanAdmitFromLobby,omitzero"`
	LobbyChat                                        policy.Field[string] `json:"LobbyChat,omitzero"`
	BackroomChat                                     policy.Field[string] `json:"BackroomChat,omitzero"`
	AnonymousUserAuthenticationMethod                policy.Field[string] `json:"AnonymousUserAuthenticationMethod,omitzero"`
	NoiseSuppressionForDialInParticipants            policy.Field[string] `json:"NoiseSuppressionForDialInParticipants,omitzero"`
	RealTimeText                                     policy.Field[string] `json:"RealTimeText,omitzero"`
	AIInterpreter                                    policy.Field[string] `json:"AIInterpreter,omitzero"`
	VoiceSimulationInInterpreter                     policy.Field[string] `json:"VoiceSimulationInInterpreter,omitzero"`
	ParticipantSlideControl                          policy.Field[string] `json:"ParticipantSlideControl,omitzero"`
	PasscodeComplexity                               policy.Field[string] `json:"PasscodeComplexity,omitzero"`
	ExternalBotAccessMode                            policy.Field[string] `json:"ExternalBotAccessMode,omitzero"`
	SyntheticMediaDetection                          policy.Field[string] `json:"SyntheticMediaDetection,omitzero"`
	SyntheticMediaDetectionAppId                     policy.Field[string] `json:"SyntheticMediaDetectionAppId,omitzero"`
	DisableAudioAnnouncementsForResourceAccounts     policy.Field[bool]   `json:"DisableAudioAnnouncementsForResourceAccounts,omitzero"`
	EnablePreMeetingConsent                          policy.Field[bool]   `json:"EnablePreMeetingConsent,omitzero"`
	PreMeetingConsentContentIdentifier               policy.Field[string] `json:"PreMeetingConsentContentIdentifier,omitzero"`
	EnableExternalRecordingDetection                 policy.Field[bool]   `json:"EnableExternalRecordingDetection,omitzero"`
	IntelligentRecapDocxFileExpirationDays           policy.Field[int64]  `json:"IntelligentRecapDocxFileExpirationDays,omitzero"`
	AllowIntelligentRecap                            policy.Field[bool]   `json:"AllowIntelligentRecap,omitzero"`
}

// CreateRequest permits no settings, matching the compiled New processing.
type CreateRequest struct {
	Settings
	policy.Controls
}

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "meeting policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "meeting policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
