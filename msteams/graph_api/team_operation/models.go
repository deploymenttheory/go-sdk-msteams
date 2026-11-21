// Package teamoperation manages the pinned Graph beta team-operation status
// resource. Resource mutations initiate operations through their own services.
package teamoperation

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Status remains open to future read values. Only known states can establish a
// polling outcome; an unknown value is never interpreted as completion.
type Status string

const (
	NotStarted Status = "notStarted"
	InProgress Status = "inProgress"
	Succeeded  Status = "succeeded"
	Failed     Status = "failed"
)

// Type is the server's operation kind, retained without a closed enum validator.
type Type string

// Operation is a native status snapshot, including nullable/future properties.
// TargetResourceLocation is opaque metadata and is never followed or parsed.
// RawFields retains property presence and the original values for inspection.
type Operation struct {
	ID                     string                     `json:"id"`
	OperationType          *Type                      `json:"operationType,omitempty"`
	Status                 *Status                    `json:"status,omitempty"`
	CreatedDateTime        *time.Time                 `json:"createdDateTime,omitempty"`
	LastActionDateTime     *time.Time                 `json:"lastActionDateTime,omitempty"`
	AttemptsCount          *int64                     `json:"attemptsCount,omitempty"`
	TargetResourceID       *string                    `json:"targetResourceId,omitempty"`
	TargetResourceLocation *string                    `json:"targetResourceLocation,omitempty"`
	Error                  *OperationError            `json:"error,omitempty"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

func (Operation) String() string     { return "team operation snapshot (redacted)" }
func (o Operation) GoString() string { return o.String() }

func (o *Operation) UnmarshalJSON(data []byte) error {
	type snapshot Operation
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*o = Operation(value)
	return nil
}

// OperationError retains server diagnostics without rendering them by default.
// Missing fields and future nested diagnostics remain distinguishable in RawFields.
type OperationError struct {
	Code      *string                    `json:"code,omitempty"`
	Message   *string                    `json:"message,omitempty"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (OperationError) String() string     { return "team operation error (redacted)" }
func (e OperationError) GoString() string { return e.String() }

func (e *OperationError) UnmarshalJSON(data []byte) error {
	type snapshot OperationError
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*e = OperationError(value)
	return nil
}

// State explicitly interprets a snapshot's status. Missing/unknown states remain
// available in the read model but fail this interpretation.
func (o Operation) State() (client.OperationState, error) {
	if o.Status == nil {
		return "", client.ErrOperationState
	}
	switch *o.Status {
	case NotStarted:
		return client.OperationAccepted, nil
	case InProgress:
		return client.OperationRunning, nil
	case Succeeded:
		return client.OperationSucceeded, nil
	case Failed:
		return client.OperationFailed, nil
	default:
		return "", client.ErrOperationState
	}
}
