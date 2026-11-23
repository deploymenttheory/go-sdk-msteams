package applicationaccesspolicyacc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const applicationCleanupFixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-app-access"

func runApplicationCleanupCase(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "application_access_lab_cleanup"
	c.Source = "tests/integration/lab/application_access_policy/application_access_policy_cleanup_test.go"
	c.Scope = "service"
	discovery := labkit.CleanupExchange(
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
					labkit.CleanupExchange("GET", "/Skype.Policy/configurations/ApplicationAccessPolicy", 200, tc.body),
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
				labkit.CleanupExchange("DELETE", path, 204, ""),
				labkit.CleanupExchange("GET", path, 403, `{"error":{"code":"Forbidden"}}`),
			},
			want: auth.ErrForbidden,
		},
		{
			name: "verified-absence",
			steps: []contracttest.Exchange{
				labkit.CleanupExchange("DELETE", path, 204, ""),
				labkit.CleanupExchange("GET", path, 404, `{"error":{"code":"NotFound"}}`),
				labkit.CleanupExchange("GET", "/Skype.Policy/configurations/ApplicationAccessPolicy", 200, `[]`),
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

func TestUnitPolicyNamesNormalizeTagOnEitherSide(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, left, right string
		want              bool
	}{
		{"plain", "Fixture", "fixture", true},
		{"left-tag", "Tag:Fixture", "fixture", true},
		{"right-tag", "fixture", "TAG:Fixture", true},
		{"both-tags", "Tag:Fixture", "tag:fixture", true},
		{"different", "Tag:Other", "Tag:Fixture", false},
		{"no-whitespace-trim", " Fixture", "Fixture", false},
		{"single-tag-only", "Tag:Tag:Fixture", "Fixture", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if labkit.SamePolicyName(test.left, test.right) != test.want {
				t.Fatal("policy identity normalization differs")
			}
		})
	}
}
