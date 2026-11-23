package migration

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("migration input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected migration response")

// Meeting migration states and types documented by Get-CsMeetingMigrationStatus.
// Other values pass through unchanged; the service owns validation.
const (
	StatePending    = "Pending"
	StateInProgress = "InProgress"
	StateFailed     = "Failed"
	StateSucceeded  = "Succeeded"

	MigrationTypeSfbToSfb     = "SfbToSfb"
	MigrationTypeSfbToTeams   = "SfbToTeams"
	MigrationTypeTeamsToTeams = "TeamsToTeams"
	MigrationTypeAllToTeams   = "AllToTeams"
	MigrationTypeToSameType   = "ToSameType"
	MigrationTypeUnknown      = "Unknown"
)

// Move options documented by Move-CsTenantServiceInstance.
const (
	MoveOptionPrepForMove   = "PrepForMove"
	MoveOptionStartDualSync = "StartDualSync"
	MoveOptionFinalize      = "Finalize"
)

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// MeetingMigrationStatus is one user's meeting migration record
// (Get-CsMeetingMigrationStatus). Absent and null fields stay nil; RawFields
// keeps the complete service object.
type MeetingMigrationStatus struct {
	UserID            *string                    `json:"UserId"`
	UserPrincipalName *string                    `json:"UserPrincipalName"`
	State             *string                    `json:"State"`
	MigrationType     *string                    `json:"MigrationType"`
	CreateDate        *string                    `json:"CreateDate"`
	ModifiedDate      *string                    `json:"ModifiedDate"`
	RetryCount        *int64                     `json:"RetryCount"`
	CorrelationID     *string                    `json:"CorrelationId"`
	TotalMeetings     *int64                     `json:"TotalMeetings"`
	SucceededMeetings *int64                     `json:"SucceededMeetings"`
	FailedMeetings    *int64                     `json:"FailedMeetings"`
	InvitesUpdates    *int64                     `json:"InvitesUpdates"`
	LastMessage       *string                    `json:"LastMessage"`
	EnqueueReason     *string                    `json:"EnqueueReason"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (v *MeetingMigrationStatus) UnmarshalJSON(raw []byte) error {
	type plain MeetingMigrationStatus
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MeetingMigrationStatus(value)
	return nil
}

// MeetingMigrationStatusSummary counts users per migration state
// (Get-CsMeetingMigrationStatus -SummaryOnly).
type MeetingMigrationStatusSummary struct {
	MigrationType *string                       `json:"MigrationType"`
	SummaryOutput []*MeetingMigrationStateCount `json:"SummaryOutput"`
	RawFields     map[string]json.RawMessage    `json:"-"`
}

func (v *MeetingMigrationStatusSummary) UnmarshalJSON(raw []byte) error {
	type plain MeetingMigrationStatusSummary
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MeetingMigrationStatusSummary(value)
	return nil
}

// MeetingMigrationStateCount is one state's user count in a summary.
type MeetingMigrationStateCount struct {
	State     *string                    `json:"State"`
	UserCount *int64                     `json:"UserCount"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *MeetingMigrationStateCount) UnmarshalJSON(raw []byte) error {
	type plain MeetingMigrationStateCount
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MeetingMigrationStateCount(value)
	return nil
}

// MeetingMigrationTransaction is one meeting migration history entry
// (Get-CsMeetingMigrationTransactionHistory).
type MeetingMigrationTransaction struct {
	UserID               *string                    `json:"UserId"`
	CorrelationID        *string                    `json:"CorrelationId"`
	ModifiedDate         *string                    `json:"ModifiedDate"`
	Operation            *string                    `json:"Operation"`
	TriggerSource        *string                    `json:"TriggerSource"`
	State                *string                    `json:"State"`
	TotalMeetings        *int64                     `json:"TotalMeetings"`
	SucceededMeetings    *int64                     `json:"SucceededMeetings"`
	FailedMeetings       *int64                     `json:"FailedMeetings"`
	InvitesUpdated       *int64                     `json:"InvitesUpdated"`
	LastErrorMessage     *string                    `json:"LastErrorMessage"`
	InternalErrorMessage *string                    `json:"InternalErrorMessage"`
	RawFields            map[string]json.RawMessage `json:"-"`
}

func (v *MeetingMigrationTransaction) UnmarshalJSON(raw []byte) error {
	type plain MeetingMigrationTransaction
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MeetingMigrationTransaction(value)
	return nil
}

// MeetingMigrationEnqueueResult is the HTTP 201 receipt of
// Start-CsExMeetingMigration.
type MeetingMigrationEnqueueResult struct {
	OperationID *string                    `json:"OperationId"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *MeetingMigrationEnqueueResult) UnmarshalJSON(raw []byte) error {
	type plain MeetingMigrationEnqueueResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MeetingMigrationEnqueueResult(value)
	return nil
}

// TenantMoveTaskStatus is one free-form tenant migration detail record
// (Get-CsMoveTenantServiceInstanceTaskStatus). The module declares no fixed
// properties and outputs every field, so they are all kept in Fields.
type TenantMoveTaskStatus struct {
	Fields map[string]json.RawMessage
}

func (v *TenantMoveTaskStatus) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	v.Fields = fields
	return nil
}

// TenantMigrationResult is the receipt of a tenant move
// (Move-CsTenantCrossRegion and Move-CsTenantServiceInstance).
type TenantMigrationResult struct {
	TenantID  *string                    `json:"TenantId"`
	JobName   *string                    `json:"JobName"`
	OrderID   *string                    `json:"OrderId"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *TenantMigrationResult) UnmarshalJSON(raw []byte) error {
	type plain TenantMigrationResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TenantMigrationResult(value)
	return nil
}

// PartitionMoveResult is one document's outcome in the HTTP 207 multi-status
// reply of Move-CsAvsTenantPartition.
type PartitionMoveResult struct {
	DocumentID         *string                    `json:"documentId"`
	SourcePartitionKey *string                    `json:"sourcePartitionKey"`
	TargetPartitionKey *string                    `json:"targetPartitionKey"`
	IsSuccessful       *bool                      `json:"isSuccessful"`
	LastCompletedStage *string                    `json:"lastCompletedStage"`
	Exception          *ProcessingException       `json:"exception"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *PartitionMoveResult) UnmarshalJSON(raw []byte) error {
	type plain PartitionMoveResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = PartitionMoveResult(value)
	return nil
}

// ProcessingException describes why a document move failed.
type ProcessingException struct {
	ErrorID       *int32                     `json:"ErrorId"`
	CorrelationID *string                    `json:"CorrelationId"`
	Message       *string                    `json:"Message"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

func (v *ProcessingException) UnmarshalJSON(raw []byte) error {
	type plain ProcessingException
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = ProcessingException(value)
	return nil
}
