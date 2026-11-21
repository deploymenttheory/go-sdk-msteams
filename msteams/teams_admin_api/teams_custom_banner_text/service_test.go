package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func TestUnitBannerTypedProviderAndSelectors(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		if f.Name == "create-default-id" {
			continue
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Observed.Request
			set := "Identity"
			if f.Selector == "relative" {
				set = "ParentAndRelativeKey"
			}
			command := x.CmdletName
			if x.Method == "" {
				command = setCommand
				if f.Operation == "create" {
					command = newCommand
				}
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}}
			if x.Method != "" {
				var body json.RawMessage
				if x.Body != "" {
					body = json.RawMessage(x.Body)
				}
				c.Steps = []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), command, body, f.Status, f.Response)}
			}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if x.Method == "" {
					c.Want.Error = resource.ErrInput
				}
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Operation == "read" || f.Operation == "list" {
				if f.Response == "" {
					c.Want.JSON = json.RawMessage("null")
				} else {
					c.Want.JSON = json.RawMessage(f.Response)
				}
			}
			if len(c.Steps) != 0 {
				c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				var settings resource.Settings
				if json.Unmarshal([]byte(f.Input), &settings) != nil {
					return contracttest.Observation{Err: resource.ErrInput}
				}
				switch f.Operation {
				case "read":
					if f.Selector == "relative" {
						v, r, e := s.GetByID(ctx, f.ID)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.GetByIdentity(ctx, f.Identity)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "create":
					text, _ := settings.Text.Value()
					input := resource.CreateRequest{Text: text, Description: settings.Description, Controls: controls}
					if f.Selector == "relative" {
						v, r, e := s.CreateByID(ctx, f.ID, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.CreateByIdentity(ctx, f.Identity, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "patch":
					input := resource.PatchRequest{Settings: settings, Controls: controls}
					if f.Selector == "relative" {
						input.BodyID = policy.Null[string]()
						if f.ReadID != "" {
							input.BodyID = policy.Set(f.ReadID)
						}
						v, r, e := s.PatchByID(ctx, f.ID, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					input.BodyIdentity = policy.Set("read-model-identity")
					v, r, e := s.PatchByIdentity(ctx, f.Identity, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					if f.Selector == "relative" {
						v, r, e := s.DeleteByID(ctx, f.ID, controls)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.DeleteByIdentity(ctx, f.Identity, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				return contracttest.Observation{Err: resource.ErrInput}
			}
			run(t, c)
		}
	}
}
func TestUnitBannerFiltersUseRealResponses(t *testing.T) {
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
			Name: "real-filter-" + tc.Name, Family: "teams_custom_banner_text", Scope: "service", Source: "msteams/teams_admin_api/teams_custom_banner_text/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/CustomBannerText", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
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
