package teamsappsetuppolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsAppSetupPolicy"
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"AppPresetList\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AppPresetList\"\n    }\n  ],\n  \"PinnedAppBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedAppBarApps\",\n      \"Order\": 9007199254740993\n    }\n  ],\n  \"PinnedMessageBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedMessageBarApps\",\n      \"Order\": 9007199254740993\n    }\n  ],\n  \"AppPresetMeetingList\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AppPresetMeetingList\"\n    }\n  ],\n  \"AdditionalCustomizationApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AdditionalCustomizationApps\",\n      \"AdditionalCustomizationId\": \"fixture-customization\"\n    }\n  ],\n  \"PinnedCallingBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedCallingBarApps\"\n    }\n  ],\n  \"Description\": \"\",\n  \"AllowSideLoading\": false,\n  \"AllowUserPinning\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "New-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "New-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AppPresetList\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AppPresetList\"\n    }\n  ],\n  \"PinnedAppBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedAppBarApps\",\n      \"Order\": 9007199254740993\n    }\n  ],\n  \"PinnedMessageBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedMessageBarApps\",\n      \"Order\": 9007199254740993\n    }\n  ],\n  \"AppPresetMeetingList\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AppPresetMeetingList\"\n    }\n  ],\n  \"AdditionalCustomizationApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-AdditionalCustomizationApps\",\n      \"AdditionalCustomizationId\": \"fixture-customization\"\n    }\n  ],\n  \"PinnedCallingBarApps\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-PinnedCallingBarApps\"\n    }\n  ],\n  \"Description\": \"\",\n  \"AllowSideLoading\": false,\n  \"AllowUserPinning\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"AppPresetList\": [],\n  \"PinnedAppBarApps\": [],\n  \"PinnedMessageBarApps\": [],\n  \"AppPresetMeetingList\": [],\n  \"AdditionalCustomizationApps\": [],\n  \"PinnedCallingBarApps\": [],\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "patch-scalars",
				Operation: "patch",
				Fields:    "scalars",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"AllowSideLoading\": false,\n  \"AllowUserPinning\": null,\n  \"Force\": \"False\"\n}",

					CmdletName: "Set-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsAppSetupPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsAppSetupPolicy",
					Body:   "{\n  \"PolicyName\": null,\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsAppSetupPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsAppSetupPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsAppSetupPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
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
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsAppSetupPolicy",
					Body:   "{\n  \"policyName\": \"\",\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "read-null-lists",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "read-null-id",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "read-int64-order",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "read-order-overflow",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
				},
			},
			{
				Name:      "read-null-customization",
				Operation: "read",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
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
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAppSetupPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAppSetupPolicy",
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
		t.Fatalf("app setup contract differs: %v %+v", e, r.Differences())
	}
	return r
}
