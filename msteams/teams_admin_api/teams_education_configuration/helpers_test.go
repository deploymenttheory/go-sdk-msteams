package teamseducationconfiguration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/teams_education_configuration/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsEducationConfiguration"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration",
					CmdletName: "Get-CsTeamsEducationConfiguration",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"ParentGuardianPreferredContactMethod\":\"FutureContact\",\"EduGenerativeAIEnhancements\":\"FutureEnhancement\",\"UpdateParentInformation\":\"FutureUpdate\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentGuardianPreferredContactMethod\": \"FutureContact\",\n  \"EduGenerativeAIEnhancements\": \"FutureEnhancement\",\n  \"UpdateParentInformation\": \"FutureUpdate\"\n}",

					CmdletName: "Set-CsTeamsEducationConfiguration",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"ParentGuardianPreferredContactMethod\":null,\"EduGenerativeAIEnhancements\":null,\"UpdateParentInformation\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentGuardianPreferredContactMethod\": null,\n  \"EduGenerativeAIEnhancements\": null,\n  \"UpdateParentInformation\": null\n}",

					CmdletName: "Set-CsTeamsEducationConfiguration",
				},
			},
			{
				Name:      "patch-empty",
				Operation: "patch",
				Input:     "{\"ParentGuardianPreferredContactMethod\":\"\",\"EduGenerativeAIEnhancements\":\"\",\"UpdateParentInformation\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentGuardianPreferredContactMethod\": \"\",\n  \"EduGenerativeAIEnhancements\": \"\",\n  \"UpdateParentInformation\": \"\"\n}",

					CmdletName: "Set-CsTeamsEducationConfiguration",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"UpdateParentInformation\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UpdateParentInformation\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsEducationConfiguration",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"UpdateParentInformation\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UpdateParentInformation\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsEducationConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration",
					CmdletName: "Get-CsTeamsEducationConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationConfiguration",
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
	c.Family = "teams_education_configuration"
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
		t.Fatalf("education configuration contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}

const getCommand = "Get-CsTeamsEducationConfiguration"
const setCommand = "Set-CsTeamsEducationConfiguration"
