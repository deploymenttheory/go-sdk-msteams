package team_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

const (
	teamID = "11111111-1111-4111-8111-111111111111"
	userID = "22222222-2222-4222-8222-222222222222"
	opID   = "33333333-3333-4333-8333-333333333333"
)

type exchange struct {
	method, path string
	query        url.Values
	json, binary string
	contentType  string
	status       int
	body         string
	header       http.Header
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type provider struct {
	audience string
	calls    int
}

func (p *provider) AcquireToken(ctx context.Context, req auth.TokenRequest) (auth.Token, error) {
	if err := ctx.Err(); err != nil {
		return auth.Token{}, err
	}
	if req.Audience != p.audience {
		return auth.Token{}, client.ErrDestination
	}
	p.calls++
	return auth.Token{AccessToken: "synthetic-team-token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

// All destinations remain production-shaped HTTPS URLs. This transport has no
// socket or fallback: any request beyond the declared sequence is a failure.
func fixture(t *testing.T, cloud string, mode auth.Mode, steps ...exchange) (*msteams.Client, *provider, *int) {
	t.Helper()
	env, err := auth.ResolveEnvironment(cloud, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := &provider{audience: env.Endpoints.Graph}
	calls := 0
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ExternalProvider,
		TenantID: "fixture-tenant", IdentityID: "fixture-identity", Cloud: cloud, TokenMode: mode, Provider: p},
		msteams.WithHTTPClient(&http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
			i := calls
			calls++
			if i >= len(steps) {
				t.Error("unexpected HTTP attempt")
				return nil, client.ErrRequest
			}
			step := steps[i]
			if req.Method != step.method || req.URL.EscapedPath() != step.path || req.URL.Scheme+"://"+req.URL.Host != env.Endpoints.Graph {
				t.Error("request method, escaped path or Graph cloud differs")
			}
			if len(step.query) != len(req.URL.Query()) || len(step.query) != 0 && !reflect.DeepEqual(req.URL.Query(), step.query) {
				t.Error("query differs")
			}
			if req.Header.Get("Authorization") != "Bearer synthetic-team-token" {
				t.Error("authorization missing")
			}
			var body []byte
			if req.Body != nil {
				var readErr error
				body, readErr = io.ReadAll(req.Body)
				closeErr := req.Body.Close()
				if readErr != nil || closeErr != nil {
					return nil, client.ErrUploadRead
				}
			}
			if step.json != "" {
				var got, want any
				if json.Unmarshal(body, &got) != nil || json.Unmarshal([]byte(step.json), &want) != nil || !reflect.DeepEqual(got, want) {
					t.Error("JSON payload differs")
				}
			} else if string(body) != step.binary {
				t.Error("binary or absent body differs")
			}
			if step.contentType != "" && req.Header.Get("Content-Type") != step.contentType {
				t.Error("content type differs")
			}
			header := step.header.Clone()
			if header == nil {
				header = make(http.Header)
			}
			header.Set("Content-Type", "application/json")
			header.Set("request-id", "fixture-response")
			return &http.Response{StatusCode: step.status, Header: header, Body: io.NopCloser(strings.NewReader(step.body)), Request: req}, nil
		})}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error(err)
		}
		if calls != len(steps) {
			t.Error("missing HTTP attempts")
		}
	})
	return sdk, p, &calls
}

