package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type tokenRecorder struct {
	mu       sync.Mutex
	requests []auth.TokenRequest
}

func (p *tokenRecorder) AcquireToken(ctx context.Context, r auth.TokenRequest) (auth.Token, error) {
	if err := ctx.Err(); err != nil {
		return auth.Token{}, err
	}
	p.mu.Lock()
	p.requests = append(p.requests, r)
	p.mu.Unlock()
	return auth.Token{AccessToken: "fixture-token-" + r.Audience, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func transportScope(t *testing.T, backend client.Backend, httpClient *http.Client, cloud string) (*client.Scope, *client.Session, *tokenRecorder) {
	t.Helper()
	provider := &tokenRecorder{}
	session, err := client.NewSession(&auth.Config{
		Method: auth.ExternalProvider, TenantID: "fixture-tenant", Cloud: cloud,
		IdentityID: "fixture-identity", TokenMode: auth.Application, Provider: provider,
	}, httpClient, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	scope, err := client.NewScope(backend, session)
	if err != nil {
		t.Fatal(err)
	}
	return scope, session, provider
}

// The SDK sees production HTTPS destinations; only this injected test transport
// maps them to httptest. No unrestricted endpoint option is needed in production.
func localBackend(t *testing.T, handler http.HandlerFunc, inspect func(*http.Request)) *http.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	destination, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	base := http.DefaultTransport.(*http.Transport).Clone()
	t.Cleanup(base.CloseIdleConnections)
	return &http.Client{Transport: transportFunc(func(request *http.Request) (*http.Response, error) {
		if inspect != nil {
			inspect(request)
		}
		forward := request.Clone(request.Context())
		forward.URL.Scheme, forward.URL.Host = destination.Scheme, destination.Host
		forward.Host = destination.Host
		return base.RoundTrip(forward)
	})}
}

func TestBackendRoutingEscapingAndResponseMetadata(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		backend        client.Backend
		host, audience string
		path           []string
	}{
		{client.Graph, "graph.microsoft.com", auth.GraphAudience, []string{"v1.0", "teams", "id/with ?#%雪", "channels"}},
		{client.TeamsAdmin, "api.interfaces.records.teams.microsoft.com", auth.TeamsAdminAudience, []string{"fixture", "policy/with ?#%雪"}},
	} {
		t.Run(string(tc.backend), func(t *testing.T) {
			headers := http.Header{"X-Fixture-Header": {"one", "two"}, "If-Match": {`W/"fixture"`}}
			query := url.Values{"$filter": {"displayName eq 'A & B'"}, "$select": {"id", "displayName"}}
			httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
				if !reflect.DeepEqual(r.URL.Query(), query) {
					t.Error("query changed")
				}
				if !reflect.DeepEqual(r.Header.Values("X-Fixture-Header"), []string{"one", "two"}) {
					t.Error("backend headers lost")
				}
				if r.Header.Get("If-Match") != `W/"fixture"` {
					t.Error("conditional header lost")
				}
				if r.Header.Get("Authorization") != "Bearer fixture-token-"+tc.audience {
					t.Error("wrong token audience")
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("request-id", "fixture-response-id")
				_, _ = io.WriteString(w, `{"id":"fixture-result","unknownFutureField":true}`)
			}, func(r *http.Request) {
				if r.URL.Scheme != "https" || r.URL.Host != tc.host {
					t.Error("wrong destination")
				}
				parts := make([]string, len(tc.path))
				for i, p := range tc.path {
					parts[i] = url.PathEscape(p)
				}
				if r.URL.EscapedPath() != "/"+strings.Join(parts, "/") {
					t.Error("identity changed route segments")
				}
			})
			scope, _, provider := transportScope(t, tc.backend, httpClient, "commercial")
			var result struct {
				ID string `json:"id"`
			}
			response, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: tc.path, Query: query, Header: headers}, &result)
			if err != nil {
				t.Fatal(err)
			}
			if result.ID != "fixture-result" || response.StatusCode() != 200 || response.Header().Get("request-id") != "fixture-response-id" || response.Size() == 0 {
				t.Fatal("lost response result or metadata")
			}
			if response.RawResponse.Request.Header.Get("Authorization") != "" || response.Request.Header.Get("Authorization") != "" {
				t.Fatal("response retained authorization")
			}
			if len(provider.requests) != 1 || provider.requests[0].Audience != tc.audience {
				t.Fatal("token binding lost")
			}
			if headers.Get("Authorization") != "" || len(headers) != 2 {
				t.Fatal("caller headers mutated")
			}
		})
	}
}

