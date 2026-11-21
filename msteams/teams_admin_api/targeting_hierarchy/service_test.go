package targetinghierarchy_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
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
		c.Source = "contracts/endpoints/8.1.0/special-service-profiles.json"
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
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("targeting-hierarchy-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("hierarchy evidence write failed")
	}
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
	var results []contracttest.Result
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
			results = append(results, run(t, contracttest.Case{Name: tc.name, Mappings: mapping("Get-TeamTargetingHierarchyStatus"), Steps: []contracttest.Exchange{wire(tc.version, "GET", tc.path, 200, `{"id":"operation","status":"FutureStatus","filename":"","future":{"v":8}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
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
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"operation","status":"FutureStatus","raw":{"id":"operation","status":"FutureStatus","filename":"","future":{"v":8}}}`)}}))
		})
	}
	results = append(results, run(t, contracttest.Case{Name: "v2-opaque-get", Steps: []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 200, "opaque-response-schema")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().GetHierarchy(ctx)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 200, Body: []byte("opaque-response-schema")}}))
	results = append(results, run(t, contracttest.Case{Name: "v2-native-create", Steps: []contracttest.Exchange{createWire(201, "opaque-create-schema")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().CreateHierarchy(ctx)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 201, Body: []byte("opaque-create-schema")}}))
	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		path := "/api/rest/v1/nodes/hierarchy"
		if version == hierarchy.V2 {
			path = hierarchyPath()
		}
		results = append(results, run(t, contracttest.Case{Name: "delete-v" + string(version), Mappings: mapping("Remove-TeamTargetingHierarchy"), Steps: []contracttest.Exchange{wire(version, "DELETE", path, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingHierarchy.Delete(ctx, version)
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 204}}))
	}
	evidence(t, results)
}

func TestUnitNativeImportsStreamOriginalCSVWithoutPreparation(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		t.Run(string(version), func(t *testing.T) {
			body := `{"requestId":"operation","future":{"v":8}}`
			if version == hierarchy.V2 {
				body = `{"id":"operation","future":{"v":8}}`
			}
			results = append(results, run(t, contracttest.Case{Name: "native-import-v" + string(version), Steps: []contracttest.Exchange{uploadWire(version, 202, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
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
			}, Want: contracttest.Expected{Status: 202, JSON: json.RawMessage(`"operation"`)}}))
		})
	}
	evidence(t, results)
}

func TestUnitUploadCompositionCreatesOnlyAfter404AndRetainsPartialState(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name    string
		version hierarchy.Version
		steps   []contracttest.Exchange
		status  int
		want    string
		err     error
	}{
		{"default-v1", "", []contracttest.Exchange{uploadWire(hierarchy.V1, 202, `{}`)}, 202, `{"found":false,"create_attempted":false,"created":false,"import_attempted":true,"accepted":true,"request_id":null}`, nil},
		{"v2-existing", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 200, `{}`), uploadWire(hierarchy.V2, 202, `{"id":"operation"}`)}, 202, `{"found":true,"create_attempted":false,"created":false,"import_attempted":true,"accepted":true,"request_id":"operation"}`, nil},
		{"v2-created", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 404, `{"code":"NotFound"}`), createWire(201, `{}`), uploadWire(hierarchy.V2, 202, `{"id":"operation"}`)}, 202, `{"found":false,"create_attempted":true,"created":true,"import_attempted":true,"accepted":true,"request_id":"operation"}`, nil},
		{"v2-read-denied", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 403, `{"code":"Denied"}`)}, 403, `{"found":false,"create_attempted":false,"created":false,"import_attempted":false,"accepted":false,"request_id":null}`, &client.APIError{Backend: client.Hierarchy, StatusCode: 403, Codes: []string{"Denied"}}},
		{"v2-create-ambiguous", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 404, `{"code":"NotFound"}`), createWire(503, `{"code":"Unavailable"}`)}, 503, `{"found":false,"create_attempted":true,"created":false,"import_attempted":false,"accepted":false,"request_id":null}`, &client.APIError{Backend: client.Hierarchy, StatusCode: 503, Codes: []string{"Unavailable"}}},
		{"v2-import-failed-after-create", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 404, `{"code":"NotFound"}`), createWire(201, `{}`), uploadWire(hierarchy.V2, 503, `{"code":"Unavailable"}`)}, 503, `{"found":false,"create_attempted":true,"created":true,"import_attempted":true,"accepted":false,"request_id":null}`, &client.APIError{Backend: client.Hierarchy, StatusCode: 503, Codes: []string{"Unavailable"}}},
		{"v2-accepted-missing-id", hierarchy.V2, []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath(), 200, `{}`), uploadWire(hierarchy.V2, 202, `{"future":false}`)}, 202, `{"found":true,"create_attempted":false,"created":false,"import_attempted":true,"accepted":true,"request_id":null}`, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := contracttest.Expected{Status: tc.status, JSON: json.RawMessage(tc.want)}
			var apiError *client.APIError
			if errors.As(tc.err, &apiError) {
				want.APIError = apiError
			} else {
				want.Error = tc.err
			}
			results = append(results, run(t, contracttest.Case{Name: tc.name, Mappings: mapping("Set-TeamTargetingHierarchy"), Steps: tc.steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s, input := source()
				v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, tc.version)
				if s.closes != 1 {
					t.Error("composition lost source ownership")
				}
				if v.CreationHTTPAccepted && (v.LookupResponse == nil || v.LookupResponse.StatusCode() != 404 || v.CreationResponse == nil || v.CreationResponse.StatusCode() != 201) {
					t.Error("created hierarchy lost preparation metadata")
				}
				return uploaded(v, err)
			}, Want: want}))
		})
	}
	evidence(t, results)
}
