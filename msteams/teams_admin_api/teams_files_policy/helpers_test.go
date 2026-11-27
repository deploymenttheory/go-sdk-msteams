package teamsfilespolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsFilesPolicy"
const item = collection + "/configuration/Global"
const getCommand = "Get-CsTeamsFilesPolicy"
const setCommand = "Set-CsTeamsFilesPolicy"
const grantCommand = "Grant-CsTeamsFilesPolicy"
const scoped = "scope /Fixture / Child"

type requestInputCase struct {
	Name, Operation, Input, Method, Path, Command string
	ForceMode                                     int
	PolicyName                                    *string
	Rank                                          int32
	Body                                          json.RawMessage
}

// requestInputs are ordinary typed-input/request-shape vectors. They contain no
// successful server replies; request-only tests terminate with controlled HTTP 400.
func requestInputs() []requestInputCase {
	return []requestInputCase{
		{Name: "read", Operation: "read", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "list-empty", Operation: "list", Input: "{}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "create-all", Operation: "create", Input: "{\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "POST", Path: "/Skype.Policy/configurations/TeamsFilesPolicy", Command: "New-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"NativeFileEntryPoints\": \"Future-NativeFileEntryPoints\",\n  \"SPChannelFilesTab\": \"Future-SPChannelFilesTab\",\n  \"DefaultFileUploadAppId\": \"Future-DefaultFileUploadAppId\",\n  \"FileSharingInChatswithExternalUsers\": \"Future-FileSharingInChatswithExternalUsers\"\n}")},
		{Name: "create-empty", Operation: "create", Input: "{}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "POST", Path: "/Skype.Policy/configurations/TeamsFilesPolicy", Command: "New-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"scope /Fixture / Child\"\n}")},
		{Name: "patch-all", Operation: "patch", Input: "{\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Set-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"read-model-identity\",\n  \"NativeFileEntryPoints\": \"Future-NativeFileEntryPoints\",\n  \"SPChannelFilesTab\": \"Future-SPChannelFilesTab\",\n  \"DefaultFileUploadAppId\": \"Future-DefaultFileUploadAppId\",\n  \"FileSharingInChatswithExternalUsers\": \"Future-FileSharingInChatswithExternalUsers\"\n}")},
		{Name: "patch-null", Operation: "patch", Input: "{\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Set-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"read-model-identity\",\n  \"NativeFileEntryPoints\": null,\n  \"SPChannelFilesTab\": null,\n  \"DefaultFileUploadAppId\": null,\n  \"FileSharingInChatswithExternalUsers\": null\n}")},
		{Name: "patch-empty", Operation: "patch", Input: "{\"NativeFileEntryPoints\":\"\",\"SPChannelFilesTab\":\"\",\"DefaultFileUploadAppId\":\"\",\"FileSharingInChatswithExternalUsers\":\"\"}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "PATCH", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Set-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"read-model-identity\",\n  \"NativeFileEntryPoints\": \"\",\n  \"SPChannelFilesTab\": \"\",\n  \"DefaultFileUploadAppId\": \"\",\n  \"FileSharingInChatswithExternalUsers\": \"\"\n}")},
		{Name: "patch-force-false", Operation: "patch", Input: "{\"NativeFileEntryPoints\":\"\"}", ForceMode: 0, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Set-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"read-model-identity\",\n  \"NativeFileEntryPoints\": \"\",\n  \"Force\": \"False\"\n}")},
		{Name: "patch-force-true", Operation: "patch", Input: "{\"NativeFileEntryPoints\":\"\"}", ForceMode: 1, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Set-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"read-model-identity\",\n  \"NativeFileEntryPoints\": \"\",\n  \"Force\": \"True\"\n}")},
		{Name: "delete", Operation: "delete", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "DELETE", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Remove-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"scope /Fixture / Child\"\n}")},
		{Name: "delete-force-false", Operation: "delete", Input: "{}", ForceMode: 0, PolicyName: nil, Rank: -1, Method: "DELETE", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Remove-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}")},
		{Name: "read-empty", Operation: "read", Input: "{}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "read-null", Operation: "read", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "list-envelope", Operation: "list", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "read-denied", Operation: "read", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "GET", Path: "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/scope%20/Fixture%20%2F%20Child", Command: "Get-CsTeamsFilesPolicy", Body: json.RawMessage("")},
		{Name: "grant-user-null", Operation: "user", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/users/user%2Ffixture/policies/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}")},
		{Name: "grant-tenant-null", Operation: "tenant", Input: "{}", ForceMode: 0, PolicyName: nil, Rank: -1, Method: "PATCH", Path: "/Skype.Policy/tenants/policies/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"PolicyName\": null,\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}")},
		{Name: "grant-group-null", Operation: "group", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: -1, Method: "POST", Path: "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": null,\n  \"priority\": -1\n}")},
		{Name: "grant-user-empty", Operation: "user", Input: "{}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "PATCH", Path: "/Skype.Policy/users/user%2Ffixture/policies/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": \"\",\n  \"additionalParameters\": null\n}")},
		{Name: "grant-tenant-empty", Operation: "tenant", Input: "{}", ForceMode: 1, PolicyName: new(""), Rank: -1, Method: "PATCH", Path: "/Skype.Policy/tenants/policies/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": true\n}")},
		{Name: "grant-group-empty", Operation: "group", Input: "{}", ForceMode: -1, PolicyName: new(""), Rank: -1, Method: "POST", Path: "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": \"\",\n  \"priority\": -1\n}")},
		{Name: "grant-group-rank-zero", Operation: "group", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: 0, Method: "POST", Path: "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": null,\n  \"priority\": 0\n}")},
		{Name: "grant-group-rank-two", Operation: "group", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: 2, Method: "POST", Path: "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": null,\n  \"priority\": 2\n}")},
		{Name: "grant-group-accepted", Operation: "group", Input: "{}", ForceMode: -1, PolicyName: nil, Rank: 2, Method: "POST", Path: "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", Command: "Grant-CsTeamsFilesPolicy", Body: json.RawMessage("{\n  \"policyName\": null,\n  \"priority\": 2\n}")},
	}
}

func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if len(body) != 0 {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	if len(body) == 0 {
		body = nil
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("files policy behavior differs: %v %+v", err, result.Differences())
	}
	return result
}
