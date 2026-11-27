package directoryuser_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/go-resty/resty/v2"
)

const (
	userID  = "11111111-1111-4111-8111-111111111111"
	otherID = "22222222-2222-4222-8222-222222222222"
)

func runCase(t *testing.T, c testutil.Case) {
	t.Helper()
	if c.Name == "" {
		c.Name = "resolver-case"
	}

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("directory resolution differs: %+v", result.Differences())
	}
}

func wire(origin, path string, status int, body string) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + path, Audience: origin, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func observed(value directoryuser.Resolution, response *resty.Response, err error) testutil.Observation {
	return testutil.Observation{Value: map[string]any{"id": value.ID, "source": value.Source, "snapshot": value.User}, Response: response, Err: err}
}

func directPath(selector string) string {
	return "/v1.0/users/" + url.PathEscape(selector) + "?$select=id"
}
func mailPath(literal string) string {
	return "/v1.0/users?" + url.Values{"$filter": {"mail eq '" + literal + "'"}, "$select": {"id"}, "$top": {"2"}}.Encode()
}

func TestUnitResolverExplicitUUIDIsUnverifiedAndCredentialFree(t *testing.T) {
	t.Parallel()
	runCase(t, testutil.Case{Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		return observed(sdk.GraphAPI.UserResolver.Resolve(ctx, userID, directoryuser.ResolveOptions{}))
	},
		Want: testutil.Expected{JSON: json.RawMessage(`{"id":"` + userID + `","source":"identifier","snapshot":null}`)}})
}

func TestUnitResolverDirectAndCurrentReads(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, selector, path, source string
		mode                         auth.Mode
		current                      bool
	}{
		{"direct-upn", "fixture@example.invalid", directPath("fixture@example.invalid"), "direct", auth.Application, false},
		{"b2b-upn", "fixture#EXT#@example.invalid", directPath("fixture#EXT#@example.invalid"), "direct", auth.Delegated, false},
		{"leading-dollar", "$o'brien@example.invalid", "/v1.0/" + url.PathEscape("users('$o''brien@example.invalid')") + "?$select=id", "direct", auth.Delegated, false},
		{"current", "", "/v1.0/me?$select=id", "current", auth.Delegated, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runCase(t, testutil.Case{Mode: tc.mode, Steps: []testutil.Exchange{wire(auth.GraphAudience, tc.path, 200, `{"id":"`+userID+`"}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					return observed(sdk.GraphAPI.UserResolver.Resolve(ctx, tc.selector, directoryuser.ResolveOptions{AllowCurrent: tc.current}))
				},
				Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + userID + `","source":"` + tc.source + `","snapshot":{"id":"` + userID + `"}}`)}})
		})
	}
}

func TestUnitResolverOnly404AllowsEscapedCanonicalMailFallback(t *testing.T) {
	t.Parallel()
	selector := "o'brien+tag#ext@example.invalid"
	runCase(t, testutil.Case{Steps: []testutil.Exchange{
		wire(auth.GraphAudience, directPath(selector), 404, `{"error":{"code":"Request_ResourceNotFound"}}`),
		wire(auth.GraphAudience, mailPath("o''brien+tag#ext@example.invalid"), 200, `{"value":[{"id":"`+userID+`"}]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			return observed(sdk.GraphAPI.UserResolver.Resolve(ctx, selector, directoryuser.ResolveOptions{}))
		},
		Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + userID + `","source":"mail","snapshot":{"id":"` + userID + `"}}`)}})
	for _, status := range []int{400, 401, 403} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			runCase(t, testutil.Case{Steps: []testutil.Exchange{wire(auth.GraphAudience, directPath(selector), status, `{"error":{"code":"FixtureError","message":"private"}}`)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					_, response, err := sdk.GraphAPI.UserResolver.Resolve(ctx, selector, directoryuser.ResolveOptions{})
					return testutil.Observation{Response: response, Err: err}
				},
				Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureError"}}}})
		})
	}
}

func TestUnitResolverMailUniquenessRequiresCompleteTraversal(t *testing.T) {
	t.Parallel()
	selector := "fixture@example.invalid"
	first := wire(auth.GraphAudience, directPath(selector), 404, `{"error":{"code":"NotFound"}}`)
	next := auth.GraphAudience + "/v1.0/users?$skiptoken=opaque%2B%2F%3D"
	for _, tc := range []struct {
		name, firstBody, lastBody string
		second                    bool
		max                       int
		want                      error
		id, source                string
	}{
		{"single", `{"value":[{"id":"` + userID + `"}]}`, "", false, 0, nil, userID, "mail"},
		{"none", `{"value":[]}`, "", false, 0, directoryuser.ErrNotFound, "", ""},
		{"ambiguous-page", `{"value":[{"id":"` + userID + `"},{"id":"` + otherID + `"}]}`, "", false, 0, directoryuser.ErrAmbiguous, "", ""},
		{"ambiguous-later", `{"value":[{"id":"` + userID + `"}],"@odata.nextLink":"` + next + `"}`, `{"value":[{"id":"` + otherID + `"}]}`, true, 0, directoryuser.ErrAmbiguous, "", ""},
		{"repeated-id", `{"value":[{"id":"` + userID + `"}],"@odata.nextLink":"` + next + `"}`, `{"value":[{"id":"` + userID + `"}]}`, true, 0, nil, userID, "mail"},
		{"bound-not-unique", `{"value":[{"id":"` + userID + `"}],"@odata.nextLink":"` + next + `"}`, "", false, 1, client.ErrPageLimit, "", ""},
		{"bound-not-absent", `{"value":[],"@odata.nextLink":"` + next + `"}`, "", false, 1, client.ErrPageLimit, "", ""},
		{"foreign-next", `{"value":[{"id":"` + userID + `"}],"@odata.nextLink":"https://foreign.invalid/private"}`, "", false, 0, client.ErrDestination, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{first, wire(auth.GraphAudience, mailPath(selector), 200, tc.firstBody)}
			if tc.second {
				steps = append(steps, wire(auth.GraphAudience, "/v1.0/users?$skiptoken=opaque%2B%2F%3D", 200, tc.lastBody))
			}
			snapshot := "null"
			if tc.id != "" {
				snapshot = `{"id":"` + tc.id + `"}`
			}
			runCase(t, testutil.Case{Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return observed(sdk.GraphAPI.UserResolver.Resolve(ctx, selector, directoryuser.ResolveOptions{MaxPages: tc.max}))
			},
				Want: testutil.Expected{Status: 200, Error: tc.want, JSON: json.RawMessage(`{"id":"` + tc.id + `","source":"` + tc.source + `","snapshot":` + snapshot + `}`)}})
		})
	}
}

func TestUnitResolverPendingAndMalformedNeverFallBack(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"accepted-empty", "", 202, directoryuser.ErrPending},
		{"accepted-snapshot", `{"id":"` + userID + `"}`, 202, directoryuser.ErrPending},
		{"missing-id", `{}`, 200, client.ErrDecode},
		{"wrong-id-type", `{"id":[]}`, 200, client.ErrDecode},
		{"object-id", `{"id":{}}`, 200, client.ErrDecode},
		{"unexpected-status", `{"id":"` + userID + `"}`, 201, directoryuser.ErrUnexpectedStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runCase(t, testutil.Case{Steps: []testutil.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), tc.status, tc.body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, response, err := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
					return testutil.Observation{Value: value.ID, Response: response, Err: err}
				},
				Want: testutil.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`""`)}})
		})
	}
}

