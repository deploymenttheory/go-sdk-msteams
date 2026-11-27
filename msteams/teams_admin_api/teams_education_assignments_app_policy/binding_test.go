package teamseducationassignmentsapppolicy_test

import (
	"context"
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
)

type bindingRequest struct {
	Name, Command string
	Arguments     map[string]json.RawMessage
	Observed      struct {
		Failure *string

		Output *struct {
			Bound     map[string]json.RawMessage
			Specified bool
		}
	}
}

func TestUnitEducationAssignmentsRequestdBindingValues(t *testing.T) {
	t.Parallel()

	fixture := struct{ Cases []bindingRequest }{
		Cases: []bindingRequest{
			{
				Name:      "list",
				Command:   "Get-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{},
					},
				},
			},
			{
				Name:    "identity",
				Command: "Get-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Global\""),
						},
					},
				},
			},
			{
				Name:    "empty-identity",
				Command: "Get-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "filter",
				Command: "Get-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"Fixture*\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Filter": json.RawMessage("\"Fixture*\""),
						},
					},
				},
			},
			{
				Name:    "empty-filter",
				Command: "Get-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "identity-only",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Global\""),
						},
						Specified: false,
					},
				},
			},
			{
				Name:    "force-only",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Global\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
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
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"\""),
					"MakeCodeEnabledType": json.RawMessage("\"Future\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},
			{
				Name:    "numeric-string",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiKey": json.RawMessage("0"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiKey": json.RawMessage("\"0\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "url-http-binding",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiUrl": json.RawMessage("\"http://fixture.invalid\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiUrl": json.RawMessage("\"http://fixture.invalid\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "url-upper",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiUrl": json.RawMessage("\"HTTPS://fixture.invalid\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiUrl": json.RawMessage("\"HTTPS://fixture.invalid\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "ParentDigestEnabledType-null",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Global\""),
					"ParentDigestEnabledType": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Global\""),
							"ParentDigestEnabledType": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "ParentDigestEnabledType-empty",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Global\""),
					"ParentDigestEnabledType": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Global\""),
							"ParentDigestEnabledType": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "MakeCodeEnabledType-null",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Global\""),
					"MakeCodeEnabledType": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Global\""),
							"MakeCodeEnabledType": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "MakeCodeEnabledType-empty",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Global\""),
					"MakeCodeEnabledType": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Global\""),
							"MakeCodeEnabledType": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInEnabledType-null",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Global\""),
					"TurnItInEnabledType": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Global\""),
							"TurnItInEnabledType": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInEnabledType-empty",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Global\""),
					"TurnItInEnabledType": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Global\""),
							"TurnItInEnabledType": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInApiUrl-null",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiUrl": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiUrl": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInApiUrl-empty",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiUrl": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiUrl": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInApiKey-null",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiKey": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiKey": json.RawMessage("null"),
						},
						Specified: true,
					},
				},
			},
			{
				Name:    "TurnItInApiKey-empty",
				Command: "Set-CsTeamsEducationAssignmentsAppPolicy",
				Arguments: map[string]json.RawMessage{
					"Identity":       json.RawMessage("\"Global\""),
					"TurnItInApiKey": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified bool
					}{
						Bound: map[string]json.RawMessage{
							"Identity":       json.RawMessage("\"Global\""),
							"TurnItInApiKey": json.RawMessage("\"\""),
						},
						Specified: true,
					},
				},
			},
		},
	}

	for _, f := range fixture.Cases {

		bound := f.Arguments

		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound

		}

		var identity, pattern string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		verb, _, _ := strings.Cut(f.Command, "-")
		c := testutil.Case{Name: "binding-" + strings.ToLower(verb+"-"+f.Name)}
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

		invalid := f.Observed.Failure != nil || f.Command == setCommand && (f.Observed.Output == nil || !f.Observed.Output.Specified || f.Name == "url-http-binding" || f.Name == "TurnItInApiUrl-empty")
		if invalid {
			c.Want.Error = resource.ErrInput
		} else if f.Command == setCommand {
			body := map[string]json.RawMessage{"Identity": bound["Identity"]}
			for _, name := range []string{"ParentDigestEnabledType", "MakeCodeEnabledType", "TurnItInEnabledType", "TurnItInApiUrl", "TurnItInApiKey"} {
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
			c.Steps = []testutil.Exchange{compiledWire("PATCH", item, setCommand, encoded, 400, "")}
		} else {
			path, reply := collection, "[]"
			if identity != "" {
				path = item
				reply = `{"Identity":"Global"}`
			}
			c.Steps = []testutil.Exchange{compiledWire("GET", path, getCommand, nil, 400, reply)}
			c.Want.JSON = json.RawMessage(reply)
		}
		if len(c.Steps) != 0 {
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsEducationAssignmentsAppPolicies
			if f.Command == setCommand {
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: settings, Controls: controls})
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			if _, ok := bound["Filter"]; ok {
				v, r, e := s.ListFiltered(ctx, pattern)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			if _, ok := bound["Identity"]; ok {
				v, r, e := s.GetByIdentity(ctx, identity)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			v, r, e := s.List(ctx)
			return testutil.Observation{Value: v, Response: r, Err: e}
		}
		run(t, c)
	}
}

func TestUnitEducationAssignmentsFiltersUseRealResponses(t *testing.T) {
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
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "real-filter-" + tc.Name,
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsEducationAssignmentsAppPolicy", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.TeamsEducationAssignmentsAppPolicies
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return testutil.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}
