package shiftsconnections

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("shifts connection input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected shifts connection response")

func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// Connector is a workforce management system the Shifts connector supports
// (Get-CsTeamsShiftsConnectionConnector). SupportedSyncScenarios keeps the
// service's per-scenario values as raw JSON.
type Connector struct {
	ID                     *string                    `json:"id"`
	Name                   *string                    `json:"name"`
	Version                *string                    `json:"version"`
	SupportedSyncScenarios map[string]json.RawMessage `json:"supportedSyncScenarios"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

func (v *Connector) UnmarshalJSON(raw []byte) error {
	type plain Connector
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Connector(value)
	return nil
}

// Connection is a WFM connection (Get-CsTeamsShiftsConnection). The service
// returns connector-specific settings without secrets; they are kept as raw
// JSON. Timestamps keep their service text.
type Connection struct {
	ID                        *string                    `json:"id"`
	TenantID                  *string                    `json:"tenantId"`
	ConnectorID               *string                    `json:"connectorId"`
	Name                      *string                    `json:"name"`
	Etag                      *string                    `json:"etag"`
	State                     *string                    `json:"state"`
	CreatedDateTime           *string                    `json:"createdDateTime"`
	LastModifiedDateTime      *string                    `json:"lastModifiedDateTime"`
	ConnectorSpecificSettings map[string]json.RawMessage `json:"connectorSpecificSettings"`
	RawFields                 map[string]json.RawMessage `json:"-"`
}

func (v *Connection) UnmarshalJSON(raw []byte) error {
	type plain Connection
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Connection(value)
	return nil
}

// SyncScenarios selects how each Shifts entity synchronizes with the WFM
// system. Values are opaque service strings (the module does not constrain
// them); nil fields are omitted from requests.
type SyncScenarios struct {
	OfferShiftRequest    *string `json:"offerShiftRequest,omitempty"`
	OpenShift            *string `json:"openShift,omitempty"`
	OpenShiftRequest     *string `json:"openShiftRequest,omitempty"`
	Shift                *string `json:"shift,omitempty"`
	SwapRequest          *string `json:"swapRequest,omitempty"`
	TimeCard             *string `json:"timeCard,omitempty"`
	TimeOff              *string `json:"timeOff,omitempty"`
	TimeOffRequest       *string `json:"timeOffRequest,omitempty"`
	UserShiftPreferences *string `json:"userShiftPreferences,omitempty"`
}

// Instance is a connector instance (Get-CsTeamsShiftsConnectionInstance): a
// scheduled synchronization between one connection and Teams Shifts.
type Instance struct {
	ID                     *string                    `json:"id"`
	TenantID               *string                    `json:"tenantId"`
	ConnectionID           *string                    `json:"connectionId"`
	ConnectorID            *string                    `json:"connectorId"`
	ConnectorAdminEmails   []string                   `json:"connectorAdminEmails"`
	DesignatedActorID      *string                    `json:"designatedActorId"`
	Name                   *string                    `json:"name"`
	SyncFrequencyInMin     *int32                     `json:"syncFrequencyInMin"`
	SyncScenarios          *SyncScenarios             `json:"syncScenarios"`
	WorkforceIntegrationID *string                    `json:"workforceIntegrationId"`
	Etag                   *string                    `json:"etag"`
	State                  *string                    `json:"state"`
	CreatedDateTime        *string                    `json:"createdDateTime"`
	LastModifiedDateTime   *string                    `json:"lastModifiedDateTime"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

func (v *Instance) UnmarshalJSON(raw []byte) error {
	type plain Instance
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Instance(value)
	return nil
}

// Settings are connector-specific settings, for example UKG Dimensions
// apiUrl, ssoUrl, appKey, clientId and clientSecret. They can carry secrets, so
// request types that hold them never print their contents.
type Settings map[string]any

// CreateConnectionRequest is New-CsTeamsShiftsConnection. Every field is
// required by the module.
type CreateConnectionRequest struct {
	Name                      string
	ConnectorID               string
	State                     string
	ConnectorSpecificSettings Settings
}

func (CreateConnectionRequest) String() string     { return "shifts connection request (redacted)" }
func (v CreateConnectionRequest) GoString() string { return v.String() }

// ConnectionUpdate is the body of Set-CsTeamsShiftsConnection (PUT) and
// Update-CsTeamsShiftsConnection (PATCH). Nil fields are omitted. Etag is the
// body _etag; IfMatch is sent as the If-Match header when non-nil.
type ConnectionUpdate struct {
	IfMatch                   *string
	Etag                      *string
	Name                      *string
	ConnectorID               *string
	State                     *string
	ConnectorSpecificSettings Settings
}

func (ConnectionUpdate) String() string     { return "shifts connection update (redacted)" }
func (v ConnectionUpdate) GoString() string { return v.String() }

// InstanceRequest is the body of New-CsTeamsShiftsConnectionInstance. Nil
// fields are omitted; a non-nil empty ConnectorAdminEmails is sent as [].
type InstanceRequest struct {
	ConnectionID         *string
	ConnectorAdminEmails []string
	DesignatedActorID    *string
	Name                 *string
	State                *string
	SyncFrequencyInMin   *int32
	SyncScenarios        *SyncScenarios
}

// InstanceUpdate is the body of Set-CsTeamsShiftsConnectionInstance (PUT)
// and Update-CsTeamsShiftsConnectionInstance (PATCH). IfMatch is required by
// the module for both; Etag is the body _etag.
type InstanceUpdate struct {
	IfMatch string
	Etag    *string
	InstanceRequest
}

// ValidateRequest is Test-CsTeamsShiftsConnectionValidate: it asks the
// service to sign in to the WFM system with the given settings.
type ValidateRequest struct {
	Name                      string
	ConnectorID               string
	ConnectorSpecificSettings Settings
}

func (ValidateRequest) String() string     { return "shifts connection validation request (redacted)" }
func (v ValidateRequest) GoString() string { return v.String() }
