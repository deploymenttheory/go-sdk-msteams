// Package appinstallation manages Teams apps installed in a team or user's
// personal scope. It implements the routes used by the pinned preview cmdlets;
// their absence from the root module does not establish runtime parameter sets.
package appinstallation

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	// ErrInput identifies an invalid local request.
	ErrInput = errors.New("invalid app installation request")
	// ErrResponse identifies an undeclared successful response status.
	ErrResponse = errors.New("unexpected app installation response")
	// ErrNotFound means no installation matched the catalog app ID.
	ErrNotFound = errors.New("app installation not found")
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
	ID            string                     `json:"id"`
	AppDefinition json.RawMessage            `json:"teamsAppDefinition"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

func (v *Installation) UnmarshalJSON(raw []byte) error {
	type plain Installation
	var value plain
	if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(raw, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = Installation(value)
	return nil
}

// Permissions is the pinned module's grantedResourceSpecificApplicationPermissions
// field. Nil omits the property; a pointer to an empty slice sends an empty list.
// Callers must obtain consent and authorize the app separately.
type Permissions struct {
	Granted *[]string `json:"grantedResourceSpecificApplicationPermissions,omitempty"`
}

// InstallRequest refers to an existing catalog app; it does not publish one.
type InstallRequest struct {
	AppID       string
	Permissions Permissions
}

// ListOptions optionally selects a catalog app ID and bounds Graph pagination.
// Each request expands teamsAppDefinition, as the pinned cmdlets do.
type ListOptions struct {
	AppID  string
	Paging client.PagingOptions
}
