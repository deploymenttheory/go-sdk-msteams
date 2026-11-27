package core_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
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

	count := 0
	results := []testutil.Result{}
	for _, fixture := range cases {
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
			want := testutil.Expected{Status: 403, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}}
			c := testutil.Case{Name: fixture.Name,

				Steps: []testutil.Exchange{{Request: testutil.WireRequest{Method: fixture.Method,
					URL: fixture.URL, Audience: auth.TeamsAdminAudience, JSON: expectedBody},
					Reply: testutil.Reply{Status: 403, Headers: http.Header{"Content-Type": {"application/json"}}}}},
				Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					var value json.RawMessage
					response, err := sdk.TeamsAdminAPI.Scope().Do(ctx, client.Request{Method: fixture.Method,
						Path: segments, Query: parsed.Query(), Body: requestBody}, &value)
					return testutil.Observation{Value: value, Response: response, Err: err}
				}}
			result, err := testutil.Run(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			results = append(results, result)
			if !result.Passed() {
				t.Errorf("template request differed: %+v", result.Differences())
			}
		})
	}
	if count != len(cases) || len(results) != count {
		t.Fatal("one or more template request vectors disappeared")
	}
}
