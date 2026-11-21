package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitValidationUpdatePreparationFailures(t *testing.T) {
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
			steps := []contracttest.Exchange{wire("GET", item, "Set-CsOnlineVoicemailValidationConfiguration", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"Global","AudioFileValidationEnabled":false}`), tc.patch, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/composition_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false)}})
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

func TestUnitValidationListFilteringAndGlobalOrderFromRealResponses(t *testing.T) {
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
			Name: "real-filter-" + tc.Name, Family: "online_voicemail_validation_configuration", Scope: "service", Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations
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
