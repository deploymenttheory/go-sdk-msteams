package teamsfeedbackpolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsFeedbackPolicy"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "create-all",
				Operation: "create",
				Input:     "{\"UserInitiatedMode\":\"Future-UserInitiatedMode\",\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\",\"AllowScreenshotCollection\":false,\"AllowEmailCollection\":false,\"AllowLogCollection\":false,\"EnableFeatureSuggestions\":false}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"UserInitiatedMode\": \"Future-UserInitiatedMode\",\n  \"ReceiveSurveysMode\": \"Future-ReceiveSurveysMode\",\n  \"AllowScreenshotCollection\": false,\n  \"AllowEmailCollection\": false,\n  \"AllowLogCollection\": false,\n  \"EnableFeatureSuggestions\": false\n}",

					CmdletName: "New-CsTeamsFeedbackPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "New-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"UserInitiatedMode\":\"Future-UserInitiatedMode\",\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\",\"AllowScreenshotCollection\":false,\"AllowEmailCollection\":false,\"AllowLogCollection\":false,\"EnableFeatureSuggestions\":false}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UserInitiatedMode\": \"Future-UserInitiatedMode\",\n  \"ReceiveSurveysMode\": \"Future-ReceiveSurveysMode\",\n  \"AllowScreenshotCollection\": false,\n  \"AllowEmailCollection\": false,\n  \"AllowLogCollection\": false,\n  \"EnableFeatureSuggestions\": false\n}",

					CmdletName: "Set-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"UserInitiatedMode\":null,\"ReceiveSurveysMode\":null,\"AllowScreenshotCollection\":null,\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"EnableFeatureSuggestions\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UserInitiatedMode\": null,\n  \"ReceiveSurveysMode\": null,\n  \"AllowScreenshotCollection\": null,\n  \"AllowEmailCollection\": null,\n  \"AllowLogCollection\": null,\n  \"EnableFeatureSuggestions\": null\n}",

					CmdletName: "Set-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "patch-empty-true",
				Operation: "patch",
				Input:     "{\"UserInitiatedMode\":\"\",\"ReceiveSurveysMode\":\"\",\"AllowScreenshotCollection\":true,\"AllowEmailCollection\":true,\"AllowLogCollection\":true,\"EnableFeatureSuggestions\":true}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UserInitiatedMode\": \"\",\n  \"ReceiveSurveysMode\": \"\",\n  \"AllowScreenshotCollection\": true,\n  \"AllowEmailCollection\": true,\n  \"AllowLogCollection\": true,\n  \"EnableFeatureSuggestions\": true\n}",

					CmdletName: "Set-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"UserInitiatedMode\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UserInitiatedMode\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsFeedbackPolicy",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"UserInitiatedMode\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"UserInitiatedMode\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsFeedbackPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsFeedbackPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsFeedbackPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsFeedbackPolicy",
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
		t.Fatalf("feedback policy contract differs: %v %+v", e, r.Differences())
	}
	return r
}

const getCommand = "Get-CsTeamsFeedbackPolicy"
const setCommand = "Set-CsTeamsFeedbackPolicy"
const grantCommand = "Grant-CsTeamsFeedbackPolicy"
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

const item = collection + "/configuration/Global"
const newCommand = "New-CsTeamsFeedbackPolicy"
const deleteCommand = "Remove-CsTeamsFeedbackPolicy"
