package teamsexternalaccessconfiguration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"

type requestCase struct {
	Name     string `json:"name"`
	Status   int    `json:"status"`
	Response string `json:"response"`
	Request  struct {
		Method string
		Url    string
		Body   string

		CmdletName  string
		FailureType *string
	} `json:"request"`
}

func requestInputs(t *testing.T) []requestCase {
	t.Helper()

	v := struct {
		Cases []requestCase `json:"cases"`
	}{
		Cases: []requestCase{
			{
				Name:   "read",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "create",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"\",\n  \"EnableFederationAccess\": false\n}",

					CmdletName: "New-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "update",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"EnableFederationAccess\": false\n}",

					CmdletName: "Set-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "delete",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "user",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/ExternalAccessPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "tenant",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/ExternalAccessPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "group",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/ExternalAccessPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "acs-read",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAcsFederationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAcsFederationConfiguration",
				},
			},
			{
				Name:   "acs-list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAcsFederationConfiguration",
					CmdletName: "Get-CsTeamsAcsFederationConfiguration",
				},
			},
			{
				Name:   "acs-update",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAcsFederationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AllowedAcsResources\": [],\n  \"EnableAcsUsers\": false\n}",

					CmdletName: "Set-CsTeamsAcsFederationConfiguration",
				},
			},
			{
				Name:   "external-read",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsExternalAccessConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsExternalAccessConfiguration",
				},
			},
			{
				Name:   "external-list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsExternalAccessConfiguration",
					CmdletName: "Get-CsTeamsExternalAccessConfiguration",
				},
			},
			{
				Name:   "external-update",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsExternalAccessConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"BlockedUsers\": [],\n  \"BlockExternalUserAccess\": null\n}",

					CmdletName: "Set-CsTeamsExternalAccessConfiguration",
				},
			},
			{
				Name:   "delete-force-false",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "read-no-content",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "group-accepted-opaque",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/ExternalAccessPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "list-envelope",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "read-tagged",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/tag%3AFixture%20/%20Policy",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "read-future-enum",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "read-null-domain-list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsExternalAccessPolicy",
				},
			},
			{
				Name:   "acs-null-list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAcsFederationConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAcsFederationConfiguration",
				},
			},
			{
				Name:   "external-null-list",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsExternalAccessConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsExternalAccessConfiguration",
				},
			},
			{
				Name:   "delete-denied",
				Status: 400,
				Request: struct {
					Method string
					Url    string
					Body   string

					CmdletName  string
					FailureType *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ExternalAccessPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsExternalAccessPolicy",
				},
			},
		},
	}

	return v.Cases
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}, "X-Ms-Configapi-Powershell-Warningmessage": {"fixture-warning"}}, Body: []byte(reply)}}
}
func exchange(c requestCase) testutil.Exchange {
	var body json.RawMessage
	if c.Request.Body != "" {
		body = json.RawMessage(c.Request.Body)
	}
	return wire(c.Request.Method, c.Request.Url[len(origin):], c.Request.CmdletName, body, c.Status, c.Response)
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) > 0 {
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}

	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("provider differs: %v %+v", err, r.Differences())
	}
	return r
}
