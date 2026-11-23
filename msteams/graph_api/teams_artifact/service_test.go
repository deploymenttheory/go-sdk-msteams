package teamsartifact_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ta "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/teams_artifact"
)

const (
	graph   = "https://graph.microsoft.com/v1.0"
	source  = "msteams/graph_api/teams_artifact/service_test.go"
	command = "Get-TeamsArtifacts"
	site    = "fixture.example.invalid,00000000-0000-4000-8000-000000000011,00000000-0000-4000-8000-000000000012"
	drive   = "b!fixture-drive"
	user    = "00000000-0000-4000-8000-000000000021"
)

func get(target string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: target, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func children(drivePath, item string) string {
	segments := strings.Split(drivePath, "/")
	for index := range segments {
		segments[index] = url.PathEscape(segments[index])
	}
	return graph + "/" + strings.Join(segments, "/") + "/items/" + url.PathEscape(item) + "/children"
}

func folder(id, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"folder":{"childCount":1}}`, id, name)
}

func file(id, name, modified string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"file":{"mimeType":"application/octet-stream"},"@microsoft.graph.downloadUrl":"https://download.example.invalid/%s","lastModifiedDateTime":%q}`, id, name, id, modified)
}

func page(items ...string) string {
	body := `{"value":[`
	for index, item := range items {
		if index > 0 {
			body += ","
		}
		body += item
	}
	return body + `]}`
}

type projection struct {
	Artifacts []map[string]any `json:"artifacts"`
	Failures  []map[string]any `json:"failures"`
}

func project(result ta.Result) projection {
	out := projection{Artifacts: []map[string]any{}, Failures: []map[string]any{}}
	for _, item := range result.Artifacts {
		out.Artifacts = append(out.Artifacts, map[string]any{"name": item.Name, "mime": item.MimeType, "url": item.DownloadURL, "modified": item.LastModifiedDateTime, "drive": item.Drive, "folder": item.Folder})
	}
	for _, failure := range result.Failures {
		var api *client.APIError
		status := 0
		if errors.As(failure.Err, &api) {
			status = api.StatusCode
		}
		out.Failures = append(out.Failures, map[string]any{"stage": failure.Stage, "id": failure.ID, "status": status})
	}
	return out
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "teams_artifacts"
	c.Mappings = []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("teams artifact contract %s differed: %+v", c.Name, result.Differences())
	}
}

// The default scan visits SharePoint site drives, then every user's OneDrive,
// applying the module's folder recursion, extension and time-window rules.
// A failed drive is recorded and skipped.
func TestUnitTeamsArtifactsScanFollowsTheModuleRules(t *testing.T) {
	t.Parallel()
	siteDrive := "drives/" + drive
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	denied := `{"error":{"code":"accessDenied","message":"Fixture message"}}`
	want := []byte(`{"artifacts":[
		{"name":"a.mp4","mime":"application/octet-stream","url":"https://download.example.invalid/f1","modified":"2026-01-01T00:00:00Z","drive":"drives/b!fixture-drive","folder":["Recordings"]},
		{"name":"c.MP4","mime":"application/octet-stream","url":"https://download.example.invalid/f3","modified":"2026-01-01T00:00:00Z","drive":"drives/b!fixture-drive","folder":["Recordings","View Only"]},
		{"name":"n.loop","mime":"application/octet-stream","url":"https://download.example.invalid/f4","modified":"2026-02-01T10:00:00.5Z","drive":"drives/b!fixture-drive","folder":["Documents","Meetings"]}],
		"failures":[{"stage":"drive","id":"users/00000000-0000-4000-8000-000000000021/drive","status":403}]}`)
	run(t, contracttest.Case{Name: "default-scan",
		Steps: []contracttest.Exchange{
			get(graph+"/sites?search=%2A", 200, page(`{"id":""}`, fmt.Sprintf(`{"id":%q,"displayName":"Fixture"}`, site))),
			get(graph+"/sites/"+url.PathEscape(site)+"/drives", 200, page(fmt.Sprintf(`{"id":%q}`, drive))),
			get(children(siteDrive, "root"), 200, page(folder("r1", "Recordings"), folder("d1", "Documents"), file("f0", "root.mp4", "2026-01-01T00:00:00Z"))),
			get(children(siteDrive, "r1"), 200, page(file("f1", "a.mp4", "2026-01-01T00:00:00Z"), file("f2", "b.txt", "2026-01-01T00:00:00Z"), folder("v1", "View Only"))),
			get(children(siteDrive, "v1"), 200, page(file("f3", "c.MP4", "2026-01-01T00:00:00Z"))),
			get(children(siteDrive, "d1"), 200, page(folder("m1", "Meetings"))),
			get(children(siteDrive, "m1"), 200, page(file("f4", "n.loop", "2026-02-01T10:00:00.5Z"), file("f5", "old.loop", "2020-01-01T00:00:00Z"), folder("s1", "sub"))),
			get(graph+"/users", 200, page(fmt.Sprintf(`{"id":%q}`, user))),
			get(children("users/"+user+"/drive", "root"), 403, denied),
		},
		Want: contracttest.Expected{JSON: want},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, ta.Options{Start: &start})
			if err != nil || !errors.Is(result.Failures[0].Err, auth.ErrForbidden) || result.Artifacts[0].Item.RawFields["file"] == nil {
				t.Fatal("scan failed or lost the failure classification")
			}
			if fmt.Sprintf("%v %+v %#v", result.Artifacts[0], result.Artifacts[0], result.Artifacts[0]) != "teams artifact (redacted) teams artifact (redacted) teams artifact (redacted)" {
				t.Fatal("artifact formatting exposed its download URL")
			}
			return contracttest.Observation{Value: project(result)}
		}})
}

