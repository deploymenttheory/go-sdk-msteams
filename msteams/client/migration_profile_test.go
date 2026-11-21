package client_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitPrivateMigrationRejectsOtherCloudPrefixesOnSameHost(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc"} {
		provider := &tokenRecorder{}
		session, err := client.NewSession(&auth.Config{Method: auth.ExternalProvider, TenantID: "fixture", Cloud: cloud, IdentityID: "fixture", TokenMode: auth.Application, Provider: provider}, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
			t.Error("rejected namespace reached HTTP")
			return nil, errors.New("unexpected")
		})}, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(session.Close)
		scope, err := client.NewScope(client.PrivateChannelMigration, session)
		if err != nil {
			t.Fatal(err)
		}
		wrong := "https://teams.microsoft.com/fabric/gcc/templates/api/channels/privateChannelMigrationStatus"
		if cloud == "gcc" {
			wrong = "https://teams.microsoft.com/fabric/templates/api/channels/privateChannelMigrationStatus"
		}
		for _, url := range []string{wrong, "https://teams.microsoft.com/fabric/templates/api-foreign/channels/privateChannelMigrationStatus", "https://foreign.invalid/fabric/templates/api/channels/privateChannelMigrationStatus"} {
			if _, err := scope.Do(t.Context(), client.Request{Method: "GET", URL: url}, nil); !errors.Is(err, client.ErrDestination) {
				t.Fatal("foreign namespace accepted")
			}
		}
		if len(provider.requests) != 0 {
			t.Fatal("rejected namespace acquired token")
		}
	}
}

func TestUnitRegionScopeSurvivesRejectedTokenRefresh(t *testing.T) {
	t.Parallel()
	const audience = "https://migrationsvc.teams.microsoft.com"
	const requestedScope = audience + "/Region.Read.All"
	binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture", Mode: auth.Delegated}
	var exchanges atomic.Int32
	provider, err := auth.NewRefreshProvider(binding, func(ctx context.Context, request auth.TokenRequest) (auth.Token, error) {
		if request != binding.RequestScope(audience, requestedScope) {
			t.Error("refresh changed exact scope/binding")
		}
		return auth.Token{AccessToken: fmt.Sprintf("opaque-%d", exchanges.Add(1)), ExpiresAt: time.Now().Add(time.Hour)}, nil
	}, auth.RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { provider.Close(); provider.Wait() })
	calls := 0
	session, err := client.NewSession(&auth.Config{Method: auth.ExternalProvider, TenantID: "fixture", IdentityID: "fixture", TokenMode: auth.Delegated, Provider: provider}, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		status, body := 200, `{"Region":"NAM"}`
		if calls == 1 {
			status, body = 401, `{"error":{"code":"Expired"}}`
		}
		if r.URL.Host != "teams.cloud.microsoft" || r.Header.Get("Authorization") != fmt.Sprintf("Bearer opaque-%d", calls) {
			t.Error("region route/token changed")
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	scope, err := client.NewScope(client.Regions, session)
	if err != nil {
		t.Fatal(err)
	}
	response, err := scope.Do(t.Context(), client.Request{Method: "GET", Path: []string{"api", "tms", "prod", "users", "fixture", "region"}, Replay: client.ReplaySafe, RefreshOnUnauthorized: true}, nil)
	if err != nil || response.StatusCode() != 200 || calls != 2 || exchanges.Load() != 2 {
		t.Fatal("scoped rejected-token refresh failed")
	}
}
