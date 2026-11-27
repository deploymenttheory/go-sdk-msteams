package onlinevoicemailvalidationconfiguration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"
const collection = "/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration"
const item = collection + "/configuration/Global"

type inputVectors struct {
	Name, Operation, Response string
	Status                    int
	Force                     bool
	NullURI                   bool `json:"null_uri"`
	Request                   struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	value := struct {
		Cases []inputVectors `json:"cases"`
	}{
		Cases: []inputVectors{
			{
				Name:      "read",
				Operation: "read",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "patch-null",
				Operation: "patch",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AudioFileValidationEnabled\": false,\n  \"AudioFileValidationUri\": null\n}",

					CmdletName: "Set-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "patch-uri",
				Operation: "patch",
				Status:    400,
				Force:     false,
				NullURI:   false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AudioFileValidationEnabled\": false,\n  \"AudioFileValidationUri\": \"HTTPS://example.invalid/validate\"\n}",

					CmdletName: "Set-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Status:    400,
				Force:     true,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AudioFileValidationEnabled\": false,\n  \"AudioFileValidationUri\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "read-invalid-uri",
				Operation: "read",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "read-case-invalid-uri",
				Operation: "read",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Status:    400,
				Force:     false,
				NullURI:   true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailValidationConfiguration",
				},
			},
		},
	}

	return value.Cases
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
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
		t.Fatalf("voicemail contract differs: %v %+v", e, r.Differences())
	}
	return r
}
