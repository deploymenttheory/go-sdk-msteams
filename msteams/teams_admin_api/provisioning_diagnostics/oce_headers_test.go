package provisioningdiagnostics_test

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	pd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/provisioning_diagnostics"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// OCE commands skip forest discovery and the generated-cmdlet headers, send a
// fresh v4 correlation ID per call and make exactly one attempt.
func TestUnitOCERequestsSkipDiscoveryAndCorrelateEachCall(t *testing.T) {
	t.Parallel()
	uuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	var correlations []string
	transport := roundTrip(func(r *http.Request) (*http.Response, error) {
		id := r.Header.Get("X-MS-Correlation-Id")
		correlations = append(correlations, id)
		if !uuid.MatchString(id) || r.Header.Get("X-MS-Tenant-Id") != tenant || r.Header.Get("X-MS-CmdletName") != "" || r.Header.Get("X-MS-Target-Uri") != "" ||
			!strings.HasPrefix(r.URL.Path, "/Teams.InternalSupport/") || r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("OCE request headers, route or credential changed")
		}
		return &http.Response{StatusCode: 503, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-auth-tenant", IdentityID: "fixture-operator", TokenMode: auth.Delegated,
		AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "fixture-only", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sdk.Close() }()
	s := sdk.TeamsAdminAPI.ProvisioningDiagnostics
	if _, _, err := s.GetUserProvisioningHistory(t.Context(), strings.ToUpper(tenant), probe, pd.RegionAPAC); err == nil {
		t.Fatal("503 was accepted")
	}
	if _, err := s.InvokeNgtProvHandler(t.Context(), pd.NgtProvHandlerRequest{TenantID: tenant, Region: pd.DeploymentLocal, HandlerName: "h", Target: "t", Payload: payload}); err == nil {
		t.Fatal("503 was accepted")
	}
	if len(correlations) != 2 || correlations[0] == correlations[1] {
		t.Fatal("OCE calls were retried or shared a correlation ID")
	}
}
