package teamschannelspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
)

func TestUnitChannelsPolicyTypedCompiledCRUD(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			if len(c.Steps) != 0 {
				c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsChannelsPolicies
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return contracttest.Observation{Err: e}
				}
				if f.Operation == "create" {
					v, r, e := s.CreateByIdentity(ctx, scoped, resource.CreateRequest{Settings: settings, Controls: controls})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}

func TestUnitChannelsPolicyTypedModernGrant(t *testing.T) {
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
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\"}"),
			},

			{
				Name:      "user-clear-empty-policy",
				Operation: "UserGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"\"}"),
			},

			{
				Name:      "tenant-assignment",
				Operation: "TenantGrantPolicy",
				Method:    "PATCH",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"forceSwitchPresent\":false,\"policyName\":\"Fixture\"}"),
			},

			{
				Name:      "group-assignment-accepted",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},

			{
				Name:      "group-assignment-no-content",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":2}"),
			},

			{
				Name:      "group-assignment-zero-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsChannelsPolicy",
				Status:    400,
				Body:      json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":0}"),
			},

			{
				Name:      "group-clear-without-priority",
				Operation: "GroupGrantPolicyAssignment",
				Method:    "POST",
				URL:       "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsChannelsPolicy",
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
		set := "Identity"
		if f.Operation == "TenantGrantPolicy" {
			set = "GrantToTenant"
		}
		if f.Operation == "GroupGrantPolicyAssignment" {
			set = "GrantToGroup"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			run(t, contracttest.Case{Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: f.Name + "-" + string(mode), Mode: mode, Source: grantSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(t), modernWire(f.Method, strings.Join(parts, "/"), f.Body, f.Status, f.Reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsChannelsPolicies
				if f.Operation == "UserGrantPolicy" {
					name := "Fixture"
					if f.Name == "user-clear-empty-policy" {
						name = ""
					}
					r, e := s.AssignToUser(ctx, "fixture/user", name)
					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Operation == "TenantGrantPolicy" {
					r, e := s.AssignTenantDefault(ctx, "Fixture", false)
					return contracttest.Observation{Response: r, Err: e}
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
				return contracttest.Observation{Response: r, Err: e}
			}})
		}
	}
}

func TestUnitChannelsPolicyFiltersUseRealResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil {
		t.Fatal("invalid real list fixture")
	}
	identities := func(row json.RawMessage) string {
		var value struct {
			Identity string `json:"Identity"`
		}
		if json.Unmarshal(row, &value) != nil {
			t.Fatal("invalid real list identity")
		}
		return value.Identity
	}
	slices.SortStableFunc(rows, func(a, b json.RawMessage) int {
		left, right := strings.EqualFold(identities(a), "Global"), strings.EqualFold(identities(b), "Global")
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	cases := []struct {
		Name, Pattern, OnlyID string
		Empty                 bool
	}{
		{Name: "list"}, {Name: "all", Pattern: "*"}, {Name: "global", Pattern: "Global", OnlyID: "Global"},
		{Name: "missing", Pattern: "__sdk_missing_fixture__", Empty: true},
		{Name: "whitespace", Pattern: " ", Empty: true},
	}
	for _, row := range rows {
		id := identities(row)
		if strings.HasPrefix(id, "Tag:") && !strings.ContainsAny(id, "*?[]`") {
			cases = append(cases, struct {
				Name, Pattern, OnlyID string
				Empty                 bool
			}{Name: "tag-prefix", Pattern: strings.TrimPrefix(id, "Tag:"), OnlyID: id})
			break
		}
	}
	for _, tc := range cases {
		want := make([]json.RawMessage, 0, len(rows))
		for _, row := range rows {
			if !tc.Empty && (tc.OnlyID == "" || strings.EqualFold(identities(row), tc.OnlyID)) {
				want = append(want, row)
			}
		}
		expected, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		result, err := contracttest.Run(t.Context(), contracttest.Case{
			Name: "real-filter-" + tc.Name, Family: "teams_channels_policy", Scope: "service", Source: "msteams/teams_admin_api/teams_channels_policy/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsChannelsPolicy", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.TeamsChannelsPolicies
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return contracttest.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}
