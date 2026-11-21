package policyassignments_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

type oceTransport func(*http.Request) (*http.Response, error)

func (f oceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOCEHeadersShareOneCorrelationPerOperationWithoutDiscovery(t *testing.T) {
	t.Parallel()
	uuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	var correlations []string
	transport := oceTransport(func(r *http.Request) (*http.Response, error) {
		id := r.Header.Get("X-MS-Correlation-Id")
		correlations = append(correlations, id)
		if !uuid.MatchString(id) || r.Header.Get("X-MS-Tenant-Id") != oceTenant || r.Header.Get("MPACmdlet") != "" || r.Header.Get("X-MS-CmdletName") != "" || r.Header.Get("X-MS-Target-Uri") != "" || r.URL.Path == "/Teams.Tenant/serviceDiscovery" {
			t.Error("OCE sender headers or discovery boundary changed")
		}
		if r.URL.Host != "api.interfaces.records.teams.microsoft.com" || r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("OCE origin or credential audience changed")
		}
		status, body := 200, `{"OperationId":"`+oceOperation+`"}`
		if len(correlations) == 1 {
			status, body = 404, `{}`
		}
		if len(correlations) == 3 {
			body = `[]`
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-auth-tenant", IdentityID: "fixture-operator", TokenMode: auth.Delegated, AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "fixture-only", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sdk.Close() }()
	s := sdk.TeamsAdminAPI.PolicyAssignments
	if _, _, err := s.GetOCEBatchStatus(t.Context(), resource.OCEContext{TenantID: oceTenant}, oceOperation, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ListOCEGroup(t.Context(), resource.OCEContext{TenantID: oceTenant}, oceGroup); err != nil {
		t.Fatal(err)
	}
	if len(correlations) != 3 || correlations[0] != correlations[1] || correlations[0] == correlations[2] {
		t.Fatal("OCE correlation lifecycle differs from command initialization")
	}
}

func TestOCEContextValidationPreventsCredentialForwarding(t *testing.T) {
	t.Parallel()
	r := run(t, contracttest.Case{Name: "oce-context-before-tokens", Source: oceSource, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.PolicyAssignments
		for _, input := range []resource.OCEContext{{}, {TenantID: "fixture"}, {TenantID: oceTenant, Environment: "unknown"}} {
			if _, _, e := s.ListOCEGroups(ctx, input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid OCE context accepted")}
			}
		}
		if _, _, e := s.ListOCEGroups(ctx, resource.OCEContext{TenantID: oceTenant, Environment: resource.OCEDOD}); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("OCE operation changed the configured cloud")}
		}
		input := resource.OCEContext{TenantID: oceTenant}
		if _, _, e := s.GetOCEBatchStatus(ctx, input, "not-a-guid", nil); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid regional operation ID accepted")}
		}
		if _, _, e := s.GetOCEBatchStatus(ctx, input, oceOperation, new(resource.Region("unknown"))); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid explicit region accepted")}
		}
		if _, _, e := s.RefreshOCEGroupUsers(ctx, input, oceGroup, resource.RegionNOAM, resource.OCERefreshOptions{UserID: new("invalid")}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid refresh user accepted")}
		}
		if _, _, e := s.ReprocessOCEGroupAssignment(ctx, input, oceGroup, "..", resource.RegionNOAM); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("unsafe OCE policy segment accepted")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).ListOCEGroups(ctx, input); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("OCE wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.ListOCEGroups(ctx, input); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing OCE service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.ReprocessOCEBatch(canceled, input, oceOperation, resource.RegionNOAM); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("OCE cancellation lost")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{r})
}
