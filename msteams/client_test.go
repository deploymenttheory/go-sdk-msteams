package msteams_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	msteams "github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

type recordingTransport struct {
	requests atomic.Int32
	closes   atomic.Int32
}

func (r *recordingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	r.requests.Add(1)
	return nil, errors.New("unexpected HTTP request")
}

func (r *recordingTransport) CloseIdleConnections() { r.closes.Add(1) }

func closeClient(t *testing.T, c *msteams.Client) {
	t.Helper()
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestExplicitClientIgnoresEnvironmentAndBorrowsTransport(t *testing.T) {
	t.Setenv("MSTEAMS_AUTH_METHOD", "invalid")
	t.Setenv("MSTEAMS_HTTP_TIMEOUT", "invalid")
	transport := &recordingTransport{}
	httpClient := &http.Client{Transport: transport, Timeout: time.Minute}
	c, err := msteams.NewClient(&auth.Config{Method: auth.ManagedIdentity, TenantID: "fixture.example"},
		msteams.WithHTTPClient(httpClient), msteams.WithHTTPTimeout(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, c)
	if c.GraphAPI.Scope().Backend() != client.Graph || c.TeamsAdminAPI.Scope().Backend() != client.TeamsAdmin {
		t.Fatal("backend families were not retained")
	}
	if c.GraphAPI.Scope().Session() != c.TeamsAdminAPI.Scope().Session() {
		t.Fatal("one client must share its session across service collections")
	}
	if httpClient.Timeout != time.Minute {
		t.Fatal("constructor mutated the caller's HTTP client")
	}
	configured, err := c.GraphAPI.Scope().Session().HTTPClient()
	if err != nil || configured.Timeout != 2*time.Second {
		t.Fatalf("timeout was not applied: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if transport.requests.Load() != 0 || transport.closes.Load() != 0 {
		t.Fatal("construction/cleanup used caller-owned HTTP resources")
	}
}

func TestEnvironmentLoadsThenExplicitOptionsOverride(t *testing.T) {
	t.Setenv("MSTEAMS_AUTH_METHOD", "invalid")
	t.Setenv("MSTEAMS_TENANT_ID", "environment-tenant")
	t.Setenv("MSTEAMS_ACCESS_TOKENS", "malformed JSON")
	t.Setenv("MSTEAMS_HTTP_TIMEOUT", "invalid duration")
	c, err := msteams.NewClientFromEnv(
		msteams.WithAuthConfig(&auth.Config{Method: auth.ManagedIdentity, TenantID: "explicit.example"}),
		msteams.WithHTTPTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, c)
	cfg, err := c.GraphAPI.Scope().Session().AuthSnapshot()
	if err != nil || cfg.TenantID != "explicit.example" {
		t.Fatalf("explicit configuration did not win: %v", err)
	}
	if _, err := msteams.NewClientFromEnv(); err == nil {
		t.Fatal("invalid environment should fail without overrides")
	}
}

func TestSuppliedTokensFromEnvironment(t *testing.T) {
	for _, name := range []string{"MSTEAMS_CLIENT_SECRET", "MSTEAMS_USERNAME", "MSTEAMS_PASSWORD", "MSTEAMS_CERTIFICATE_PEM", "MSTEAMS_PRIVATE_KEY_PEM", "MSTEAMS_ASSERTION_FILE", "MSTEAMS_HTTP_TIMEOUT"} {
		t.Setenv(name, "")
	}
	t.Setenv("MSTEAMS_AUTH_METHOD", "supplied-tokens")
	t.Setenv("MSTEAMS_TENANT_ID", "tenant")
	t.Setenv("MSTEAMS_IDENTITY_ID", "identity")
	t.Setenv("MSTEAMS_CLOUD", "commercial")
	t.Setenv("MSTEAMS_TOKEN_MODE", "application")
	t.Setenv("MSTEAMS_ACCESS_TOKENS", `{"audience":{"access_token":"test-token","expires_at":"2099-01-01T00:00:00Z"}}`)
	c, err := msteams.NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, c)
	cfg, err := c.GraphAPI.Scope().Session().AuthSnapshot()
	if err != nil || cfg.AccessTokens["audience"].AccessToken != "test-token" || cfg.Cloud != "commercial" {
		t.Fatalf("audience tokens did not load: %v", err)
	}
}

func TestInvalidClientOptions(t *testing.T) {
	t.Parallel()
	for _, option := range []msteams.ClientOption{nil, msteams.WithHTTPTimeout(0), msteams.WithHTTPTimeout(-time.Second), msteams.WithHTTPClient(nil), msteams.WithLogger(nil), msteams.WithTracerProvider(nil), msteams.WithHTTPClient(&http.Client{Transport: (*recordingTransport)(nil)}), msteams.WithHTTPClient(&http.Client{Timeout: -time.Second})} {
		_, err := msteams.NewClient(&auth.Config{Method: auth.ManagedIdentity, TenantID: "fixture.example"}, option)
		if !errors.Is(err, msteams.ErrInvalidOption) {
			t.Fatalf("expected invalid option, got %v", err)
		}
	}
	if _, err := msteams.NewClient(nil); !errors.Is(err, auth.ErrInvalidConfig) {
		t.Fatalf("expected missing authentication configuration, got %v", err)
	}
}

func TestConcurrentClientsAndSnapshotsAreIndependent(t *testing.T) {
	t.Parallel()
	var group sync.WaitGroup
	for _, tenant := range []string{"first-tenant", "second-tenant"} {
		group.Add(1)
		go func() {
			defer group.Done()
			original := &auth.Config{Method: auth.SuppliedTokens, TenantID: tenant, IdentityID: tenant,
				TokenMode: auth.Application, AccessTokens: map[string]auth.Token{"audience": {AccessToken: tenant, ExpiresAt: time.Now().Add(time.Hour)}}}
			c, err := msteams.NewClient(original)
			if err != nil {
				t.Error(err)
				return
			}
			defer func() {
				if err := c.Close(); err != nil {
					t.Error(err)
				}
			}()
			original.AccessTokens["audience"] = auth.Token{}
			for range 20 {
				snapshot, err := c.GraphAPI.Scope().Session().AuthSnapshot()
				if err != nil {
					t.Error(err)
					return
				}
				if snapshot.TenantID != tenant || snapshot.AccessTokens["audience"].AccessToken != tenant {
					t.Error("tenant or token configuration was shared")
				}
				snapshot.AccessTokens["audience"] = auth.Token{}
			}
		}()
	}
	group.Wait()
}

func TestSessionCancellationAndClose(t *testing.T) {
	t.Parallel()
	c, err := msteams.NewClient(&auth.Config{Method: auth.ManagedIdentity, TenantID: "fixture.example"})
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, c)
	session := c.GraphAPI.Scope().Session()
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	operation, cancelOperation, err := session.Begin(parent)
	if err != nil {
		t.Fatal(err)
	}
	defer cancelOperation()
	cancelParent()
	select {
	case <-operation.Done():
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not propagate")
	}
	operation, cancelOperation, err = session.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancelOperation()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-operation.Done():
	case <-time.After(time.Second):
		t.Fatal("close did not cancel active local operation")
	}
	if _, _, err := session.Begin(context.Background()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("closed session accepted work")
	}
	if _, err := session.AuthSnapshot(); !errors.Is(err, client.ErrClosed) {
		t.Fatal("closed session exposed credentials")
	}
}
