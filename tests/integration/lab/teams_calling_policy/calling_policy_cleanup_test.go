package teamscallingpolicyacc

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestUnitCallingCleanupReconcilesRejectedCreate(t *testing.T) {
	t.Parallel()
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-calling-policy"
	const base = "/Skype.Policy/configurations/TeamsCallingPolicy"
	list, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_calling_policy/testdata/list.json")
	if err != nil {
		t.Fatal(err)
	}
	missing, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_calling_policy/testdata/get_deleted_404.json")
	if err != nil {
		t.Fatal(err)
	}
	wire := func(path string, status int, body []byte) testutil.Exchange {
		return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {"Get-CsTeamsCallingPolicy"}, "X-MS-Target-Uri": nil}}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
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
			result, err := testutil.Run(t.Context(), testutil.Case{Name: scenario, Steps: []testutil.Exchange{wire(base, 200, list), wire(base+"/configuration/"+name, status, body)}, Want: testutil.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				id, err := (callingPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallingPolicies}).Resolve(ctx, lab.Entry{Fixture: name})
				if id != "" {
					t.Fatal("missing owned policy resolved to a different object")
				}
				return testutil.Observation{Err: err}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("rejected-create cleanup differs: %v %+v", err, result.Differences())
			}
		})
	}
}
