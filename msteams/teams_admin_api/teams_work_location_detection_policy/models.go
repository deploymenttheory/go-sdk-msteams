// Package teamsworklocationdetectionpolicy manages compiled Teams work location detection policy.
package teamsworklocationdetectionpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// EnableWorkLocationDetection is a nullable Boolean; UserSettingsDefault is an
// opaque nullable string (the pinned model setter runs no enum rule).
type Policy struct {
	Identity                    *string                    `json:"Identity"`
	EnableWorkLocationDetection *bool                      `json:"EnableWorkLocationDetection"`
	UserSettingsDefault         *string                    `json:"UserSettingsDefault"`
	RawFields                   map[string]json.RawMessage `json:"-"`
	RawJSON                     json.RawMessage            `json:"-"`
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
func (Policy) String() string     { return "work location detection policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty values.
// The pinned New/Set cmdlets bind EnableWorkLocationDetection as Nullable<Boolean>
// and UserSettingsDefault as an opaque System.String with no enum or range rule.
type Settings struct {
	EnableWorkLocationDetection policy.Field[bool]   `json:"EnableWorkLocationDetection,omitzero"`
	UserSettingsDefault         policy.Field[string] `json:"UserSettingsDefault,omitzero"`
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

func (UpdateResult) String() string     { return "work location detection policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "work location detection policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
