package policyassignments_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func TestUnitUserReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	fixtures, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixtures.Close(); err != nil {
			t.Error(err)
		}
	})
	const user = "00000000-0000-4000-8000-000000000010"
	for _, name := range []string{"before", "assigned", "restored"} {
		for _, inputObject := range []bool{false, true} {
			label := name + "-identity"
			if inputObject {
				label = name + "-input-object"
			}
			t.Run(label, func(t *testing.T) {
				raw, err := fixtures.ReadFile("user_application_access_" + name + ".json")
				if err != nil || !json.Valid(raw) {
					t.Fatal("invalid actual user-assignment response")
				}
				var expected []map[string]json.RawMessage
				if json.Unmarshal(raw, &expected) != nil || expected == nil {
					t.Fatal("actual user response must remain a bare array")
				}
				run(t, testutil.Case{Name: label, Steps: []testutil.Exchange{discovery(t), wire("GET", origin+"/Skype.User/users/"+user+"/policies?policyType=ApplicationAccessPolicy", "Get-CsUserPolicyAssignment", nil, 200, string(raw))}, Want: testutil.Expected{Status: 200, Body: raw, JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					var values []*resource.EffectivePolicy
					if inputObject {
						values, response, e := sdk.TeamsAdminAPI.PolicyAssignments.GetUserByIdentity(ctx, resource.UserIdentity{Identity: user}, "ApplicationAccessPolicy")
						if e != nil {
							return testutil.Observation{Response: response, Err: e}
						}
						verifyUserRead(t, values, expected)
						return testutil.Observation{Value: values, Response: response}
					}
					values, response, e := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, user, "ApplicationAccessPolicy")
					if e != nil {
						return testutil.Observation{Response: response, Err: e}
					}
					verifyUserRead(t, values, expected)
					return testutil.Observation{Value: values, Response: response}
				}})
			})
		}
	}
}

func verifyUserRead(t *testing.T, values []*resource.EffectivePolicy, expected []map[string]json.RawMessage) {
	t.Helper()
	if len(values) != len(expected) {
		t.Fatal("actual user-policy count changed")
	}
	for i, value := range values {
		if value == nil || !reflect.DeepEqual(value.RawFields, expected[i]) {
			t.Fatal("complete actual policy snapshot changed")
		}
		var fields struct{ PolicyType, PolicyName string }
		body, err := json.Marshal(expected[i])
		if err != nil || json.Unmarshal(body, &fields) != nil || value.PolicyName == nil || *value.PolicyName != fields.PolicyName || value.PolicyType == nil || *value.PolicyType != fields.PolicyType {
			t.Fatal("typed effective-policy fields differ from the actual response")
		}
	}
}
