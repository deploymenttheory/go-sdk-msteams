// Package teamsapppolicyconfiguration manages the conditional TeamsAppPolicyConfiguration compiled configuration.
package teamsapppolicyconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ErrInput identifies a setting rejected by pinned model validation before I/O.
var ErrInput = policy.ErrInput

// Configuration retains nullable typed properties and the complete response.
// RawFields/RawJSON distinguish omitted, null and future fields.
type Configuration struct {
	Identity      *string                    `json:"Identity"`
	AppCatalogURI *string                    `json:"AppCatalogUri"`
	ResourceURI   *string                    `json:"ResourceUri"`
	RawFields     map[string]json.RawMessage `json:"-"`
	RawJSON       json.RawMessage            `json:"-"`
}

// UnmarshalJSON retains response presence and enforces known model boundaries.
func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	if next.AppCatalogURI != nil && !validURI(*next.AppCatalogURI) {
		return client.ErrDecode
	}
	if next.ResourceURI != nil && !validURI(*next.ResourceURI) {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(next)
	return nil
}

// MarshalJSON returns an owned response snapshot, preserving unknown fields.
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string {
	return "conditional TeamsAppPolicyConfiguration configuration (redacted)"
}
func (v Configuration) GoString() string { return v.String() }

// Settings distinguishes omission from explicit null and selected scalar values.
// Pinned setter constraints apply to selected values only.
type Settings struct {
	AppCatalogURI policy.Field[string] `json:"AppCatalogUri,omitzero"`
	ResourceURI   policy.Field[string] `json:"ResourceUri,omitzero"`
}

func (Settings) String() string     { return "conditional TeamsAppPolicyConfiguration settings (redacted)" }
func (v Settings) GoString() string { return v.String() }

// PatchRequest sends selected settings without an implicit preparation read.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the compiled read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and acceptance when a later stage fails.
// HTTP acceptance is not evidence that a setting persisted or became effective.
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string {
	return "conditional TeamsAppPolicyConfiguration update result (redacted)"
}
func (v UpdateResult) GoString() string { return v.String() }