func TestUnitResolverConfiguredCloudRoutes(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			t.Run(cloud+"/"+string(mode), func(t *testing.T) {
				env, err := auth.ResolveEnvironment(cloud, nil)
				if err != nil {
					t.Fatal(err)
				}
				runCase(t, testutil.Case{Cloud: cloud, Mode: mode, Steps: []testutil.Exchange{wire(env.Endpoints.Graph, directPath("fixture@example.invalid"), 200, `{"id":"`+userID+`"}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						_, response, err := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
						return testutil.Observation{Response: response, Err: err}
					}, Want: testutil.Expected{Status: 200}})
			})
		}
	}
}

func TestUnitResolverValidationAndCurrentModeAreLocal(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"missing", "whitespace", "control", "backslash", "dot", "negative-pages", "app-current", "cancelled", "closed", "wrong-backend", "nil-resolver", "bad-select"} {
		want := client.ErrRequest
		switch action {
		case "app-current":
			want = directoryuser.ErrCurrentUser
		case "negative-pages":
			want = client.ErrPagination
		case "cancelled":
			want = context.Canceled
		case "closed", "nil-resolver":
			want = client.ErrClosed
		case "wrong-backend":
			want = client.ErrDestination
		}
		runCase(t, testutil.Case{Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			var err error
			switch action {
			case "missing":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, "", directoryuser.ResolveOptions{})
			case "whitespace":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, " ", directoryuser.ResolveOptions{})
			case "control":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, "private\n", directoryuser.ResolveOptions{})
			case "backslash":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, "private\\selector", directoryuser.ResolveOptions{})
			case "dot":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, "..", directoryuser.ResolveOptions{})
			case "negative-pages":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, userID, directoryuser.ResolveOptions{MaxPages: -1})
			case "app-current":
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, "", directoryuser.ResolveOptions{AllowCurrent: true})
			case "cancelled":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(cancelled, userID, directoryuser.ResolveOptions{})
			case "closed":
				if e := sdk.Close(); e != nil {
					return testutil.Observation{Err: e}
				}
				_, _, err = sdk.GraphAPI.UserResolver.Resolve(ctx, userID, directoryuser.ResolveOptions{})
			case "wrong-backend":
				_, _, err = directoryuser.NewResolver(sdk.TeamsAdminAPI.Scope()).Resolve(ctx, userID, directoryuser.ResolveOptions{})
			case "nil-resolver":
				var r *directoryuser.Resolver
				_, _, err = r.Resolve(ctx, userID, directoryuser.ResolveOptions{})
			case "bad-select":
				_, _, err = sdk.GraphAPI.UserResolver.Get(ctx, userID, directoryuser.ReadOptions{Select: []string{""}})
			}
			return testutil.Observation{Err: err}
		}, Want: testutil.Expected{Error: want}})
	}
}
