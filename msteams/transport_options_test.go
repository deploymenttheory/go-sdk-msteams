package msteams_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	msteams "github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func optionsAuth(graph string) *auth.Config {
	configuration := &auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture", IdentityID: "fixture", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{graph: {AccessToken: "fixture-bearer", ExpiresAt: time.Now().Add(time.Hour)}}}
	if graph != auth.GraphAudience {
		configuration.Endpoints = &auth.Endpoints{Graph: graph}
	}
	return configuration
}

func TestDefaultHeadersPreservePrecedenceAcrossRequestKinds(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"json", "upload", "download", "pager", "operation"} {
		t.Run(kind, func(t *testing.T) {
			defaults := map[string]string{"Accept": "application/fixture", "X-Fixture": "original"}
			calls := 0
			transport := diagnosticTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				wantAccept := "application/fixture"
				if calls == 2 {
					wantAccept = "application/explicit"
				}
				if r.Header.Get("Accept") != wantAccept || r.Header.Get("User-Agent") != "fixture-client/1" || r.Header.Get("X-Fixture") != "original" {
					t.Error("defaults or request precedence changed")
				}
				if r.Body != nil {
					_, _ = io.Copy(io.Discard, r.Body)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"items":[],"status":"succeeded"}`)), Request: r}, nil
			})
			option := msteams.WithGlobalHeaders(defaults)
			defaults["X-Fixture"] = "mutated"
			sdk, err := msteams.NewClient(optionsAuth(auth.GraphAudience), msteams.WithTransport(transport), option, msteams.WithUserAgent("fixture-client/1"))
			if err != nil {
				t.Fatal(err)
			}
			closeClient(t, sdk)
			scope := sdk.GraphAPI.Scope()
			for attempt := range 2 {
				request := client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}
				if attempt == 1 {
					request.Header = http.Header{"accept": {"application/explicit"}}
				}
				switch kind {
				case "json":
					_, err = scope.Do(t.Context(), request, nil)
				case "upload":
					request.Method = http.MethodPost
					_, err = scope.Upload(t.Context(), request, io.NopCloser(strings.NewReader("fixture")), "application/octet-stream", 1024, nil)
				case "download":
					_, _, err = scope.Download(t.Context(), request, io.Discard, 1024)
				case "pager":
					request.Replay = client.ReplaySafe
					var pager *client.Pager[string]
					pager, err = client.NewPager(scope, request, func([]byte) (client.PageContent[string], error) { return client.PageContent[string]{}, nil }, client.PagingOptions{Style: client.LinkPaging})
					if err == nil {
						_, _, err = pager.Next(t.Context())
					}
				case "operation":
					request.Replay = client.ReplaySafe
					var operation *client.Operation
					operation, err = client.NewOperation(scope, request, "fixture", func([]byte) (client.OperationStatus, error) {
						return client.OperationStatus{ID: "fixture", State: client.OperationSucceeded}, nil
					}, 0, nil)
					if err == nil {
						_, err = operation.Poll(t.Context())
					}
				}
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestTransportOptionsRejectUnsafeOrIncompatibleConfiguration(t *testing.T) {
	t.Parallel()
	for _, option := range []msteams.ClientOption{
		msteams.WithTransport((*recordingTransport)(nil)), msteams.WithUserAgent(""),
		msteams.WithGlobalHeader("Authorization", "fixture-secret"), msteams.WithGlobalHeader("Cookie", "fixture-secret"), msteams.WithGlobalHeader("Host", "fixture-secret"),
		msteams.WithGlobalHeader("X-Fixture", "fixture\r\nsecret"), msteams.WithGlobalHeaders(map[string]string{"X-Fixture": "one", "x-fixture": "two"}),
		msteams.WithProxy("http://user:fixture-secret@proxy.test/path"), msteams.WithProxy("file:///fixture-secret"),
		msteams.WithTLSClientConfig(nil), msteams.WithTLSClientConfig(&tls.Config{InsecureSkipVerify: true}), //nolint:gosec // G402: intentionally invalid fixture must be rejected before any TLS connection.
		msteams.WithTLSClientConfig(&tls.Config{KeyLogWriter: io.Discard}),
		msteams.WithTLSClientConfig(&tls.Config{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS12}), //nolint:gosec // G402: inverted version bounds are a rejection fixture, never used by a connection.
	} {
		_, err := msteams.NewClient(optionsAuth(auth.GraphAudience), option)
		if err == nil || strings.Contains(err.Error(), "fixture-secret") {
			t.Fatal("unsafe config accepted or value leaked")
		}
	}
	_, err := msteams.NewClient(optionsAuth(auth.GraphAudience), msteams.WithTransport(&recordingTransport{}), msteams.WithProxy("http://proxy.test"))
	if !errors.Is(err, msteams.ErrInvalidOption) {
		t.Fatal("proxy silently ignored opaque transport")
	}
}

func TestManagementHeadersNeverReachIdentityExchange(t *testing.T) {
	t.Parallel()
	const tenant = "11111111-1111-4111-8111-111111111111"
	const authority = "https://login.microsoftonline.com/" + tenant
	identity, backend := 0, 0
	transport := diagnosticTransport(func(r *http.Request) (*http.Response, error) {
		var payload any
		switch r.URL.Host {
		case "login.microsoftonline.com":
			identity++
			if r.Header.Get("X-Management") != "" || r.Header.Get("User-Agent") == "fixture-management-agent" {
				t.Error("management defaults leaked to token authority")
			}
			if r.Method == http.MethodGet {
				payload = map[string]string{"issuer": authority + "/v2.0", "token_endpoint": authority + "/oauth2/v2.0/token", "authorization_endpoint": authority + "/oauth2/v2.0/authorize"}
			} else {
				payload = map[string]any{"access_token": "fixture-token", "expires_in": 3600, "token_type": "Bearer"}
			}
		case "graph.microsoft.com":
			backend++
			if r.Header.Get("X-Management") != "fixture" || r.Header.Get("User-Agent") != "fixture-management-agent" {
				t.Error("management defaults missing")
			}
			payload = map[string]any{}
		default:
			t.Fatal("unexpected request authority")
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(body)), Request: r}, nil
	})
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ClientSecret, TenantID: tenant, ClientID: "22222222-2222-4222-8222-222222222222", ClientSecret: "fixture-secret"}, msteams.WithTransport(transport), msteams.WithGlobalHeader("X-Management", "fixture"), msteams.WithUserAgent("fixture-management-agent"))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	if _, err := sdk.GraphAPI.Scope().Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil); err != nil {
		t.Fatal(err)
	}
	if identity < 2 || backend != 1 {
		t.Fatal("identity/backend separation was not exercised")
	}
}

func TestTLSOptionSnapshotsCertificateAndECHData(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(server.Close)
	leaf := server.Certificate()
	configuration := &tls.Config{MinVersion: tls.VersionTLS13, EncryptedClientHelloConfigList: []byte{1, 2, 3}, Certificates: []tls.Certificate{{Certificate: [][]byte{leaf.Raw}, Leaf: leaf}}}
	option := msteams.WithTLSClientConfig(configuration)
	configuration.EncryptedClientHelloConfigList[0] = 9
	configuration.Certificates[0].Leaf.Subject.CommonName = "mutated"
	configuration.Certificates[0].Certificate[0][0] ^= 0xff
	sdk, err := msteams.NewClient(optionsAuth(auth.GraphAudience), option)
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	configured, err := sdk.GraphAPI.Scope().Session().HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	copied := configured.Transport.(*http.Transport).TLSClientConfig
	if copied.EncryptedClientHelloConfigList[0] != 1 || copied.Certificates[0].Leaf.Subject.CommonName == "mutated" || copied.Certificates[0].Certificate[0][0] == configuration.Certificates[0].Certificate[0][0] {
		t.Fatal("TLS snapshot retained mutable caller data")
	}
}

func TestTLSAndProxyOptionsReachVerifiedBackendWithoutMutatingTransport(t *testing.T) {
	t.Parallel()
	var backendCalls, proxyCalls atomic.Int32
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendCalls.Add(1)
		if r.Header.Get("Authorization") != "Bearer fixture-bearer" || r.Header.Get("Proxy-Authorization") != "" {
			t.Error("backend/proxy credential boundary failed")
		}
		w.WriteHeader(204)
	}))
	t.Cleanup(backend.Close)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyCalls.Add(1)
		if r.Method != http.MethodConnect || r.Host != backend.Listener.Addr().String() || r.Header.Get("Authorization") != "" {
			t.Error("proxy destination or auth boundary failed")
			w.WriteHeader(400)
			return
		}
		dialer := &net.Dialer{Timeout: time.Second}
		upstream, err := dialer.DialContext(r.Context(), "tcp", backend.Listener.Addr().String())
		if err != nil {
			t.Error(err)
			w.WriteHeader(502)
			return
		}
		downstream, buffered, err := w.(http.Hijacker).Hijack()
		if err != nil {
			_ = upstream.Close()
			t.Error(err)
			return
		}
		defer func() { _ = downstream.Close(); _ = upstream.Close() }()
		if _, err := buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
			t.Error(err)
			return
		}
		if err := buffered.Flush(); err != nil {
			t.Error(err)
			return
		}
		done := make(chan struct{})
		go func() { _, _ = io.Copy(upstream, buffered); _ = upstream.Close(); close(done) }()
		_, _ = io.Copy(downstream, upstream)
		_ = downstream.Close()
		<-done
	}))
	t.Cleanup(proxy.Close)
	roots := x509.NewCertPool()
	roots.AddCert(backend.Certificate())
	tlsConfig := &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	borrowed := &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13}}
	t.Cleanup(borrowed.CloseIdleConnections)
	sdk, err := msteams.NewClient(optionsAuth(backend.URL), msteams.WithTransport(borrowed), msteams.WithTLSClientConfig(tlsConfig), msteams.WithProxy(proxy.URL))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	tlsConfig.ServerName = "mutated.invalid"
	if _, err := sdk.GraphAPI.Scope().Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil); err != nil {
		t.Fatalf("verified proxy/TLS request failed: %v", err)
	}
	if proxyCalls.Load() != 1 || backendCalls.Load() != 1 || borrowed.Proxy != nil || borrowed.TLSClientConfig.MinVersion != tls.VersionTLS13 || borrowed.TLSClientConfig.RootCAs != nil {
		t.Fatal("transport mutation or proxy bypass")
	}
	configured, err := sdk.GraphAPI.Scope().Session().HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	owned := configured.Transport.(*http.Transport)
	if owned == borrowed || owned.TLSClientConfig == tlsConfig || owned.TLSClientConfig.RootCAs == roots {
		t.Fatal("mutable TLS configuration was borrowed")
	}
}

func TestZeroSessionDiagnosticsReturnClosed(t *testing.T) {
	t.Parallel()
	scope, err := client.NewScope(client.Graph, &client.Session{})
	if err != nil {
		t.Fatal(err)
	}
	request := client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}
	if _, err := scope.Do(context.Background(), request, nil); !errors.Is(err, client.ErrClosed) {
		t.Fatal("zero session JSON did not fail closed")
	}
	if _, _, err := scope.Download(context.Background(), request, &bytes.Buffer{}, 1024); !errors.Is(err, client.ErrClosed) {
		t.Fatal("zero session download did not fail closed")
	}
	request.Method = http.MethodPost
	if _, err := scope.Upload(context.Background(), request, io.NopCloser(strings.NewReader("fixture")), "application/octet-stream", 1024, nil); !errors.Is(err, client.ErrClosed) {
		t.Fatal("zero session upload did not fail closed")
	}
}
