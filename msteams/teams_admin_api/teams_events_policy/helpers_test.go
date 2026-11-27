package teamseventspolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsEventsPolicy"
const item = collection + "/configuration/Global"
const setCommand = "Set-CsTeamsEventsPolicy"
const scoped = "scope /Fixture / Child"

type inputVectors struct {
	Name, Operation, Input, Response string
	Status                           int
	ForceMode                        int     `json:"force_mode"`
	PolicyName                       *string `json:"policy_name"`
	Rank                             int32   `json:"rank"`
	Request                          struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
		Sends                         int
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
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "list-empty",
				Operation:  "list",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "create-all",
				Operation: "create",
				Input:     "{\"AllowWebinars\":\"Future-AllowWebinars\",\"EventAccessType\":\"Future-EventAccessType\",\"AllowTownhalls\":\"Future-AllowTownhalls\",\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\",\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\",\"AllowEmailEditing\":\"Future-AllowEmailEditing\",\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\",\"AllowEventIntegrations\":false,\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\",\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\",\"RecordingForTownhall\":\"Future-RecordingForTownhall\",\"RecordingForWebinar\":\"Future-RecordingForWebinar\",\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\",\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\",\"TownhallChatExperience\":\"Future-TownhallChatExperience\",\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\",\"UseMicrosoftECDN\":false,\"ImmersiveEvents\":\"Future-ImmersiveEvents\",\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\",\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\",\"BackroomChat\":\"Future-BackroomChat\",\"Registration\":\"Future-Registration\",\"AllowEngagementReport\":\"Future-AllowEngagementReport\",\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\",\"Description\":\"Future-Description\"}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"AllowWebinars\": \"Future-AllowWebinars\",\n  \"EventAccessType\": \"Future-EventAccessType\",\n  \"AllowTownhalls\": \"Future-AllowTownhalls\",\n  \"TownhallEventAttendeeAccess\": \"Future-TownhallEventAttendeeAccess\",\n  \"ExternalPresenterJoinVerification\": \"Future-ExternalPresenterJoinVerification\",\n  \"AllowEmailEditing\": \"Future-AllowEmailEditing\",\n  \"AllowedQuestionTypesInRegistrationForm\": \"Future-AllowedQuestionTypesInRegistrationForm\",\n  \"AllowEventIntegrations\": false,\n  \"AllowedWebinarTypesForRecordingPublish\": \"Future-AllowedWebinarTypesForRecordingPublish\",\n  \"AllowedTownhallTypesForRecordingPublish\": \"Future-AllowedTownhallTypesForRecordingPublish\",\n  \"RecordingForTownhall\": \"Future-RecordingForTownhall\",\n  \"RecordingForWebinar\": \"Future-RecordingForWebinar\",\n  \"TranscriptionForTownhall\": \"Future-TranscriptionForTownhall\",\n  \"TranscriptionForWebinar\": \"Future-TranscriptionForWebinar\",\n  \"TownhallChatExperience\": \"Future-TownhallChatExperience\",\n  \"BroadcastPremiumApps\": \"Future-BroadcastPremiumApps\",\n  \"UseMicrosoftECDN\": false,\n  \"ImmersiveEvents\": \"Future-ImmersiveEvents\",\n  \"TownhallMaxResolution\": \"Future-TownhallMaxResolution\",\n  \"HighBitrateForTownhall\": \"Future-HighBitrateForTownhall\",\n  \"BackroomChat\": \"Future-BackroomChat\",\n  \"Registration\": \"Future-Registration\",\n  \"AllowEngagementReport\": \"Future-AllowEngagementReport\",\n  \"InfoShownInReportMode\": \"Future-InfoShownInReportMode\",\n  \"Description\": \"Future-Description\"\n}",

					CmdletName: "New-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "create-empty",
				Operation:  "create",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "New-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"AllowWebinars\":\"Future-AllowWebinars\",\"EventAccessType\":\"Future-EventAccessType\",\"AllowTownhalls\":\"Future-AllowTownhalls\",\"TownhallEventAttendeeAccess\":\"Future-TownhallEventAttendeeAccess\",\"ExternalPresenterJoinVerification\":\"Future-ExternalPresenterJoinVerification\",\"AllowEmailEditing\":\"Future-AllowEmailEditing\",\"AllowedQuestionTypesInRegistrationForm\":\"Future-AllowedQuestionTypesInRegistrationForm\",\"AllowEventIntegrations\":false,\"AllowedWebinarTypesForRecordingPublish\":\"Future-AllowedWebinarTypesForRecordingPublish\",\"AllowedTownhallTypesForRecordingPublish\":\"Future-AllowedTownhallTypesForRecordingPublish\",\"RecordingForTownhall\":\"Future-RecordingForTownhall\",\"RecordingForWebinar\":\"Future-RecordingForWebinar\",\"TranscriptionForTownhall\":\"Future-TranscriptionForTownhall\",\"TranscriptionForWebinar\":\"Future-TranscriptionForWebinar\",\"TownhallChatExperience\":\"Future-TownhallChatExperience\",\"BroadcastPremiumApps\":\"Future-BroadcastPremiumApps\",\"UseMicrosoftECDN\":false,\"ImmersiveEvents\":\"Future-ImmersiveEvents\",\"TownhallMaxResolution\":\"Future-TownhallMaxResolution\",\"HighBitrateForTownhall\":\"Future-HighBitrateForTownhall\",\"BackroomChat\":\"Future-BackroomChat\",\"Registration\":\"Future-Registration\",\"AllowEngagementReport\":\"Future-AllowEngagementReport\",\"InfoShownInReportMode\":\"Future-InfoShownInReportMode\",\"Description\":\"Future-Description\"}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowWebinars\": \"Future-AllowWebinars\",\n  \"EventAccessType\": \"Future-EventAccessType\",\n  \"AllowTownhalls\": \"Future-AllowTownhalls\",\n  \"TownhallEventAttendeeAccess\": \"Future-TownhallEventAttendeeAccess\",\n  \"ExternalPresenterJoinVerification\": \"Future-ExternalPresenterJoinVerification\",\n  \"AllowEmailEditing\": \"Future-AllowEmailEditing\",\n  \"AllowedQuestionTypesInRegistrationForm\": \"Future-AllowedQuestionTypesInRegistrationForm\",\n  \"AllowEventIntegrations\": false,\n  \"AllowedWebinarTypesForRecordingPublish\": \"Future-AllowedWebinarTypesForRecordingPublish\",\n  \"AllowedTownhallTypesForRecordingPublish\": \"Future-AllowedTownhallTypesForRecordingPublish\",\n  \"RecordingForTownhall\": \"Future-RecordingForTownhall\",\n  \"RecordingForWebinar\": \"Future-RecordingForWebinar\",\n  \"TranscriptionForTownhall\": \"Future-TranscriptionForTownhall\",\n  \"TranscriptionForWebinar\": \"Future-TranscriptionForWebinar\",\n  \"TownhallChatExperience\": \"Future-TownhallChatExperience\",\n  \"BroadcastPremiumApps\": \"Future-BroadcastPremiumApps\",\n  \"UseMicrosoftECDN\": false,\n  \"ImmersiveEvents\": \"Future-ImmersiveEvents\",\n  \"TownhallMaxResolution\": \"Future-TownhallMaxResolution\",\n  \"HighBitrateForTownhall\": \"Future-HighBitrateForTownhall\",\n  \"BackroomChat\": \"Future-BackroomChat\",\n  \"Registration\": \"Future-Registration\",\n  \"AllowEngagementReport\": \"Future-AllowEngagementReport\",\n  \"InfoShownInReportMode\": \"Future-InfoShownInReportMode\",\n  \"Description\": \"Future-Description\"\n}",

					CmdletName: "Set-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"AllowWebinars\":null,\"EventAccessType\":null,\"AllowTownhalls\":null,\"TownhallEventAttendeeAccess\":null,\"ExternalPresenterJoinVerification\":null,\"AllowEmailEditing\":null,\"AllowedQuestionTypesInRegistrationForm\":null,\"AllowEventIntegrations\":null,\"AllowedWebinarTypesForRecordingPublish\":null,\"AllowedTownhallTypesForRecordingPublish\":null,\"RecordingForTownhall\":null,\"RecordingForWebinar\":null,\"TranscriptionForTownhall\":null,\"TranscriptionForWebinar\":null,\"TownhallChatExperience\":null,\"BroadcastPremiumApps\":null,\"UseMicrosoftECDN\":null,\"ImmersiveEvents\":null,\"TownhallMaxResolution\":null,\"HighBitrateForTownhall\":null,\"BackroomChat\":null,\"Registration\":null,\"AllowEngagementReport\":null,\"InfoShownInReportMode\":null,\"Description\":null}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowWebinars\": null,\n  \"EventAccessType\": null,\n  \"AllowTownhalls\": null,\n  \"TownhallEventAttendeeAccess\": null,\n  \"ExternalPresenterJoinVerification\": null,\n  \"AllowEmailEditing\": null,\n  \"AllowedQuestionTypesInRegistrationForm\": null,\n  \"AllowEventIntegrations\": null,\n  \"AllowedWebinarTypesForRecordingPublish\": null,\n  \"AllowedTownhallTypesForRecordingPublish\": null,\n  \"RecordingForTownhall\": null,\n  \"RecordingForWebinar\": null,\n  \"TranscriptionForTownhall\": null,\n  \"TranscriptionForWebinar\": null,\n  \"TownhallChatExperience\": null,\n  \"BroadcastPremiumApps\": null,\n  \"UseMicrosoftECDN\": null,\n  \"ImmersiveEvents\": null,\n  \"TownhallMaxResolution\": null,\n  \"HighBitrateForTownhall\": null,\n  \"BackroomChat\": null,\n  \"Registration\": null,\n  \"AllowEngagementReport\": null,\n  \"InfoShownInReportMode\": null,\n  \"Description\": null\n}",

					CmdletName: "Set-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "patch-empty-true",
				Operation: "patch",
				Input:     "{\"AllowWebinars\":\"\",\"EventAccessType\":\"\",\"AllowTownhalls\":\"\",\"TownhallEventAttendeeAccess\":\"\",\"ExternalPresenterJoinVerification\":\"\",\"AllowEmailEditing\":\"\",\"AllowedQuestionTypesInRegistrationForm\":\"\",\"AllowEventIntegrations\":true,\"AllowedWebinarTypesForRecordingPublish\":\"\",\"AllowedTownhallTypesForRecordingPublish\":\"\",\"RecordingForTownhall\":\"\",\"RecordingForWebinar\":\"\",\"TranscriptionForTownhall\":\"\",\"TranscriptionForWebinar\":\"\",\"TownhallChatExperience\":\"\",\"BroadcastPremiumApps\":\"\",\"UseMicrosoftECDN\":true,\"ImmersiveEvents\":\"\",\"TownhallMaxResolution\":\"\",\"HighBitrateForTownhall\":\"\",\"BackroomChat\":\"\",\"Registration\":\"\",\"AllowEngagementReport\":\"\",\"InfoShownInReportMode\":\"\",\"Description\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowWebinars\": \"\",\n  \"EventAccessType\": \"\",\n  \"AllowTownhalls\": \"\",\n  \"TownhallEventAttendeeAccess\": \"\",\n  \"ExternalPresenterJoinVerification\": \"\",\n  \"AllowEmailEditing\": \"\",\n  \"AllowedQuestionTypesInRegistrationForm\": \"\",\n  \"AllowEventIntegrations\": true,\n  \"AllowedWebinarTypesForRecordingPublish\": \"\",\n  \"AllowedTownhallTypesForRecordingPublish\": \"\",\n  \"RecordingForTownhall\": \"\",\n  \"RecordingForWebinar\": \"\",\n  \"TranscriptionForTownhall\": \"\",\n  \"TranscriptionForWebinar\": \"\",\n  \"TownhallChatExperience\": \"\",\n  \"BroadcastPremiumApps\": \"\",\n  \"UseMicrosoftECDN\": true,\n  \"ImmersiveEvents\": \"\",\n  \"TownhallMaxResolution\": \"\",\n  \"HighBitrateForTownhall\": \"\",\n  \"BackroomChat\": \"\",\n  \"Registration\": \"\",\n  \"AllowEngagementReport\": \"\",\n  \"InfoShownInReportMode\": \"\",\n  \"Description\": \"\"\n}",

					CmdletName: "Set-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"Description\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"Description\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "delete-force-false",
				Operation: "delete",
				Input:     "{}",
				Status:    400,
				ForceMode: 0,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "read-empty",
				Operation:  "read",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "read-null",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEventsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-user-null",
				Operation: "user",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-tenant-null",
				Operation: "tenant",
				Input:     "{}",
				Status:    400,
				ForceMode: 0,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsEventsPolicy",
					Body:   "{\n  \"PolicyName\": null,\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-group-null",
				Operation: "group",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "grant-user-empty",
				Operation:  "user",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "grant-tenant-empty",
				Operation:  "tenant",
				Input:      "{}",
				Status:     400,
				ForceMode:  1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsEventsPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": true\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:       "grant-group-empty",
				Operation:  "group",
				Input:      "{}",
				Status:     400,
				ForceMode:  -1,
				PolicyName: new(""),
				Rank:       -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-group-rank-zero",
				Operation: "group",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": 0\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-group-rank-two",
				Operation: "group",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      2,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": 2\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
			{
				Name:      "grant-group-accepted",
				Operation: "group",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Rank:      2,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
					Sends                         int
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": 2\n}",

					CmdletName: "Grant-CsTeamsEventsPolicy",
				},
			},
		},
	}

	return v.Cases
}
func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("events policy contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}
