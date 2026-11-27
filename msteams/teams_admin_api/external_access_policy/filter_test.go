package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
)

func TestUnitProviderFilteringIsLocalAndGlobalOrderIsExactFromRealResponses(t *testing.T) {
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
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/ExternalAccessPolicy", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.ExternalAccessPolicies
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
func TestUnitWrongBackendAndNilServiceRejectLocally(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "wrong-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).GetByIdentity(ctx, "Global"); !errors.Is(e, client.ErrDestination) {
			return testutil.Observation{Err: errors.New("wrong backend accepted")}
		}
		var missing *resource.Service
		if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
			return testutil.Observation{Err: errors.New("nil service accepted")}
		}
		return testutil.Observation{}
	}})
}