func TestUnitTeamsArtifactsSharePointDenialUsesLabResponse(t *testing.T) {
	t.Parallel()
	denied, err := os.ReadFile("testdata/sites_403.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, contracttest.Case{Name: "sharepoint-forbidden", Source: "msteams/graph_api/teams_artifact/testdata/sites_403.json",
		Steps: []contracttest.Exchange{get(graph+"/sites?search=%2A", 403, string(denied))},
		Want:  contracttest.Expected{JSON: []byte(`{"artifacts":[],"failures":[{"stage":"sites","id":"","status":403}]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, ta.Options{SharePoint: true})
			var api *client.APIError
			if err != nil || !errors.As(result.Failures[0].Err, &api) || api.Codes[0] != "accessDenied" || !errors.Is(api, auth.ErrForbidden) {
				t.Fatal("site search denial was not classified")
			}
			return contracttest.Observation{Value: project(result)}
		}})
}

func TestUnitTeamsArtifactsSelectorsAndPaging(t *testing.T) {
	t.Parallel()
	filter := url.Values{"$filter": {"mail eq 'o''neil@example.invalid'"}}.Encode()
	run(t, contracttest.Case{Name: "onedrive-mail",
		Steps: []contracttest.Exchange{get(graph+"/users?"+filter, 200, page())},
		Want:  contracttest.Expected{JSON: []byte(`{"artifacts":[],"failures":[]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, ta.Options{OneDrive: "o'neil@example.invalid"})
			if err != nil {
				t.Fatal(err)
			}
			return contracttest.Observation{Value: project(result)}
		}})

	whiteboard := ta.Whiteboard
	siteDrive := "drives/" + drive
	next := children(siteDrive, "w1") + "?$skiptoken=fixture"
	run(t, contracttest.Case{Name: "type-filter-and-next-link",
		Steps: []contracttest.Exchange{
			get(graph+"/sites?search=%2A", 200, page(fmt.Sprintf(`{"id":%q}`, site))),
			get(graph+"/sites/"+url.PathEscape(site)+"/drives", 200, page(fmt.Sprintf(`{"id":%q}`, drive))),
			get(children(siteDrive, "root"), 200, page(folder("r1", "Recordings"), folder("w1", "Whiteboards"))),
			get(children(siteDrive, "w1"), 200, `{"value":[`+file("f1", "one.whiteboard", "2026-01-01T00:00:00Z")+`],"@odata.nextLink":"`+next+`"}`),
			get(next, 200, page(file("f2", "two.WhiteBoard", "2026-01-01T00:00:00Z"), file("f3", "late.whiteboard", "2027-01-01T00:00:00Z"))),
		},
		Want: contracttest.Expected{JSON: []byte(`{"artifacts":[
			{"name":"one.whiteboard","mime":"application/octet-stream","url":"https://download.example.invalid/f1","modified":"2026-01-01T00:00:00Z","drive":"drives/b!fixture-drive","folder":["Whiteboards"]},
			{"name":"two.WhiteBoard","mime":"application/octet-stream","url":"https://download.example.invalid/f2","modified":"2026-01-01T00:00:00Z","drive":"drives/b!fixture-drive","folder":["Whiteboards"]}],"failures":[]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			end := time.Date(2026, 6, 1, 0, 0, 0, 0, time.FixedZone("x", 3600))
			result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, ta.Options{SharePoint: true, Type: &whiteboard, End: &end})
			if err != nil {
				t.Fatal(err)
			}
			return contracttest.Observation{Value: project(result)}
		}})
}

func TestUnitTeamsArtifactsRejectsInvalidOptions(t *testing.T) {
	t.Parallel()
	bad := ta.Type(7)
	for name, options := range map[string]ta.Options{
		"both-sources": {OneDrive: "a@example.invalid", SharePoint: true},
		"blank-mail":   {OneDrive: " "},
		"control-mail": {OneDrive: "a\n@example.invalid"},
		"bad-type":     {Type: &bad},
		"bad-paging":   {Paging: client.PagingOptions{Style: client.TokenPaging}},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Steps: []contracttest.Exchange{},
			Want: contracttest.Expected{Error: ta.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, options)
				return contracttest.Observation{Err: err}
			}})
	}
	var s *ta.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, err := s.List(nil, ta.Options{}); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.List(t.Context(), ta.Options{}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
}
