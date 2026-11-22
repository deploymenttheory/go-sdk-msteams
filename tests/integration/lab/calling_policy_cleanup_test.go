package lab_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitCallingCleanupReconcilesRejectedCreate(t *testing.T) {
	t.Parallel()
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-calling-policy"
	const base = "/Skype.Policy/configurations/TeamsCallingPolicy"
	list, err := os.ReadFile("../../../msteams/teams_admin_api/teams_calling_policy/testdata/list.json")
	if err != nil {
		t.Fatal(err)
	}
	missing, err := os.ReadFile("../../../msteams/teams_admin_api/teams_calling_policy/testdata/get_deleted_404.json")
	if err != nil {
		t.Fatal(err)
	}
	wire := func(path string, status int, body []byte) contracttest.Exchange {
		return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {"Get-CsTeamsCallingPolicy"}, "X-MS-Target-Uri": nil}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
	}
	for _, denied := range []bool{false, true} {
		scenario := "verified-absence"
		status, body := 404, missing
		var want error
		if denied {
			// Inject denial to ensure a list miss cannot hide a failed second read.
			scenario, status, body, want = "independent-read-denied", 403, nil, auth.ErrForbidden
		}
		t.Run(scenario, func(t *testing.T) {
			result, err := contracttest.Run(t.Context(), contracttest.Case{Name: scenario, Family: "calling_policy_cleanup", Scope: "service", Source: "tests/integration/lab/calling_policy_cleanup_test.go", Steps: []contracttest.Exchange{wire(base, 200, list), wire(base+"/configuration/"+name, status, body)}, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				id, err := (callingPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallingPolicies}).Resolve(ctx, lab.Entry{Fixture: name})
				if id != "" {
					t.Fatal("missing owned policy resolved to a different object")
				}
				return contracttest.Observation{Err: err}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("rejected-create cleanup differs: %v %+v", err, result.Differences())
			}
		})
	}
}
