package auth_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// TestAccWorkloadIdentity is explicitly enabled only by the manual lab workflow.
// Ordinary go test runs remain offline and never discover developer credentials.
func TestAccWorkloadIdentity(t *testing.T) {
	if os.Getenv("MSTEAMS_LIVE_AUTH") != "1" {
		t.Skip("live workload authentication is opt-in")
	}
	tenant, application, file := os.Getenv("MSTEAMS_TENANT_ID"), os.Getenv("MSTEAMS_CLIENT_ID"), os.Getenv("MSTEAMS_ASSERTION_FILE")
	if tenant == "" || application == "" || file == "" {
		t.Fatal("explicit lab identity and assertion file are required")
	}
	for _, source := range []string{"file", "callback"} {
		t.Run(source, func(t *testing.T) {
			cfg := &auth.Config{Method: auth.WorkloadIdentity, TenantID: tenant, ClientID: application, AssertionFile: file}
			if source == "callback" {
				cfg.AssertionFile = ""
				cfg.Assertion = func(ctx context.Context) (string, error) {
					if err := ctx.Err(); err != nil {
						return "", err
					}
					//nolint:gosec // G304: explicit opt-in lab input, created in private runner temporary storage by the trusted workflow.
					value, err := os.ReadFile(file)
					if err != nil {
						return "", auth.ErrAssertionUnavailable
					}
					defer clear(value)
					return string(value), ctx.Err()
				}
			}
			sdk, err := msteams.NewClient(cfg, msteams.WithHTTPTimeout(30*time.Second))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := sdk.Close(); err != nil {
					t.Error(err)
				}
			})
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
			defer cancel()
			for _, audience := range []string{auth.GraphAudience, auth.TeamsAdminAudience} {
				token, err := sdk.GraphAPI.Scope().Session().AcquireToken(ctx, audience)
				if err != nil {
					t.Fatalf("resource acquisition failed: %v", err)
				}
				if token.AccessToken == "" || !token.ExpiresAt.After(time.Now()) {
					t.Fatal("invalid acquired token")
				}
			}
			response, err := sdk.GraphAPI.Scope().Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "organization"}, Query: map[string][]string{"$select": {"id"}}, Replay: client.ReplaySafe}, nil)
			if err != nil {
				t.Fatalf("read-only authorization failed: %v", err)
			}
			if response.StatusCode() != http.StatusOK {
				t.Fatal("unexpected organization response status")
			}
		})
	}
}
