package teamoperation_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	operation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
)

func TestStatusInterpretationNeverGuessesFutureCompletion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		status *operation.Status
		state  client.OperationState
		valid  bool
	}{
		{name: "absent"}, {name: "future", status: statusPointer("future")}, {name: "not-started", status: statusPointer(operation.NotStarted), state: client.OperationAccepted, valid: true},
		{name: "running", status: statusPointer(operation.InProgress), state: client.OperationRunning, valid: true}, {name: "succeeded", status: statusPointer(operation.Succeeded), state: client.OperationSucceeded, valid: true}, {name: "failed", status: statusPointer(operation.Failed), state: client.OperationFailed, valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := operation.Operation{ID: "fixture", Status: tc.status}
			state, e := v.State()
			if tc.valid {
				if e != nil || state != tc.state {
					t.Fatal("known operation state differed")
				}
			} else if !errors.Is(e, client.ErrOperationState) || state != "" {
				t.Fatal("unknown operation state was guessed")
			}
		})
	}
	var v operation.Operation
	if json.Unmarshal([]byte(`{"id":"fixture","error":null,"attemptsCount":0,"status":"future"}`), &v) != nil || v.Error != nil || v.AttemptsCount == nil || *v.AttemptsCount != 0 || string(v.RawFields["error"]) != "null" {
		t.Fatal("null or zero state was lost")
	}
	if _, exists := v.RawFields["operationType"]; exists {
		t.Fatal("missing operation type was invented")
	}
}

func statusPointer(value operation.Status) *operation.Status { return &value }
