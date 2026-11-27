// Package tenantconfiguration manages the conditional TenantConfiguration compiled configuration.
package tenantconfiguration

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
	Identity          *string                    `json:"Identity"`
	MaxAllowedDomains *uint32                    `json:"MaxAllowedDomains"`
	MaxBlockedDomains *uint32                    `json:"MaxBlockedDomains"`
	RawFields         map[string]json.RawMessage `json:"-"`
	RawJSON           json.RawMessage            `json:"-"`
}

// UnmarshalJSON retains response presence and enforces known model boundaries.
func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	if next.MaxAllowedDomains != nil && *next.MaxAllowedDomains < 1 {
		return client.ErrDecode
	}
	if next.MaxBlockedDomains != nil && *next.MaxBlockedDomains < 1 {
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
	return "conditional TenantConfiguration configuration (redacted)"
}
func (v Configuration) GoString() string { return v.String() }

// Settings distinguishes omission from explicit null and selected scalar values.
// Pinned setter constraints apply to selected values only.
type Settings struct {
	MaxAllowedDomains policy.Field[uint32] `json:"MaxAllowedDomains,omitzero"`
	MaxBlockedDomains policy.Field[uint32] `json:"MaxBlockedDomains,omitzero"`
}

func (Settings) String() string     { return "conditional TenantConfiguration settings (redacted)" }
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
	return "conditional TenantConfiguration update result (redacted)"
}
func (v UpdateResult) GoString() string { return v.String() }
