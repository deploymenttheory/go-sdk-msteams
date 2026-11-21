package core_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// TestUnitTemplatePrimitiveRequestShapes checks scope routing independently of
// successful template responses, which require tenant acceptance fixtures.
func TestUnitTemplatePrimitiveRequestShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name, Method, URL string
		Body              json.RawMessage
	}{
		{"template-read-opaque-id", "GET", "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", nil},
		{"template-list", "GET", "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/api/TeamTemplates/v1.0/en-US", nil},
		{"template-delete-empty", "DELETE", "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", nil},
		{"template-create", "POST", "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/api/TeamTemplates/v1.0/Tenant/en-US", json.RawMessage(`{"displayName":"Fixture","shortDescription":"Summary"}`)},
		{"template-update", "PUT", "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", json.RawMessage(`{"displayName":"Fixture","shortDescription":"Summary"}`)},
	}
	mappings := map[string]contracttest.Mapping{
		"template-read-opaque-id": {Command: "Get-CsTeamTemplate", ParameterSet: "Get"},
		"template-list":           {Command: "Get-CsTeamTemplateList", ParameterSet: "DefaultLocaleOverride"},
		"template-delete-empty":   {Command: "Remove-CsTeamTemplate", ParameterSet: "Delete"},
		"template-create":         {Command: "New-CsTeamTemplate", ParameterSet: "NewExpanded"},
		"template-update":         {Command: "Update-CsTeamTemplate", ParameterSet: "UpdateExpanded"},
	}
	verifyMappings(t, mappings)
	count := 0
	results := []contracttest.Result{}
	for _, fixture := range cases {
		mapping, selected := mappings[fixture.Name]
		if !selected {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			parsed, err := url.Parse(fixture.URL)
			if err != nil {
				t.Fatal("fixture URL is invalid")
			}
			segments := strings.Split(strings.TrimPrefix(parsed.EscapedPath(), "/"), "/")
			for index, segment := range segments {
				segments[index], err = url.PathUnescape(segment)
				if err != nil {
					t.Fatal("fixture path encoding is invalid")
				}
			}
			var requestBody any
			var expectedBody json.RawMessage
			if len(fixture.Body) != 0 {
				requestBody, expectedBody = fixture.Body, fixture.Body
			}
			want := contracttest.Expected{Status: 403, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}}
			c := contracttest.Case{Name: fixture.Name, Family: "team_templates", Scope: "primitive",
				Source:   "tests/contract/core/fixtures_test.go",
				Mappings: []contracttest.Mapping{mapping},
				Steps: []contracttest.Exchange{{Request: contracttest.WireRequest{Method: fixture.Method,
					URL: fixture.URL, Audience: auth.TeamsAdminAudience, JSON: expectedBody},
					Reply: contracttest.Reply{Status: 403, Headers: http.Header{"Content-Type": {"application/json"}}}}},
				Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					var value json.RawMessage
					response, err := sdk.TeamsAdminAPI.Scope().Do(ctx, client.Request{Method: fixture.Method,
						Path: segments, Query: parsed.Query(), Body: requestBody}, &value)
					return contracttest.Observation{Value: value, Response: response, Err: err}
				}}
			result, err := contracttest.Run(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			results = append(results, result)
			if !result.Passed() {
				t.Errorf("template request differed: %+v", result.Differences())
			}
		})
	}
	if count != len(mappings) || len(results) != count {
		t.Fatal("one or more template request vectors disappeared")
	}
}

func verifyMappings(t *testing.T, mappings map[string]contracttest.Mapping) {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var runtime struct {
		Commands []struct {
			Name string `json:"name"`
			Sets []struct {
				Name string `json:"name"`
			} `json:"parameter_sets"`
		} `json:"commands"`
	}
	if json.Unmarshal(raw, &runtime) != nil {
		t.Fatal("pinned runtime metadata is invalid")
	}
	known := map[contracttest.Mapping]bool{}
	for _, command := range runtime.Commands {
		for _, set := range command.Sets {
			known[contracttest.Mapping{Command: command.Name, ParameterSet: set.Name}] = true
		}
	}
	for _, mapping := range mappings {
		if !known[mapping] {
			t.Fatal("fixture command/parameter-set mapping is absent from the pinned runtime")
		}
	}
}
