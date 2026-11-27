package policyassignments_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
const authoredOCEGroup = `{"groupId":"` + oceGroup + `","priority":0,"policyType":"Fixture","policyName":"","createdBy":"fixture-creator","createdTime":"2026-01-01T12:00:00Z","Future":false}`
const authoredOCESimpleStatus = `{"OperationId":"` + oceOperation + `","CreatedBy":"` + oceTenant + `","OperationName":"Fixture","OverallStatus":"InProgress","NotStartedCount":0,"InProgressCount":1,"CompletedCount":0,"ErrorCount":0,"Future":false}`
const authoredOCEStatus = `{"OperationId":"` + oceOperation + `","CreatedBy":"` + oceTenant + `","OperationName":"Fixture","OverallStatus":"InProgress","NotStartedCount":0,"InProgressCount":1,"CompletedCount":0,"ErrorCount":0,"UserState":[{"Id":"fixture-user","AlternateId":null,"State":"InProgress","Result":""}],"Future":false}`
const authoredOCEDefinition = `{"OperationId":"` + oceOperation + `","OperationName":"Fixture","OperationType":"FixtureType","Commands":[{"Name":"FixtureCommand","RetryCount":0,"Properties":{"Flag":false,"Count":0},"Description":""}],"Ids":["fixture-user"],"AcceptTime":9007199254740993,"Future":false}`

func verifyAuthoredOCEGroup(t *testing.T, value *resource.GroupAssignment) {
	t.Helper()
	if value == nil || value.GroupID == nil || *value.GroupID != oceGroup || value.Priority == nil || *value.Priority != 0 || value.PolicyType == nil || *value.PolicyType != "Fixture" || value.PolicyName == nil || *value.PolicyName != "" || string(value.RawFields["Future"]) != "false" {
		t.Fatal("OCE group lost its identity, zero rank, empty policy or raw extension")
	}
}

func verifyAuthoredOCEStatus(t *testing.T, value *resource.OCEBatchStatus) {
	t.Helper()
	if value == nil || value.Region != resource.RegionNOAM || value.OperationID == nil || *value.OperationID != oceOperation || value.NotStartedCount == nil || *value.NotStartedCount != 0 || value.InProgressCount == nil || *value.InProgressCount != 1 || string(value.RawFields["Future"]) != "false" {
		t.Fatal("regional status lost its operation identity, origin region or typed counters")
	}
}

