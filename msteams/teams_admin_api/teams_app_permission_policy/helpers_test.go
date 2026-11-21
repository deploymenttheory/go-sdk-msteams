package teamsapppermissionpolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/teams_app_permission_policy/testdata/get_global.json"
const sourceModels = "msteams/teams_admin_api/teams_app_permission_policy/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsAppPermissionPolicy"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"

type inputVectors struct {
	Name, Operation, Fields, Response string
	Status                            int
	Force                             bool
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
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "create",
				Operation: "create",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"DefaultCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-default\"\n    }\n  ],\n  \"GlobalCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-global\"\n    }\n  ],\n  \"PrivateCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-private\"\n    }\n  ],\n  \"Description\": \"\",\n  \"DefaultCatalogAppsType\": \"FutureDefaultMode\",\n  \"GlobalCatalogAppsType\": \"\",\n  \"PrivateCatalogAppsType\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "New-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "create-empty-settings",
				Operation: "create",
				Fields:    "none",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "New-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "patch",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"DefaultCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-default\"\n    }\n  ],\n  \"GlobalCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-global\"\n    }\n  ],\n  \"PrivateCatalogApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-private\"\n    }\n  ],\n  \"Description\": \"\",\n  \"DefaultCatalogAppsType\": \"FutureDefaultMode\",\n  \"GlobalCatalogAppsType\": \"\",\n  \"PrivateCatalogAppsType\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "patch-empty-lists",
				Operation: "patch",
				Fields:    "empty-lists",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"DefaultCatalogApps\": [],\n  \"GlobalCatalogApps\": [],\n  \"PrivateCatalogApps\": [],\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "patch-scalar-only",
				Operation: "patch",
				Fields:    "description",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"DefaultCatalogAppsType\": \"FutureDefaultMode\",\n  \"GlobalCatalogAppsType\": \"\",\n  \"PrivateCatalogAppsType\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Fields:    "none",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "grant-user",
				Operation: "user",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsAppPermissionPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "grant-tenant",
				Operation: "tenant",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsAppPermissionPolicy",
					Body:   "{\n  \"PolicyName\": null,\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "grant-group",
				Operation: "group",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsAppPermissionPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:       "grant-user-empty",
				Operation:  "user",
				Fields:     "none",
				Status:     400,
				Force:      false,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsAppPermissionPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:       "grant-tenant-empty",
				Operation:  "tenant",
				Fields:     "none",
				Status:     400,
				Force:      false,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsAppPermissionPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:       "grant-group-empty",
				Operation:  "group",
				Fields:     "none",
				Status:     400,
				Force:      false,
				PolicyName: new(""),
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsAppPermissionPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "read-null-catalogs",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "read-null-app-id",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppPermissionPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppPermissionPolicy",
				},
			},
		},
	}

	return v.Cases
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
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
	c.Family = "teams_app_permission_policy"
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
		t.Fatalf("app permission contract differs: %v %+v", e, r.Differences())
	}
	return r
}
