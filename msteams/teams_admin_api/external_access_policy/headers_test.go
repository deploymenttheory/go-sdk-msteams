package externalaccesspolicy_test

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCompiledCallsUseFreshCorrelationAndNoForestBootstrap(t *testing.T) {
	t.Parallel()
	uuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	seen := map[string]bool{}
	calls := 0
	transport := transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		id := r.Header.Get("X-MS-Correlation-Id")
		if !uuid.MatchString(id) || seen[id] || r.Header.Get("MPACmdlet") != "true" || r.Header.Get("X-MS-Target-Uri") != "" || r.URL.Path == "/Teams.Tenant/serviceDiscovery" {
			t.Error("compiled protocol headers/bootstrap changed")
		}
		seen[id] = true
		if r.URL.Host != "api.interfaces.records.teams.microsoft.com" || r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("compiled destination/audience changed")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/plain"}}, Body: io.NopCloser(strings.NewReader(`{"Identity":"Global"}`)), Request: r}, nil
	})
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "fixture-only", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sdk.Close() }()
	if _, _, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(t.Context(), "Global"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sdk.TeamsAdminAPI.ACSFederationConfigurations.GetByIdentity(t.Context(), "Global"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.GetByIdentity(t.Context(), "Global"); err != nil {
		t.Fatal(err)
	}
	if calls != 3 || len(seen) != 3 {
		t.Fatal("unexpected compiled request count")
	}
}