func TestUnitOCEAuthoredTypedResponses(t *testing.T) {
	t.Parallel()
	for _, vector := range oceRequestVectors() {
		t.Run(vector.name, func(t *testing.T) {
			status, reply := 200, authoredOCEGroup
			switch vector.name {
			case "batch-status":
				reply = authoredOCEStatus
			case "batch-list":
				reply = "[" + authoredOCESimpleStatus + "]"
			case "batch-definition":
				reply = authoredOCEDefinition
			case "batch-reprocess":
				status, reply = 202, `"`+oceOperation+`"`
			case "group-list", "group-list-type", "group-list-all":
				reply = "[" + authoredOCEGroup + "]"
			case "group-members", "group-members-all":
				reply = `[{"@odata.type":"#microsoft.graph.user","id":"` + oceTenant + `","Future":false}]`
			case "user-membership":
				reply = `[{"id":"` + oceGroup + `","groupTypes":["Unified"],"mailEnabled":false,"securityEnabled":true,"visibility":"Private","createdDateTime":"2026-01-01T12:00:00Z","deletedDateTime":null,"Future":false}]`
			case "group-refresh":
				reply = "null"
			case "group-reprocess":
				reply = `""`
			}
			want := testutil.Expected{Status: status, Body: []byte(reply)}
			if (strings.HasPrefix(vector.name, "group-") && vector.name != "group-refresh" && vector.name != "group-reprocess") || vector.name == "user-membership" {
				want.JSON = json.RawMessage(reply)
			}
			r, err := testutil.Run(t.Context(), testutil.Case{Name: "typed-" + vector.name, Steps: []testutil.Exchange{oceWire(vector.method, vector.url, status, reply)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				observed := invokeOCE(vector.name, ctx, sdk)
				if observed.Err != nil {
					return observed
				}
				switch vector.name {
				case "group-get":
					value, ok := observed.Value.(*resource.GroupAssignment)
					if !ok {
						t.Fatal("OCE group read returned the wrong type")
					}
					verifyAuthoredOCEGroup(t, value)
				case "group-list", "group-list-type", "group-list-all":
					values, ok := observed.Value.([]*resource.GroupAssignment)
					if !ok || len(values) != 1 {
						t.Fatal("OCE group list did not return a nonempty typed slice")
					}
					verifyAuthoredOCEGroup(t, values[0])
				case "group-members", "group-members-all":
					values, ok := observed.Value.([]*resource.OCEGroupMember)
					if !ok || len(values) != 1 || values[0] == nil || values[0].ID == nil || *values[0].ID != oceTenant || values[0].DataType == nil || *values[0].DataType != "#microsoft.graph.user" || string(values[0].RawFields["Future"]) != "false" {
						t.Fatal("member projection lost typed identity/type or raw data")
					}
				case "user-membership":
					values, ok := observed.Value.([]*resource.OCEUserMembership)
					if !ok || len(values) != 1 || values[0] == nil || values[0].ID == nil || *values[0].ID != oceGroup || values[0].MailEnabled == nil || *values[0].MailEnabled || values[0].SecurityEnabled == nil || !*values[0].SecurityEnabled || values[0].DeletedDateTime != nil || string(values[0].RawFields["deletedDateTime"]) != "null" || len(values[0].GroupTypes) != 1 {
						t.Fatal("membership lost identity, explicit false/null or group types")
					}
				case "batch-status":
					value, ok := observed.Value.(*resource.RegionalLookup[resource.OCEBatchStatus])
					if !ok || value == nil || len(value.Attempts) != 1 || value.Attempts[0].Err != nil || value.Region != resource.RegionNOAM {
						t.Fatal("status lookup lost its typed region/attempt receipt")
					}
					verifyAuthoredOCEStatus(t, value.Value)
					if len(value.Value.UserState) != 1 || value.Value.UserState[0] == nil || value.Value.UserState[0].AlternateID != nil || string(value.Value.UserState[0].RawFields["AlternateId"]) != "null" || value.Value.UserState[0].Result == nil || *value.Value.UserState[0].Result != "" {
						t.Fatal("extended status lost explicit-null alternate identity or empty outcome")
					}
				case "batch-list":
					value, ok := observed.Value.(*resource.RegionalList[resource.OCEBatchStatus])
					if !ok || value == nil || len(value.Values) != 1 || len(value.Attempts) != 1 || value.Attempts[0].Err != nil {
						t.Fatal("status list lost its typed results or regional receipt")
					}
					verifyAuthoredOCEStatus(t, value.Values[0])
				case "batch-definition":
					value, ok := observed.Value.(*resource.RegionalLookup[resource.OCEBatchDefinition])
					if !ok || value == nil || value.Value == nil || value.Region != resource.RegionNOAM || value.Value.Region != resource.RegionNOAM || value.Value.OperationID == nil || *value.Value.OperationID != oceOperation || value.Value.AcceptTime == nil || *value.Value.AcceptTime != 9007199254740993 || len(value.Value.Commands) != 1 || value.Value.Commands[0] == nil || value.Value.Commands[0].RetryCount == nil || *value.Value.Commands[0].RetryCount != 0 || string(value.Value.Commands[0].Properties["Flag"]) != "false" {
						t.Fatal("definition lost its operation, region, exact Int64 or command dictionary")
					}
				case "batch-reprocess":
					value, ok := observed.Value.(*resource.OCEBatchReprocessResult)
					if !ok || value == nil || !value.HTTPAccepted || value.OperationID == nil || *value.OperationID != oceOperation {
						t.Fatal("batch reprocess receipt lost its typed GUID handle")
					}
				default:
					value, ok := observed.Value.(*resource.OCEActionResult)
					if !ok || value == nil || !value.HTTPAccepted || (vector.name == "group-refresh" && value.Value != nil) || (vector.name == "group-reprocess" && (value.Value == nil || *value.Value != "")) {
						t.Fatal("GPA action confused null, empty string and batch handle")
					}
				}
				return observed
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("authored typed OCE response differs: %v %+v", err, r.Differences())
			}
		})
	}
}

// Source GPATorusCmdletBase.SubmitHTTPRequest<T> IL111 accepts every 2xx.
// 206 here is an authored transport branch, not an observed service response.
// Explicit LROS status requirements remain independent of that generic helper.
func TestUnitOCEGenericSuccessRangePreservesTypedAndDecodeResults(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, operation, reply string
		wantErr                error
	}{
		{name: "generic-typed", operation: "group-get", reply: authoredOCEGroup},
		{name: "generic-malformed", operation: "group-get", reply: `{"groupId":`, wantErr: client.ErrDecode},
		{name: "lros-still-exact", operation: "batch-status", reply: authoredOCEStatus, wantErr: resource.ErrResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var vector oceRequestVector
			for _, candidate := range oceRequestVectors() {
				if candidate.name == tc.operation {
					vector = candidate
				}
			}
			c := testutil.Case{Name: tc.name, Steps: []testutil.Exchange{oceWire(vector.method, vector.url, 206, tc.reply)}, Want: testutil.Expected{Status: 206, Body: []byte(tc.reply), Error: tc.wantErr}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				observed := invokeOCE(tc.operation, ctx, sdk)
				if !errors.Is(observed.Err, tc.wantErr) || observed.Response == nil || observed.Response.StatusCode() != 206 || string(observed.Response.Body()) != tc.reply {
					t.Fatal("source success/error branch lost its response or error")
				}
				if tc.wantErr == nil {
					value, ok := observed.Value.(*resource.GroupAssignment)
					if !ok {
						t.Fatal("generic successful status did not reach typed decoding")
					}
					verifyAuthoredOCEGroup(t, value)
				} else if tc.operation == "batch-status" {
					value, ok := observed.Value.(*resource.RegionalLookup[resource.OCEBatchStatus])
					if !ok || value == nil || value.Value != nil || len(value.Attempts) != 1 || !errors.Is(value.Attempts[0].Err, resource.ErrResponse) {
						t.Fatal("explicit LROS status list accepted generic 2xx or invented a status value")
					}
				}
				return observed
			}}
			result, err := testutil.Run(t.Context(), c)
			if err != nil || !result.Passed() {
				t.Fatalf("source success range differs: %v %+v", err, result.Differences())
			}
		})
	}
}

