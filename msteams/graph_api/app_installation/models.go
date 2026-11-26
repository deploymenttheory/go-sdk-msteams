// Package appinstallation manages Teams apps installed in a team or user's
// personal scope. It implements the routes used by the pinned preview cmdlets;
// their absence from the root module does not establish runtime parameter sets.
package appinstallation

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	// ErrInput identifies an invalid local request.
	ErrInput = errors.New("invalid app installation request")
	// ErrResponse identifies an undeclared successful response status.
	ErrResponse = errors.New("unexpected app installation response")
	// ErrNotFound means no installation matched the catalog app ID.
	ErrNotFound = errors.New("app installation not found")
	// ErrIdentityMismatch means an expanded app definition contradicts the requested catalog ID.
	ErrIdentityMismatch = errors.New("app installation catalog identity mismatch")
	// ErrAmbiguous means more than one installation matched the catalog app ID.
	ErrAmbiguous = errors.New("ambiguous app installation")
)

// Target selects exactly one team or user's personal scope. Chat scope is not
// supported by the pinned cmdlets and is not accepted here.
type Target struct {
	TeamID string
	UserID string
}

// Installation.ID is an installation ID, distinct from the catalog app ID.
// RawFields retains expanded definitions, explicit nulls and future fields.
type Installation struct {
	ID                     string                     `json:"id"`
	AppDefinition          *AppDefinition             `json:"teamsAppDefinition"`
	ConsentedPermissionSet *PermissionSet             `json:"consentedPermissionSet"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON rejects missing installation IDs and retains the full response.
func (v *Installation) UnmarshalJSON(raw []byte) error {
	type plain Installation
	var value plain
	if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(raw, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = Installation(value)
	return nil
}

// AppDefinition exposes the known expanded app version fields. ID addresses
// the definition; TeamsAppID addresses the associated catalog app. AzureADAppID
// is a third identity. RawFields preserves nulls and unknown future members.
type AppDefinition struct {
	ID               *string                    `json:"id"`
	TeamsAppID       *string                    `json:"teamsAppId"`
	AzureADAppID     *string                    `json:"azureADAppId"`
	DisplayName      *string                    `json:"displayName"`
	Version          *string                    `json:"version"`
	PublishingState  *string                    `json:"publishingState"`
	Description      *string                    `json:"description"`
	ShortDescription *string                    `json:"shortDescription"`
	Authorization    *AppAuthorization          `json:"authorization"`
	RawFields        map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON retains typed expanded properties and their original presence.
func (v *AppDefinition) UnmarshalJSON(raw []byte) error {
	type plain AppDefinition
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AppDefinition(value)
	return nil
}

// AppAuthorization declares the permissions required by an app definition.
type AppAuthorization struct {
	RequiredPermissionSet *PermissionSet `json:"requiredPermissionSet"`
}

// PermissionType names the published Graph resource-specific permission kinds.
type PermissionType string

const (
	PermissionApplication PermissionType = "application"
	PermissionDelegated   PermissionType = "delegated"
)

// ResourcePermission is the native Graph v1.0 RSC permission representation.
// PermissionType is "application" or "delegated" for requests.
type ResourcePermission struct {
	PermissionValue string         `json:"permissionValue"`
	PermissionType  PermissionType `json:"permissionType"`
}

// PermissionSet models the native Graph v1.0 consentedPermissionSet value.
// A nil request pointer omits consent; a nonnil empty slice sends an empty list.
type PermissionSet struct {
	ResourceSpecificPermissions []ResourcePermission `json:"resourceSpecificPermissions"`
}

// CreateRequest refers to an existing catalog app; it does not publish one.
// Consent uses the published v1.0 schema, not the pinned preview legacy field.
type CreateRequest struct {
	CatalogID string
	Consent   *PermissionSet
}

// CatalogMutationResult retains the resolved installation receipt even when
// the action fails or its response is lost. ActionInvoked means the resource
// method was called, not proof that the service received the HTTP request.
type CatalogMutationResult struct {
	CatalogID      string
	InstallationID string
	LookupResponse *resty.Response
	ActionResponse *resty.Response
	LookupComplete bool
	ActionInvoked  bool
	HTTPAccepted   bool
}

func (CatalogMutationResult) String() string     { return "app installation mutation result (redacted)" }
func (v CatalogMutationResult) GoString() string { return v.String() }

// ListOptions optionally selects a catalog app ID and bounds Graph pagination.
// Each request expands teamsAppDefinition, as the pinned cmdlets do.
type ListOptions struct {
	CatalogID string
	Paging    client.PagingOptions
}
