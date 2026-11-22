// Package teamspersonalattendantpolicy manages compiled Teams personal attendant policy.
package teamspersonalattendantpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// All source setting strings remain opaque and nullable.
type Policy struct {
	Identity               *string                    `json:"Identity"`
	PersonalAttendant      *string                    `json:"PersonalAttendant"`
	CallScreening          *string                    `json:"CallScreening"`
	CalendarBookings       *string                    `json:"CalendarBookings"`
	InboundInternalCalls   *string                    `json:"InboundInternalCalls"`
	InboundFederatedCalls  *string                    `json:"InboundFederatedCalls"`
	InboundPSTNCalls       *string                    `json:"InboundPSTNCalls"`
	AutomaticTranscription *string                    `json:"AutomaticTranscription"`
	AutomaticRecording     *string                    `json:"AutomaticRecording"`
	RawFields              map[string]json.RawMessage `json:"-"`
	RawJSON                json.RawMessage            `json:"-"`
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
func (Policy) String() string     { return "personal attendant policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	PersonalAttendant      policy.Field[string] `json:"PersonalAttendant,omitzero"`
	CallScreening          policy.Field[string] `json:"CallScreening,omitzero"`
	CalendarBookings       policy.Field[string] `json:"CalendarBookings,omitzero"`
	InboundInternalCalls   policy.Field[string] `json:"InboundInternalCalls,omitzero"`
	InboundFederatedCalls  policy.Field[string] `json:"InboundFederatedCalls,omitzero"`
	InboundPSTNCalls       policy.Field[string] `json:"InboundPSTNCalls,omitzero"`
	AutomaticTranscription policy.Field[string] `json:"AutomaticTranscription,omitzero"`
	AutomaticRecording     policy.Field[string] `json:"AutomaticRecording,omitzero"`
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

func (UpdateResult) String() string     { return "personal attendant policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "personal attendant policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
