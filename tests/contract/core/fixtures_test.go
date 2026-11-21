package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// TestUnitPinnedTemplatePrimitiveFixtures proves a representative captured contract
// can use the shared runner. It exercises the shared ConfigAPI scope only, not an
// implemented template service or complete PowerShell command composition.
func TestUnitPinnedTemplatePrimitiveFixtures(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Version string `json:"baseline_version"`
		Cases   []struct {
			Name     string          `json:"name"`
			Method   string          `json:"method"`
			URL      string          `json:"generated_url"`
			Body     json.RawMessage `json:"body"`
			Status   int             `json:"response_status"`
			Response string          `json:"response_body"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &capture) != nil || capture.Version != "8.1.0" {
		t.Fatal("pinned request capture is invalid")
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
	for index, fixture := range capture.Cases {
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
			if !bytes.Equal(fixture.Body, []byte("null")) {
				requestBody, expectedBody = fixture.Body, fixture.Body
			}
			want := contracttest.Expected{Status: fixture.Status, Body: []byte(fixture.Response)}
			if fixture.Response != "" {
				want.JSON = json.RawMessage(fixture.Response)
			}
			c := contracttest.Case{Name: fixture.Name, Family: "team_templates", Scope: "primitive",
				Source:   "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json#/cases/" + strconv.Itoa(index),
				Mappings: []contracttest.Mapping{mapping},
				Steps: []contracttest.Exchange{{Request: contracttest.WireRequest{Method: fixture.Method,
					URL: fixture.URL, Audience: auth.TeamsAdminAudience, JSON: expectedBody},
					Reply: contracttest.Reply{Status: fixture.Status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(fixture.Response)}}},
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
				t.Errorf("pinned template primitive differed: %+v", result.Differences())
			}
		})
	}
	if count != len(mappings) || len(results) != count {
		t.Fatal("one or more pinned template fixtures disappeared")
	}
	// CI passes the exact checked-out implementation revision. Ordinary local
	// tests still run every case, but do not fabricate a revision/run artifact.
	if revision := os.Getenv("MSTEAMS_CONTRACT_REVISION"); revision != "" {
		root, err := os.OpenRoot(t.ArtifactDir())
		if err != nil {
			t.Fatal("contract artifact directory unavailable")
		}
		defer func() { _ = root.Close() }()
		file, err := root.Create("core-primitives.json")
		if err != nil {
			t.Fatal("contract artifact file unavailable")
		}
		writeErr := contracttest.WriteEvidence(file, revision, results)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			t.Fatal("contract evidence could not be written")
		}
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
