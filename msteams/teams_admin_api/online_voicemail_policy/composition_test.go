package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitVoicemailUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expectedBefore map[string]json.RawMessage
	if json.Unmarshal(before, &expectedBefore) != nil {
		t.Fatal("invalid real preparation fixture")
	}
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases,
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-denied", 200, 403, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-throttled", 200, 429, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-unavailable", 200, 503, string(before)},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", item, "Set-CsOnlineVoicemailPolicy", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsOnlineVoicemailPolicy", json.RawMessage(`{"Identity":"Global","EnableTranscription":false,"MaximumRecordingLength":null}`), tc.patch, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/online_voicemail_policy/composition_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{EnableTranscription: policy.Set(false), MaximumRecordingLength: policy.Null[resource.TimeSpan]()}})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.patch != 0 && (v.Before == nil || !reflect.DeepEqual(v.Before.RawFields, expectedBefore)) {
					t.Error("failed update changed or lost the actual preparation snapshot")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitVoicemailCreateComparisonDoesNotCloneDefaultsOrBlockSubmissionOnDeniedSubmission(t *testing.T) {
	t.Parallel()
	for _, readStatus := range []int{200, 403} {
		t.Run(strconv.Itoa(readStatus), func(t *testing.T) {
			readBody := string(readResponseFixture(t, "get_global.json"))
			if readStatus != 200 {
				readBody = ""
			}
			run(t, contracttest.Case{Name: "comparison-denial", Source: "msteams/teams_admin_api/online_voicemail_policy/composition_test.go", Steps: []contracttest.Exchange{wire("GET", item, "New-CsOnlineVoicemailPolicy", nil, readStatus, readBody), wire("POST", collection, "New-CsOnlineVoicemailPolicy", json.RawMessage(`{"Identity":"Fixture","Description":""}`), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: resource.Settings{Description: policy.Set("")}})
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != 403 || v == nil || !v.CreateInvoked || v.HTTPAccepted || v.Created != nil || (v.ComparisonError != nil) != (readStatus != 200) {
					t.Error("comparison and denied submission outcomes were conflated")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitVoicemailListFilteringAndExactGlobalOrderFromRealResponses(t *testing.T) {
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
			Name: "real-filter-" + tc.Name, Family: "online_voicemail_policy", Scope: "service", Source: "msteams/teams_admin_api/online_voicemail_policy/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/OnlineVoicemailPolicy", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
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

func TestUnitVoicemailDeniedBindingAndOperationReadReuseDiscovery(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "denied-group-and-operation", Source: "msteams/teams_admin_api/online_voicemail_policy/composition_test.go", Steps: []contracttest.Exchange{
		discovery(t),
		generatedWire("POST", origin+"/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy", "Grant-CsOnlineVoicemailPolicy", json.RawMessage(`{"policyName":""}`), 403, ""),
		generatedWire("GET", origin+"/Skype.Policy/assignments/operations/fixture%2Foperation", "Get-CsBatchPolicyAssignmentOperation", nil, 403, ""),
	}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
		zero := int32(0)
		v, _, e := s.AssignWithPowerShellBinding(ctx, resource.Assignment{PolicyName: policy.Null[string](), Group: &resource.GroupAssignment{ID: "fixture/group", Priority: &zero}})
		if !errors.Is(e, auth.ErrForbidden) || v != nil && v.Accepted {
			t.Fatal("denied group binding invented an accepted operation")
		}
		status, r, e := s.GetAssignmentOperation(ctx, "fixture/operation")
		if status != nil {
			t.Error("denied operation read fabricated status")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
