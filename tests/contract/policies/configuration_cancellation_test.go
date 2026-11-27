package policies_test

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
	clientconfig "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

type configurationRoundTrip func(*http.Request) (*http.Response, error)

func (f configurationRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type configurationCancelBody struct {
	io.Reader
	cancel context.CancelFunc
}

func (b configurationCancelBody) Close() error { b.cancel(); return nil }

// Cancel at response-body close, without sleeps or a concurrent timing race.
// The HTTP read receipt exists, but cancellation prevents typed preparation and
// PATCH. A fresh caller context can still use another service on the same client.
func TestUnitConfigurationCanceledReadRetainsReceiptAndAllowsIndependentRead(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	before := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_client_configuration/testdata/get_global.json")
	const independent = `{"Identity":"Global","EnablePrivacyMode":false}`
	calls := 0
	transport := configurationRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.Host != "api.interfaces.records.teams.microsoft.com" {
			t.Fatal("cancellation allowed a mutation or unexpected destination")
		}
		var body io.ReadCloser
		switch calls {
		case 1:
			if r.URL.Path != "/Skype.Policy/configurations/TeamsClientConfiguration/configuration/Global" || r.Header.Get("X-MS-CmdletName") != "Set-CsTeamsClientConfiguration" {
				t.Fatal("update did not start with its own preparation read")
			}
			body = configurationCancelBody{Reader: strings.NewReader(string(before)), cancel: cancel}
		case 2:
			if r.URL.Path != "/Skype.Policy/configurations/PrivacyConfiguration/configuration/Global" || r.Header.Get("X-MS-CmdletName") != "Get-CsPrivacyConfiguration" {
				t.Fatal("fresh context did not reach the independent resource")
			}
			body = io.NopCloser(strings.NewReader(independent))
		default:
			t.Fatal("canceled read was replayed or triggered an extra request")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: body, Request: r}, nil
	})
	sdk, err := msteams.NewClient(&auth.Config{
		Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application,
		AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "synthetic-token", ExpiresAt: time.Now().Add(time.Hour)}},
	}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal("synthetic root client construction failed")
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error("synthetic root client close failed")
		}
	})
	result, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.UpdateByIdentity(ctx, "Global", clientconfig.UpdateRequest{Settings: clientconfig.Settings{AllowEmailIntoChannel: policy.Set(true)}})
	if !errors.Is(err, context.Canceled) || calls != 1 || result == nil || response == nil || result.ReadResponse != response {
		t.Fatal("canceled update lost the read receipt or sent another request")
	}
	if response.StatusCode() != http.StatusOK || string(response.Body()) != string(before) || result.Before != nil || result.After != nil || result.PatchInvoked || result.HTTPAccepted {
		t.Fatal("canceled preparation invented a typed snapshot, write or acceptance")
	}
	value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(t.Context(), "Global")
	if err != nil || response == nil || value == nil || value.Identity == nil || *value.Identity != "Global" || value.EnablePrivacyMode == nil || *value.EnablePrivacyMode || calls != 2 {
		t.Fatal("canceled operation poisoned an independent context or resource")
	}
	if result.ReadResponse == nil || string(result.ReadResponse.Body()) != string(before) || result.PatchInvoked || result.HTTPAccepted {
		t.Fatal("later read changed the earlier cancellation outcome")
	}
}
