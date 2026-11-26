package extintegration

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
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	appconfig "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_policy_configuration"
	tenantconfig "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_configuration"
)

type cancelTransport func(*http.Request) (*http.Response, error)

func (f cancelTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestUnitConditionalSettingsCancellationRetainsPreparation(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"tenant", "app"} {
		t.Run(family, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			reads, writes := 0, 0
			transport := cancelTransport(func(request *http.Request) (*http.Response, error) {
				switch request.Method {
				case http.MethodGet:
					reads++
					body := `{"Identity":"Global","MaxAllowedDomains":1,"MaxBlockedDomains":null}`
					if family == "app" {
						body = `{"Identity":"Global","AppCatalogUri":"https://catalog.example.invalid/","ResourceUri":null}`
					}
					return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
				case http.MethodPatch:
					writes++
					cancel()
					return nil, context.Canceled
				default:
					t.Fatal("conditional setting used an unexpected HTTP method")
					return nil, context.Canceled
				}
			})
			sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "synthetic-offline-token", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := sdk.Close(); err != nil {
					t.Error(err)
				}
			})
			before, prepared, invoked, accepted := false, false, false, false
			if family == "tenant" {
				result, _, operationErr := sdk.TeamsAdminAPI.TenantConfigurations.UpdateByIdentity(ctx, "Global", tenantconfig.UpdateRequest{Settings: tenantconfig.Settings{MaxAllowedDomains: policy.Null[uint32]()}})
				err = operationErr
				if result != nil {
					before = result.Before != nil
					prepared = result.ReadResponse != nil
					invoked = result.PatchInvoked
					accepted = result.HTTPAccepted
				}
			} else {
				result, _, operationErr := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", appconfig.UpdateRequest{Settings: appconfig.Settings{AppCatalogURI: policy.Null[string]()}})
				err = operationErr
				if result != nil {
					before = result.Before != nil
					prepared = result.ReadResponse != nil
					invoked = result.PatchInvoked
					accepted = result.HTTPAccepted
				}
			}
			if !errors.Is(err, context.Canceled) || !before || !prepared || !invoked || accepted || reads != 1 || writes != 1 {
				t.Fatal("cancelled conditional set lost preparation, action state or no-replay guarantee")
			}
		})
	}
}
