package msteams_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitSessionInformationAndCloseRemainClientLocal(t *testing.T) {
	t.Parallel()
	configs := []*auth.Config{
		{Method: auth.SuppliedTokens, TenantID: "tenant-a", IdentityID: "identity-a", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "fixture-a", ExpiresAt: time.Now().Add(time.Hour)}}},
		{Method: auth.SuppliedTokens, TenantID: "tenant-b", IdentityID: "identity-b", TokenMode: auth.Delegated, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "fixture-b", ExpiresAt: time.Now().Add(time.Hour)}}},
	}
	clients := make([]*msteams.Client, len(configs))
	for i, cfg := range configs {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		clients[i] = sdk
		t.Cleanup(func() {
			if err := sdk.Close(); err != nil {
				t.Error(err)
			}
		})
	}
	// Caller replacement/mutation cannot rebind an existing client.
	configs[0].TenantID = "replacement"
	configs[0].AccessTokens[auth.GraphAudience] = auth.Token{}
	var wg sync.WaitGroup
	for range 20 {
		for i, sdk := range clients {
			wg.Go(func() {
				info, err := sdk.SessionInfo()
				if err != nil {
					t.Error(err)
					return
				}
				wantTenant, wantIdentity, wantToken := "tenant-a", "identity-a", "fixture-a"
				if i == 1 {
					wantTenant, wantIdentity, wantToken = "tenant-b", "identity-b", "fixture-b"
				}
				if info.TenantID != wantTenant || info.IdentityID != wantIdentity || info.Cloud != "commercial" || info.Method != auth.SuppliedTokens {
					t.Error("client configuration crossed session boundary")
				}
				credential, err := sdk.GraphAPI.Scope().Session().AcquireToken(context.Background(), auth.GraphAudience)
				if err != nil || credential.AccessToken != wantToken {
					t.Error("client token state crossed session boundary")
				}
				info.TenantID = "mutated-copy"
			})
		}
	}
	wg.Wait()
	if err := clients[0].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := clients[0].SessionInfo(); !errors.Is(err, client.ErrClosed) {
		t.Fatal("closed client retained inspectable identity")
	}
	info, err := clients[1].SessionInfo()
	if err != nil || info.TenantID != "tenant-b" || info.Mode != auth.Delegated {
		t.Fatal("closing one client changed another")
	}
	for _, formatted := range []string{fmt.Sprintf("%v", info), fmt.Sprintf("%+v", info), fmt.Sprintf("%#v", info)} {
		if strings.Contains(formatted, "tenant-b") || strings.Contains(formatted, "identity-b") {
			t.Fatal("session formatting exposed identity")
		}
	}
	if _, err := clients[1].GraphAPI.Scope().Session().AcquireToken(t.Context(), auth.GraphAudience); err != nil {
		t.Fatal("closing another session invalidated this client's token")
	}
}

func TestUnitSessionInfoIsLazyAndNilSafe(t *testing.T) {
	t.Parallel()
	var sdk *msteams.Client
	if _, err := sdk.SessionInfo(); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil client information did not fail safely")
	}
	var session *client.Session
	if _, err := session.Info(); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil session information did not fail safely")
	}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.Certificate, TenantID: "11111111-1111-1111-1111-111111111111", ClientID: "22222222-2222-2222-2222-222222222222", Certificate: []byte("intentionally-unparsed"), PrivateKey: []byte("intentionally-unparsed")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := sdk.Close(); err != nil {
			t.Error(err)
		}
	}()
	info, err := sdk.SessionInfo()
	if err != nil || info.Method != auth.Certificate || info.Mode != auth.Application {
		t.Fatal("inspection authenticated or parsed credentials")
	}
}
