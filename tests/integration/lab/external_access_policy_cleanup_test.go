package lab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const externalCleanupName = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-external-access"

func externalCleanupWire(method, path, command string, status int, body []byte) contracttest.Exchange {
	headers := http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": nil}
	var input json.RawMessage
	if method == "DELETE" {
		input = json.RawMessage(`{"Identity":"` + externalCleanupName + `"}`)
		headers.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func TestUnitExternalCleanupMissingOwnedPolicyStaysUnresolved(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/external_access_policy/testdata/list.json")
	if err != nil {
		t.Fatal(err)
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "unresolved-create", Family: "external_policy_cleanup", Scope: "service", Source: "tests/integration/lab/external_access_policy_cleanup_test.go", Steps: []contracttest.Exchange{externalCleanupWire("GET", "/Skype.Policy/configurations/ExternalAccessPolicy", "Get-CsExternalAccessPolicy", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := (externalPolicyCleanup{sdk.TeamsAdminAPI.ExternalAccessPolicies}).Resolve(ctx, lab.Entry{Fixture: externalCleanupName})
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: lab.ErrUnresolved}})
	if err != nil || !result.Passed() {
		t.Fatalf("uncertain creation incorrectly resolved: %v %+v", err, result.Differences())
	}
}

func TestUnitExternalCleanupDeleteRequiresIndependentAbsence(t *testing.T) {
	t.Parallel()
	path := "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/" + externalCleanupName
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "get-denied-after-delete", Family: "external_policy_cleanup", Scope: "service", Source: "tests/integration/lab/external_access_policy_cleanup_test.go", Steps: []contracttest.Exchange{
		externalCleanupWire("DELETE", path, "Remove-CsExternalAccessPolicy", 404, nil),
		externalCleanupWire("GET", path, "Get-CsExternalAccessPolicy", 403, nil),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		err := (externalPolicyCleanup{sdk.TeamsAdminAPI.ExternalAccessPolicies}).Delete(ctx, lab.Entry{Fixture: externalCleanupName, ResourceID: externalCleanupName})
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: auth.ErrForbidden}})
	if err != nil || !result.Passed() {
		t.Fatalf("delete response incorrectly proved absence: %v %+v", err, result.Differences())
	}
}
