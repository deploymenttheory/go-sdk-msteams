package targetinghierarchy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
)

func statusObservation(value *hierarchy.Status) any {
	if value == nil {
		return nil
	}
	return map[string]any{"id": value.ID, "status": value.Status, "last_known_status": value.LastKnownStatus, "error_message": value.ErrorMessage, "last_updated_at": value.LastUpdatedAt, "last_modified_by": value.LastModifiedBy, "filename": value.Filename}
}

func TestStatusProjectionMatchesPinnedCLRStateConstructors(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name  string
			Input json.RawMessage
			Wire  json.RawMessage
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixtures) != nil {
		t.Fatal("invalid pinned model fixture")
	}
	var results []contracttest.Result
	for _, fixture := range fixtures.Cases {
		if !strings.HasPrefix(fixture.Name, "hierarchy-status-projection-") {
			continue
		}
		var input string
		if json.Unmarshal(fixture.Input, &input) != nil {
			t.Fatal("invalid selected model input")
		}
		results = append(results, run(t, contracttest.Case{Name: strings.ToLower(fixture.Name), Source: "contracts/endpoints/8.1.0/fixtures/core-model-serialization.json", Mappings: mapping("Get-TeamTargetingHierarchyStatus"), Steps: []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath()+"/operations/latest", 200, input)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, "", hierarchy.V2)
			var actual any
			if v != nil {
				if v.RawStatus == nil || v.Hierarchy == nil || v.Hierarchy.Status != v.RawStatus {
					t.Error("native status was lost during projection")
				}
				actual = map[string]any{"id": v.ID, "status": v.Status, "lastUpdatedAt": v.LastUpdatedAt, "fileName": v.Filename}
			}
			return contracttest.Observation{Value: actual, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 200, JSON: fixture.Wire}}))
	}
	if len(results) != 6 {
		t.Fatal("pinned state constructor count changed")
	}
	evidence(t, results)
}

func TestStatusErrorProjectionMatchesPinnedCLRConstructors(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/hierarchy-status-projections.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Version string `json:"baseline_version"`
		Hash    string `json:"assembly_sha256"`
		Cases   []struct {
			Name       string
			Version    hierarchy.Version
			Input      string
			Projection map[string]any
		}
	}
	if json.Unmarshal(raw, &fixtures) != nil || fixtures.Version != "8.1.0" || fixtures.Hash != "3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24" {
		t.Fatal("invalid pinned error projection fixture")
	}
	var results []contracttest.Result
	for _, fixture := range fixtures.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			if message, ok := fixture.Projection["error_message"].(string); ok {
				header, body, found := strings.Cut(message, "\r\n")
				if !found {
					t.Fatal("pinned error header line ending differs")
				}
				body = strings.ReplaceAll(body, "\r\n", "\n")
				if runtime.GOOS == "windows" {
					body = strings.ReplaceAll(body, "\n", "\r\n")
				}
				fixture.Projection["error_message"] = header + "\r\n" + body
			}
			want, err := json.Marshal(fixture.Projection)
			if err != nil {
				t.Fatal(err)
			}
			path := "/api/rest/v1/nodes/upload/status"
			if fixture.Version == hierarchy.V2 {
				path = hierarchyPath() + "/operations/latest"
			}
			results = append(results, run(t, contracttest.Case{Name: fixture.Name, Source: "contracts/endpoints/8.1.0/fixtures/hierarchy-status-projections.json", Mappings: mapping("Get-TeamTargetingHierarchyStatus"), Steps: []contracttest.Exchange{wire(fixture.Version, "GET", path, 200, fixture.Input)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, "", fixture.Version)
				return contracttest.Observation{Value: statusObservation(v), Response: r, Err: err}
			}, Want: contracttest.Expected{Status: 200, JSON: want}}))
		})
	}
	if len(results) != 8 {
		t.Fatal("pinned error constructor count changed")
	}
	evidence(t, results)
}

func TestMalformedNestedErrorsRetainReadableNativeOperation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		version    hierarchy.Version
	}{
		{"malformed-json-text", `{"id":"operation","status":"Failed","errorDetails":{"lastKnownStatus":"Ingesting","errorMessage":"not-json"}}`, hierarchy.V1},
		{"missing-v2-status", `{"id":"operation","future":5}`, hierarchy.V2},
	} {
		path := "/api/rest/v1/nodes/upload/status"
		if tc.version == hierarchy.V2 {
			path = hierarchyPath() + "/operations/latest"
		}
		run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{wire(tc.version, "GET", path, 200, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, "", tc.version)
			if v == nil || v.ID != "operation" || v.Retail == nil && v.Hierarchy == nil {
				t.Error("projection failure lost native read")
			}
			for _, format := range []string{"%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, v), "operation\"") || strings.Contains(fmt.Sprintf(format, v), "not-json") {
					t.Error("projection formatting leaked native data")
				}
			}
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
}
