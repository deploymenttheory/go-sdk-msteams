package appinstallation_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ai "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_installation"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestUnitCatalogMutationRetainsResolvedIDOnLostResponseAndCancellation(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"lost-response", "cancelled-write"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			reads, writes := 0, 0
			roundTrip := transport(func(request *http.Request) (*http.Response, error) {
				if request.Method == http.MethodGet {
					reads++
					return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"value":[` + item + `]}`)), Request: request}, nil
				}
				if request.Method != http.MethodDelete {
					t.Fatal("unexpected installation mutation method")
				}
				writes++
				if scenario == "cancelled-write" {
					cancel()
					return nil, context.Canceled
				}
				return nil, io.ErrUnexpectedEOF
			})
			sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "synthetic-offline-token", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: roundTrip}))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := sdk.Close(); err != nil {
					t.Error(err)
				}
			})
			result, err := sdk.GraphAPI.AppInstallations.DeleteByCatalogID(ctx, ai.Target{TeamID: "owned-team"}, catalogID)
			want := client.ErrTransport
			if scenario == "cancelled-write" {
				want = context.Canceled
			}
			if !errors.Is(err, want) || result == nil || result.CatalogID != catalogID || result.InstallationID != installedID || !result.LookupComplete || !result.ActionInvoked || result.HTTPAccepted || result.LookupResponse == nil || result.LookupResponse.StatusCode() != 200 || reads != 1 || writes != 1 {
				t.Fatalf("catalog mutation lost receipt: reads=%d writes=%d errtype=%T transport=%t cancelled=%t result=%t resolved=%t complete=%t invoked=%t accepted=%t lookup=%t", reads, writes, err, errors.Is(err, client.ErrTransport), errors.Is(err, context.Canceled), result != nil, result != nil && result.InstallationID == installedID, result != nil && result.LookupComplete, result != nil && result.ActionInvoked, result != nil && result.HTTPAccepted, result != nil && result.LookupResponse != nil)
			}
		})
	}
}