func TestRejectedDestinationsAndRequestsNeverAcquireCredentials(t *testing.T) {
	t.Parallel()
	badURLs := []string{
		"https://attacker.invalid/v1.0/teams", "http://graph.microsoft.com/v1.0/teams",
		"https://graph.microsoft.com.attacker.invalid/v1.0/teams", "https://graph.microsoft.com:443/v1.0/teams",
		"https://fixture-secret@graph.microsoft.com/v1.0/teams", "https://graph.microsoft.com/v1.0/teams#fragment",
		"//graph.microsoft.com/v1.0/teams", "https:graph.microsoft.com/v1.0/teams", "https://graph.microsoft.com/other/teams",
		"https://graph.microsoft.com/v1.0/../teams", "https://graph.microsoft.com/v1.0/%2e%2e/teams",
		"https://graph.microsoft.com/v1.0/%2f..%2fteams", "https://graph.microsoft.com//v1.0/teams",
		"https://graph.microsoft.com/v1.0/%5csecret", "https://graph.microsoft.com/v1.0/%00secret",
		"https://graph.microsoft.com/v1.0/teams?invalid=%xx", "https://api.interfaces.records.teams.microsoft.com/fixture",
	}
	requests := []client.Request{}
	for _, value := range badURLs {
		requests = append(requests, client.Request{Method: http.MethodGet, URL: value})
	}
	requests = append(requests,
		client.Request{Method: http.MethodGet, Path: []string{"v1.0", ".."}},
		client.Request{Method: "CONNECT", Path: []string{"v1.0", "teams"}},
		client.Request{Method: http.MethodGet, URL: "https://graph.microsoft.com/v1.0/teams", Path: []string{"v1.0"}},
		client.Request{Method: http.MethodGet, URL: "https://graph.microsoft.com/v1.0/teams", Query: url.Values{"x": {"y"}}},
		client.Request{Method: http.MethodPost, Path: []string{"v1.0", "teams"}, Body: make(chan int)},
	)
	for _, header := range []http.Header{
		{"aUtHoRiZaTiOn": {"fixture-secret"}}, {"Host": {"attacker.invalid"}}, {"Cookie": {"fixture-secret"}},
		{"Connection": {"Authorization"}}, {"Transfer-Encoding": {"chunked"}}, {"Content-Length": {"0"}},
		{"X-Invalid\nName": {"x"}}, {"X-Value": {"value\r\nAuthorization: fixture-secret"}},
	} {
		requests = append(requests, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "teams"}, Header: header})
	}
	scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Error("invalid request reached transport")
		return nil, errors.New("unexpected")
	})}, "commercial")
	for _, request := range requests {
		if _, err := scope.Do(t.Context(), request, nil); err == nil {
			t.Errorf("accepted invalid request: method %s", request.Method)
		}
	}
	if len(provider.requests) != 0 {
		t.Fatal("invalid request acquired credentials")
	}
	for _, tc := range []struct{ cloud, destination string }{
		{"government", "https://graph.microsoft.com/v1.0/teams"},
		{"commercial", "https://msit.api.interfaces.records.teams.microsoft.com/fixture"},
		{"commercial", "https://dev.api.interfaces.records.teams.microsoft.com/fixture"},
		{"commercial", "https://api.interfaces.records.gov.teams.microsoft.us/fixture"},
	} {
		s, _, p := transportScope(t, client.TeamsAdmin, &http.Client{}, tc.cloud)
		if _, err := s.Do(t.Context(), client.Request{Method: http.MethodGet, URL: tc.destination}, nil); !errors.Is(err, client.ErrDestination) {
			t.Fatal("unsupported profile accepted")
		}
		if len(p.requests) != 0 {
			t.Fatal("profile rejection happened after acquisition")
		}
	}
}

