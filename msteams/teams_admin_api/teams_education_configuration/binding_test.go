package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

const bindingSource = "msteams/teams_admin_api/teams_education_configuration/testdata/get_global.json"

type bindingRequest struct {
	Name, Command string
	Arguments     map[string]json.RawMessage
	Observed      struct {
		Failure *string

		Output *struct {
			ParameterSet string `json:"parameter_set"`
			Bound        map[string]json.RawMessage
			Specified    bool
		}
	}
}

// HTTP branches preserve controlled 400 request probes; local binding failures
// remain zero-request checks. No HTTP success or model decoding is asserted here.
func TestUnitEducationConfigurationRejectedRequestsAndLocalBinding(t *testing.T) {
	t.Parallel()

	fixture := struct{ Cases []bindingRequest }{
		Cases: []bindingRequest{
			{
				Name:      "list",
				Command:   "Get-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound:        map[string]json.RawMessage{},
					},
				},
			},
			{
				Name:    "identity",
				Command: "Get-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Global\""),
						},
					},
				},
			},
			{
				Name:    "empty-identity",
				Command: "Get-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "filter",
				Command: "Get-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"Fixture*\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Filter",
						Bound: map[string]json.RawMessage{
							"Filter": json.RawMessage("\"Fixture*\""),
						},
					},
				},
			},
			{
				Name:    "empty-filter",
				Command: "Get-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "identity-only",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Global\""),
						},
						Specified: false,
					},
				},
			},
			{
				Name:    "force-only",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Force":    json.RawMessage("{\"IsPresent\":false}"),
							"Identity": json.RawMessage("\"Global\""),
						},
						Specified: false,
					},
				},
			},
			{
				Name:    "empty-identity",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"\""),
					"UpdateParentInformation": json.RawMessage("\"Future\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "numeric-string",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Global\""),
					"UpdateParentInformation": json.RawMessage("0"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Global\""),
							"UpdateParentInformation": json.RawMessage("\"0\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "all-settings",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"EduGenerativeAIEnhancements":          json.RawMessage("\"FutureEnhancement\""),
					"Identity":                             json.RawMessage("\"Global\""),
					"ParentGuardianPreferredContactMethod": json.RawMessage("\"FutureContact\""),
					"UpdateParentInformation":              json.RawMessage("\"FutureUpdate\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"EduGenerativeAIEnhancements":          json.RawMessage("\"FutureEnhancement\""),
							"Identity":                             json.RawMessage("\"Global\""),
							"ParentGuardianPreferredContactMethod": json.RawMessage("\"FutureContact\""),
							"UpdateParentInformation":              json.RawMessage("\"FutureUpdate\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "force-true",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Force":                   json.RawMessage("true"),
					"Identity":                json.RawMessage("\"Global\""),
					"UpdateParentInformation": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Force":                   json.RawMessage("{\"IsPresent\":true}"),
							"Identity":                json.RawMessage("\"Global\""),
							"UpdateParentInformation": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "ParentGuardianPreferredContactMethod-null",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                             json.RawMessage("\"Global\""),
					"ParentGuardianPreferredContactMethod": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity":                             json.RawMessage("\"Global\""),
							"ParentGuardianPreferredContactMethod": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "ParentGuardianPreferredContactMethod-empty",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                             json.RawMessage("\"Global\""),
					"ParentGuardianPreferredContactMethod": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity":                             json.RawMessage("\"Global\""),
							"ParentGuardianPreferredContactMethod": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "EduGenerativeAIEnhancements-null",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"EduGenerativeAIEnhancements": json.RawMessage("null"),
					"Identity":                    json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"EduGenerativeAIEnhancements": json.RawMessage("null"),
							"Identity":                    json.RawMessage("\"Global\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "EduGenerativeAIEnhancements-empty",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"EduGenerativeAIEnhancements": json.RawMessage("\"\""),
					"Identity":                    json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"EduGenerativeAIEnhancements": json.RawMessage("\"\""),
							"Identity":                    json.RawMessage("\"Global\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "UpdateParentInformation-null",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Global\""),
					"UpdateParentInformation": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Global\""),
							"UpdateParentInformation": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "UpdateParentInformation-empty",
				Command: "Set-CsTeamsEducationConfiguration",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Global\""),
					"UpdateParentInformation": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}
				}{
					Output: &struct {
						ParameterSet string `json:"parameter_set"`
						Bound        map[string]json.RawMessage
						Specified    bool
					}{
						ParameterSet: "Identity",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Global\""),
							"UpdateParentInformation": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
		},
	}

	for _, f := range fixture.Cases {

		bound := f.Arguments
		set := "Identity"
		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound
			set = f.Observed.Output.ParameterSet
		} else if _, ok := bound["Filter"]; ok {
			set = "Filter"
		}
		var identity, pattern string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		verb, _, _ := strings.Cut(f.Command, "-")
		c := contracttest.Case{Name: "binding-" + strings.ToLower(verb+"-"+f.Name), Source: bindingSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: set}}}
		var settings resource.Settings
		controls := policy.Controls{}
		if f.Command == setCommand {
			encoded, _ := json.Marshal(bound)
			if json.Unmarshal(encoded, &settings) != nil {
				t.Fatal("invalid typed bound setting")
			}
			if force, ok := bound["Force"]; ok {
				var value struct{ IsPresent bool }
				if json.Unmarshal(force, &value) != nil {
					t.Fatal("invalid source switch")
				}
				controls.Force = new(value.IsPresent)
			}
		}

		invalid := f.Observed.Failure != nil || f.Command == setCommand && (f.Observed.Output == nil || !f.Observed.Output.Specified)
		if invalid {
			c.Want.Error = resource.ErrInput
		} else if f.Command == setCommand {
			body := map[string]json.RawMessage{"Identity": bound["Identity"]}
			for _, name := range []string{"ParentGuardianPreferredContactMethod", "EduGenerativeAIEnhancements", "UpdateParentInformation"} {
				if value, ok := bound[name]; ok {
					body[name] = value
				}
			}
			if controls.Force != nil {
				value := "False"
				if *controls.Force {
					value = "True"
				}
				body["Force"], _ = json.Marshal(value)
			}
			encoded, _ := json.Marshal(body)
			c.Steps = []contracttest.Exchange{compiledWire("PATCH", item, setCommand, encoded, 400, "")}
		} else {
			path, reply := collection, "[]"
			if identity != "" {
				path = item
				reply = `{"Identity":"Global"}`
			}
			c.Steps = []contracttest.Exchange{compiledWire("GET", path, getCommand, nil, 400, reply)}
		}
		if len(c.Steps) != 0 {
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
			if f.Command == setCommand {
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: settings, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			if _, ok := bound["Filter"]; ok {
				v, r, e := s.ListFiltered(ctx, pattern)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			if _, ok := bound["Identity"]; ok {
				v, r, e := s.GetByIdentity(ctx, identity)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			v, r, e := s.List(ctx)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}
		run(t, c)
	}
}

func TestUnitEducationConfigurationFiltersUseRealResponses(t *testing.T) {
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
			Name: "real-filter-" + tc.Name, Family: "teams_education_configuration", Scope: "service", Source: "msteams/teams_admin_api/teams_education_configuration/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsEducationConfiguration", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.TeamsEducationConfigurations
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
