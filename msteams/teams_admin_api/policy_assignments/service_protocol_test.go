package policyassignments_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// Primary-authored ordinary generated batch responses, not OCE or captures.
// Pinned private.dll 8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6:
// GetCsBatchPolicyAssignmentOperation_Get/Get1 onOk bind SimpleBatchJobStatus[]
// and BatchJobStatus; NewExpanded onAccepted binds BatchOperationId whose ToJson
// IL64 key is OperationId. Status keys are lowercamel, IDs are opaque strings.
// This differs from OCE's Pascal-key, GUID-native regional response protocol.
const authoredSimpleBatch = `{"operationId":"fixture/operation:@","operationName":"Fixture","overallStatus":"InProgress","notStartedCount":0,"inProgressCount":1,"completedCount":0,"errorCount":0,"createdBy":"fixture-creator","createdTime":"2026-01-01T12:00:00Z","completedTime":null,"Future":false}`
const authoredBatch = `{"operationId":"fixture/operation:@","operationName":"Fixture","overallStatus":"InProgress","notStartedCount":0,"inProgressCount":1,"completedCount":0,"errorCount":0,"createdBy":"fixture-creator","createdTime":"2026-01-01T12:00:00Z","completedTime":null,"userState":[{"id":"fixture/user:@","state":"InProgress","result":""}],"Future":false}`

func verifyAuthoredBatch(t *testing.T, value *resource.BatchStatus) {
	t.Helper()
	if value == nil || value.OperationID == nil || *value.OperationID != "fixture/operation:@" || value.NotStartedCount == nil || *value.NotStartedCount != 0 || value.InProgressCount == nil || *value.InProgressCount != 1 || value.CompletedTime != nil || string(value.RawFields["completedTime"]) != "null" || string(value.RawFields["Future"]) != "false" {
		t.Fatal("batch status lost opaque identity, counters, null or future snapshot")
	}
}

func TestUnitBatchAuthoredTypedResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, suffix, reply string }{
		{"batch-list", "", "[" + authoredSimpleBatch + "]"},
		{"batch-list-filter", "?status=Future%20%2F%3F%26%3A%40", "[" + authoredSimpleBatch + "]"},
		{"batch-get", "/fixture%2Foperation%3A%40", authoredBatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: "typed-" + tc.name, Family: "policy_assignments", Scope: "service", Source: "msteams/teams_admin_api/policy_assignments/service_protocol_test.go", Mappings: mapping(tc.name), Steps: []contracttest.Exchange{discovery(t), wire("GET", origin+"/Skype.Policy/assignments/operations"+tc.suffix, "Get-CsBatchPolicyAssignmentOperation", nil, 200, tc.reply)}, Want: contracttest.Expected{Status: 200, Body: []byte(tc.reply), JSON: json.RawMessage(tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observed := invoke(tc.name, ctx, sdk)
				if observed.Err != nil {
					return observed
				}
				if tc.name == "batch-get" {
					value, ok := observed.Value.(*resource.BatchStatus)
					if !ok {
						t.Fatal("batch item returned the wrong type")
					}
					verifyAuthoredBatch(t, value)
					if len(value.UserState) != 1 || value.UserState[0] == nil || value.UserState[0].ID == nil || *value.UserState[0].ID != "fixture/user:@" || value.UserState[0].Result == nil || *value.UserState[0].Result != "" {
						t.Fatal("batch detail lost distinct user identity or explicit empty outcome")
					}
				} else {
					values, ok := observed.Value.([]*resource.BatchStatus)
					if !ok || len(values) != 1 {
						t.Fatal("batch collection did not return a nonempty typed slice")
					}
					verifyAuthoredBatch(t, values[0])
				}
				return observed
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("authored batch response differs: %v %+v", err, r.Differences())
			}
		})
	}
}

func TestUnitBatchAuthoredAcceptedHandlePresence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, reply string
		id          *string
		decode      bool
	}{
		{"opaque", `{"OperationId":"fixture/operation:@","Future":false}`, new("fixture/operation:@"), false},
		{"missing", `{}`, nil, false},
		{"null", `{"OperationId":null}`, nil, false},
		{"empty", `{"OperationId":""}`, new(""), false},
		{"malformed", `{"OperationId":`, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := contracttest.Expected{Status: 202, Body: []byte(tc.reply)}
			if tc.decode {
				want.Error = client.ErrDecode
			}
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: "batch-handle-" + tc.name, Family: "policy_assignments", Scope: "service", Source: "msteams/teams_admin_api/policy_assignments/service_protocol_test.go", Mappings: mapping("batch-create-empty"), Steps: []contracttest.Exchange{discovery(t), wire("POST", origin+"/Skype.Policy/assignments/operations", "New-CsBatchPolicyAssignmentOperation", json.RawMessage(`{"identities":[]}`), 202, tc.reply)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.CreateBatch(ctx, resource.BatchCreateRequest{Identities: []string{}})
				if value == nil || !value.HTTPAccepted || (tc.id == nil && value.OperationID != nil) || (tc.id != nil && (value.OperationID == nil || *value.OperationID != *tc.id)) || (!tc.decode && value.RawFields == nil) {
					t.Fatal("accepted batch conflated missing, null, empty or opaque handle")
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("batch receipt presence differs: %v %+v", err, r.Differences())
			}
		})
	}
}
