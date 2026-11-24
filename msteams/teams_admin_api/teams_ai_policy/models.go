// Package teamsaipolicy manages compiled Teams AI policy.
package teamsaipolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// The five pinned TeamsAIPolicy settings are opaque nullable strings; the
// model setters run no enum, range or throwing length validation.
type Policy struct {
	Identity                  *string                    `json:"Identity"`
	Description               *string                    `json:"Description"`
	EnrollFace                *string                    `json:"EnrollFace"`
	EnrollVoice               *string                    `json:"EnrollVoice"`
	SpeakerAttributionForBYOD *string                    `json:"SpeakerAttributionForBYOD"`
	PassiveVoiceEnrollment    *string                    `json:"PassiveVoiceEnrollment"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
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
func (Policy) String() string     { return "AI policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty values.
// The pinned New/Set cmdlets bind all five settings as optional System.String
// parameters with no enum; tenant validation stays authoritative.
type Settings struct {
	Description               policy.Field[string] `json:"Description,omitzero"`
	EnrollFace                policy.Field[string] `json:"EnrollFace,omitzero"`
	EnrollVoice               policy.Field[string] `json:"EnrollVoice,omitzero"`
	SpeakerAttributionForBYOD policy.Field[string] `json:"SpeakerAttributionForBYOD,omitzero"`
	PassiveVoiceEnrollment    policy.Field[string] `json:"PassiveVoiceEnrollment,omitzero"`
}

// CreateRequest carries optional settings for one compiled New POST.
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

func (UpdateResult) String() string     { return "AI policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "AI policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
