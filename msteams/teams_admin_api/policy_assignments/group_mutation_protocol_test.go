package policyassignments_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// The effective public NewExpanded binding requires nonempty PolicyName. The
// lower-level GroupAssignPayload permits omitted fields; that does not make
// omission a valid New invocation or an assignment-clear operation.
func TestUnitGroupCreateRequiresPolicyNameBeforeCredentials(t *testing.T) {
	t.Parallel()
	for _, name := range []policy.Field[string]{{}, policy.Null[string](), policy.Set("")} {
		run(t, contracttest.Case{Name: "missing-policy-name", Source: "msteams/teams_admin_api/policy_assignments/group_mutation_protocol_test.go", Mappings: mapping("group-create"), Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.CreateGroup(ctx, "fixture-group", "TeamsMeetingPolicy", resource.GroupSettings{PolicyName: name})
			if value != nil || response != nil {
				t.Fatal("missing mandatory name reached discovery or fabricated acceptance")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	}
	// Whitespace is not empty. Preserve its literal bytes in a rejected-wire
	// proof, without claiming it names a valid policy or tenant operation.
	run(t, contracttest.Case{Name: "literal-space-name", Source: "msteams/teams_admin_api/policy_assignments/group_mutation_protocol_test.go", Mappings: mapping("group-create"), Steps: []contracttest.Exchange{discovery(t), wire("PUT", origin+"/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/TeamsMeetingPolicy", "New-CsGroupPolicyAssignment", json.RawMessage(`{"policyName":" "}`), 403, "")}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.CreateGroup(ctx, "fixture-group", "TeamsMeetingPolicy", resource.GroupSettings{PolicyName: policy.Set(" ")})
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})
}

// Primary-authored created/accepted responses from pinned private.dll's
// NewExpanded.onCreated201 and Remove.onAccepted202 Task<IBatchOperationId>.
// They share the string OperationId model with ordinary batch acceptance;
// this does not prove a common polling route or effective assignment outcome.
func TestUnitGroupMutationHandlesPreserveAcceptanceWithoutPolling(t *testing.T) {
	t.Parallel()
	const path = "/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/TeamsMeetingPolicy"
	const reply = `{"OperationId":"fixture/operation:@","Future":false}`
	for _, tc := range []struct {
		name, method, command string
		status                int
		body                  json.RawMessage
	}{
		{"group-create", "PUT", "New-CsGroupPolicyAssignment", 201, json.RawMessage(`{"policyName":"fixture-policy"}`)},
		{"group-delete", "DELETE", "Remove-CsGroupPolicyAssignment", 202, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "opaque-handle-" + tc.name, Source: "msteams/teams_admin_api/policy_assignments/group_mutation_protocol_test.go", Mappings: mapping(tc.name), Steps: []contracttest.Exchange{discovery(t), wire(tc.method, origin+path, tc.command, tc.body, tc.status, reply)}, Want: contracttest.Expected{Body: []byte(reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.PolicyAssignments
				var observed contracttest.Observation
				var value *resource.OperationResult
				if tc.method == "PUT" {
					value, observed.Response, observed.Err = s.CreateGroup(ctx, "fixture-group", "TeamsMeetingPolicy", resource.GroupSettings{PolicyName: policy.Set("fixture-policy")})
				} else {
					value, observed.Response, observed.Err = s.DeleteGroup(ctx, "fixture-group", "TeamsMeetingPolicy")
				}
				if observed.Err != nil || observed.Response == nil || observed.Response.StatusCode() != tc.status || value == nil || !value.HTTPAccepted || value.OperationID == nil || *value.OperationID != "fixture/operation:@" || string(value.RawFields["Future"]) != "false" {
					t.Fatal("group mutation lost its independent opaque receipt or acceptance")
				}
				observed.Value = value
				return observed
			}})
		})
	}
}