func TestContinuationURLPreservesOpaqueQuery(t *testing.T) {
	t.Parallel()
	const rawQuery = "$skiptoken=a%2Bb%2fC&x=1&x=2&space=%20"
	httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != rawQuery {
			t.Error("continuation query changed")
		}
		w.WriteHeader(http.StatusNoContent)
	}, nil)
	scope, _, _ := transportScope(t, client.Graph, httpClient, "commercial")
	if _, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, URL: "https://graph.microsoft.com/beta/teams?" + rawQuery}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRedirectsAndBorrowedCookiesAreNotForwarded(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("Cookie") != "" {
			t.Error("borrowed cookie leaked")
		}
		w.Header().Set("Set-Cookie", "fixture=secret")
		w.Header().Set("Location", "https://attacker.invalid/fixture-secret")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}, nil)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse("https://graph.microsoft.com")
	if err != nil {
		t.Fatal(err)
	}
	jar.SetCookies(u, []*http.Cookie{{Name: "fixture", Value: "borrowed-secret", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}})
	httpClient.Jar = jar
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { t.Error("borrowed redirect callback used"); return nil }
	scope, _, _ := transportScope(t, client.Graph, httpClient, "commercial")
	response, err := scope.Do(t.Context(), client.Request{Method: http.MethodPost, Path: []string{"v1.0", "teams"}, Body: map[string]string{"id": "fixture"}}, nil)
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 307 || response.Header().Get("Location") == "" || hits.Load() != 1 {
		t.Fatal("redirect was not returned as one HTTP failure")
	}
	if httpClient.Jar != jar || httpClient.CheckRedirect == nil || jar.Cookies(u)[0].Value != "borrowed-secret" {
		t.Fatal("borrowed HTTP client mutated")
	}
}

