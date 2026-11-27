package teamsnotificationandfeedspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []testutil.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []testutil.Exchange
	return append(steps, testutil.Exchange{
		Request: testutil.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-CsTeamsNotificationAndFeedsPolicy"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitGetGlobalResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	var expected struct {
		Field *string `json:"SuggestedFeedsEnabledType"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	result, err := testutil.Run(t.Context(), testutil.Case{
		Name:  "get-global",
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsNotificationAndFeedsPolicy/configuration/Global", 200, raw),
		Want:  testutil.Expected{Status: http.StatusOK, Body: raw, JSON: raw},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.GetByIdentity(ctx, "Global")
			if err == nil {
				if value == nil || value.Identity == nil || *value.Identity != "Global" || !reflect.DeepEqual(value.SuggestedFeedsEnabledType, expected.Field) {
					t.Fatal("typed identity or SuggestedFeedsEnabledType differs from the real response")
				}
				return testutil.Observation{Value: value.RawFields, Response: response}
			}
			return testutil.Observation{Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}

func TestUnitListResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var expected []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil || expected == nil {
		t.Fatal("actual list fixture must retain a JSON array")
	}
	wantFields := make(map[string]*string, len(expected))
	for _, snapshot := range expected {
		var item struct {
			Identity string  `json:"Identity"`
			Field    *string `json:"SuggestedFeedsEnabledType"`
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil || json.Unmarshal(encoded, &item) != nil || item.Identity == "" {
			t.Fatal("invalid typed list fixture projection")
		}
		wantFields[item.Identity] = item.Field
	}
	// The native List contract stably places Global before other identities.
	slices.SortStableFunc(expected, func(a, b map[string]json.RawMessage) int {
		var left, right string
		if json.Unmarshal(a["Identity"], &left) != nil || json.Unmarshal(b["Identity"], &right) != nil {
			t.Fatal("real list fixture has no comparable identity")
		}
		globalLeft, globalRight := strings.EqualFold(left, "Global"), strings.EqualFold(right, "Global")
		if globalLeft == globalRight {
			return 0
		}
		if globalLeft {
			return -1
		}
		return 1
	})
	want, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	result, err := testutil.Run(t.Context(), testutil.Case{
		Name:  "list",
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsNotificationAndFeedsPolicy", 200, raw),
		Want:  testutil.Expected{Status: http.StatusOK, Body: raw, JSON: want},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.List(ctx)
			if err == nil {
				if len(values) != len(expected) {
					t.Fatal("real list item count changed")
				}
				for _, value := range values {
					if value == nil || value.Identity == nil || *value.Identity == "" {
						t.Fatal("list lost its typed resource identity")
					}
					field, present := wantFields[*value.Identity]
					if !present || !reflect.DeepEqual(value.SuggestedFeedsEnabledType, field) {
						t.Fatal("typed SuggestedFeedsEnabledType differs from the real list response")
					}
				}
			}
			return testutil.Observation{Value: values, Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}
