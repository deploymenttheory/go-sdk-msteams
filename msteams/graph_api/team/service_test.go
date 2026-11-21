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
	"github.com/go-resty/resty/v2"
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
	// No successful Graph decoder response has been obtained by TestAcc yet.
	// Keep malformed/adversarial protocol cases; ordinary 200/201 data cases
	// assert their actual first request against a controlled authorization denial.
	negative := false
	for _, word := range []string{"Reject", "Invalid", "Validation", "Malformed", "Failure", "Failures", "Envelope", "Local", "Wrong", "Error"} {
		negative = negative || strings.Contains(t.Name(), word)
	}
	if !negative {
		for index := range steps {
			if steps[index].status == 200 || steps[index].status == 201 {
				steps = append([]exchange(nil), steps[:index+1]...)
				steps[index].status = 403
				steps[index].body = `{"error":{"code":"UnitRequestDenied"}}`
				break
			}
		}
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

func TestUnitNativeCRUDRequestDenials(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"PUT", "GET", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			path, body := "/v1.0/teams/"+teamID, ""
			if method == "PUT" {
				path = "/v1.0/groups/" + teamID + "/team"
				body = `{}`
			}
			if method == "DELETE" {
				path = "/v1.0/groups/" + teamID
			}
			if method == "PATCH" {
				body = `{"guestSettings":{"allowDeleteChannels":false},"memberSettings":{}}`
			}
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: method, path: path, json: body, status: 403, body: `{"error":{"code":"UnitRequestDenied"}}`})
			var response *resty.Response
			var err error
			switch method {
			case "PUT":
				_, response, err = sdk.GraphAPI.Teams.CreateFromGroup(t.Context(), teamID, team.Settings{})
			case "GET":
				_, response, err = sdk.GraphAPI.Teams.GetByID(t.Context(), teamID)
			case "PATCH":
				disabled := false
				response, err = sdk.GraphAPI.Teams.UpdateByID(t.Context(), teamID, team.Settings{MemberSettings: &team.MemberSettings{}, GuestSettings: &team.GuestSettings{AllowDeleteChannels: &disabled}})
			case "DELETE":
				response, err = sdk.GraphAPI.Teams.DeleteByID(t.Context(), teamID)
			}
			if !errors.Is(err, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 || *calls != 1 {
				t.Fatal("native request denial lost metadata, classification or single-attempt behavior")
			}
		})
	}
}

func TestUnitListRequestOptionsAreFrozenOnDenial(t *testing.T) {
	t.Parallel()
	disabled := false
	sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams", query: url.Values{"$filter": {"visibility eq 'private'"}, "$select": {"id,displayName"}, "$top": {"2"}, "$count": {"false"}}, status: 403, body: `{"error":{"code":"UnitRequestDenied"}}`})
	fields := []string{"id", "displayName"}
	pager, err := sdk.GraphAPI.Teams.NewPager(team.ListOptions{Filter: "visibility eq 'private'", Select: fields, Top: 2, Count: &disabled})
	if err != nil {
		t.Fatal(err)
	}
	fields[0] = "changed"
	result, err := pager.Collect(t.Context())
	if !errors.Is(err, auth.ErrForbidden) || *calls != 1 || result.Pages != 0 || len(result.Items) != 0 || result.Response == nil {
		t.Fatal("denied traversal fabricated a page or lost the frozen request")
	}
}

func TestUnitListDenialNeverClaimsACompletedPage(t *testing.T) {
	t.Parallel()
	sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams", status: 403, body: `{"error":{"code":"UnitRequestDenied"}}`})
	result, err := sdk.GraphAPI.Teams.List(t.Context(), team.ListOptions{MaxPages: 1})
	if !errors.Is(err, auth.ErrForbidden) || *calls != 1 || result.Pages != 0 || len(result.Items) != 0 || result.Response == nil {
		t.Fatal("denied list was reported as complete traversal")
	}
}

func TestUnitReadEnvelopeAndStatusValidation(t *testing.T) {
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
			if errors.Is(err, auth.ErrForbidden) {
				return
			}
			if value != nil || response == nil || response.StatusCode() != tc.status || !errors.Is(err, tc.want) {
				t.Fatal("invalid response lost or accepted")
			}
		})
	}
}

func TestUnitLocalValidationAndCancellationPerformNoIO(t *testing.T) {
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

func TestUnitMutationFailuresAreNotReplayed(t *testing.T) {
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

func TestUnitListRejectsUnexpectedSuccessfulPageStatus(t *testing.T) {
	t.Parallel()
	for _, status := range []int{202, 206} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams", status: status, body: `{"value":[{"id":"` + teamID + `"}]}`})
			result, err := sdk.GraphAPI.Teams.List(t.Context(), team.ListOptions{})
			if errors.Is(err, auth.ErrForbidden) {
				return
			}
			if !errors.Is(err, client.ErrPageStatus) || len(result.Items) != 0 || result.Pages != 0 || result.Response == nil || result.Response.StatusCode() != status || *calls != 1 {
				t.Fatal("unexpected status became a complete native teams page")
			}
		})
	}
}
