package teamseducationassignmentsapppolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/teams_education_assignments_app_policy/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"ParentDigestEnabledType\":\"FutureParent\",\"MakeCodeEnabledType\":\"FutureCode\",\"TurnItInEnabledType\":\"FutureTurnItIn\",\"TurnItInApiUrl\":\"HTTPS://fixture.invalid\",\"TurnItInApiKey\":\"offline-fixture-key\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentDigestEnabledType\": \"FutureParent\",\n  \"MakeCodeEnabledType\": \"FutureCode\",\n  \"TurnItInEnabledType\": \"FutureTurnItIn\",\n  \"TurnItInApiUrl\": \"HTTPS://fixture.invalid\",\n  \"TurnItInApiKey\": \"offline-fixture-key\"\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"ParentDigestEnabledType\":null,\"MakeCodeEnabledType\":null,\"TurnItInEnabledType\":null,\"TurnItInApiUrl\":null,\"TurnItInApiKey\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentDigestEnabledType\": null,\n  \"MakeCodeEnabledType\": null,\n  \"TurnItInEnabledType\": null,\n  \"TurnItInApiUrl\": null,\n  \"TurnItInApiKey\": null\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-empty",
				Operation: "patch",
				Input:     "{\"ParentDigestEnabledType\":\"\",\"MakeCodeEnabledType\":\"\",\"TurnItInEnabledType\":\"\",\"TurnItInApiKey\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"ParentDigestEnabledType\": \"\",\n  \"MakeCodeEnabledType\": \"\",\n  \"TurnItInEnabledType\": \"\",\n  \"TurnItInApiKey\": \"\"\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"TurnItInApiKey\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"TurnItInApiKey\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"TurnItInApiKey\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"TurnItInApiKey\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-url-final-lf",
				Operation: "patch",
				Input:     "{\"TurnItInApiUrl\":\"https://opaque\\n\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"TurnItInApiUrl\": \"https://opaque\\n\"\n}",

					CmdletName: "Set-CsTeamsEducationAssignmentsAppPolicy",
				},
			},
			{
				Name:      "patch-url-empty",
				Operation: "patch",
				Input:     "{\"TurnItInApiUrl\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{},
			},
			{
				Name:      "patch-url-http",
				Operation: "patch",
				Input:     "{\"TurnItInApiUrl\":\"http://fixture.invalid\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{},
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsEducationAssignmentsAppPolicy",
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
	c.Family = "teams_education_assignments_app_policy"
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
		t.Fatalf("education assignments contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}

const getCommand = "Get-CsTeamsEducationAssignmentsAppPolicy"
const setCommand = "Set-CsTeamsEducationAssignmentsAppPolicy"
