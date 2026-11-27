package appcatalog_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ac "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_catalog"
)

const (
	graph = "https://graph.microsoft.com"
	route = graph + "/v1.0/appCatalogs/teamsApps"

	appID   = "00000000-0000-4000-8000-000000000001"
	created = `{"id":"00000000-0000-4000-8000-000000000001","externalId":"00000000-0000-4000-8000-000000000002","displayName":"Fixture App","distributionMethod":"organization","future":1}`
)

var pkg = []byte{0x50, 0x4b, 0x05, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	body, err := root.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func reply(status int, body []byte) testutil.Reply {
	if body == nil {
		return testutil.Reply{Status: status}
	}
	return testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}
}

func get(target string, status int, body []byte) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: target, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: reply(status, body)}
}

func upload(method, target string, status int, body []byte) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: target, Audience: auth.GraphAudience,
		Headers: http.Header{"Accept": {"application/json"}, "Content-Type": {"application/zip"}}, Binary: pkg}, Reply: reply(status, body)}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("app catalog contract %s differed: %+v", c.Name, result.Differences())
	}
}

func filtered(filter string) string {
	return route + "?" + url.Values{"$filter": {filter}}.Encode()
}

func TestUnitAppCatalogListBuildsTheModuleFilterAndFollowsPages(t *testing.T) {
	t.Parallel()
	next := route + "?$skiptoken=fixture"
	first := []byte(`{"value":[` + created + `],"@odata.nextLink":"` + next + `"}`)
	second := []byte(`{"value":[{"id":"com.microsoft.teamspace.tab.fixture","distributionMethod":"store"}]}`)
	filter := "id eq 'a' and externalId eq 'b' and displayName eq 'o''brien' and distributionMethod eq 'organization'"
	run(t, testutil.Case{Name: "list-filter-pages",
		Steps: []testutil.Exchange{get(filtered(filter), 200, first), get(next, 200, second)},
		Want:  testutil.Expected{Status: 200, Body: second},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{ID: "a", ExternalID: "b", DisplayName: "o'brien", DistributionMethod: "organization"})
			if err != nil || len(result.Items) != 2 || result.Pages != 2 || result.Items[0] == nil || result.Items[1] == nil || result.Items[0].DisplayName == nil || *result.Items[0].DisplayName != "Fixture App" || result.Items[0].RawFields["future"] == nil ||
				result.Items[1].ID == nil || *result.Items[1].ID != "com.microsoft.teamspace.tab.fixture" || result.Items[1].ExternalID != nil {
				t.Fatal("catalog pages were not collected")
			}
			return testutil.Observation{Response: result.Response, Err: err}
		}})
	empty := []byte(`{"value":[]}`)
	run(t, testutil.Case{Name: "list-unfiltered",
		Steps: []testutil.Exchange{get(route, 200, empty)},
		Want:  testutil.Expected{Status: 200, Body: empty},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{DisplayName: "  "})
			return testutil.Observation{Response: result.Response, Err: err}
		}})
	if (ac.ListOptions{DisplayName: "x"}).Filter() != "displayName eq 'x'" || (ac.ListOptions{}).Filter() != "" {
		t.Fatal("filter text changed")
	}
}

