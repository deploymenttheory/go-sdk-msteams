// Package privacyconfiguration manages native privacy configuration reads/updates.
package privacyconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration preserves received Boolean presence, identity and future fields.
type Configuration struct {
	Identity                     *string                    `json:"Identity"`
	EnablePrivacyMode            *bool                      `json:"EnablePrivacyMode"`
	AutoInitiateContacts         *bool                      `json:"AutoInitiateContacts"`
	PublishLocationDataDefault   *bool                      `json:"PublishLocationDataDefault"`
	DisplayPublishedPhotoDefault *bool                      `json:"DisplayPublishedPhotoDefault"`
	RawFields                    map[string]json.RawMessage `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var wire struct {
		Identity                     policyrp.ReadString `json:"Identity"`
		EnablePrivacyMode            policyrp.ReadBool   `json:"EnablePrivacyMode"`
		AutoInitiateContacts         policyrp.ReadBool   `json:"AutoInitiateContacts"`
		PublishLocationDataDefault   policyrp.ReadBool   `json:"PublishLocationDataDefault"`
		DisplayPublishedPhotoDefault policyrp.ReadBool   `json:"DisplayPublishedPhotoDefault"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	next := Configuration{
		Identity:                     wire.Identity.Value,
		EnablePrivacyMode:            wire.EnablePrivacyMode.Value,
		AutoInitiateContacts:         wire.AutoInitiateContacts.Value,
		PublishLocationDataDefault:   wire.PublishLocationDataDefault.Value,
		DisplayPublishedPhotoDefault: wire.DisplayPublishedPhotoDefault.Value,
		RawFields:                    fields,
	}
	*v = next
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type plain Configuration
	return json.Marshal(plain(v))
}

// Settings distinguishes omitted values, explicit null, true and false.
type Settings struct {
	EnablePrivacyMode            policy.Field[bool] `json:"EnablePrivacyMode,omitzero"`
	AutoInitiateContacts         policy.Field[bool] `json:"AutoInitiateContacts,omitzero"`
	PublishLocationDataDefault   policy.Field[bool] `json:"PublishLocationDataDefault,omitzero"`
	DisplayPublishedPhotoDefault policy.Field[bool] `json:"DisplayPublishedPhotoDefault,omitzero"`
}

// PatchRequest sends only selected settings. BodyIdentity overrides the body
// identity independently of the route, when preparing a source-compatible call.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit read/modify/PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// HTTPAccepted does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before       *Configuration
	After        *Configuration
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}

func (Configuration) String() string     { return "privacy configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }
func (UpdateResult) String() string      { return "privacy configuration update result (redacted)" }
func (v UpdateResult) GoString() string  { return v.String() }