// GPA actions decode nullable String; batch reprocess constructs a Guid after
// its exact202 check. An object is deliberately adversarial for all three
// contracts, not a fabricated upstream success sample or business outcome.
func TestUnitOCEActionDecodeFailuresPreserveAcceptedResponse(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"group-refresh", "group-reprocess", "batch-reprocess"} {
		t.Run(operation, func(t *testing.T) {
			var vector oceRequestVector
			for _, candidate := range oceRequestVectors() {
				if candidate.name == operation {
					vector = candidate
				}
			}
			status := 200
			if operation == "batch-reprocess" {
				status = 202
			}
			c := testutil.Case{Name: "decode-" + operation, Steps: []testutil.Exchange{oceWire(vector.method, vector.url, status, `{}`)}, Want: testutil.Expected{Status: status, Body: []byte(`{}`), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				observed := invokeOCE(operation, ctx, sdk)
				if !errors.Is(observed.Err, client.ErrDecode) || observed.Response == nil || observed.Response.StatusCode() != status || string(observed.Response.Body()) != `{}` {
					t.Fatal("accepted malformed action lost decode error or response")
				}
				switch value := observed.Value.(type) {
				case *resource.OCEActionResult:
					if value == nil || !value.HTTPAccepted || value.Value != nil {
						t.Fatal("malformed scalar erased acceptance or fabricated a result")
					}
				case *resource.OCEBatchReprocessResult:
					if value == nil || !value.HTTPAccepted || value.OperationID != nil {
						t.Fatal("malformed batch handle erased acceptance or fabricated an operation ID")
					}
				default:
					t.Fatal("action returned an unexpected result type")
				}
				return observed
			}}
			result, err := testutil.Run(t.Context(), c)
			if err != nil || !result.Passed() {
				t.Fatalf("action protocol failure differs: %v %+v", err, result.Differences())
			}
		})
	}
}
