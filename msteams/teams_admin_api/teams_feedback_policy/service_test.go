package teamsfeedbackpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestUnitFeedbackPolicyTypedCompiledCRUD(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{compiledWire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			if len(c.Steps) != 0 {
				c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "delete":
					v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return testutil.Observation{Err: e}
				}
				if f.Operation == "create" {
					v, r, e := s.CreateByIdentity(ctx, scoped, resource.CreateRequest{Settings: settings, Controls: controls})
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}

func TestUnitFeedbackPolicyTypedModernGrant(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name, Operation, Method string
			URL                     string `json:"generated_url"`
			Status                  int    `json:"response_status"`
			Reply                   string `json:"response_body"`
			Body                    json.RawMessage
		}
	}{
		Cases: []struct {
			Name, Operation, Method string
			URL                     string `json:"generated_url"`
			Status                  int    `json:"response_status"`
			Reply                   string `json:"response_body"`
			Body                    json.RawMessage
		}{

			{
				Name:      "user-assignment",
				Operation: "UserGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\"}"),
			},

			{
				Name:      "user-clear-empty-policy",
				Operation: "UserGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"\"}"),
			},

			{
				Name:      "tenant-assignment",
				Operation: "TenantGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"forceSwitchPresent\":false,\"policyName\":\"Fixture\"}"),
			},

			{
				Name:      "group-assignment-accepted",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},

			{
				Name:      "group-assignment-no-content",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},

			{
				Name:      "group-assignment-zero-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":0}"),
			},

			{
				Name:      "group-clear-without-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsFeedbackPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"\"}"),
			},
		},
	}

	for _, f := range fixture.Cases {

		u, e := url.Parse(f.URL)
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

		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			run(t, testutil.Case{Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{discovery(t), modernWire(f.Method, strings.Join(parts, "/"), f.Body, f.Status, f.Reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
				if f.Operation == "UserGrantPolicy" {
					name := "Fixture"
					if f.Name == "user-clear-empty-policy" {
						name = ""
					}
					r, e := s.AssignToUser(ctx, "fixture/user", name)
					return testutil.Observation{Response: r, Err: e}
				}
				if f.Operation == "TenantGrantPolicy" {
					r, e := s.AssignTenantDefault(ctx, "Fixture", false)
					return testutil.Observation{Response: r, Err: e}
				}
				name, priority := "Fixture", new(int32(2))
				if f.Name == "group-assignment-zero-priority" {
					priority = new(int32(0))
				}
				if f.Name == "group-clear-without-priority" {
					name, priority = "", nil
				}
				v, r, e := s.AssignToGroup(ctx, "fixture/group", name, priority)
				if e == nil && (v == nil || v.Accepted != (f.Status == 202)) {
					e = errors.New("modern group acceptance changed")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		}
	}
}
