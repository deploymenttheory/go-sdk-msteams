package applicationaccesspolicyacc

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
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
	steps := []testutil.Exchange{
		labkit.CleanupExchange("GET", "/Teams.Tenant/serviceDiscovery", http.StatusOK, string(discovery)),
		labkit.CleanupExchange("GET", "/Skype.Policy/groupPolicyAssignments/"+group+"/policyTypes/ApplicationAccessPolicy", http.StatusOK, string(assigned)),
	}
	result, err := testutil.Run(t.Context(), testutil.Case{
		Name:  "group-policy-cleanup-dependency",
		Steps: steps,
		Want:  testutil.Expected{Error: lab.ErrUnresolved},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			handler := groupOwnedPolicyCleanup{policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, groupAssignmentCleanup{sdk, group}}
			return testutil.Observation{Err: handler.Delete(ctx, lab.Entry{Fixture: applicationCleanupFixture, ResourceID: applicationCleanupFixture})}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("policy deletion did not respect group restoration: %v %+v", err, result.Differences())
	}
}