func TestNormalizedErrorEnvelopes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                                        string
		backend                                     client.Backend
		body                                        string
		codes, requestIDs, correlations, operations []string
		diagnostics                                 []int
	}{
		{"graph", client.Graph, `{"error":{"code":"BadRequest","message":"fixture-secret","innerError":{"code":"InvalidIdentity","request-id":"body-request","client-request-id":"body-client"}}}`, []string{"BadRequest", "InvalidIdentity"}, []string{"header-request", "body-request"}, nil, nil, nil},
		{"teams-flat", client.TeamsAdmin, `{"code":"BadRequest","message":"fixture-secret","action":"fixture-secret"}`, []string{"BadRequest"}, []string{"header-request"}, nil, nil, nil},
		{"teams-default", client.TeamsAdmin, `{"code":"BadRequest","message":{"diagnostics":{"code":42,"subCode":7,"correlationID":"body-correlation","reason":"fixture-secret","debugContent":"fixture-secret"}}}`, []string{"BadRequest"}, []string{"header-request"}, []string{"body-correlation"}, nil, []int{42}},
		{"teams-ncs", client.TeamsAdmin, `{"diagnostics":{"code":42,"correlationID":"body-correlation","reason":"fixture-secret"}}`, nil, []string{"header-request"}, []string{"body-correlation"}, nil, []int{42}},
		{"teams-array", client.TeamsAdmin, `{"operationId":"body-operation","errors":[{"code":"One","message":"fixture-secret"},{"code":"Two","message":"fixture-secret"}]}`, []string{"One", "Two"}, []string{"header-request"}, nil, []string{"body-operation"}, nil},
		{"malformed", client.Graph, `{"error":fixture-secret`, nil, []string{"header-request"}, nil, nil, nil},
		{"non-json", client.TeamsAdmin, `<html>fixture-secret</html>`, nil, []string{"header-request"}, nil, nil, nil},
		{"empty", client.Graph, ``, nil, []string{"header-request"}, nil, nil, nil},
		{"malicious-code", client.Graph, `{"error":{"code":"fixture-secret","message":"fixture-secret"}}`, []string{"fixture-secret"}, []string{"header-request"}, nil, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := localBackend(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("request-id", "header-request")
				w.WriteHeader(400)
				_, _ = io.WriteString(w, tc.body)
			}, nil)
			scope, _, _ := transportScope(t, tc.backend, httpClient, "commercial")
			response, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("missing typed error: %v", err)
			}
			if apiErr.StatusCode != 400 || apiErr.Backend != tc.backend || !reflect.DeepEqual(apiErr.Codes, tc.codes) || !reflect.DeepEqual(apiErr.RequestIDs, tc.requestIDs) || !reflect.DeepEqual(apiErr.CorrelationIDs, tc.correlations) || !reflect.DeepEqual(apiErr.OperationIDs, tc.operations) || !reflect.DeepEqual(apiErr.DiagnosticCodes, tc.diagnostics) {
				t.Fatal("lost normalized diagnostics")
			}
			if tc.name == "graph" && !reflect.DeepEqual(apiErr.ClientRequestIDs, []string{"body-client"}) {
				t.Fatal("client request ID lost")
			}
			if tc.name == "teams-default" && !reflect.DeepEqual(apiErr.DiagnosticSubCodes, []int{7}) {
				t.Fatal("diagnostic subcode lost")
			}
			if string(response.Body()) != tc.body {
				t.Fatal("raw response body lost")
			}
			for _, format := range []string{"%s", "%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, err), "fixture-secret") {
					t.Fatal("error formatting exposed server data")
				}
				if strings.Contains(fmt.Sprintf(format, *apiErr), "fixture-secret") {
					t.Fatal("copied error formatting exposed server data")
				}
			}
		})
	}
}

func TestWrongAudienceHasNoFallback(t *testing.T) {
	t.Parallel()
	session, err := client.NewSession(&auth.Config{
		Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application,
		AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "fixture-graph-token", ExpiresAt: time.Now().Add(time.Hour)}},
	}, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Error("wrong-audience request was sent")
		return nil, errors.New("unexpected")
	})}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	scope, err := client.NewScope(client.TeamsAdmin, session)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"fixture"}}, nil); !errors.Is(err, auth.ErrTokenMissing) {
		t.Fatalf("wrong audience did not fail: %v", err)
	}
}

func TestConcurrentRequestsKeepHeadersAndResultsIsolated(t *testing.T) {
	t.Parallel()
	httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"id": r.Header.Get("X-Fixture-ID")}); err != nil {
			t.Error(err)
		}
	}, nil)
	scope, _, provider := transportScope(t, client.Graph, httpClient, "commercial")
	var group sync.WaitGroup
	for i := range 24 {
		group.Go(func() {
			identity := fmt.Sprint(i)
			var result struct {
				ID string `json:"id"`
			}
			_, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Header: http.Header{"X-Fixture-Id": {identity}}}, &result)
			if err != nil || result.ID != identity {
				t.Errorf("concurrent request lost its result: %v", err)
			}
		})
	}
	group.Wait()
	if len(provider.requests) != 24 {
		t.Fatal("concurrent acquisition count differs")
	}
}

