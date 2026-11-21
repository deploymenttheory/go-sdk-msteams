package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
	if !json.Valid(raw) {
		t.Fatal("invalid permanent response fixture")
	}
	return raw
}

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []contracttest.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []contracttest.Exchange
	return append(steps, contracttest.Exchange{
		Request: contracttest.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-CsTeamsCustomBannerText"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitListResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var expected []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil || expected == nil {
		t.Fatal("actual list fixture must retain a JSON array")
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
	result, err := contracttest.Run(t.Context(), contracttest.Case{
		Name: "list", Family: "teams_custom_banner_text", Scope: "service", Source: "msteams/teams_admin_api/teams_custom_banner_text/testdata/list.json",
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/CustomBannerText", 200, raw),
		Want:  contracttest.Expected{Status: http.StatusOK, Body: raw, JSON: want},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.List(ctx)
			if err == nil {
				if len(values) != len(expected) {
					t.Fatal("real list item count changed")
				}
				for _, value := range values {
					if value == nil || value.Identity == nil || *value.Identity == "" {
						t.Fatal("list lost its typed resource identity")
					}
				}
			}
			return contracttest.Observation{Value: values, Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}
