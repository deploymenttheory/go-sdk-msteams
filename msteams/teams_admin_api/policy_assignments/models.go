// Package policyassignments manages policy assignment resources and operations.
package policyassignments

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	ErrInput    = errors.New("invalid policy assignment input")
	ErrResponse = errors.New("unexpected policy assignment response")
)

// GroupAssignment retains received rank and timestamps without applying local
// precedence rules or manufacturing absent CLR defaults. RawFields is a snapshot
// of the received object, including explicit nulls and future fields.
type GroupAssignment struct {
	GroupID     *string                    `json:"groupId"`
	Priority    *int32                     `json:"priority"`
	PolicyType  *string                    `json:"policyType"`
	PolicyName  *string                    `json:"policyName"`
	CreatedBy   *string                    `json:"createdBy"`
	CreatedTime *string                    `json:"createdTime"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

// SourceEntry describes a server-reported assignment source. AssignmentType and
// Reference remain open strings; the SDK does not calculate an effective winner.
type SourceEntry struct {
	PolicyName     *string                    `json:"policyName"`
	PolicyType     *string                    `json:"policyType"`
	AssignmentType *string                    `json:"assignmentType"`
	Reference      *string                    `json:"reference"`
	RawFields      map[string]json.RawMessage `json:"-"`
}

type EffectivePolicy struct {
	PolicyType   *string                    `json:"policyType"`
	PolicyName   *string                    `json:"policyName"`
	PolicySource []*SourceEntry             `json:"policySource"`
	RawFields    map[string]json.RawMessage `json:"-"`
}

// UserOutcome retains per-user results independently of the overall job state.
type UserOutcome struct {
	ID        *string                    `json:"id"`
	State     *string                    `json:"state"`
	Result    *string                    `json:"result"`
	RawFields map[string]json.RawMessage `json:"-"`
}

// BatchStatus represents the generated policy-assignment operation protocol.
// List responses contain the simple status fields; Get additionally exposes
// UserState. Status names are open strings, and timestamps retain their wire text.
// This type does not describe the separate regional OCE batch protocol.
type BatchStatus struct {
	OperationID     *string                    `json:"operationId"`
	OperationName   *string                    `json:"operationName"`
	OverallStatus   *string                    `json:"overallStatus"`
	NotStartedCount *int32                     `json:"notStartedCount"`
	InProgressCount *int32                     `json:"inProgressCount"`
	CompletedCount  *int32                     `json:"completedCount"`
	ErrorCount      *int32                     `json:"errorCount"`
	CreatedBy       *string                    `json:"createdBy"`
	CreatedTime     *string                    `json:"createdTime"`
	CompletedTime   *string                    `json:"completedTime"`
	UserState       []*UserOutcome             `json:"userState"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

// OperationResult separates a recognized successful HTTP status from an optional
// operation identifier. It survives a malformed accepted body and implies neither
// completed processing nor effective policy convergence.
type OperationResult struct {
	HTTPAccepted bool
	OperationID  *string
	RawFields    map[string]json.RawMessage
}

func snapshot(data []byte, value any, raw *map[string]json.RawMessage) error {
	if json.Unmarshal(data, value) != nil || json.Unmarshal(data, raw) != nil || *raw == nil {
		return client.ErrDecode
	}
	return nil
}
func marshalSnapshot(value any, raw map[string]json.RawMessage) ([]byte, error) {
	if raw != nil {
		return json.Marshal(raw)
	}
	return json.Marshal(value)
}

func (v *GroupAssignment) UnmarshalJSON(data []byte) error {
	type plain GroupAssignment
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = GroupAssignment(value)
	return nil
}
func (v GroupAssignment) MarshalJSON() ([]byte, error) {
	type plain GroupAssignment
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *SourceEntry) UnmarshalJSON(data []byte) error {
	type plain SourceEntry
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = SourceEntry(value)
	return nil
}
func (v SourceEntry) MarshalJSON() ([]byte, error) {
	type plain SourceEntry
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *EffectivePolicy) UnmarshalJSON(data []byte) error {
	type plain EffectivePolicy
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = EffectivePolicy(value)
	return nil
}
func (v EffectivePolicy) MarshalJSON() ([]byte, error) {
	type plain EffectivePolicy
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *UserOutcome) UnmarshalJSON(data []byte) error {
	type plain UserOutcome
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = UserOutcome(value)
	return nil
}
func (v UserOutcome) MarshalJSON() ([]byte, error) {
	type plain UserOutcome
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *BatchStatus) UnmarshalJSON(data []byte) error {
	type plain BatchStatus
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BatchStatus(value)
	return nil
}
func (v BatchStatus) MarshalJSON() ([]byte, error) {
	type plain BatchStatus
	return marshalSnapshot(plain(v), v.RawFields)
}

func (GroupAssignment) String() string     { return "group policy assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }
func (SourceEntry) String() string         { return "policy assignment source (redacted)" }
func (v SourceEntry) GoString() string     { return v.String() }
func (EffectivePolicy) String() string     { return "effective policy assignment (redacted)" }
func (v EffectivePolicy) GoString() string { return v.String() }
func (UserOutcome) String() string         { return "policy assignment user outcome (redacted)" }
func (v UserOutcome) GoString() string     { return v.String() }
func (BatchStatus) String() string         { return "policy assignment batch status (redacted)" }
func (v BatchStatus) GoString() string     { return v.String() }
func (OperationResult) String() string     { return "policy assignment operation result (redacted)" }
func (v OperationResult) GoString() string { return v.String() }
