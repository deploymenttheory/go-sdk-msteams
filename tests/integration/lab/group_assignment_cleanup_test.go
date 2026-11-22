package lab_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitGroupRecoveryKeepsPoliciesUntilAssignmentIsAbsent(t *testing.T) {
	t.Parallel()
	discovery, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	assigned, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/policy_assignments/testdata/group_application_access_created_get.json")
	if err != nil {
		t.Fatal(err)
	}
	const group = "00000000-0000-4000-8000-000000000020"
	steps := []contracttest.Exchange{
		applicationCleanupExchange("GET", "/Teams.Tenant/serviceDiscovery", http.StatusOK, string(discovery)),
		applicationCleanupExchange("GET", "/Skype.Policy/groupPolicyAssignments/"+group+"/policyTypes/ApplicationAccessPolicy", http.StatusOK, string(assigned)),
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{
		Name: "group-policy-cleanup-dependency", Family: "application_access_lab_cleanup", Scope: "service",
		Source: "tests/integration/lab/group_assignment_cleanup_test.go", Steps: steps,
		Want: contracttest.Expected{Error: lab.ErrUnresolved},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			handler := groupOwnedPolicyCleanup{policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, groupAssignmentCleanup{sdk, group}}
			return contracttest.Observation{Err: handler.Delete(ctx, lab.Entry{Fixture: applicationCleanupFixture, ResourceID: applicationCleanupFixture})}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("policy deletion did not respect group restoration: %v %+v", err, result.Differences())
	}
}
