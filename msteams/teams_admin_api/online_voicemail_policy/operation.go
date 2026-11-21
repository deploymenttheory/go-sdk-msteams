package onlinevoicemailpolicy

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

// AssignmentStatus is a generated operation snapshot. Status names/timestamps
// remain open strings; count fields retain absence/null/zero. No terminal-state
// classification or propagation guarantee is inferred from OverallStatus.
type AssignmentStatus struct {
	OperationID     *string                    `json:"operationId"`
	OperationName   *string                    `json:"operationName"`
	OverallStatus   *string                    `json:"overallStatus"`
	CreatedBy       *string                    `json:"createdBy"`
	CreatedTime     *string                    `json:"createdTime"`
	CompletedTime   *string                    `json:"completedTime"`
	CompletedCount  *int32                     `json:"completedCount"`
	ErrorCount      *int32                     `json:"errorCount"`
	InProgressCount *int32                     `json:"inProgressCount"`
	NotStartedCount *int32                     `json:"notStartedCount"`
	UserState       []json.RawMessage          `json:"userState"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *AssignmentStatus) UnmarshalJSON(raw []byte) error {
	type snapshot AssignmentStatus
	var value snapshot
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AssignmentStatus(value)
	return nil
}
func (v AssignmentStatus) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot AssignmentStatus
	return json.Marshal(snapshot(v))
}
func (AssignmentStatus) String() string     { return "voicemail assignment status (redacted)" }
func (v AssignmentStatus) GoString() string { return v.String() }

// GetAssignmentOperation reads one opaque operation ID through the generated
// route. It does not poll, retry submissions, delete jobs or infer completion.
func (s *Service) GetAssignmentOperation(ctx context.Context, id string) (*AssignmentStatus, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(id) {
		return nil, nil, ErrInput
	}
	r, err := s.generated(ctx, "Get-CsBatchPolicyAssignmentOperation", client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "assignments", "operations", id}, Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	var value *AssignmentStatus
	if json.Unmarshal(r.Body(), &value) != nil {
		return nil, r, client.ErrDecode
	}
	return value, r, nil
}
