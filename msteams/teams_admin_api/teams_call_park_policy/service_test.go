package teamscallparkpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func fixtureSettings() resource.Settings {
	return resource.Settings{Description: new(""), AllowCallPark: new(false), ParkTimeoutSeconds: new(int64(0)), PickupRangeStart: new(int64(9007199254740993)), PickupRangeEnd: new(int64(9223372036854775807))}
}
func TestUnitCallParkTypedGeneratedRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			command, set := getCommand, "Identity"
			switch f.Operation {
			case "ConfigurationCreateConfiguration":
				command = newCommand
			case "ConfigurationUpdateConfiguration":
				command = setCommand
			case "ConfigurationDeleteConfiguration":
				command = deleteCommand
			case "UserGrantPolicy":
				command = grantCommand
			case "TenantGrantPolicy":
				command, set = grantCommand, "GrantToTenant"
			case "GroupGrantPolicyAssignment":
				command, set = grantCommand, "GrantToGroup"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}}
			if !f.Canceled {
				c.Steps = []contracttest.Exchange{discovery(t), requestWire(t, f, command)}
			} else {
				c.Want.Error = context.Canceled
			}
			if f.Method == "GET" && f.Status == 200 && !f.Canceled {
				c.Want.JSON = json.RawMessage(f.Reply)
			}
			if len(c.Steps) != 0 {
				c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsCallParkPolicies
				if f.Canceled {
					canceled, cancel := context.WithCancel(ctx)
					cancel()
					v, r, e := s.List(canceled)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				var o contracttest.Observation
				targetID, settings := identity, fixtureSettings()
				if f.Name == "policy-identity-only-put" {
					targetID, settings = "Global", resource.Settings{}
				}
				if f.Name == "policy-signed-int64-put" {
					targetID = "Global"
					settings = resource.Settings{Description: new(""), AllowCallPark: new(false), ParkTimeoutSeconds: new(int64(-9223372036854775808)), PickupRangeStart: new(int64(0)), PickupRangeEnd: new(int64(-1))}
				}
				switch f.Operation {
				case "ConfigurationGetConfigurations":
					v, r, e := s.List(ctx)
					o = contracttest.Observation{Value: v, Response: r, Err: e}
				case "ConfigurationGetConfiguration":
					id := identity
					if strings.Contains(f.URL, "/configuration/Fixture") {
						id = "Fixture"
					}
					v, r, e := s.GetByIdentity(ctx, id)
					o = contracttest.Observation{Value: v, Response: r, Err: e}
				case "ConfigurationCreateConfiguration":
					r, e := s.CreateByIdentity(ctx, identity, fixtureSettings())
					o = contracttest.Observation{Response: r, Err: e}
				case "ConfigurationUpdateConfiguration":
					r, e := s.UpdateByIdentity(ctx, targetID, settings)
					o = contracttest.Observation{Response: r, Err: e}
				case "ConfigurationDeleteConfiguration":
					r, e := s.DeleteByIdentity(ctx, identity)
					o = contracttest.Observation{Response: r, Err: e}
				case "UserGrantPolicy":
					name := "Fixture"
					if f.Name == "user-clear-empty-policy" {
						name = ""
					}
					r, e := s.AssignToUser(ctx, "fixture/user", name)
					o = contracttest.Observation{Response: r, Err: e}
				case "TenantGrantPolicy":
					r, e := s.AssignTenantDefault(ctx, "Fixture", false)
					o = contracttest.Observation{Response: r, Err: e}
				case "GroupGrantPolicyAssignment":
					name, priority := "Fixture", new(int32(2))
					if f.Name == "group-assignment-zero-priority" {
						priority = new(int32(0))
					}
					if f.Name == "group-clear-without-priority" {
						name, priority = "", nil
					}
					v, r, e := s.AssignToGroup(ctx, "fixture/group", name, priority)
					if e == nil && (v == nil || v.Accepted != (f.Status == 202)) {
						e = errors.New("group acceptance status changed")
					}
					o = contracttest.Observation{Response: r, Err: e}
				default:
					o.Err = errors.New("unknown pinned operation")
				}
				if f.Status >= 400 {
					var api *client.APIError
					if !errors.As(o.Err, &api) || api.StatusCode != f.Status || strings.Contains(o.Err.Error(), "Fixture message") {
						return contracttest.Observation{Err: errors.New("generated error status/redaction changed")}
					}
				}
				return o
			}
			run(t, c)
		}
	}
}

func TestUnitCallParkFiltersUseRealResponses(t *testing.T) {
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
		{Name: "whitespace", Pattern: " ", Empty: false},
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
			Name: "real-filter-" + tc.Name, Family: "teams_call_park_policy", Scope: "service", Source: "msteams/teams_admin_api/teams_call_park_policy/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsCallParkPolicy", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.TeamsCallParkPolicies
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
