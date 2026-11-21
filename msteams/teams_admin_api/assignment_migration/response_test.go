package assignmentmigration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func readResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directory.Close(); err != nil {
			t.Error("testdata directory close failed")
		}
	})
	raw, err := directory.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatal("invalid permanent response fixture")
	}
	return raw
}

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []contracttest.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []contracttest.Exchange
	return append(steps, contracttest.Exchange{
		Request: contracttest.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-DirectToGroupAssignmentsMigrationStatus"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitListResponseFixtureError(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list_400.json")
	var envelope struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	var codes []string
	if envelope.Code != "" {
		codes = []string{envelope.Code}
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{
		Name: "list-400", Family: "assignment_migration", Scope: "service", Source: "msteams/teams_admin_api/assignment_migration/testdata/list_400.json",
		Steps: responseFixtureSteps(t, "/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration", 400, raw),
		Want:  contracttest.Expected{Status: 400, Body: raw, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400, Codes: codes}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AssignmentMigrations.List(ctx)
			var apiErr *client.APIError
			if value != nil || !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
				t.Fatal("actual error response did not remain a typed API failure")
			}
			return contracttest.Observation{Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}
