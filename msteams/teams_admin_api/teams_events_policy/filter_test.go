package teamseventspolicy_test

import (
	"context"
	"encoding/json"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const getCommand = "Get-CsTeamsEventsPolicy"
const grantCommand = "Grant-CsTeamsEventsPolicy"

func TestUnitEventsPolicyFiltersUseRealResponses(t *testing.T) {
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
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsEventsPolicy", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.TeamsEventsPolicies
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

func TestUnitEventsPolicyParameterlessGrantHasNoDispatch(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name, Command string
			Observed      struct {
				Failure       *string
				Sends, Tokens int
				Output        struct {
					NoScope bool `json:"no_scope_result"`
				}
			}
		}
	}{
		Cases: []struct {
			Name, Command string
			Observed      struct {
				Failure       *string
				Sends, Tokens int
				Output        struct {
					NoScope bool `json:"no_scope_result"`
				}
			}
		}{

			{
				Name:    "list",
				Command: "Get-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "filter",
				Command: "Get-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "empty-filter",
				Command: "Get-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Name:    "empty-identity",
				Command: "Get-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Name:    "identity-only",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "empty-identity",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ParameterBindingValidationException"),
				},
			},

			{
				Name:    "identity-only",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "force-only",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "empty-identity",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Name:    "force-false",
				Command: "Remove-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "no-scope",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Exception"),
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: true,
					},
				},
			},

			{
				Name:    "user-omitted-name",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "user-null-name",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "user-empty-name",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "tenant-false-selector",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "tenant-true-force",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "group-rank-zero",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "group-rank-negative",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "group-rank-two",
				Command: "Grant-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{
						NoScope: false,
					},
				},
			},

			{
				Name:    "AllowWebinars-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowWebinars-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "EventAccessType-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "EventAccessType-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowTownhalls-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowTownhalls-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallEventAttendeeAccess-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallEventAttendeeAccess-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ExternalPresenterJoinVerification-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ExternalPresenterJoinVerification-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEmailEditing-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEmailEditing-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedQuestionTypesInRegistrationForm-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedQuestionTypesInRegistrationForm-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEventIntegrations-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEventIntegrations-false",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedWebinarTypesForRecordingPublish-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedWebinarTypesForRecordingPublish-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedTownhallTypesForRecordingPublish-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedTownhallTypesForRecordingPublish-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForTownhall-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForTownhall-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForWebinar-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForWebinar-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForTownhall-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForTownhall-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForWebinar-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForWebinar-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallChatExperience-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallChatExperience-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BroadcastPremiumApps-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BroadcastPremiumApps-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "UseMicrosoftECDN-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "UseMicrosoftECDN-false",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ImmersiveEvents-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ImmersiveEvents-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallMaxResolution-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallMaxResolution-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "HighBitrateForTownhall-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "HighBitrateForTownhall-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BackroomChat-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BackroomChat-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Registration-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Registration-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEngagementReport-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEngagementReport-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "InfoShownInReportMode-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "InfoShownInReportMode-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Description-null",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Description-empty",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "all-settings",
				Command: "New-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowWebinars-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowWebinars-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "EventAccessType-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "EventAccessType-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowTownhalls-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowTownhalls-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallEventAttendeeAccess-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallEventAttendeeAccess-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ExternalPresenterJoinVerification-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ExternalPresenterJoinVerification-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEmailEditing-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEmailEditing-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedQuestionTypesInRegistrationForm-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedQuestionTypesInRegistrationForm-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEventIntegrations-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEventIntegrations-false",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedWebinarTypesForRecordingPublish-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedWebinarTypesForRecordingPublish-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedTownhallTypesForRecordingPublish-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowedTownhallTypesForRecordingPublish-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForTownhall-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForTownhall-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForWebinar-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "RecordingForWebinar-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForTownhall-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForTownhall-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForWebinar-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TranscriptionForWebinar-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallChatExperience-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallChatExperience-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BroadcastPremiumApps-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BroadcastPremiumApps-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "UseMicrosoftECDN-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "UseMicrosoftECDN-false",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ImmersiveEvents-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "ImmersiveEvents-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallMaxResolution-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "TownhallMaxResolution-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "HighBitrateForTownhall-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "HighBitrateForTownhall-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BackroomChat-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "BackroomChat-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Registration-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Registration-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEngagementReport-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "AllowEngagementReport-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "InfoShownInReportMode-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "InfoShownInReportMode-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Description-null",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "Description-empty",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},

			{
				Name:    "all-settings",
				Command: "Set-CsTeamsEventsPolicy",
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        struct {
						NoScope bool `json:"no_scope_result"`
					}
				}{
					Output: struct {
						NoScope bool `json:"no_scope_result"`
					}{},
				},
			},
		},
	}

	found := false
	for _, x := range fixture.Cases {
		if x.Command == grantCommand && x.Name == "no-scope" {
			found = x.Observed.Failure != nil && x.Observed.Sends == 0 && x.Observed.Tokens == 0 && x.Observed.Output.NoScope
		}
	}
	if !found {
		t.Fatal("source no-scope failure not observed")
	}
	run(t, testutil.Case{Name: "grant-no-scope", Want: testutil.Expected{Error: policy.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsEventsPolicies.Assign(ctx, policy.Assignment{})
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}
