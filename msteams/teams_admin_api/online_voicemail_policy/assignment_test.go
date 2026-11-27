package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const forest = "https://admin.example.invalid/"

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func generatedWire(method, url, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("Accept", "application/json")
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func TestUnitVoicemailGeneratedAssignmentRequests(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name    string          `json:"name"`
			Method  string          `json:"method"`
			URL     string          `json:"generated_url"`
			Reply   string          `json:"response_body"`
			Status  int             `json:"response_status"`
			Body    json.RawMessage `json:"body"`
			Failure *string         `json:"decoder_failure"`
		} `json:"cases"`
	}{
		Cases: []struct {
			Name    string          `json:"name"`
			Method  string          `json:"method"`
			URL     string          `json:"generated_url"`
			Reply   string          `json:"response_body"`
			Status  int             `json:"response_status"`
			Body    json.RawMessage `json:"body"`
			Failure *string         `json:"decoder_failure"`
		}{

			{
				Name:   "user",
				Method: "PATCH",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"policyName\":\"Fixture\"}"),
			},

			{
				Name:   "user-empty",
				Method: "PATCH",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"policyName\":\"\"}"),
			},

			{
				Name:   "user-null",
				Method: "PATCH",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{}"),
			},

			{
				Name:   "tenant",
				Method: "PATCH",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"forceSwitchPresent\":false,\"policyName\":\"Fixture\"}"),
			},

			{
				Name:   "tenant-null-force",
				Method: "PATCH",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"forceSwitchPresent\":true}"),
			},

			{
				Name:   "group-accepted",
				Method: "POST",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"policyName\":\"Fixture\"}"),
			},

			{
				Name:   "group-zero-priority",
				Method: "POST",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":0}"),
			},

			{
				Name:   "group-null-negative",
				Method: "POST",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"priority\":-1}"),
			},

			{
				Name:    "group-malformed-accepted",
				Method:  "POST",
				URL:     "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status:  400,
				Body:    json.RawMessage("{\"priority\":-1}"),
				Failure: new("decode-rejected"),
			},

			{
				Name:   "group-empty-id",
				Method: "POST",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"priority\":-1}"),
			},

			{
				Name:   "group-forbidden",
				Method: "POST",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy",
				Status: 400,
				Body:   json.RawMessage("{\"priority\":-1}"),
			},

			{
				Name:   "operation-read",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations/fixture%2Foperation",
				Status: 400,
			},

			{
				Name:    "operation-malformed",
				Method:  "GET",
				URL:     "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations/fixture%2Foperation",
				Status:  400,
				Failure: new("decode-rejected"),
			},

			{
				Name:   "operation-unknown-fields",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations/fixture%2Foperation",
				Status: 400,
			},
		},
	}

	for _, f := range fixture.Cases {
		command := "Grant-CsOnlineVoicemailPolicy"

		if strings.HasPrefix(f.Name, "operation") {
			command = "Get-CsBatchPolicyAssignmentOperation"
			f.Body = nil // A JSON null fixture body denotes absent content here.
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{discovery(t), generatedWire(f.Method, f.URL, command, f.Body, f.Status, f.Reply)}}
			if f.Failure != nil {
				c.Want.Error = client.ErrDecode
			}
			if f.Status == 403 {
				c.Want.Error = auth.ErrForbidden
			}
			if len(c.Steps) != 0 {
				c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
				if strings.HasPrefix(f.Name, "operation") {
					v, r, e := s.GetAssignmentOperation(ctx, "fixture/operation")
					if e == nil && (v == nil || v.OverallStatus == nil || *v.OverallStatus != "Future") {
						e = errors.New("open operation status lost")
					}
					return testutil.Observation{Response: r, Err: e}
				}
				name := policy.Set("Fixture")
				if strings.Contains(f.Name, "null") || strings.HasPrefix(f.Name, "group-malformed") || f.Name == "group-empty-id" || f.Name == "group-forbidden" {
					name = policy.Null[string]()
				}
				if f.Name == "user-empty" {
					name = policy.Set("")
				}
				var v *resource.AssignmentResult
				if strings.HasPrefix(f.Name, "user") {
					value, r, e := s.AssignUser(ctx, "fixture/user", name)
					return testutil.Observation{Value: value, Response: r, Err: e}
				}
				if strings.HasPrefix(f.Name, "tenant") {
					value, r, e := s.AssignTenantDefault(ctx, name, strings.Contains(f.Name, "force"))
					return testutil.Observation{Value: value, Response: r, Err: e}
				}
				var priority *int32
				if f.Name == "group-zero-priority" {
					priority = new(int32)
				}
				if f.Name != "group-accepted" && f.Name != "group-zero-priority" {
					n := int32(-1)
					priority = &n
				}
				v, r, e := s.AssignGroup(ctx, "fixture/group", name, priority)
				if r != nil && r.StatusCode() == 202 && (v == nil || !v.Accepted) {
					return testutil.Observation{Err: errors.New("group acceptance lost")}
				}
				if e == nil && f.Name == "group-empty-id" && v.OperationID != nil {
					e = errors.New("fabricated operation ID")
				}
				return testutil.Observation{Response: r, Err: e}
			}
			run(t, c)
		}
	}
}
