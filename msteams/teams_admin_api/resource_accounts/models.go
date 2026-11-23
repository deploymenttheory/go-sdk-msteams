package resourceaccounts

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Well-known application IDs a resource account can be bound to.
const (
	AutoAttendantApplicationID = "ce933385-9390-45d1-9512-c8d228074e07"
	CallQueueApplicationID     = "11cd3e2e-fccb-42ad-ad00-878b93575e07"
)

// Configuration types accepted by association create. Other values pass
// through unchanged, as the module sends them as plain strings.
const (
	ConfigurationTypeAutoAttendant = "AutoAttendant"
	ConfigurationTypeCallQueue     = "CallQueue"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("resource account input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected resource account response")

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// ApplicationInstance is a resource account as returned by the platform
// service (Get/New/Set-CsOnlineApplicationInstance). Absent and null fields stay
// nil; RawFields keeps the complete service object.
type ApplicationInstance struct {
	ObjectID          *string                    `json:"objectId"`
	TenantID          *string                    `json:"tenantId"`
	UserPrincipalName *string                    `json:"userPrincipalName"`
	ApplicationID     *string                    `json:"applicationId"`
	AcsResourceID     *string                    `json:"acsResourceId"`
	DisplayName       *string                    `json:"displayName"`
	PhoneNumber       *string                    `json:"phoneNumber"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (v *ApplicationInstance) UnmarshalJSON(raw []byte) error {
	type plain ApplicationInstance
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = ApplicationInstance(value)
	return nil
}

// SearchResult is one Find-CsOnlineApplicationInstance match from the voice
// apps service. Its property names differ from ApplicationInstance on the wire.
type SearchResult struct {
	ID                *string                    `json:"Id"`
	TenantID          *string                    `json:"TenantId"`
	DisplayName       *string                    `json:"DisplayName"`
	PhoneNumber       *string                    `json:"PhoneNumber"`
	UserPrincipalName *string                    `json:"UserPrincipalName"`
	SipURI            *string                    `json:"SipUri"`
	ApplicationID     *string                    `json:"ApplicationId"`
	IsOnPrem          *bool                      `json:"IsOnPrem"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (v *SearchResult) UnmarshalJSON(raw []byte) error {
	type plain SearchResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = SearchResult(value)
	return nil
}

// Association binds a resource account (ID) to an auto attendant or call
// queue configuration.
type Association struct {
	ID                           *string                    `json:"Id"`
	ApplicationConfigurationType *string                    `json:"ApplicationConfigurationType"`
	ConfigurationID              *string                    `json:"ConfigurationId"`
	CallPriority                 *int32                     `json:"CallPriority"`
	RawFields                    map[string]json.RawMessage `json:"-"`
}

func (v *Association) UnmarshalJSON(raw []byte) error {
	type plain Association
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Association(value)
	return nil
}

// StatusRecord reports the outcome for one resource account. ErrorCode carries
// the service's provider code, for example
// CreateApplicationInstanceAssociation_AppEndpointNotFound; an empty or nil
// ErrorCode means no error was reported. AuxiliaryData is kept as raw JSON.
type StatusRecord struct {
	ID              *string                    `json:"Id"`
	Status          *int32                     `json:"Status"`
	ErrorCode       *string                    `json:"ErrorCode"`
	Message         *string                    `json:"Message"`
	StatusTimestamp *string                    `json:"StatusTimestamp"`
	AuxiliaryData   json.RawMessage            `json:"AuxiliaryData"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *StatusRecord) UnmarshalJSON(raw []byte) error {
	type plain StatusRecord
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = StatusRecord(value)
	return nil
}

// Failed reports whether the record carries a service error code.
func (v *StatusRecord) Failed() bool {
	return v != nil && v.ErrorCode != nil && *v.ErrorCode != ""
}

// AssociationOperationResult is the per-account outcome of an association
// create or remove. The service answers HTTP 200 even when every account
// failed; inspect Status.Failed for each result.
type AssociationOperationResult struct {
	Association *Association  `json:"ApplicationInstanceAssociation"`
	Status      *StatusRecord `json:"StatusRecord"`
}

// ListOptions pages Get-CsOnlineApplicationInstance. Nil fields are omitted.
type ListOptions struct {
	ResultSize *int32
	Skip       *int32
}

// SearchRequest is Find-CsOnlineApplicationInstance. SearchQuery is required.
// Nil switches are omitted; a non-nil false is sent explicitly, as a bound
// -Switch:$false is by the module.
type SearchRequest struct {
	SearchQuery      string
	MaxResults       *int32
	ExactMatchOnly   *bool
	AssociatedOnly   *bool
	UnAssociatedOnly *bool
}

// CreateRequest is New-CsOnlineApplicationInstance. UserPrincipalName must use
// a domain verified in the tenant. ApplicationID is a GUID such as
// CallQueueApplicationID; nil fields are omitted.
type CreateRequest struct {
	UserPrincipalName string
	ApplicationID     *string
	DisplayName       *string
}

// UpdateRequest is Set-CsOnlineApplicationInstance. Nil fields are omitted.
// The service requires ApplicationID on every update and rejects changing it
// under application (app-only) authentication. OnpremPhoneNumber is rejected by
// the service as deprecated in favour of phone number assignment.
type UpdateRequest struct {
	ApplicationID     *string
	DisplayName       *string
	OnpremPhoneNumber *string
	AcsResourceID     *string
}

// SyncRequest is Sync-CsOnlineApplicationInstance. The service requires at
// least one of the GUIDs; nil fields are omitted.
type SyncRequest struct {
	ApplicationID *string
	AcsResourceID *string
}

// AssociationRequest is New-CsOnlineApplicationInstanceAssociation.
// EndpointIDs are resource account object IDs. CallPriority is optional; the
// module rejects values outside 0-5 and the service accepts 1-5.
type AssociationRequest struct {
	EndpointIDs       []string
	ConfigurationID   string
	ConfigurationType string
	CallPriority      *int16
}
