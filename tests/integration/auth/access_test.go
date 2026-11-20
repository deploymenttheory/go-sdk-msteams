package auth_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// TestApplicationAccessLab uses only the explicitly selected dedicated lab app.
// Its expected grant is Organization.Read.All, with no User.Read.All or directory
// role. Requests are read-only; assertions never log response bodies or identifiers.
func TestApplicationAccessLab(t *testing.T) {
	if os.Getenv("MSTEAMS_LIVE_ACCESS") != "1" {
		t.Skip("live operation access checks are opt-in")
	}
	file := os.Getenv("MSTEAMS_LAB_IDENTITY_FILE")
	if file == "" {
		t.Fatal("explicit private lab identity file is required")
	}
	var state struct {
		Tenant      string `json:"tenant_id"`
		Client      string `json:"client_id"`
		Secret      string `json:"client_secret"`
		Certificate string `json:"certificate_file"`
		Key         string `json:"private_key_file"`
	}
	raw := readPrivateLabFile(t, file)
	defer clear(raw)
	if json.Unmarshal(raw, &state) != nil {
		t.Fatal("invalid private lab identity file")
	}
	certificate, key := readPrivateLabFile(t, state.Certificate), readPrivateLabFile(t, state.Key)
	defer clear(key)
	for _, cfg := range []*auth.Config{
		{Method: auth.ClientSecret, TenantID: state.Tenant, ClientID: state.Client, ClientSecret: state.Secret},
		{Method: auth.Certificate, TenantID: state.Tenant, ClientID: state.Client, Certificate: certificate, PrivateKey: key},
	} {
		t.Run(string(cfg.Method), func(t *testing.T) {
			sdk, err := msteams.NewClient(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := sdk.Close(); err != nil {
					t.Error(err)
				}
			})
			ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
			defer cancel()
			scope := sdk.GraphAPI.Scope()
			for _, audience := range []string{auth.GraphAudience, auth.TeamsAdminAudience} {
				credential, err := scope.Session().AcquireToken(ctx, audience)
				if err != nil {
					t.Fatal(err)
				}
				if credential.AccessToken == "" || !credential.ExpiresAt.After(time.Now()) {
					t.Fatal("invalid acquired credential")
				}
			}
			response, err := scope.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "organization"},
				Query: map[string][]string{"$select": {"id"}}}, nil)
			if err != nil || response == nil || response.StatusCode() != http.StatusOK {
				t.Fatal("expected permitted organization read")
			}
			response, err = scope.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "users"},
				Query: map[string][]string{"$select": {"id"}, "$top": {"1"}}}, nil)
			if response == nil || response.StatusCode() != http.StatusForbidden || !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("expected forbidden user read for the deliberately limited lab app")
			}
			if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrTenantCapabilityUnavailable) || errors.Is(err, auth.ErrUnsupportedOperationMode) {
				t.Fatal("forbidden access was misclassified")
			}
		})
	}
	// A deliberately invalid opaque token exercises the actual server rejection,
	// without revoking or modifying the lab application's credentials.
	t.Run("invalid-supplied-token", func(t *testing.T) {
		sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: state.Tenant,
			IdentityID: state.Client, TokenMode: auth.Application, AccessTokens: map[string]auth.Token{
				auth.GraphAudience: {AccessToken: "invalid-" + rand.Text(), ExpiresAt: time.Now().Add(time.Minute)},
			}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := sdk.Close(); err != nil {
				t.Error(err)
			}
		})
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()
		response, err := sdk.GraphAPI.Scope().Do(ctx, client.Request{Method: http.MethodGet,
			Path: []string{"v1.0", "organization"}, Query: map[string][]string{"$select": {"id"}}}, nil)
		if response == nil || response.StatusCode() != http.StatusUnauthorized || !errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatal("expected unauthenticated result for invalid supplied token")
		}
	})
}

func readPrivateLabFile(t *testing.T, file string) []byte {
	t.Helper()
	//nolint:gosec // G304: explicit opt-in operator-owned lab file and its credential paths; no service input reaches this helper.
	value, err := os.ReadFile(file)
	if err != nil {
		t.Fatal("private lab file unavailable")
	}
	return value
}
