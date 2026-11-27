package privacyconfiguration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"
const collection = "/Skype.Policy/configurations/PrivacyConfiguration"
const item = collection + "/configuration/Global"

type inputVectors struct {
	Name, Operation, Response string
	Status                    int
	Force                     bool
	NullLocation              bool `json:"null_location"`
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
				Name:         "read",
				Operation:    "read",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "list",
				Operation:    "list",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration",
					CmdletName: "Get-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "patch-null",
				Operation:    "patch",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"EnablePrivacyMode\": false,\n  \"AutoInitiateContacts\": true,\n  \"PublishLocationDataDefault\": null,\n  \"DisplayPublishedPhotoDefault\": false\n}",

					CmdletName: "Set-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "patch-false",
				Operation:    "patch",
				Status:       400,
				Force:        false,
				NullLocation: false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"EnablePrivacyMode\": false,\n  \"AutoInitiateContacts\": true,\n  \"PublishLocationDataDefault\": false,\n  \"DisplayPublishedPhotoDefault\": false\n}",

					CmdletName: "Set-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "patch-force-false",
				Operation:    "patch",
				Status:       400,
				Force:        true,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"EnablePrivacyMode\": false,\n  \"AutoInitiateContacts\": true,\n  \"PublishLocationDataDefault\": null,\n  \"DisplayPublishedPhotoDefault\": false,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "read-empty",
				Operation:    "read",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "list-envelope",
				Operation:    "list",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration",
					CmdletName: "Get-CsPrivacyConfiguration",
				},
			},
			{
				Name:         "read-denied",
				Operation:    "read",
				Status:       400,
				Force:        false,
				NullLocation: true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/PrivacyConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsPrivacyConfiguration",
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
		t.Fatalf("privacy contract differs: %v %+v", e, r.Differences())
	}
	return r
}
