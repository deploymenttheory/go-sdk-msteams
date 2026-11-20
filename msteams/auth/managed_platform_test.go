package auth

import (
	"context"
	"crypto/sha1" //nolint:gosec // G505: Service Fabric publishes SHA-1 certificate thumbprints; this fixture tests platform pinning.
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/managedidentity"
)

func TestManagedIdentityServiceFabricKeepsCertificatePinning(t *testing.T) {
	clearManagedEnvironment(t)
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("Secret") != "synthetic-platform-secret" {
			t.Error("missing Service Fabric secret")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "synthetic-fabric", "expires_on": strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10), "token_type": "Bearer", "resource": r.URL.Query().Get("resource")})
	}))
	defer server.Close()
	t.Setenv("IDENTITY_ENDPOINT", server.URL+"/token")
	t.Setenv("IDENTITY_HEADER", "synthetic-platform-secret")
	//nolint:gosec // G401: Service Fabric's endpoint certificate pin is defined as SHA-1 by the platform.
	thumbprint := sha1.Sum(server.Certificate().Raw)
	t.Setenv("IDENTITY_SERVER_THUMBPRINT", hex.EncodeToString(thumbprint[:]))
	// A plain transport lets the Azure library install its certificate pin; the
	// self-signed test certificate is not trusted by the default system roots.
	base := http.DefaultTransport.(*http.Transport).Clone()
	defer base.CloseIdleConnections()
	for _, wrong := range []bool{false, true} {
		t.Run(strconv.FormatBool(wrong), func(t *testing.T) {
			if wrong {
				t.Setenv("IDENTITY_SERVER_THUMBPRINT", strings.Repeat("0", 40))
			}
			p, err := NewManagedIdentityProvider(&Config{Method: ManagedIdentity, TenantID: appTenant}, &http.Client{Transport: base})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { p.Close(); p.Wait() }()
			token, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience))
			if wrong {
				if !errors.Is(err, ErrManagedIdentityUnavailable) {
					t.Fatal(err)
				}
			} else if err != nil || token.AccessToken != "synthetic-fabric" {
				t.Fatalf("result: %v", err)
			}
		})
	}
	if calls.Load() != 1 {
		t.Fatal("bad certificate pin reached handler")
	}
}

func TestManagedIdentityArcRejectsUntrustedSecretFileChallenge(t *testing.T) {
	clearManagedEnvironment(t)
	t.Setenv("IDENTITY_ENDPOINT", "http://identity.test/token")
	t.Setenv("IMDS_ENDPOINT", "http://identity.test")
	calls := 0
	base := tokenRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 401, Header: http.Header{"Www-Authenticate": {"Basic realm=/untrusted/path/secret.key"}}, Body: http.NoBody, Request: r}, nil
	})
	p, err := NewManagedIdentityProvider(&Config{Method: ManagedIdentity, TenantID: appTenant}, &http.Client{Transport: base})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { p.Close(); p.Wait() }()
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); !errors.Is(err, ErrManagedIdentityRejected) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("untrusted challenge sent a second request")
	}
}

func TestManagedIdentityGuardRejectsChangedDestinationAndIdentity(t *testing.T) {
	clearManagedEnvironment(t)
	env := map[string]string{"IDENTITY_ENDPOINT": "http://identity.test/token"}
	t.Setenv("IDENTITY_ENDPOINT", env["IDENTITY_ENDPOINT"])
	for _, destination := range []string{"http://other.test/token?resource=" + GraphAudience, "http://identity.test/other?resource=" + GraphAudience, "http://identity.test/token?resource=other", "http://identity.test/token?resource=" + GraphAudience + "&client_id=wrong"} {
		guard, err := newManagedGuard(managedidentity.AppService, env, appID, GraphAudience)
		if err != nil {
			t.Fatal(err)
		}
		request, err := runtime.NewRequest(context.Background(), http.MethodGet, destination)
		if err != nil {
			t.Fatal(err)
		}
		response, err := guard.Do(request)
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		if !errors.Is(err, ErrAuthorityDestination) {
			t.Fatal("untrusted request reached empty pipeline")
		}
	}
	for _, endpoint := range []string{"http://user:secret@identity.test/token", "http://identity.test/token?client_id=wrong", "http://identity.test/token#fragment", "file:///token"} {
		if _, err := newManagedGuard(managedidentity.AppService, map[string]string{"IDENTITY_ENDPOINT": endpoint}, "", GraphAudience); !errors.Is(err, ErrAuthorityDestination) {
			t.Fatal("invalid endpoint accepted")
		}
	}
}
