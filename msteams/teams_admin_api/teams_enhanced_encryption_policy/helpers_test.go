package teamsenhancedencryptionpolicy_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://admin.example.invalid/"
const collection = "/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy"
const item = collection + "/configuration/Global"
const identity = "tag:Fixture / Policy"
const getCommand = "Get-CsTeamsEnhancedEncryptionPolicy"
const newCommand = "New-CsTeamsEnhancedEncryptionPolicy"
const setCommand = "Set-CsTeamsEnhancedEncryptionPolicy"
const grantCommand = "Grant-CsTeamsEnhancedEncryptionPolicy"
const deleteCommand = "Remove-CsTeamsEnhancedEncryptionPolicy"

type requestInput struct {
	Name, Operation, Method string
	URL                     string `json:"generated_url"`
	Status                  int    `json:"response_status"`
	Reply                   string `json:"response_body"`

	Canceled bool
	Body     json.RawMessage
}

func requestInputs(t *testing.T) []requestInput {
	t.Helper()

	v := struct{ Cases []requestInput }{
		Cases: []requestInput{
			{
				Name:      "policy-list",
				Operation: "ConfigurationGetConfigurations",
				Method:    "GET",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
			},
			{
				Name:      "policy-read-escaped",
				Operation: "ConfigurationGetConfiguration",
				Method:    "GET",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/tag%3AFixture%20%2F%20Policy",
				Status:    400,
				Canceled:  false,
			},
			{
				Name:      "policy-create",
				Operation: "ConfigurationCreateConfiguration",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"FutureCalling\",\"Description\":\"\",\"Identity\":\"tag:Fixture / Policy\",\"MeetingEndToEndEncryption\":\"FutureMeeting\"}"),
			},
			{
				Name:      "policy-update-put",
				Operation: "ConfigurationUpdateConfiguration",
				Method:    "PUT",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/tag%3AFixture%20%2F%20Policy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"FutureCalling\",\"Description\":\"\",\"Identity\":\"tag:Fixture / Policy\",\"MeetingEndToEndEncryption\":\"FutureMeeting\"}"),
			},
			{
				Name:      "policy-delete",
				Operation: "ConfigurationDeleteConfiguration",
				Method:    "DELETE",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/tag%3AFixture%20%2F%20Policy",
				Status:    400,
				Canceled:  false,
			},
			{
				Name:      "policy-forbidden",
				Operation: "ConfigurationGetConfiguration",
				Method:    "GET",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/Fixture",
				Status:    400,
				Canceled:  false,
			},
			{
				Name:      "policy-malformed-error",
				Operation: "ConfigurationGetConfiguration",
				Method:    "GET",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/Fixture",
				Status:    400,
				Canceled:  false,
			},
			{
				Name:      "policy-pre-canceled",
				Operation: "ConfigurationGetConfigurations",
				Status:    400,
				Canceled:  true,
			},
			{
				Name:      "policy-identity-only-put",
				Operation: "ConfigurationUpdateConfiguration",
				Method:    "PUT",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/Global",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"Identity\":\"Global\"}"),
			},
			{
				Name:      "policy-whitespace-put",
				Operation: "ConfigurationUpdateConfiguration",
				Method:    "PUT",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/%20",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"Description\":\"\",\"Identity\":\" \"}"),
			},
			{
				Name:      "policy-empty-string-put",
				Operation: "ConfigurationUpdateConfiguration",
				Method:    "PUT",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/Global",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Description\":\"\",\"Identity\":\"Global\",\"MeetingEndToEndEncryption\":\"\"}"),
			},
			{
				Name:      "user-assignment",
				Operation: "UserGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\"}"),
			},
			{
				Name:      "user-clear-empty-policy",
				Operation: "UserGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"\"}"),
			},
			{
				Name:      "tenant-assignment",
				Operation: "TenantGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"forceSwitchPresent\":false,\"policyName\":\"Fixture\"}"),
			},
			{
				Name:      "group-assignment-accepted",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},
			{
				Name:      "group-assignment-no-content",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},
			{
				Name:      "group-assignment-zero-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":0}"),
			},
			{
				Name:      "group-clear-without-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsEnhancedEncryptionPolicy",
				Status:    400,
				Canceled:  false,
				Body:      json.RawMessage("{\"policyName\":\"\"}"),
			},
		},
	}

	return v.Cases
}
func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil && string(body) != "null" {
		h.Set("Content-Type", "application/json")
	} else {
		body = nil
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func requestWire(t *testing.T, c requestInput, command string) testutil.Exchange {
	t.Helper()
	u, e := url.Parse(c.URL)
	if e != nil {
		t.Fatal(e)
	}
	parts := strings.Split(u.EscapedPath(), "/")
	for i, p := range parts {
		v, e := url.PathUnescape(p)
		if e != nil {
			t.Fatal(e)
		}
		parts[i] = url.PathEscape(v)
	}
	return wire(c.Method, strings.Join(parts, "/"), command, c.Body, c.Status, c.Reply)
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("enhanced encryption policy contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}
