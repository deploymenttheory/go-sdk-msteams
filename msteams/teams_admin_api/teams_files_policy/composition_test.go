package teamsfilespolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

// Without a successful mutation fixture, Set tests prove preparation, exact sparse
// PATCH requests, failure preservation and non-replay; they do not invent acceptance.
func TestUnitFilesPolicySetRetainsPreparationOnFailure(t *testing.T) {
	t.Parallel()
	realRead := readResponseFixture(t, "get_global.json")
	for _, c := range []struct {
		Name         string
		Read, Patch  int
		InvalidRead  bool
		NullIdentity bool
	}{
		{"patch-rejected", 200, 400, false, false}, {"patch-throttled", 200, 429, false, false},
		{"patch-unavailable", 200, 503, false, false}, {"nullable-identity-request", 200, 400, false, true},
		{"read-denied", 403, 0, false, false}, {"read-empty", 204, 0, false, false},
		{"read-malformed", 200, 0, true, false},
	} {
		body := append([]byte(nil), realRead...)
		expectedIdentity := json.RawMessage(`"Global"`)
		if c.NullIdentity {
			var fields map[string]json.RawMessage
			if json.Unmarshal(body, &fields) != nil {
				t.Fatal("invalid real preparation fixture")
			}
			fields["Identity"] = json.RawMessage("null")
			var err error
			body, err = json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			expectedIdentity = json.RawMessage("null")
		}
		if c.InvalidRead {
			body = []byte("not-json")
		}
		if c.Read != 200 {
			body = nil
		}
		steps := []testutil.Exchange{compiledWire("GET", item, setCommand, nil, c.Read, string(body))}
		if c.Patch != 0 {
			payload, err := json.Marshal(map[string]json.RawMessage{"Identity": expectedIdentity, "NativeFileEntryPoints": json.RawMessage("null")})
			if err != nil {
				t.Fatal(err)
			}
			steps = append(steps, compiledWire("PATCH", item, setCommand, payload, c.Patch, ""))
		}
		run(t, testutil.Case{Name: "set-" + c.Name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{NativeFileEntryPoints: policy.Null[string]()}})
			if value == nil || value.ReadResponse == nil || value.PatchInvoked != (c.Patch != 0) || value.HTTPAccepted || err == nil {
				t.Fatal("Set preparation/failure outcome changed")
			}
			if c.Patch != 0 {
				if value.Before == nil || c.NullIdentity != (value.Before.Identity == nil) {
					t.Fatal("preparation snapshot or nullable identity changed")
				}
				var api *client.APIError
				if !errors.As(err, &api) || api.StatusCode != c.Patch {
					t.Fatal("failed patch lost typed status")
				}
			}
			return testutil.Observation{Response: response}
		}})
	}
}

func TestUnitFilesPolicyCompiledGrantRetainsExplicitZeroAndSignedRank(t *testing.T) {
	t.Parallel()
	for _, rank := range []int32{0, -2147483648, 2147483647} {
		body, err := json.Marshal(map[string]any{"policyName": "Fixture", "priority": rank})
		if err != nil {
			t.Fatal(err)
		}
		run(t, testutil.Case{Name: fmt.Sprintf("group-rank-%d", rank), Steps: []testutil.Exchange{compiledWire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", grantCommand, body, 400, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, response, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.AssignGroup(ctx, "group/fixture", new("Fixture"), rank)
			return testutil.Observation{Response: response, Err: err}
		}})
	}
}
