package teamsclientconfiguration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsClientConfiguration"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsClientConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration",
					CmdletName: "Get-CsTeamsClientConfiguration",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Input:     "{\"AllowEmailIntoChannel\":false,\"RestrictedSenderList\":\"Future-RestrictedSenderList\",\"AllowDropBox\":false,\"AllowBox\":false,\"AllowGoogleDrive\":false,\"AllowShareFile\":false,\"AllowEgnyte\":false,\"AllowOrganizationTab\":false,\"AllowSkypeBusinessInterop\":false,\"ContentPin\":\"Future-ContentPin\",\"AllowResourceAccountSendMessage\":false,\"ResourceAccountContentAccess\":\"Future-ResourceAccountContentAccess\",\"AllowGuestUser\":false,\"AllowScopedPeopleSearchandAccess\":false,\"AllowRoleBasedChatPermissions\":false,\"ExtendedWorkInfoInPeopleSearch\":false,\"UseUnifiedDomain\":\"Future-UseUnifiedDomain\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowEmailIntoChannel\": false,\n  \"RestrictedSenderList\": \"Future-RestrictedSenderList\",\n  \"AllowDropBox\": false,\n  \"AllowBox\": false,\n  \"AllowGoogleDrive\": false,\n  \"AllowShareFile\": false,\n  \"AllowEgnyte\": false,\n  \"AllowOrganizationTab\": false,\n  \"AllowSkypeBusinessInterop\": false,\n  \"ContentPin\": \"Future-ContentPin\",\n  \"AllowResourceAccountSendMessage\": false,\n  \"ResourceAccountContentAccess\": \"Future-ResourceAccountContentAccess\",\n  \"AllowGuestUser\": false,\n  \"AllowScopedPeopleSearchandAccess\": false,\n  \"AllowRoleBasedChatPermissions\": false,\n  \"ExtendedWorkInfoInPeopleSearch\": false,\n  \"UseUnifiedDomain\": \"Future-UseUnifiedDomain\"\n}",

					CmdletName: "Set-CsTeamsClientConfiguration",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Input:     "{\"RestrictedSenderList\":null,\"AllowEmailIntoChannel\":null,\"UseUnifiedDomain\":null}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowEmailIntoChannel\": null,\n  \"RestrictedSenderList\": null,\n  \"UseUnifiedDomain\": null\n}",

					CmdletName: "Set-CsTeamsClientConfiguration",
				},
			},
			{
				Name:      "patch-empty",
				Operation: "patch",
				Input:     "{\"RestrictedSenderList\":\"\",\"ContentPin\":\"\"}",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"RestrictedSenderList\": \"\",\n  \"ContentPin\": \"\"\n}",

					CmdletName: "Set-CsTeamsClientConfiguration",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Input:     "{\"RestrictedSenderList\":\"\"}",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"RestrictedSenderList\": \"\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsClientConfiguration",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Input:     "{\"RestrictedSenderList\":\"\"}",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"RestrictedSenderList\": \"\",\n  \"Force\": \"True\"\n}",

					CmdletName: "Set-CsTeamsClientConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsClientConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsClientConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration",
					CmdletName: "Get-CsTeamsClientConfiguration",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsClientConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsClientConfiguration",
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
		t.Fatalf("client configuration contract differs: %v %+v", e, r.Differences())
	}
	return r
}

const setCommand = "Set-CsTeamsClientConfiguration"
