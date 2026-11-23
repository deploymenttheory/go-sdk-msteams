// Package teamsvdipolicy manages compiled Teams VDI policy.
package teamsvdipolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// Source strings remain opaque; booleans and strings remain nullable.
type Policy struct {
	Identity                            *string                    `json:"Identity"`
	DisableCallsAndMeetings             *bool                      `json:"DisableCallsAndMeetings"`
	DisableAudioVideoInCallsAndMeetings *bool                      `json:"DisableAudioVideoInCallsAndMeetings"`
	VDI2Optimization                    *string                    `json:"VDI2Optimization"`
	RawFields                           map[string]json.RawMessage `json:"-"`
	RawJSON                             json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "VDI policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	DisableCallsAndMeetings             policy.Field[bool]   `json:"DisableCallsAndMeetings,omitzero"`
	DisableAudioVideoInCallsAndMeetings policy.Field[bool]   `json:"DisableAudioVideoInCallsAndMeetings,omitzero"`
	VDI2Optimization                    policy.Field[string] `json:"VDI2Optimization,omitzero"`
}

// CreateRequest permits no settings, matching the compiled New processing.
type CreateRequest struct {
	Settings
	policy.Controls
}

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "VDI policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "VDI policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
