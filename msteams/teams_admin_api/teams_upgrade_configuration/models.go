// Package teamsupgradeconfiguration manages compiled Teams upgrade configuration.
package teamsupgradeconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration preserves nullable known settings and an owned complete snapshot.
// Source strings remain opaque; Boolean and string settings are nullable.
type Configuration struct {
	Identity                 *string                    `json:"Identity"`
	DownloadTeams            *bool                      `json:"DownloadTeams"`
	SfBMeetingJoinUx         *string                    `json:"SfBMeetingJoinUx"`
	BlockLegacyAuthorization *bool                      `json:"BlockLegacyAuthorization"`
	RawFields                map[string]json.RawMessage `json:"-"`
	RawJSON                  json.RawMessage            `json:"-"`
}

func (p *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Configuration(next)
	return nil
}
func (p Configuration) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(p))
}
func (Configuration) String() string     { return "Teams upgrade configuration (redacted)" }
func (p Configuration) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	DownloadTeams            policy.Field[bool]   `json:"DownloadTeams,omitzero"`
	SfBMeetingJoinUx         policy.Field[string] `json:"SfBMeetingJoinUx,omitzero"`
	BlockLegacyAuthorization policy.Field[bool]   `json:"BlockLegacyAuthorization,omitzero"`
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
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string {
	return "Teams upgrade configuration update result (redacted)"
}
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "Teams upgrade configuration settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
