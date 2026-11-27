package teamschannelspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func readResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directory.Close(); err != nil {
			t.Error("testdata directory close failed")
		}
	})
	raw, err := directory.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 && strings.HasSuffix(name, ".body") {
		return raw
	}
	if !json.Valid(raw) {
		t.Fatal("invalid permanent response fixture")
	}
	return raw
}

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []testutil.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []testutil.Exchange
	return append(steps, testutil.Exchange{
		Request: testutil.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-CsTeamsChannelsPolicy"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitGetGlobalResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	var expected struct {
		Field *bool `json:"AllowOrgWideTeamCreation"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	result, err := testutil.Run(t.Context(), testutil.Case{
		Name:  "get-global",
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsChannelsPolicy/configuration/Global", 200, raw),
		Want:  testutil.Expected{Status: http.StatusOK, Body: raw, JSON: raw},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsChannelsPolicies.GetByIdentity(ctx, "Global")
			if err == nil {
				if value == nil || value.Identity == nil || *value.Identity != "Global" || !reflect.DeepEqual(value.AllowOrgWideTeamCreation, expected.Field) {
					t.Fatal("typed identity or AllowOrgWideTeamCreation differs from the real response")
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
	wantFields := make(map[string]*bool, len(expected))
	for _, snapshot := range expected {
		var item struct {
			Identity string `json:"Identity"`
			Field    *bool  `json:"AllowOrgWideTeamCreation"`
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
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsChannelsPolicy", 200, raw),
		Want:  testutil.Expected{Status: http.StatusOK, Body: raw, JSON: want},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsChannelsPolicies.List(ctx)
			if err == nil {
				if len(values) != len(expected) {
					t.Fatal("real list item count changed")
				}
				for _, value := range values {
					if value == nil || value.Identity == nil || *value.Identity == "" {
						t.Fatal("list lost its typed resource identity")
					}
					field, present := wantFields[*value.Identity]
					if !present || !reflect.DeepEqual(value.AllowOrgWideTeamCreation, field) {
						t.Fatal("typed AllowOrgWideTeamCreation differs from the real list response")
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
