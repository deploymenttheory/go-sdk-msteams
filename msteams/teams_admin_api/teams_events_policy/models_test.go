package teamseventspolicy_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

func TestUnitEventsPolicyModelsMatchTypedSetters(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}{
		Models: []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "all-settings",
				Input: "{\"AllowWebinars\":\"Future-AllowWebinars\",\"EventAccessType\":\"Future-EventAccessType\",\"AllowTownhalls\":\"Future-AllowTownhalls\",\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\",\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\",\"AllowEmailEditing\":\"Future-AllowEmailEditing\",\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\",\"AllowEventIntegrations\":false,\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\",\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\",\"RecordingForTownhall\":\"Future-RecordingForTownhall\",\"RecordingForWebinar\":\"Future-RecordingForWebinar\",\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\",\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\",\"TownhallChatExperience\":\"Future-TownhallChatExperience\",\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\",\"UseMicrosoftECDN\":false,\"ImmersiveEvents\":\"Future-ImmersiveEvents\",\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\",\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\",\"BackroomChat\":\"Future-BackroomChat\",\"Registration\":\"Future-Registration\",\"AllowEngagementReport\":\"Future-AllowEngagementReport\",\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\",\"Description\":\"Future-Description\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":\"Future-AllowEmailEditing\",\"AllowEngagementReport\":\"Future-AllowEngagementReport\",\"AllowEventIntegrations\":false,\"AllowTownhalls\":\"Future-AllowTownhalls\",\"AllowWebinars\":\"Future-AllowWebinars\",\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\",\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\",\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\",\"BackroomChat\":\"Future-BackroomChat\",\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\",\"Description\":\"Future-Description\",\"EventAccessType\":\"Future-EventAccessType\",\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\",\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\",\"Identity\":null,\"ImmersiveEvents\":\"Future-ImmersiveEvents\",\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\",\"RecordingForTownhall\":\"Future-RecordingForTownhall\",\"RecordingForWebinar\":\"Future-RecordingForWebinar\",\"Registration\":\"Future-Registration\",\"TownhallChatExperience\":\"Future-TownhallChatExperience\",\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\",\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\",\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\",\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\",\"UseMicrosoftECDN\":false}"),
				},
			},

			{
				Name:  "AllowWebinars-null",
				Input: "{\"AllowWebinars\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowWebinars-empty",
				Input: "{\"AllowWebinars\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":\"\",\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowWebinars-future",
				Input: "{\"AllowWebinars\":\"Future-AllowWebinars\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":\"Future-AllowWebinars\",\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "EventAccessType-null",
				Input: "{\"EventAccessType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "EventAccessType-empty",
				Input: "{\"EventAccessType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":\"\",\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "EventAccessType-future",
				Input: "{\"EventAccessType\":\"Future-EventAccessType\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":\"Future-EventAccessType\",\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowTownhalls-null",
				Input: "{\"AllowTownhalls\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowTownhalls-empty",
				Input: "{\"AllowTownhalls\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":\"\",\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowTownhalls-future",
				Input: "{\"AllowTownhalls\":\"Future-AllowTownhalls\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":\"Future-AllowTownhalls\",\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallEventAttendeeAccess-null",
				Input: "{\"TownhallEventAttendeeAccess\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallEventAttendeeAccess-empty",
				Input: "{\"TownhallEventAttendeeAccess\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":\"\",\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallEventAttendeeAccess-future",
				Input: "{\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\",\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "ExternalPresenterJoinVerification-null",
				Input: "{\"ExternalPresenterJoinVerification\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "ExternalPresenterJoinVerification-empty",
				Input: "{\"ExternalPresenterJoinVerification\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":\"\",\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "ExternalPresenterJoinVerification-future",
				Input: "{\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\",\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEmailEditing-null",
				Input: "{\"AllowEmailEditing\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEmailEditing-empty",
				Input: "{\"AllowEmailEditing\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":\"\",\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEmailEditing-future",
				Input: "{\"AllowEmailEditing\":\"Future-AllowEmailEditing\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":\"Future-AllowEmailEditing\",\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedQuestionTypesInRegistrationForm-null",
				Input: "{\"AllowedQuestionTypesInRegistrationForm\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedQuestionTypesInRegistrationForm-empty",
				Input: "{\"AllowedQuestionTypesInRegistrationForm\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":\"\",\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedQuestionTypesInRegistrationForm-future",
				Input: "{\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\",\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEventIntegrations-null",
				Input: "{\"AllowEventIntegrations\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEventIntegrations-false",
				Input: "{\"AllowEventIntegrations\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":false,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEventIntegrations-true",
				Input: "{\"AllowEventIntegrations\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":true,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedWebinarTypesForRecordingPublish-null",
				Input: "{\"AllowedWebinarTypesForRecordingPublish\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedWebinarTypesForRecordingPublish-empty",
				Input: "{\"AllowedWebinarTypesForRecordingPublish\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":\"\",\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedWebinarTypesForRecordingPublish-future",
				Input: "{\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\",\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedTownhallTypesForRecordingPublish-null",
				Input: "{\"AllowedTownhallTypesForRecordingPublish\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedTownhallTypesForRecordingPublish-empty",
				Input: "{\"AllowedTownhallTypesForRecordingPublish\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":\"\",\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowedTownhallTypesForRecordingPublish-future",
				Input: "{\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\",\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForTownhall-null",
				Input: "{\"RecordingForTownhall\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForTownhall-empty",
				Input: "{\"RecordingForTownhall\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":\"\",\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForTownhall-future",
				Input: "{\"RecordingForTownhall\":\"Future-RecordingForTownhall\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":\"Future-RecordingForTownhall\",\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForWebinar-null",
				Input: "{\"RecordingForWebinar\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForWebinar-empty",
				Input: "{\"RecordingForWebinar\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":\"\",\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "RecordingForWebinar-future",
				Input: "{\"RecordingForWebinar\":\"Future-RecordingForWebinar\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":\"Future-RecordingForWebinar\",\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForTownhall-null",
				Input: "{\"TranscriptionForTownhall\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForTownhall-empty",
				Input: "{\"TranscriptionForTownhall\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":\"\",\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForTownhall-future",
				Input: "{\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\",\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForWebinar-null",
				Input: "{\"TranscriptionForWebinar\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForWebinar-empty",
				Input: "{\"TranscriptionForWebinar\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":\"\",\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TranscriptionForWebinar-future",
				Input: "{\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\",\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallChatExperience-null",
				Input: "{\"TownhallChatExperience\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallChatExperience-empty",
				Input: "{\"TownhallChatExperience\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":\"\",\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallChatExperience-future",
				Input: "{\"TownhallChatExperience\":\"Future-TownhallChatExperience\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":\"Future-TownhallChatExperience\",\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BroadcastPremiumApps-null",
				Input: "{\"BroadcastPremiumApps\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BroadcastPremiumApps-empty",
				Input: "{\"BroadcastPremiumApps\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":\"\",\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BroadcastPremiumApps-future",
				Input: "{\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\",\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "UseMicrosoftECDN-null",
				Input: "{\"UseMicrosoftECDN\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "UseMicrosoftECDN-false",
				Input: "{\"UseMicrosoftECDN\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":false}"),
				},
			},

			{
				Name:  "UseMicrosoftECDN-true",
				Input: "{\"UseMicrosoftECDN\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":true}"),
				},
			},

			{
				Name:  "ImmersiveEvents-null",
				Input: "{\"ImmersiveEvents\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "ImmersiveEvents-empty",
				Input: "{\"ImmersiveEvents\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":\"\",\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "ImmersiveEvents-future",
				Input: "{\"ImmersiveEvents\":\"Future-ImmersiveEvents\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":\"Future-ImmersiveEvents\",\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallMaxResolution-null",
				Input: "{\"TownhallMaxResolution\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallMaxResolution-empty",
				Input: "{\"TownhallMaxResolution\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":\"\",\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "TownhallMaxResolution-future",
				Input: "{\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\",\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "HighBitrateForTownhall-null",
				Input: "{\"HighBitrateForTownhall\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "HighBitrateForTownhall-empty",
				Input: "{\"HighBitrateForTownhall\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":\"\",\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "HighBitrateForTownhall-future",
				Input: "{\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\",\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BackroomChat-null",
				Input: "{\"BackroomChat\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BackroomChat-empty",
				Input: "{\"BackroomChat\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":\"\",\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "BackroomChat-future",
				Input: "{\"BackroomChat\":\"Future-BackroomChat\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":\"Future-BackroomChat\",\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Registration-null",
				Input: "{\"Registration\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Registration-empty",
				Input: "{\"Registration\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":\"\",\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Registration-future",
				Input: "{\"Registration\":\"Future-Registration\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":\"Future-Registration\",\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEngagementReport-null",
				Input: "{\"AllowEngagementReport\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEngagementReport-empty",
				Input: "{\"AllowEngagementReport\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":\"\",\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "AllowEngagementReport-future",
				Input: "{\"AllowEngagementReport\":\"Future-AllowEngagementReport\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":\"Future-AllowEngagementReport\",\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "InfoShownInReportMode-null",
				Input: "{\"InfoShownInReportMode\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "InfoShownInReportMode-empty",
				Input: "{\"InfoShownInReportMode\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":\"\",\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "InfoShownInReportMode-future",
				Input: "{\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\",\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Description-null",
				Input: "{\"Description\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Description-empty",
				Input: "{\"Description\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":\"\",\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "Description-future",
				Input: "{\"Description\":\"Future-Description\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":\"Future-Description\",\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "long-description",
				Input: "{\"Description\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Ide" +
						"ntity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"Recordi" +
						"ngForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":n" +
						"ull,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"Use" +
						"MicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "case-duplicate",
				Input: "{\"Description\":\"first\",\"description\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":\"second\",\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "duplicate-null",
				Input: "{\"UseMicrosoftECDN\":true,\"usemicrosoftecdn\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailEditing\":null,\"AllowEngagementReport\":null,\"AllowEventIntegrations\":null,\"AllowTownhalls\":null,\"AllowWebinars\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"BackroomChat\":null,\"BroadcastPremiumApps\":null,\"Description\":null,\"EventAccessType\":null,\"ExternalPresenterJoinVerification\":null,\"HighBitrateForTownhall\":null,\"Identity\":null,\"ImmersiveEvents\":null,\"InfoShownInReportMode\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"Registration\":null,\"TownhallChatExperience\":null,\"TownhallEventAttendeeAccess\":null,\"TownhallMaxResolution\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"UseMicrosoftECDN\":null}"),
				},
			},

			{
				Name:  "invalid-boolean-object",
				Input: "{\"AllowEventIntegrations\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},

			{
				Name:  "invalid-string-array",
				Input: "{\"AllowWebinars\":[]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var value resource.Policy
			var e error
			if c.Input != "constructor" {
				e = json.Unmarshal([]byte(c.Input), &value)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatal("source setter failure accepted")
				}
				return
			}
			if e != nil {
				t.Fatal("source-compatible model rejected")
			}
			type projection resource.Policy
			encoded, e := json.Marshal(projection(value))
			var got, want map[string]any
			if e != nil || json.Unmarshal(encoded, &got) != nil || json.Unmarshal(c.Observation.Value, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("source typed projection differs")
			}
		})
	}
}

func TestUnitEventsPolicySnapshotsOwnFutureData(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","UseMicrosoftECDN":false,"AllowEventIntegrations":null,"Description":"","Future":9007199254740993}`)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.UseMicrosoftECDN == nil || *value.UseMicrosoftECDN || value.AllowEventIntegrations != nil || value.Description == nil || *value.Description != "" {
		t.Fatal("nullable false/empty projection lost")
	}
	clear(raw)
	if string(value.RawFields["Future"]) != "9007199254740993" {
		t.Fatal("future snapshot lost")
	}
	value.RawFields["Future"][0] = 'x'
	encoded, e := json.Marshal(value)
	if e != nil || !json.Valid(encoded) {
		t.Fatal("original snapshot borrowed projected map storage")
	}
}
