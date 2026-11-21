package teamscallholdpolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/teams_call_hold_policy/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsCallHoldPolicy"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"

type inputVectors struct {
	Name, Operation, Fields, Response string
	Status                            int
	ForceMode                         int     `json:"force_mode"`
	PolicyName                        *string `json:"policy_name"`
	Request                           struct {
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
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "list-empty",
				Operation: "list",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "create-all",
				Operation: "create",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"\",\n  \"AudioFileId\": \"future-file\",\n  \"StreamingSourceUrl\": \"future://fixture\",\n  \"StreamingSourceAuthType\": null\n}",

					CmdletName: "New-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "create-empty",
				Operation: "create",
				Fields:    "empty",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "New-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"AudioFileId\": \"future-file\",\n  \"StreamingSourceUrl\": \"future://fixture\",\n  \"StreamingSourceAuthType\": null\n}",

					CmdletName: "Set-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Fields:    "description-null",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": null\n}",

					CmdletName: "Set-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "patch-empty",
				Operation: "patch",
				Fields:    "description-empty",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\"\n}",

					CmdletName: "Set-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"AudioFileId\": \"future-file\",\n  \"StreamingSourceUrl\": \"future://fixture\",\n  \"StreamingSourceAuthType\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"AudioFileId\": \"future-file\",\n  \"StreamingSourceUrl\": \"future://fixture\",\n  \"StreamingSourceAuthType\": null,\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Fields:    "empty",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "delete-force-false",
				Operation: "delete",
				Fields:    "empty",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "read-null",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsCallHoldPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsCallHoldPolicy",
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
	c.Family = "teams_call_hold_policy"
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
		t.Fatalf("call hold contract differs: %v %+v", e, r.Differences())
	}
	return r
}

const getCommand = "Get-CsTeamsCallHoldPolicy"
const newCommand = "New-CsTeamsCallHoldPolicy"
const setCommand = "Set-CsTeamsCallHoldPolicy"
const deleteCommand = "Remove-CsTeamsCallHoldPolicy"
const grantCommand = "Grant-CsTeamsCallHoldPolicy"
const forest = "https://admin.example.invalid/"
const grantSource = "msteams/teams_admin_api/teams_call_hold_policy/testdata/get_global.json"
const wrapperSource = "msteams/teams_admin_api/teams_call_hold_policy/testdata/get_global.json"

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func modernWire(method, path string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {grantCommand}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil && string(body) != "null" {
		h.Set("Content-Type", "application/json")
	} else {
		body = nil
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
