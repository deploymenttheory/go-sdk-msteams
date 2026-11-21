package policyassignments

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

// OCEGroupMember is the conditional internal-support member projection.
type OCEGroupMember struct {
	DataType  *string                    `json:"@odata.type"`
	ID        *string                    `json:"id"`
	RawFields map[string]json.RawMessage `json:"-"`
}

type OCEUserMembership struct {
	ID              *string                    `json:"id"`
	GroupTypes      []*string                  `json:"groupTypes"`
	MailEnabled     *bool                      `json:"mailEnabled"`
	SecurityEnabled *bool                      `json:"securityEnabled"`
	Visibility      *string                    `json:"visibility"`
	CreatedDateTime *string                    `json:"createdDateTime"`
	DeletedDateTime *string                    `json:"deletedDateTime"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

// OCEBatchStatus is the regional LROS protocol, distinct from BatchStatus.
// Region is assigned from the request, never trusted from a response payload.
// Nullable native fields preserve wire presence rather than CLR default values.
type OCEBatchStatus struct {
	Region          Region                     `json:"-"`
	OperationName   *string                    `json:"OperationName"`
	OverallStatus   *string                    `json:"OverallStatus"`
	NotStartedCount *int32                     `json:"NotStartedCount"`
	InProgressCount *int32                     `json:"InProgressCount"`
	CompletedCount  *int32                     `json:"CompletedCount"`
	ErrorCount      *int32                     `json:"ErrorCount"`
	CreatedBy       *string                    `json:"CreatedBy"`
	CreatedTime     *string                    `json:"CreatedTime"`
	CompletedTime   *string                    `json:"CompletedTime"`
	OperationID     *string                    `json:"OperationId"`
	UserState       []*OCEUserOutcome          `json:"UserState"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

type OCEUserOutcome struct {
	ID          *string                    `json:"Id"`
	AlternateID *string                    `json:"AlternateId"`
	State       *string                    `json:"State"`
	Result      *string                    `json:"Result"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

type OCEBatchDefinition struct {
	Region        Region                     `json:"-"`
	OperationID   *string                    `json:"OperationId"`
	OperationName *string                    `json:"OperationName"`
	OperationType *string                    `json:"OperationType"`
	Commands      []*OCEBatchCommand         `json:"Commands"`
	IDs           []*string                  `json:"Ids"`
	AcceptTime    *int64                     `json:"AcceptTime"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

type OCEBatchCommand struct {
	Name        *string                    `json:"Name"`
	RetryCount  *int32                     `json:"RetryCount"`
	Properties  map[string]json.RawMessage `json:"Properties"`
	Description *string                    `json:"Description"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

// RegionAttempt retains every completed attempt, including failures before a
// later successful lookup. Error/response bodies are never formatted implicitly.
type RegionAttempt struct {
	Region   Region
	Response *resty.Response
	Err      error
}

// RegionalLookup is a first-success regional lookup, with the complete attempt
// history. A later successful region does not erase earlier diagnostics.
type RegionalLookup[T any] struct {
	Value    *T
	Region   Region
	Attempts []RegionAttempt
}

// RegionalList retains partial results when another region fails. Each result's
// Region field identifies its actual request region. Errors are also per-attempt.
type RegionalList[T any] struct {
	Values   []*T
	Attempts []RegionAttempt
}

// OCEActionResult records GPA action acceptance and its nullable string response.
// Value is not treated as a polling handle or evidence of effective completion.
type OCEActionResult struct {
	HTTPAccepted bool
	Value        *string
}
type OCEBatchReprocessResult struct {
	HTTPAccepted bool
	OperationID  *string
}

func (v *OCEGroupMember) UnmarshalJSON(data []byte) error {
	type plain OCEGroupMember
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEGroupMember(value)
	return nil
}
func (v OCEGroupMember) MarshalJSON() ([]byte, error) {
	type plain OCEGroupMember
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *OCEUserMembership) UnmarshalJSON(data []byte) error {
	type plain OCEUserMembership
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEUserMembership(value)
	return nil
}
func (v OCEUserMembership) MarshalJSON() ([]byte, error) {
	type plain OCEUserMembership
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *OCEBatchStatus) UnmarshalJSON(data []byte) error {
	type plain OCEBatchStatus
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEBatchStatus(value)
	return nil
}
func (v OCEBatchStatus) MarshalJSON() ([]byte, error) {
	type plain OCEBatchStatus
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *OCEUserOutcome) UnmarshalJSON(data []byte) error {
	type plain OCEUserOutcome
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEUserOutcome(value)
	return nil
}
func (v OCEUserOutcome) MarshalJSON() ([]byte, error) {
	type plain OCEUserOutcome
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *OCEBatchDefinition) UnmarshalJSON(data []byte) error {
	type plain OCEBatchDefinition
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEBatchDefinition(value)
	return nil
}
func (v OCEBatchDefinition) MarshalJSON() ([]byte, error) {
	type plain OCEBatchDefinition
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *OCEBatchCommand) UnmarshalJSON(data []byte) error {
	type plain OCEBatchCommand
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OCEBatchCommand(value)
	return nil
}
func (v OCEBatchCommand) MarshalJSON() ([]byte, error) {
	type plain OCEBatchCommand
	return marshalSnapshot(plain(v), v.RawFields)
}

func (OCEGroupMember) String() string              { return "OCE group member (redacted)" }
func (v OCEGroupMember) GoString() string          { return v.String() }
func (OCEUserMembership) String() string           { return "OCE user membership (redacted)" }
func (v OCEUserMembership) GoString() string       { return v.String() }
func (OCEBatchStatus) String() string              { return "OCE regional batch status (redacted)" }
func (v OCEBatchStatus) GoString() string          { return v.String() }
func (OCEUserOutcome) String() string              { return "OCE user outcome (redacted)" }
func (v OCEUserOutcome) GoString() string          { return v.String() }
func (OCEBatchDefinition) String() string          { return "OCE batch definition (redacted)" }
func (v OCEBatchDefinition) GoString() string      { return v.String() }
func (OCEBatchCommand) String() string             { return "OCE batch command (redacted)" }
func (v OCEBatchCommand) GoString() string         { return v.String() }
func (RegionAttempt) String() string               { return "OCE region attempt (redacted)" }
func (v RegionAttempt) GoString() string           { return v.String() }
func (RegionalLookup[T]) String() string           { return "OCE regional lookup (redacted)" }
func (v RegionalLookup[T]) GoString() string       { return v.String() }
func (RegionalList[T]) String() string             { return "OCE regional list (redacted)" }
func (v RegionalList[T]) GoString() string         { return v.String() }
func (OCEActionResult) String() string             { return "OCE action result (redacted)" }
func (v OCEActionResult) GoString() string         { return v.String() }
func (OCEBatchReprocessResult) String() string     { return "OCE batch reprocess result (redacted)" }
func (v OCEBatchReprocessResult) GoString() string { return v.String() }
