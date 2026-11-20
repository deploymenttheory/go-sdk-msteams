package teamsadminapi

import (
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// NewBatchPolicyOperation binds an evidenced batch-policy status route to its
// operation ID. The resource service supplies the route; this helper does not
// invent one from an ID or assume every Teams job has the same envelope.
func NewBatchPolicyOperation(scope *client.Scope, request client.Request, id string, accepted *resty.Response) (*client.Operation, error) {
	if scope == nil || scope.Backend() != client.TeamsAdmin || id == "" || request.Method != http.MethodGet {
		return nil, client.ErrOperation
	}
	request.Replay = client.ReplaySafe
	return client.NewOperation(scope, request, id, DecodeBatchPolicyOperation, 0, accepted)
}

// DecodeBatchPolicyOperation maps the pinned BatchJobStatus JSON fields. Complete
// is not necessarily success: errorCount and returned user results must be checked.
// Missing errorCount or inconsistent incomplete members cannot prove success.
func DecodeBatchPolicyOperation(body []byte) (client.OperationStatus, error) {
	var wire struct {
		ID      string `json:"operationId"`
		Status  string `json:"overallStatus"`
		Errors  *int   `json:"errorCount"`
		Pending *int   `json:"notStartedCount"`
		Running *int   `json:"inProgressCount"`
		Users   []struct {
			State  string `json:"state"`
			Result string `json:"result"`
		} `json:"userState"`
	}
	if json.Unmarshal(body, &wire) != nil || wire.ID == "" {
		return client.OperationStatus{}, client.ErrDecode
	}
	status := client.OperationStatus{ID: wire.ID}
	switch wire.Status {
	case "NotStarted":
		status.State = client.OperationAccepted
	case "InProgress":
		status.State = client.OperationRunning
	case "Complete":
		if wire.Errors == nil || *wire.Errors < 0 || (wire.Pending != nil && *wire.Pending != 0) || (wire.Running != nil && *wire.Running != 0) {
			return client.OperationStatus{}, client.ErrOperationState
		}
		status.State = client.OperationSucceeded
		if *wire.Errors > 0 {
			status.State = client.OperationFailed
		}
		for _, user := range wire.Users {
			if user.State != "Completed" || user.Result == "" {
				return client.OperationStatus{}, client.ErrOperationState
			}
			if user.Result != "Success" {
				status.State = client.OperationFailed
			}
		}
	default:
		return client.OperationStatus{}, client.ErrOperationState
	}
	return status, nil
}
