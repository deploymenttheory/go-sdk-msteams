package targetinghierarchy_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	"github.com/go-resty/resty/v2"
)

const (
	tenant            = "11111111-1111-4111-8111-111111111111"
	retailOrigin      = "https://retailservices.teams.microsoft.com"
	hierarchyOrigin   = "https://hierarchy.teams.microsoft.com"
	hierarchyAudience = "https://chatsvcagg.teams.microsoft.com"
	csvBytes          = "name,parent\r\nFixture,Root\r\n"
)

func hierarchyPath() string { return "/api/v1/hierarchies/" + tenant }
func wire(version hierarchy.Version, method, path string, status int, body string) contracttest.Exchange {
	origin, audience := retailOrigin, retailOrigin
	if version == hierarchy.V2 {
		origin, audience = hierarchyOrigin, hierarchyAudience
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: audience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func uploadWire(version hierarchy.Version, status int, body string) contracttest.Exchange {
	path := "/api/rest/v1/nodes/upload"
	if version == hierarchy.V2 {
		path = hierarchyPath() + "/importCsv"
	}
	step := wire(version, "POST", path, status, body)
	step.Request.Multipart = []contracttest.MultipartField{{Name: "csvFile", Filename: "fixture.csv", ContentType: "application/json", Body: []byte(csvBytes)}}
	return step
}
func createWire(status int, body string) contracttest.Exchange {
	step := wire(hierarchy.V2, "POST", "/api/v1/hierarchies", status, body)
	step.Request.Headers.Set("Content-Type", "application/json")
	step.Request.JSON = json.RawMessage(`{"id":"` + tenant + `","displayName":"Operational Hierarchy"}`)
	return step
}
func mapping(command string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "targeting_hierarchy"
	if c.Source == "" {
		c.Source = "msteams/teams_admin_api/targeting_hierarchy/service_test.go"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	if c.TenantID == "" {
		c.TenantID = tenant
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("hierarchy contract %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
	return r
}

type csvSource struct {
	io.Reader
	closes   int
	closeErr error
}

func (s *csvSource) Close() error { s.closes++; return s.closeErr }
func source() (*csvSource, hierarchy.CSV) {
	s := &csvSource{Reader: strings.NewReader(csvBytes)}
	return s, hierarchy.CSV{Filename: `C:\private\fixture.csv`, Body: s, MaxBytes: 4096}
}
func uploaded(v hierarchy.UploadResult, err error) contracttest.Observation {
	var id *string
	if v.Import != nil {
		id = v.Import.RequestID
	}
	return contracttest.Observation{Value: map[string]any{"found": v.LookupHTTPAccepted, "create_attempted": v.CreateAttempted, "created": v.CreationHTTPAccepted, "import_attempted": v.ImportAttempted, "accepted": v.ImportHTTPAccepted, "request_id": id}, Response: v.Response, Err: err}
}

func TestUnitNativeStatusRoutesAndIndependentHierarchyPrimitives(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, id, path string
		version        hierarchy.Version
	}{
		{"v1-latest", "", "/api/rest/v1/nodes/upload/status", hierarchy.V1},
		{"v1-opaque-query", "opaque/request?x=1&y=2", "/api/rest/v1/nodes/upload/status?" + url.Values{"requestId": {"opaque/request?x=1&y=2"}}.Encode(), hierarchy.V1},
		{"v2-latest", " \t", hierarchyPath() + "/operations/latest", hierarchy.V2},
		{"v2-opaque-segment", "opaque/request?x=1", hierarchyPath() + "/operations/" + url.PathEscape("opaque/request?x=1"), hierarchy.V2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: mapping("Get-TeamTargetingHierarchyStatus"), Steps: []contracttest.Exchange{wire(tc.version, "GET", tc.path, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				if tc.version == hierarchy.V1 {
					v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV1().GetOperation(ctx, tc.id)
					var value any
					if v != nil {
						value = map[string]any{"id": v.ID, "status": v.Status, "raw": v.RawFields}
					}
					return contracttest.Observation{Value: value, Response: r, Err: err}
				}
				v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().GetOperation(ctx, tc.id)
				var value any
				if v != nil {
					value = map[string]any{"id": v.ID, "status": v.Status, "raw": v.RawFields}
				}
				return contracttest.Observation{Value: value, Response: r, Err: err}
			}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		})
	}
	run(t, contracttest.Case{Name: "v2-opaque-get", Steps: []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().GetHierarchy(ctx)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	run(t, contracttest.Case{Name: "v2-native-create", Steps: []contracttest.Exchange{createWire(403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().CreateHierarchy(ctx)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		path := "/api/rest/v1/nodes/hierarchy"
		if version == hierarchy.V2 {
			path = hierarchyPath()
		}
		run(t, contracttest.Case{Name: "delete-v" + string(version), Mappings: mapping("Remove-TeamTargetingHierarchy"), Steps: []contracttest.Exchange{wire(version, "DELETE", path, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingHierarchy.Delete(ctx, version)
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	}
}

func TestUnitNativeImportsStreamOriginalCSVWithoutPreparation(t *testing.T) {
	t.Parallel()

	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		t.Run(string(version), func(t *testing.T) {
			run(t, contracttest.Case{Name: "native-import-v" + string(version), Steps: []contracttest.Exchange{uploadWire(version, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s, input := source()
				var value *hierarchy.ImportResult
				var response *resty.Response
				var err error
				if version == hierarchy.V1 {
					value, response, err = sdk.TeamsAdminAPI.TargetingHierarchy.NativeV1().ImportCSV(ctx, input)
				} else {
					value, response, err = sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().ImportCSV(ctx, input)
				}
				if s.closes != 1 {
					t.Error("CSV source not closed exactly once")
				}
				var id *string
				if value != nil {
					id = value.RequestID
				}
				return contracttest.Observation{Value: id, Response: response, Err: err}
			}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		})
	}
}

func TestUnitUploadCompositionDenialsRetainPreparationState(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		version hierarchy.Version
		steps   []contracttest.Exchange
		want    string
		status  int
		err     error
	}{
		{"v1-import-denied", hierarchy.V1, []contracttest.Exchange{uploadWire(hierarchy.V1, 403, "")}, `{"found":false,"create_attempted":false,"created":false,"import_attempted":true,"accepted":false,"request_id":null}`, 403, auth.ErrForbidden},
		{"v2-read-denied", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 403, "")}, `{"found":false,"create_attempted":false,"created":false,"import_attempted":false,"accepted":false,"request_id":null}`, 403, auth.ErrForbidden},
		{"v2-create-failed", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 404, ""), createWire(403, "")}, `{"found":false,"create_attempted":true,"created":false,"import_attempted":false,"accepted":false,"request_id":null}`, 403, auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: mapping("Set-TeamTargetingHierarchy"), Steps: tc.steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s, input := source()
				v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, tc.version)
				if s.closes != 1 {
					t.Error("composition lost source ownership")
				}
				return uploaded(v, err)
			}, Want: contracttest.Expected{Status: tc.status, JSON: json.RawMessage(tc.want), Error: tc.err}})
		})
	}
}

// These protocol vectors follow the pinned handlers/models listed in
// authored-protocol-sources.md. Preparation responses are deliberately opaque;
// an accepted import remains distinct from successful ingestion.
func TestUnitHierarchySuccessfulPreparationImportStatusAndDelete(t *testing.T) {
	t.Parallel()
	const primary = "https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0"
	reviewed := func(step contracttest.Exchange) contracttest.Exchange {
		step.Reply.ProtocolSources = []string{primary}
		return step
	}
	for _, tc := range []struct {
		name    string
		version hierarchy.Version
		steps   []contracttest.Exchange
		want    string
		status  int
		err     error
	}{
		{"v1-optional-handle", hierarchy.V1, []contracttest.Exchange{reviewed(uploadWire(hierarchy.V1, 202, `{}`))}, `{"found":false,"create_attempted":false,"created":false,"import_attempted":true,"accepted":true,"request_id":null}`, 202, nil},
		{"v2-existing", hierarchy.V2, []contracttest.Exchange{reviewed(wire(hierarchy.V2, "GET", hierarchyPath(), 200, "opaque preparation")), reviewed(uploadWire(hierarchy.V2, 202, `{"id":"offline-import"}`))}, `{"found":true,"create_attempted":false,"created":false,"import_attempted":true,"accepted":true,"request_id":"offline-import"}`, 202, nil},
		{"v2-create", hierarchy.V2, []contracttest.Exchange{reviewed(wire(hierarchy.V2, "GET", hierarchyPath(), 404, "")), reviewed(createWire(201, "")), reviewed(uploadWire(hierarchy.V2, 202, `{"id":"offline-import"}`))}, `{"found":false,"create_attempted":true,"created":true,"import_attempted":true,"accepted":true,"request_id":"offline-import"}`, 202, nil},
		{"v2-import-denied-after-create", hierarchy.V2, []contracttest.Exchange{reviewed(wire(hierarchy.V2, "GET", hierarchyPath(), 404, "")), reviewed(createWire(201, "")), uploadWire(hierarchy.V2, 403, "")}, `{"found":false,"create_attempted":true,"created":true,"import_attempted":true,"accepted":false,"request_id":null}`, 403, auth.ErrForbidden},
	} {
		run(t, contracttest.Case{Name: tc.name, Mappings: mapping("Set-TeamTargetingHierarchy"), Steps: tc.steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			src, input := source()
			v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, tc.version)
			if src.closes != 1 {
				t.Fatal("CSV ownership was lost")
			}
			if tc.version == hierarchy.V2 && tc.name != "v2-existing" && (v.CreationResponse == nil || v.CreationResponse.StatusCode() != 201) {
				t.Fatal("successful preparation receipt was lost")
			}
			if tc.name == "v2-existing" && (v.LookupResponse == nil || v.LookupResponse.StatusCode() != 200) {
				t.Fatal("lookup receipt lost")
			}
			if v.ImportHTTPAccepted && tc.version == hierarchy.V2 && (v.Import == nil || string(v.Import.RawFields["id"]) != `"offline-import"`) {
				t.Fatal("native import handle lost")
			}
			return uploaded(v, err)
		}, Want: contracttest.Expected{Status: tc.status, JSON: json.RawMessage(tc.want), Error: tc.err}})
	}
	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		path, body, wantStatus := "/api/rest/v1/nodes/upload/status", `{"id":"offline-operation","status":"Successful","lastUpdatedAt":"2026-11-20T12:00:00Z","filename":"fixture.csv","future":false}`, "Successful"
		deletePath := "/api/rest/v1/nodes/hierarchy"
		if version == hierarchy.V2 {
			path = hierarchyPath() + "/operations/latest"
			body = `{"id":"offline-operation","status":"Succeeded","lastActionDateTime":"2026-11-20T12:00:00Z","fileName":"fixture.csv","future":false}`
			deletePath = hierarchyPath()
		}
		run(t, contracttest.Case{Name: "status-then-delete-v" + string(version), Mappings: append(mapping("Get-TeamTargetingHierarchyStatus"), mapping("Remove-TeamTargetingHierarchy")...), Steps: []contracttest.Exchange{reviewed(wire(version, "GET", path, 200, body)), reviewed(wire(version, "DELETE", deletePath, 204, ""))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, "", version)
			if err != nil {
				return contracttest.Observation{Response: r, Err: err}
			}
			if v == nil || v.ID != "offline-operation" || v.Status == nil || *v.Status != wantStatus || v.Filename == nil || *v.Filename != "fixture.csv" || v.LastUpdatedAt == nil || *v.LastUpdatedAt != "2026-11-20T12:00:00Z" {
				t.Fatal("status projection lost native operation")
			}
			if version == hierarchy.V2 {
				if v.RawStatus == nil || *v.RawStatus != "Succeeded" || v.Hierarchy == nil || string(v.Hierarchy.RawFields["future"]) != "false" {
					t.Fatal("v2 projection lost raw state")
				}
			} else if v.Retail == nil || string(v.Retail.RawFields["future"]) != "false" {
				t.Fatal("v1 projection lost native state")
			}
			r, err = sdk.TeamsAdminAPI.TargetingHierarchy.Delete(ctx, version)
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 204}})
	}
}
