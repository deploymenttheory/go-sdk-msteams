package shiftsconnections

import (
	"encoding/json"
	"time"
)

// ErrorReport is a Shifts connector synchronization error
// (Get-CsTeamsShiftsConnectionErrorReport). Timestamps keep their service
// text; ReferenceLinks keeps the service's name-to-URL map.
type ErrorReport struct {
	ID                         *string                    `json:"id"`
	TenantID                   *string                    `json:"tenantId"`
	WfmConnectorInstanceID     *string                    `json:"wfmConnectorInstanceId"`
	ConnectionID               *string                    `json:"connectionId"`
	TeamID                     *string                    `json:"teamId"`
	Operation                  *string                    `json:"operation"`
	Procedure                  *string                    `json:"procedure"`
	Code                       *string                    `json:"code"`
	Message                    *string                    `json:"message"`
	Parameters                 []string                   `json:"parameters"`
	ReferenceLinks             map[string]*string         `json:"referenceLinks"`
	RevisitIntervalInMinutes   *int32                     `json:"revisitIntervalInMinutes"`
	ErrorNotificationSent      *bool                      `json:"errorNotificationSent"`
	ResolvedAt                 *string                    `json:"resolvedAt"`
	ResolvedNotificationSentOn *string                    `json:"resolvedNotificationSentOn"`
	RevisitedAt                *string                    `json:"revisitedAt"`
	TotalIncidents             *int32                     `json:"totalIncidents"`
	IntermediateIncidents      *int32                     `json:"intermediateIncidents"`
	CreatedAt                  *string                    `json:"createdAt"`
	Culture                    *string                    `json:"culture"`
	ScheduleSequenceNumber     *int64                     `json:"scheduleSequenceNumber"`
	Severity                   *string                    `json:"severity"`
	ErrorType                  *string                    `json:"errorType"`
	TTL                        *int32                     `json:"ttl"`
	RawFields                  map[string]json.RawMessage `json:"-"`
}

func (v *ErrorReport) UnmarshalJSON(raw []byte) error {
	type plain ErrorReport
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = ErrorReport(value)
	return nil
}

// ErrorReportListOptions filters Get-CsTeamsShiftsConnectionErrorReport. Nil
// fields are omitted from the query. Before and After keep the caller's
// timestamp text, as the module does.
type ErrorReportListOptions struct {
	Activeness          *string
	ConnectorInstanceID *string
	ConnectionID        *string
	TeamID              *string
	Operation           *string
	Procedure           *string
	Code                *string
	Before              *string
	After               *string
}

// OperationDetail is one step of a Shifts connector operation. Properties
// keeps the service's values as raw JSON.
type OperationDetail struct {
	Status     *string                    `json:"status"`
	Error      *string                    `json:"error"`
	Properties map[string]json.RawMessage `json:"properties"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

func (v *OperationDetail) UnmarshalJSON(raw []byte) error {
	type plain OperationDetail
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OperationDetail(value)
	return nil
}

// Operation is an asynchronous Shifts connector operation, such as a batch
// team mapping (Get-CsTeamsShiftsConnectionOperation). The service names its
// identifier "Id".
type Operation struct {
	ID                     *string                    `json:"Id"`
	Type                   *string                    `json:"type"`
	TenantID               *string                    `json:"tenantId"`
	WfmConnectorInstanceID *string                    `json:"wfmConnectorInstanceId"`
	CreatedDateTime        *string                    `json:"createdDateTime"`
	LastActionDateTime     *string                    `json:"lastActionDateTime"`
	Status                 *string                    `json:"status"`
	Details                []*OperationDetail         `json:"details"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

func (v *Operation) UnmarshalJSON(raw []byte) error {
	type plain Operation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Operation(value)
	return nil
}

// User is a WFM or Microsoft Entra user the connector reports by ID and name
// (Get-CsTeamsShiftsConnectionWfmUser, and the failed users of a sync result).
type User struct {
	ID        *string                    `json:"id"`
	Name      *string                    `json:"name"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *User) UnmarshalJSON(raw []byte) error {
	type plain User
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = User(value)
	return nil
}

// UserMapping is a user the connector matched between Teams and the WFM system.
type UserMapping struct {
	UserID      *string                    `json:"userId"`
	UserName    *string                    `json:"userName"`
	WfmUserID   *string                    `json:"wfmUserId"`
	WfmUserName *string                    `json:"wfmUserName"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *UserMapping) UnmarshalJSON(raw []byte) error {
	type plain UserMapping
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = UserMapping(value)
	return nil
}

// SyncResult is the user synchronization result for one mapped team
// (Get-CsTeamsShiftsConnectionSyncResult).
type SyncResult struct {
	SuccessfulUsers []*UserMapping             `json:"successfulUsers"`
	FailedWfmUsers  []*User                    `json:"failedWfmUsers"`
	FailedAadUsers  []*User                    `json:"failedAadUsers"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *SyncResult) UnmarshalJSON(raw []byte) error {
	type plain SyncResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = SyncResult(value)
	return nil
}

// TeamMap maps a Teams team to a WFM team for a connector instance
// (Get-CsTeamsShiftsConnectionTeamMap).
type TeamMap struct {
	TeamID      *string                    `json:"teamId"`
	TeamName    *string                    `json:"teamName"`
	WfmTeamID   *string                    `json:"wfmTeamId"`
	WfmTeamName *string                    `json:"wfmTeamName"`
	TimeZone    *string                    `json:"timeZone"`
	State       *string                    `json:"state"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *TeamMap) UnmarshalJSON(raw []byte) error {
	type plain TeamMap
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TeamMap(value)
	return nil
}

// WfmTeam is a team in the WFM system (Get-CsTeamsShiftsConnectionWfmTeam).
// Teams read through a connection also carry the connector instance and the
// mapped Teams team, when there is one.
type WfmTeam struct {
	ID                  *string                    `json:"id"`
	Name                *string                    `json:"name"`
	ConnectorInstanceID *string                    `json:"connectorInstanceId"`
	TeamID              *string                    `json:"teamId"`
	RawFields           map[string]json.RawMessage `json:"-"`
}

func (v *WfmTeam) UnmarshalJSON(raw []byte) error {
	type plain WfmTeam
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = WfmTeam(value)
	return nil
}

// TeamMapping is one Teams-to-WFM team mapping for
// New-CsTeamsShiftsConnectionBatchTeamMap. A nil TimeZone is omitted.
type TeamMapping struct {
	TeamID    string  `json:"teamId"`
	WfmTeamID string  `json:"wfmTeamId"`
	TimeZone  *string `json:"timeZone,omitempty"`
}

// TeamMapOperation is the accepted batch team mapping. Poll its OperationID
// with GetOperationByID.
type TeamMapOperation struct {
	OperationID        *string                    `json:"operationId"`
	CreatedDateTime    *string                    `json:"createdDateTime"`
	LastActionDateTime *string                    `json:"lastActionDateTime"`
	Status             *string                    `json:"status"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *TeamMapOperation) UnmarshalJSON(raw []byte) error {
	type plain TeamMapOperation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TeamMapOperation(value)
	return nil
}

// ClearScheduleRequest is Remove-CsTeamsShiftsScheduleRecord: it deletes
// Shifts schedule records of the given entity types from a team. TeamID and
// at least one EntityTypes value are required, as in the module. The date
// range is sent only when StartDate or EndDate is set.
type ClearScheduleRequest struct {
	TeamID               string
	EntityTypes          []string
	ClearSchedulingGroup bool
	StartDate            *time.Time
	EndDate              *time.Time
	DesignatedActorID    *string
	TimeZone             *string
}
