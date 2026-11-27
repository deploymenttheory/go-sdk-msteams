package teamschannelspolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsChannelsPolicy"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "list-empty",
				Operation: "list",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy",
					CmdletName: "Get-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "create-all",
				Operation: "create",
				Input:     "{\"Description\":\"Future-Description\",\"AllowOrgWideTeamCreation\":false,\"EnablePrivateTeamDiscovery\":false,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":false,\"AllowChannelSharingToExternalUser\":false,\"AllowUserToParticipateInExternalSharedChannel\":false,\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\",\"AllowCreateChannel\":\"Future-AllowCreateChannel\",\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\",\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\",\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\",\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\",\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\",\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\",\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\",\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\",\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"Future-Description\",\n  \"AllowOrgWideTeamCreation\": false,\n  \"EnablePrivateTeamDiscovery\": false,\n  \"AllowPrivateChannelCreation\": false,\n  \"AllowSharedChannelCreation\": false,\n  \"AllowChannelSharingToExternalUser\": false,\n  \"AllowUserToParticipateInExternalSharedChannel\": false,\n  \"ThreadedChannelCreation\": \"Future-ThreadedChannelCreation\",\n  \"AllowCreateChannel\": \"Future-AllowCreateChannel\",\n  \"AllowCreateClassicChannel\": \"Future-AllowCreateClassicChannel\",\n  \"AllowCreatePrivateChannel\": \"Future-AllowCreatePrivateChannel\",\n  \"AllowCreateSharedChannel\": \"Future-AllowCreateSharedChannel\",\n  \"CreateSharedChannelsByDefault\": \"Future-CreateSharedChannelsByDefault\",\n  \"AllowUsersFromOutsideTeam\": \"Future-AllowUsersFromOutsideTeam\",\n  \"AllowGuestsFromOutsideTeam\": \"Future-AllowGuestsFromOutsideTeam\",\n  \"AllowUsersFromOutsideTeamInPrivateChannel\": \"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\n  \"AllowGuestsFromOutsideTeamInPrivateChannel\": \"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\n  \"AllowSharingWithTeamInOrg\": \"Future-AllowSharingWithTeamInOrg\",\n  \"AllowSharingPrivateChannelWithTeamInOrg\": \"Future-AllowSharingPrivateChannelWithTeamInOrg\",\n  \"DefaultChannelTypeOnCreation\": \"Future-DefaultChannelTypeOnCreation\"\n}",

					CmdletName: "New-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "create-empty",
				Operation: "create",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "New-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"Description\":\"Future-Description\",\"AllowOrgWideTeamCreation\":false,\"EnablePrivateTeamDiscovery\":false,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":false,\"AllowChannelSharingToExternalUser\":false,\"AllowUserToParticipateInExternalSharedChannel\":false,\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\",\"AllowCreateChannel\":\"Future-AllowCreateChannel\",\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\",\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\",\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\",\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\",\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\",\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\",\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\",\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\",\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"Future-Description\",\n  \"AllowOrgWideTeamCreation\": false,\n  \"EnablePrivateTeamDiscovery\": false,\n  \"AllowPrivateChannelCreation\": false,\n  \"AllowSharedChannelCreation\": false,\n  \"AllowChannelSharingToExternalUser\": false,\n  \"AllowUserToParticipateInExternalSharedChannel\": false,\n  \"ThreadedChannelCreation\": \"Future-ThreadedChannelCreation\",\n  \"AllowCreateChannel\": \"Future-AllowCreateChannel\",\n  \"AllowCreateClassicChannel\": \"Future-AllowCreateClassicChannel\",\n  \"AllowCreatePrivateChannel\": \"Future-AllowCreatePrivateChannel\",\n  \"AllowCreateSharedChannel\": \"Future-AllowCreateSharedChannel\",\n  \"CreateSharedChannelsByDefault\": \"Future-CreateSharedChannelsByDefault\",\n  \"AllowUsersFromOutsideTeam\": \"Future-AllowUsersFromOutsideTeam\",\n  \"AllowGuestsFromOutsideTeam\": \"Future-AllowGuestsFromOutsideTeam\",\n  \"AllowUsersFromOutsideTeamInPrivateChannel\": \"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\n  \"AllowGuestsFromOutsideTeamInPrivateChannel\": \"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\n  \"AllowSharingWithTeamInOrg\": \"Future-AllowSharingWithTeamInOrg\",\n  \"AllowSharingPrivateChannelWithTeamInOrg\": \"Future-AllowSharingPrivateChannelWithTeamInOrg\",\n  \"DefaultChannelTypeOnCreation\": \"Future-DefaultChannelTypeOnCreation\"\n}",

					CmdletName: "Set-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"Description\":null,\"AllowPrivateChannelCreation\":null,\"AllowCreatePrivateChannel\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": null,\n  \"AllowPrivateChannelCreation\": null,\n  \"AllowCreatePrivateChannel\": null\n}",

					CmdletName: "Set-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "patch-empty",
				Operation: "patch",
				Input:     "{\"Description\":\"\",\"AllowCreatePrivateChannel\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"AllowCreatePrivateChannel\": \"\"\n}",

					CmdletName: "Set-CsTeamsChannelsPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsChannelsPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsChannelsPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsChannelsPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsChannelsPolicy",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Input:     "{}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsChannelsPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsChannelsPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy",
					CmdletName: "Get-CsTeamsChannelsPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsChannelsPolicy",
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
		t.Fatalf("channels policy contract differs: %v %+v", e, r.Differences())
	}
	return r
}

const getCommand = "Get-CsTeamsChannelsPolicy"
const newCommand = "New-CsTeamsChannelsPolicy"
const setCommand = "Set-CsTeamsChannelsPolicy"
const deleteCommand = "Remove-CsTeamsChannelsPolicy"
const grantCommand = "Grant-CsTeamsChannelsPolicy"
const forest = "https://admin.example.invalid/"

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func modernWire(method, path string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {grantCommand}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil && string(body) != "null" {
		h.Set("Content-Type", "application/json")
	} else {
		body = nil
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
