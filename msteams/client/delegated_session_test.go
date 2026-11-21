package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitSessionOwnsDelegatedProvidersWithoutStartingLogin(t *testing.T) {
	t.Parallel()
	for _, method := range []auth.Method{auth.Browser, auth.DeviceCode, auth.Password} {
		t.Run(string(method), func(t *testing.T) {
			cfg := &auth.Config{Method: method, TenantID: "11111111-1111-4111-8111-111111111111", ClientID: "22222222-2222-4222-8222-222222222222"}
			switch method {
			case auth.Browser:
				cfg.OpenBrowser = func(context.Context, string) error { t.Error("constructor opened browser"); return nil }
			case auth.DeviceCode:
				cfg.DevicePrompt = func(context.Context, auth.DeviceCodeMessage) error { t.Error("constructor prompted user"); return nil }
			case auth.Password:
				cfg.Username, cfg.Password = "fixture@example.test", "fixture-password"
			}
			session, err := NewSession(cfg, &http.Client{Transport: forbiddenTransport{t}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			provider, ok := session.provider.(*auth.DelegatedProvider)
			if !ok {
				session.Close()
				t.Fatal("delegated provider was not wired")
			}
			session.Close()
			if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(auth.GraphAudience)); !errors.Is(err, auth.ErrProviderClosed) {
				t.Fatal("session left owned provider usable after close")
			}
			provider.Wait()
		})
	}
}