func TestNativeCRUDAndRootWiring(t *testing.T) {
	t.Parallel()
	falseValue := false
	sdk, _, calls := fixture(t, "commercial", auth.Application,
		exchange{method: "PUT", path: "/v1.0/groups/" + teamID + "/team", json: `{}`, status: 201, body: `{"id":"` + teamID + `","unknownFuture":5}`},
		exchange{method: "GET", path: "/v1.0/teams/" + teamID, status: 200, body: `{"id":"` + teamID + `","displayName":"fixture","isArchived":false,"future":{"value":7},"funSettings":{"giphyContentRating":"futureRating"}}`},
		exchange{method: "PATCH", path: "/v1.0/teams/" + teamID, json: `{"guestSettings":{"allowDeleteChannels":false},"memberSettings":{}}`, status: 204},
		exchange{method: "DELETE", path: "/v1.0/groups/" + teamID, status: 204})
	created, response, err := sdk.GraphAPI.Teams.CreateFromGroup(t.Context(), teamID, team.Settings{})
	if err != nil || created.ID != teamID || string(created.RawFields["unknownFuture"]) != "5" || response.StatusCode() != 201 {
		t.Fatal("create-from-group result lost")
	}
	value, response, err := sdk.GraphAPI.Teams.GetByID(t.Context(), teamID)
	if err != nil || response.Header().Get("request-id") != "fixture-response" || value.IsArchived == nil || *value.IsArchived || value.FunSettings.GiphyContentRating == nil || *value.FunSettings.GiphyContentRating != "futureRating" || string(value.RawFields["future"]) != `{"value":7}` {
		t.Fatal("read snapshot lost presence or future fields")
	}
	response, err = sdk.GraphAPI.Teams.UpdateByID(t.Context(), teamID, team.Settings{MemberSettings: &team.MemberSettings{}, GuestSettings: &team.GuestSettings{AllowDeleteChannels: &falseValue}})
	if err != nil || response.StatusCode() != 204 {
		t.Fatal("update failed")
	}
	response, err = sdk.GraphAPI.Teams.DeleteByID(t.Context(), teamID)
	if err != nil || response.StatusCode() != 204 || *calls != 4 {
		t.Fatal("delete failed or mutation replayed")
	}
}

func TestListBasicProjectionsAndOpaqueContinuation(t *testing.T) {
	t.Parallel()
	falseValue := false
	opaque := "https://graph.microsoft.com/v1.0/teams?$skiptoken=opaque%2B%2F%3D&$select=id%2CdisplayName"
	sdk, _, _ := fixture(t, "commercial", auth.Application,
		exchange{method: "GET", path: "/v1.0/teams", query: url.Values{"$filter": {"visibility eq 'private'"}, "$select": {"id,displayName"}, "$top": {"2"}, "$count": {"false"}}, status: 200, body: `{"value":[{"id":"` + teamID + `","displayName":"fixture","isArchived":null}],"@odata.nextLink":"` + opaque + `"}`},
		exchange{method: "GET", path: "/v1.0/teams", query: url.Values{"$skiptoken": {"opaque+/="}, "$select": {"id,displayName"}}, status: 200, body: `{"value":[]}`})
	fields := []string{"id", "displayName"}
	pager, err := sdk.GraphAPI.Teams.NewPager(team.ListOptions{Filter: "visibility eq 'private'", Select: fields, Top: 2, Count: &falseValue})
	if err != nil {
		t.Fatal(err)
	}
	fields[0] = "changed"
	result, err := pager.Collect(t.Context())
	if err != nil || result.Pages != 2 || len(result.Items) != 1 || result.Items[0].IsArchived != nil || result.Response == nil {
		t.Fatal("list omitted/null projection or paging changed")
	}
}

func TestListPartialResultsAndBounds(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		maxPages int
		next     string
		want     error
		second   bool
	}{
		{"item-preserved-on-forbidden", 0, "https://graph.microsoft.com/v1.0/teams?$skiptoken=next", auth.ErrForbidden, true},
		{"page-bound", 1, "https://graph.microsoft.com/v1.0/teams?$skiptoken=next", client.ErrPageLimit, false},
		{"foreign-next", 0, "https://foreign.invalid/v1.0/teams", client.ErrDestination, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []exchange{{method: "GET", path: "/v1.0/teams", status: 200, body: `{"value":[{"id":"` + teamID + `"}],"@odata.nextLink":"` + tc.next + `"}`}}
			if tc.second {
				steps = append(steps, exchange{method: "GET", path: "/v1.0/teams", query: url.Values{"$skiptoken": {"next"}}, status: 403, body: `{"error":{"code":"Forbidden","message":"private"}}`})
			}
			sdk, _, _ := fixture(t, "commercial", auth.Application, steps...)
			result, err := sdk.GraphAPI.Teams.List(t.Context(), team.ListOptions{MaxPages: tc.maxPages})
			if !errors.Is(err, tc.want) || len(result.Items) != 1 || result.Pages != 1 || result.Response == nil {
				t.Fatal("partial list was lost or reported complete")
			}
		})
	}
}

func TestReadEnvelopeAndStatusValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"malformed", `{"id":`, 200, client.ErrDecode},
		{"missing-identity", `{}`, 200, client.ErrDecode},
		{"null", `null`, 200, client.ErrDecode},
		{"wrong-type", `{"id":"` + teamID + `","isArchived":"false"}`, 200, client.ErrDecode},
		{"unexpected-acceptance", `{"id":"` + teamID + `"}`, 202, team.ErrUnexpectedStatus},
		{"forbidden", `{"error":{"code":"Forbidden","message":"private"}}`, 403, auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sdk, _, _ := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams/" + teamID, status: tc.status, body: tc.body})
			value, response, err := sdk.GraphAPI.Teams.GetByID(t.Context(), teamID)
			if value != nil || response == nil || response.StatusCode() != tc.status || !errors.Is(err, tc.want) {
				t.Fatal("invalid response lost or accepted")
			}
		})
	}
}

func TestLocalValidationAndCancellationPerformNoIO(t *testing.T) {
	t.Parallel()
	sdk, p, calls := fixture(t, "commercial", auth.Application)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := sdk.GraphAPI.Teams.GetByID(ctx, teamID); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
	//nolint:staticcheck // Deliberately verifies nil context rejection before any I/O.
	if _, err := sdk.GraphAPI.Teams.UpdateByID(nil, teamID, team.Settings{}); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	for _, id := range []string{"", ".", "../groups", "id/segment", "id?query", "id\\segment", "private\n"} {
		if _, _, err := sdk.GraphAPI.Teams.GetByID(t.Context(), id); !errors.Is(err, client.ErrRequest) {
			t.Fatal("invalid team identity accepted")
		}
		if _, _, err := sdk.GraphAPI.Teams.CreateFromGroup(t.Context(), id, team.Settings{}); !errors.Is(err, client.ErrRequest) {
			t.Fatal("invalid group identity accepted")
		}
		if _, err := sdk.GraphAPI.Teams.DeleteByID(t.Context(), id); !errors.Is(err, client.ErrRequest) {
			t.Fatal("invalid delete identity accepted")
		}
	}
	if _, err := sdk.GraphAPI.Teams.NewPager(team.ListOptions{Top: -1}); !errors.Is(err, client.ErrRequest) {
		t.Fatal("negative page size accepted")
	}
	if _, err := sdk.GraphAPI.Teams.NewPager(team.ListOptions{Select: []string{""}}); !errors.Is(err, client.ErrRequest) {
		t.Fatal("empty projection field accepted")
	}
	if _, err := sdk.GraphAPI.Teams.NewPager(team.ListOptions{MaxPages: -1}); !errors.Is(err, client.ErrPagination) {
		t.Fatal("negative bound accepted")
	}
	wrong := team.NewService(sdk.TeamsAdminAPI.Scope())
	if _, _, err := wrong.GetByID(t.Context(), teamID); !errors.Is(err, client.ErrDestination) {
		t.Fatal("wrong backend accepted")
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sdk.GraphAPI.Teams.GetByID(t.Context(), teamID); !errors.Is(err, client.ErrClosed) {
		t.Fatal("closed session accepted")
	}
	if p.calls != 0 || *calls != 0 {
		t.Fatal("local validation acquired credentials or performed IO")
	}
}

func TestMutationFailuresAreNotReplayed(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			path := "/v1.0/groups/" + teamID
			body := ""
			if method == "PUT" {
				path += "/team"
				body = `{}`
			}
			if method == "PATCH" {
				path = "/v1.0/teams/" + teamID
				body = `{}`
			}
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: method, path: path, json: body, status: 503, body: `{"error":{"code":"ServiceUnavailable","message":"private"}}`})
			var err error
			switch method {
			case "PUT":
				_, _, err = sdk.GraphAPI.Teams.CreateFromGroup(t.Context(), teamID, team.Settings{})
			case "PATCH":
				_, err = sdk.GraphAPI.Teams.UpdateByID(t.Context(), teamID, team.Settings{})
			case "DELETE":
				_, err = sdk.GraphAPI.Teams.DeleteByID(t.Context(), teamID)
			}
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 || *calls != 1 {
				t.Fatal("mutation failure replayed or lost diagnostics")
			}
		})
	}
}

func TestListRejectsUnexpectedSuccessfulPageStatus(t *testing.T) {
	t.Parallel()
	for _, status := range []int{202, 206} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams", status: status, body: `{"value":[{"id":"` + teamID + `"}]}`})
			result, err := sdk.GraphAPI.Teams.List(t.Context(), team.ListOptions{})
			if !errors.Is(err, client.ErrPageStatus) || len(result.Items) != 0 || result.Pages != 0 || result.Response == nil || result.Response.StatusCode() != status || *calls != 1 {
				t.Fatal("unexpected status became a complete native teams page")
			}
		})
	}
}
