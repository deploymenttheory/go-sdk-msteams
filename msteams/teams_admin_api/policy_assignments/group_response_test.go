package policyassignments_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const responseGroupID = "00000000-0000-4000-8000-000000000020"
const responseGroupPath = origin + "/Skype.Policy/groupPolicyAssignments/" + responseGroupID + "/policyTypes"

func groupResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	raw, err := root.ReadFile("group_application_access_" + name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func verifyGroupRead(t *testing.T, value *resource.GroupAssignment, raw map[string]json.RawMessage) {
	t.Helper()
	if value == nil || !reflect.DeepEqual(value.RawFields, raw) {
		t.Fatal("complete actual group assignment snapshot changed")
	}
	var expected struct {
		GroupID     string `json:"groupId"`
		Priority    int32  `json:"priority"`
		PolicyType  string `json:"policyType"`
		PolicyName  string `json:"policyName"`
		CreatedBy   string `json:"createdBy"`
		CreatedTime string `json:"createdTime"`
	}
	body, err := json.Marshal(raw)
	if err != nil || json.Unmarshal(body, &expected) != nil {
		t.Fatal("invalid group response fixture")
	}
	if value.GroupID == nil || *value.GroupID != expected.GroupID || value.Priority == nil || *value.Priority != expected.Priority || value.PolicyType == nil || *value.PolicyType != expected.PolicyType || value.PolicyName == nil || *value.PolicyName != expected.PolicyName || value.CreatedBy == nil || *value.CreatedBy != expected.CreatedBy || value.CreatedTime == nil || *value.CreatedTime != expected.CreatedTime {
		t.Fatal("typed group assignment fields differ from actual response")
	}
}

func TestUnitGroupReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"created", "updated", "granted"} {
		t.Run(state, func(t *testing.T) {
			raw := groupResponseFixture(t, state+"_get.json")
			var expected map[string]json.RawMessage
			if json.Unmarshal(raw, &expected) != nil {
				t.Fatal("invalid actual group response")
			}
			run(t, contracttest.Case{Name: state, Source: "msteams/teams_admin_api/policy_assignments/testdata/group_application_access_" + state + "_get.json", Steps: []contracttest.Exchange{discovery(t), wire("GET", responseGroupPath+"/ApplicationAccessPolicy", "Get-CsGroupPolicyAssignment", nil, 200, string(raw))}, Want: contracttest.Expected{Body: raw, JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, responseGroupID, "ApplicationAccessPolicy")
				if err == nil {
					verifyGroupRead(t, value, expected)
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
		})
	}
}

func TestUnitGroupListsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"before", "created_list", "updated_list", "granted_list", "list_deleted", "restored_list_deleted"} {
		t.Run(state, func(t *testing.T) {
			raw := groupResponseFixture(t, state+".json")
			var expected []map[string]json.RawMessage
			if json.Unmarshal(raw, &expected) != nil || expected == nil {
				t.Fatal("actual group list must remain a bare array")
			}
			run(t, contracttest.Case{Name: state, Source: "msteams/teams_admin_api/policy_assignments/testdata/group_application_access_" + state + ".json", Steps: []contracttest.Exchange{discovery(t), wire("GET", responseGroupPath, "Get-CsGroupPolicyAssignment", nil, 200, string(raw))}, Want: contracttest.Expected{Body: raw, JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, responseGroupID)
				if err == nil {
					if len(values) != len(expected) {
						t.Fatal("actual group list length changed")
					}
					for i, value := range values {
						verifyGroupRead(t, value, expected[i])
					}
				}
				return contracttest.Observation{Value: values, Response: response, Err: err}
			}})
		})
	}
}

func TestUnitGroupWritesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, method, command, policyName string }{
		{"create", "PUT", "New-CsGroupPolicyAssignment", "fixture-policy-a"},
		{"update", "PATCH", "Set-CsGroupPolicyAssignment", "fixture-policy-b"},
		{"delete", "DELETE", "Remove-CsGroupPolicyAssignment", ""},
		{"restored_delete", "DELETE", "Remove-CsGroupPolicyAssignment", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := groupResponseFixture(t, tc.name+"_204.body")
			var body json.RawMessage
			if tc.policyName != "" {
				body = json.RawMessage(`{"policyName":"` + tc.policyName + `","priority":1}`)
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/policy_assignments/group_response_test.go", Steps: []contracttest.Exchange{discovery(t), wire(tc.method, responseGroupPath+"/ApplicationAccessPolicy", tc.command, body, 204, string(raw))}, Want: contracttest.Expected{Body: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				settings := resource.GroupSettings{PolicyName: policy.Set(tc.policyName), Priority: policy.Set(int32(1))}
				var observation contracttest.Observation
				var result *resource.OperationResult
				switch tc.method {
				case "PUT":
					result, observation.Response, observation.Err = sdk.TeamsAdminAPI.PolicyAssignments.CreateGroup(ctx, responseGroupID, "ApplicationAccessPolicy", settings)
				case "PATCH":
					result, observation.Response, observation.Err = sdk.TeamsAdminAPI.PolicyAssignments.UpdateGroup(ctx, responseGroupID, "ApplicationAccessPolicy", settings)
				case "DELETE":
					result, observation.Response, observation.Err = sdk.TeamsAdminAPI.PolicyAssignments.DeleteGroup(ctx, responseGroupID, "ApplicationAccessPolicy")
				}
				if observation.Err == nil && (result == nil || !result.HTTPAccepted || result.OperationID != nil || result.RawFields != nil) {
					t.Fatal("actual HTTP204 result fabricated operation data")
				}
				return observation
			}})
		})
	}
}

func TestUnitGroupDeletedReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"get_deleted_404", "restored_get_deleted_404"} {
		t.Run(name, func(t *testing.T) {
			raw := groupResponseFixture(t, name+".json")
			run(t, contracttest.Case{Name: name, Source: "msteams/teams_admin_api/policy_assignments/testdata/group_application_access_" + name + ".json", Steps: []contracttest.Exchange{discovery(t), wire("GET", responseGroupPath+"/ApplicationAccessPolicy", "Get-CsGroupPolicyAssignment", nil, 404, string(raw))}, Want: contracttest.Expected{Body: raw, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{"ClientError"}, ServiceErrorCodes: []int{40404}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, responseGroupID, "ApplicationAccessPolicy")
				if value != nil {
					t.Fatal("deleted assignment fabricated a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}