func TestJSONBodiesSuccessDecodingAndEmptyResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		wantError  error
	}{
		{"json", `{"enabled":false,"count":0}`, 200, nil},
		{"no-content", "", 204, nil}, {"empty", "", 200, nil},
		{"malformed", `{"enabled":fixture-secret}`, 200, client.ErrDecode},
		{"html", `<html>fixture-secret</html>`, 200, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if string(body) != `{"count":0,"enabled":false,"list":[]}` || r.Header.Get("Content-Type") != "application/json" {
					t.Error("JSON request changed explicit zero values")
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}, nil)
			scope, _, _ := transportScope(t, client.Graph, httpClient, "commercial")
			var result struct {
				Enabled bool `json:"enabled"`
				Count   int  `json:"count"`
			}
			_, err := scope.Do(t.Context(), client.Request{Method: http.MethodPatch, Path: []string{"v1.0", "fixture"}, Body: map[string]any{"enabled": false, "count": 0, "list": []string{}}}, &result)
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("got %v, want %v", err, tc.wantError)
			}
		})
	}
}

func TestJSONMediaTypeIsPreservedAndNonJSONRejected(t *testing.T) {
	t.Parallel()
	const contentType = "application/merge-patch+json; charset=utf-8"
	httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != contentType {
			t.Error("backend JSON media type was overwritten")
		}
		w.WriteHeader(http.StatusNoContent)
	}, nil)
	scope, _, provider := transportScope(t, client.Graph, httpClient, "commercial")
	request := client.Request{Method: http.MethodPatch, Path: []string{"v1.0", "fixture"}, Body: map[string]bool{"enabled": false}, Header: http.Header{"Content-Type": {contentType}}}
	if _, err := scope.Do(t.Context(), request, nil); err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	if _, err := scope.Do(t.Context(), request, nil); !errors.Is(err, client.ErrRequest) {
		t.Fatal("non-JSON body contract accepted")
	}
	if len(provider.requests) != 1 {
		t.Fatal("non-JSON request acquired token")
	}
}

type trackedBody struct {
	io.Reader
	closed atomic.Bool
}

func (b *trackedBody) Close() error { b.closed.Store(true); return nil }

func TestBodiesClosedForSuccessFailureAndSizeLimit(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"success", `{}`, 200, nil}, {"error", `{}`, 500, nil},
		{"oversized", strings.Repeat("x", client.MaxResponseBytes+1), 200, client.ErrResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(tc.body)}
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: body, Request: r}, nil
			})}, "commercial")
			_, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("unexpected error %v", err)
			}
			if !body.closed.Load() {
				t.Fatal("response body left open")
			}
		})
	}
}

func TestCancellationReachesBackendAndCloseCancelsRequest(t *testing.T) {
	t.Parallel()
	for _, closeSession := range []bool{false, true} {
		t.Run(fmt.Sprint(closeSession), func(t *testing.T) {
			started := make(chan struct{})
			finished := make(chan struct{})
			httpClient := localBackend(t, func(_ http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done(); close(finished) }, nil)
			scope, session, _ := transportScope(t, client.Graph, httpClient, "commercial")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := scope.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
				done <- err
			}()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("request never started")
			}
			if closeSession {
				session.Close()
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("lost cancellation: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("request did not cancel")
			}
			select {
			case <-finished:
			case <-time.After(5 * time.Second):
				t.Fatal("backend did not observe cancellation")
			}
		})
	}
}

func TestTransportErrorsDoNotExposeCredentialsOrURLs(t *testing.T) {
	t.Parallel()
	responseBody := &trackedBody{Reader: strings.NewReader("fixture-secret")}
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 502, Body: responseBody, Header: make(http.Header)}, fmt.Errorf("fixture-secret %s", r.Header.Get("Authorization"))
	})}, "commercial")
	_, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture-secret"}}, nil)
	if !errors.Is(err, client.ErrTransport) || errors.Unwrap(err) != nil || strings.Contains(fmt.Sprintf("%#v", err), "fixture-secret") {
		t.Fatal("unsafe transport error retained")
	}
	if !responseBody.closed.Load() {
		t.Fatal("response accompanying transport error leaked body")
	}
}
