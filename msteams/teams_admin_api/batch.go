package teamsadminapi

import (
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// GrantActionResult preserves one pinned ConfigAPI GrantActionStatus. Result is
// service text, not an instruction to retry the whole operation.
type GrantActionResult struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Result string `json:"result"`
}

// BatchPolicyResult preserves counts and returned per-user outcomes from a batch
// policy operation. Missing userState means details were not supplied, not that
// there were no affected users. Summary and detail requests remain resource-owned.
type BatchPolicyResult struct {
	OperationID     string              `json:"operationId"`
	OverallStatus   string              `json:"overallStatus"`
	CompletedCount  *int                `json:"completedCount"`
	ErrorCount      *int                `json:"errorCount"`
	NotStartedCount *int                `json:"notStartedCount"`
	InProgressCount *int                `json:"inProgressCount"`
	Users           []GrantActionResult `json:"userState"`
}

func (GrantActionResult) String() string     { return "msteams batch user result (redacted)" }
func (r GrantActionResult) GoString() string { return r.String() }
func (BatchPolicyResult) String() string     { return "msteams batch policy result (redacted)" }
func (r BatchPolicyResult) GoString() string { return r.String() }

// DecodeBatchPolicyResult retains known outcomes even alongside an aggregate
// failure or inconsistent completion state. It uses HTTP-04's pinned status rules
// and never submits or retries a mutation. Unknown per-user result text is retained.
func DecodeBatchPolicyResult(body []byte) (BatchPolicyResult, error) {
	var result BatchPolicyResult
	if json.Unmarshal(body, &result) != nil {
		return result, client.ErrDecode
	}
	status, err := DecodeBatchPolicyOperation(body)
	if err != nil {
		return result, err
	}
	seen := make(map[string]bool, len(result.Users))
	for _, item := range result.Users {
		id := strings.ToLower(item.ID)
		if id == "" || seen[id] {
			return result, client.ErrDecode
		}
		seen[id] = true
	}
	if status.State == client.OperationFailed {
		return result, client.ErrOperationFailed
	}
	return result, nil
}
