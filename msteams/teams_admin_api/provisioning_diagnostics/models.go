package provisioningdiagnostics

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("provisioning diagnostics input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected provisioning diagnostics response")

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// DirectoryDiagnosticQuery selects Business Voice Directory diagnostic rows
// (Get-CsBusinessVoiceDirectoryDiagnosticData). Table, Region and PartitionKey
// are required path segments; nil optional fields are omitted from the query.
type DirectoryDiagnosticQuery struct {
	Table        string
	Region       string
	PartitionKey string
	RowKey       *string
	ResultSize   *int32
}

// DirectoryDiagnosticEntity is one Business Voice Directory table entity.
type DirectoryDiagnosticEntity struct {
	Properties []*EntityProperty          `json:"Properties"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

func (v *DirectoryDiagnosticEntity) UnmarshalJSON(raw []byte) error {
	type plain DirectoryDiagnosticEntity
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = DirectoryDiagnosticEntity(value)
	return nil
}

// EntityProperty is one named, typed value of a directory entity. The module
// flattens these into its output.
type EntityProperty struct {
	Name      *string                    `json:"Name"`
	Type      *string                    `json:"Type"`
	Value     *string                    `json:"Value"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *EntityProperty) UnmarshalJSON(raw []byte) error {
	type plain EntityProperty
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = EntityProperty(value)
	return nil
}

// ObjectChangelogQuery filters MAS object change history
// (Get-CsMasObjectChangelog). Empty fields are omitted.
type ObjectChangelogQuery struct {
	Identity   string
	SchemaName string
	Version    string
}

// ChangelogItem is one MAS object change. VersionCreatedTime keeps the
// service's text form; ModifiedValues keeps each change object as sent.
type ChangelogItem struct {
	SchemaName         *string                    `json:"SchemaName"`
	Version            *int64                     `json:"Version"`
	VersionCreatedTime *string                    `json:"VersionCreatedTime"`
	UpdatedBy          *string                    `json:"UpdatedBy"`
	OriginOfChange     *string                    `json:"OriginOfChange"`
	ModifiedValues     []json.RawMessage          `json:"ModifiedValues"`
	Orders             []string                   `json:"Orders"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *ChangelogItem) UnmarshalJSON(raw []byte) error {
	type plain ChangelogItem
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = ChangelogItem(value)
	return nil
}

// VersionedSchemaQuery selects versioned MAS schema data
// (Get-CsMasVersionedSchemaData). SchemaName is required; empty optional
// fields are omitted.
type VersionedSchemaQuery struct {
	SchemaName string
	Identity   string
	Version    string
}

// CustomHandlerOperation is the durable-function management operation sent by
// Invoke-CsCustomHandlerCallBackNgtprov.
type CustomHandlerOperation string

// Operations the module offers for custom handler callbacks.
const (
	StatusQueryGetURI     CustomHandlerOperation = "StatusQueryGetUri"
	TerminatePostURI      CustomHandlerOperation = "TerminatePostUri"
	PurgeHistoryDeleteURI CustomHandlerOperation = "PurgeHistoryDeleteUri"
	RestartPostURI        CustomHandlerOperation = "RestartPostUri"
	SendEventPostURI      CustomHandlerOperation = "SendEventPostUri"
)

// CustomHandlerCallbackRequest addresses a running custom handler instance.
// An empty EventName is omitted.
type CustomHandlerCallbackRequest struct {
	ID        string
	Operation CustomHandlerOperation
	EventName string
}

// CustomHandlerPayload starts a provisioning custom handler
// (Invoke-CsCustomHandlerNgtprov). Nil fields are omitted.
type CustomHandlerPayload struct {
	HandlerFullyQualifiedName *string `json:"HandlerFullyQualifiedName,omitempty"`
	Payload                   *string `json:"Payload,omitempty"`
}

// CustomHandlerResult identifies the started handler instance.
type CustomHandlerResult struct {
	ID        *string                    `json:"id"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *CustomHandlerResult) UnmarshalJSON(raw []byte) error {
	type plain CustomHandlerResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = CustomHandlerResult(value)
	return nil
}

// DirectObjectSyncRequest is the Invoke-CsDirectObjectSync body. Nil fields are
// omitted, as the module omits unbound parameters.
type DirectObjectSyncRequest struct {
	ServiceInstance                 *string  `json:"ServiceInstance,omitempty"`
	ObjectID                        *string  `json:"ObjectId,omitempty"`
	ScenariosToSuppress             *string  `json:"ScenariosToSuppress,omitempty"`
	SynchronizeTenantWithAllObjects *bool    `json:"SynchronizeTenantWithAllObjects,omitempty"`
	ReSyncOptions                   *int64   `json:"ReSyncOptions,omitempty"`
	DeploymentName                  *string  `json:"DeploymentName,omitempty"`
	ObjectClass                     *string  `json:"ObjectClass,omitempty"`
	IsValidationRequest             *bool    `json:"IsValidationRequest,omitempty"`
	ObjectIDs                       []string `json:"ObjectIds,omitempty"`
}

// MsodsSyncRequest is the Invoke-CsMsodsSync body. TenantID and ObjectClass are
// required; nil fields are omitted.
type MsodsSyncRequest struct {
	ServiceInstance                 *string  `json:"ServiceInstance,omitempty"`
	TenantID                        string   `json:"TenantId"`
	ScenariosToSuppress             *string  `json:"ScenariosToSuppress,omitempty"`
	SynchronizeTenantWithAllObjects *bool    `json:"SynchronizeTenantWithAllObjects,omitempty"`
	ReSyncOptions                   *int64   `json:"ReSyncOptions,omitempty"`
	ObjectClass                     string   `json:"ObjectClass"`
	IsValidationRequest             *bool    `json:"IsValidationRequest,omitempty"`
	ObjectIDs                       []string `json:"ObjectIds,omitempty"`
}

// OperationResult is the Set-CsTenantUserBackfill response.
type OperationResult struct {
	Message   *string                    `json:"message"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *OperationResult) UnmarshalJSON(raw []byte) error {
	type plain OperationResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OperationResult(value)
	return nil
}
