package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestSessionOwnsMachineProvidersWithoutAcquiring(t *testing.T) {
	t.Parallel()
	for _, method := range []auth.Method{auth.ManagedIdentity, auth.WorkloadIdentity} {
		t.Run(string(method), func(t *testing.T) {
			cfg := &auth.Config{Method: method, TenantID: "11111111-1111-4111-8111-111111111111", ClientID: "22222222-2222-4222-8222-222222222222"}
			if method == auth.WorkloadIdentity {
				cfg.Assertion = func(context.Context) (string, error) { t.Error("constructor invoked assertion source"); return "", nil }
			}
			session, err := NewSession(cfg, &http.Client{Transport: forbiddenTransport{t}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			provider, ok := session.provider.(interface {
				auth.Provider
				Binding() auth.Binding
				Wait()
			})
			if !ok {
				session.Close()
				t.Fatal("owned machine provider not wired")
			}
			session.Close()
			if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(auth.GraphAudience)); !errors.Is(err, auth.ErrProviderClosed) {
				t.Fatal("session left provider usable")
			}
			provider.Wait()
		})
	}
}