func TestUnitAppCatalogListDenialUsesLabResponse(t *testing.T) {
	t.Parallel()
	denied := readFile(t, "testdata/list_403.json")
	run(t, testutil.Case{Name: "list-forbidden",
		Steps: []testutil.Exchange{get(filtered("distributionMethod eq 'organization'"), 403, denied)},
		Want:  testutil.Expected{Status: 403, Body: denied, APIError: &client.APIError{Backend: client.Graph, StatusCode: 403, Codes: []string{"Forbidden"}, RequestIDs: []string{"00000000-0000-4000-8000-0000000000c1"}, ClientRequestIDs: []string{"00000000-0000-4000-8000-0000000000c1"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{DistributionMethod: "organization"})
			if !errors.Is(err, auth.ErrForbidden) || len(result.Items) != 0 {
				t.Fatal("403 was not classified as forbidden")
			}
			return testutil.Observation{Response: result.Response, Err: err}
		}})
	bad := readFile(t, "testdata/filter_400.json")
	run(t, testutil.Case{Name: "list-bad-request",
		Steps: []testutil.Exchange{get(filtered("displayName eq 'x'"), 400, bad)},
		Want:  testutil.Expected{Status: 400, Body: bad, APIError: &client.APIError{Backend: client.Graph, StatusCode: 400, Codes: []string{"BadRequest"}, RequestIDs: []string{"00000000-0000-4000-8000-0000000000c3"}, ClientRequestIDs: []string{"00000000-0000-4000-8000-0000000000c3"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{DisplayName: "x"})
			return testutil.Observation{Response: result.Response, Err: err}
		}})
}

func TestUnitAppCatalogPackageWrites(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "create-organization",
		Steps: []testutil.Exchange{upload(http.MethodPost, route, 201, []byte(created))},
		Want:  testutil.Expected{Status: 201, Body: []byte(created)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			app, response, err := sdk.GraphAPI.AppCatalog.Create(ctx, ac.CreateRequest{DistributionMethod: "Organization", Package: pkg})
			if err != nil || app == nil || app.ID == nil || *app.ID != appID || app.ExternalID == nil || *app.ExternalID != "00000000-0000-4000-8000-000000000002" {
				t.Fatal("created app was not decoded")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	for _, status := range []int{200, 201, 202, 204, 206, 299} {
		run(t, testutil.Case{Name: "update-" + strconv.Itoa(status),
			Steps: []testutil.Exchange{upload(http.MethodPut, route+"/"+appID, status, nil)},
			Want:  testutil.Expected{Status: status},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				response, err := sdk.GraphAPI.AppCatalog.UpdateByID(ctx, appID, pkg)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	run(t, testutil.Case{Name: "delete",
		Steps: []testutil.Exchange{{Request: testutil.WireRequest{Method: http.MethodDelete, URL: route + "/" + appID, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: reply(204, nil)}},
		Want:  testutil.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.GraphAPI.AppCatalog.DeleteByID(ctx, appID)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitAppCatalogFailuresStayTyped(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "create-accepted-malformed-field", Steps: []testutil.Exchange{upload(http.MethodPost, route, 201, []byte(`{"id":{}}`))}, Want: testutil.Expected{Status: 201, Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		app, response, err := sdk.GraphAPI.AppCatalog.Create(ctx, ac.CreateRequest{DistributionMethod: ac.OrganizationDistribution, Package: pkg})
		if app != nil || response == nil {
			t.Fatal("accepted parse failure lost receipt")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
	for _, item := range []struct {
		name, command string
		invoke        func(context.Context, *ac.Service) error
	}{
		{"create-store", "New-TeamsApp", func(ctx context.Context, s *ac.Service) error {
			_, _, err := s.Create(ctx, ac.CreateRequest{DistributionMethod: "store", Package: pkg})
			return err
		}},
		{"create-omitted-package", "New-TeamsApp", func(ctx context.Context, s *ac.Service) error {
			_, _, err := s.Create(ctx, ac.CreateRequest{DistributionMethod: "organization"})
			return err
		}},
		{"update-slash", "Set-TeamsApp", func(ctx context.Context, s *ac.Service) error { _, err := s.UpdateByID(ctx, "a/b", pkg); return err }},
		{"update-omitted-package", "Set-TeamsApp", func(ctx context.Context, s *ac.Service) error { _, err := s.UpdateByID(ctx, appID, nil); return err }},
		{"delete-blank", "Remove-TeamsApp", func(ctx context.Context, s *ac.Service) error { _, err := s.DeleteByID(ctx, " "); return err }},
	} {
		run(t, testutil.Case{Name: item.name, Want: testutil.Expected{Error: ac.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			return testutil.Observation{Err: item.invoke(ctx, sdk.GraphAPI.AppCatalog)}
		}})
	}
}

func TestUnitAppCatalogNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *ac.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, err := s.List(nil, ac.ListOptions{}); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteByID(t.Context(), appID); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	if _, err := ac.NewService(nil).NewPager(ac.ListOptions{}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("unbound pager accepted")
	}
}
