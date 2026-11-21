package lab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const applicationCleanupFixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-app-access"

func applicationCleanupExchange(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{
		Request: contracttest.WireRequest{
			Method:   method,
			URL:      "https://api.interfaces.records.teams.microsoft.com" + path,
			Audience: auth.TeamsAdminAudience,
		},
		Reply: contracttest.Reply{
			Status:  status,
			Headers: http.Header{"Content-Type": {"application/json"}},
			Body:    []byte(body),
		},
	}
}

func runApplicationCleanupCase(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "application_access_lab_cleanup"
	c.Source = "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json"
	c.Scope = "service"
	discovery := applicationCleanupExchange(
		"GET", "/Teams.Tenant/serviceDiscovery", 200,
		`{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid"}}`,
	)
	c.Steps = append([]contracttest.Exchange{discovery}, c.Steps...)
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("application policy cleanup differs: %+v", result.Differences())
	}
}

func TestUnitApplicationPolicyCleanupRequiresPositiveCreationResolution(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, identity string
		want                 error
	}{
		{name: "uncertain-empty-list", body: `[]`, want: lab.ErrUnresolved},
		{name: "owned-policy", body: `[{"Identity":"` + applicationCleanupFixture + `"}]`, identity: applicationCleanupFixture},
		{name: "ambiguous", body: `[{"Identity":"` + applicationCleanupFixture + `"},{"Identity":"` + applicationCleanupFixture + `"}]`, want: lab.ErrUnresolved},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runApplicationCleanupCase(t, contracttest.Case{
				Name: tc.name,
				Steps: []contracttest.Exchange{
					applicationCleanupExchange("GET", "/Skype.Policy/configurations/ApplicationAccessPolicy", 200, tc.body),
				},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					identity, err := (policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}).Resolve(ctx, lab.Entry{Fixture: applicationCleanupFixture})
					return contracttest.Observation{Value: identity, Err: err}
				},
				Want: contracttest.Expected{Error: tc.want, JSON: json.RawMessage(`"` + tc.identity + `"`)},
			})
		})
	}
}

func TestUnitApplicationPolicyCleanupVerifiesAbsenceAfterDelete(t *testing.T) {
	t.Parallel()
	const path = "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + applicationCleanupFixture
	for _, tc := range []struct {
		name  string
		steps []contracttest.Exchange
		want  error
	}{
		{
			name: "delete-acceptance-is-not-absence",
			steps: []contracttest.Exchange{
				applicationCleanupExchange("DELETE", path, 204, ""),
				applicationCleanupExchange("GET", path, 403, `{"error":{"code":"Forbidden"}}`),
			},
			want: auth.ErrForbidden,
		},
		{
			name: "verified-absence",
			steps: []contracttest.Exchange{
				applicationCleanupExchange("DELETE", path, 204, ""),
				applicationCleanupExchange("GET", path, 404, `{"error":{"code":"NotFound"}}`),
				applicationCleanupExchange("GET", "/Skype.Policy/configurations/ApplicationAccessPolicy", 200, `[]`),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runApplicationCleanupCase(t, contracttest.Case{
				Name: tc.name, Steps: tc.steps,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					err := (policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}).Delete(ctx, lab.Entry{
						Fixture: applicationCleanupFixture, ResourceID: applicationCleanupFixture,
					})
					return contracttest.Observation{Err: err}
				},
				Want: contracttest.Expected{Error: tc.want},
			})
		})
	}
}
