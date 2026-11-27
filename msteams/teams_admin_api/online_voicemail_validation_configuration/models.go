// Package onlinevoicemailvalidationconfiguration models the compiled voicemail
// validation configuration. URI values are settings, never local file/network I/O.
package onlinevoicemailvalidationconfiguration

import (
	"encoding/json"
	"regexp"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// The pinned setter uses this simple ASCII-case HTTPS regex. It permits one
// final LF and supplies no URI parser, length limit or reachability check.
var httpsPattern = regexp.MustCompile(`^[hH][tT][tT][pP][sS]://[^\n]+\n?$`)

type Configuration struct {
	Identity                   *string                    `json:"Identity"`
	AudioFileValidationEnabled *bool                      `json:"AudioFileValidationEnabled"`
	AudioFileValidationURI     *string                    `json:"AudioFileValidationUri"`
	RawFields                  map[string]json.RawMessage `json:"-"`
}

type readURI struct{ Value *string }

func (v *readURI) UnmarshalJSON(raw []byte) error {
	var value policyrp.ReadString
	if json.Unmarshal(raw, &value) != nil {
		return client.ErrDecode
	}
	if value.Value != nil && !httpsPattern.MatchString(*value.Value) {
		return client.ErrDecode
	}
	v.Value = value.Value
	return nil
}
func (v *Configuration) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	// Each occurrence invokes its read profile before a later alias can hide an
	// invalid setter assignment; no normalization of the raw snapshot occurs.
	var wire struct {
		Identity policyrp.ReadString `json:"Identity"`
		Enabled  policyrp.ReadBool   `json:"AudioFileValidationEnabled"`
		URI      readURI             `json:"AudioFileValidationUri"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	*v = Configuration{Identity: wire.Identity.Value, AudioFileValidationEnabled: wire.Enabled.Value, AudioFileValidationURI: wire.URI.Value, RawFields: fields}
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string     { return "voicemail validation configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

type Settings struct {
	AudioFileValidationEnabled policy.Field[bool]   `json:"AudioFileValidationEnabled,omitzero"`
	AudioFileValidationURI     policy.Field[string] `json:"AudioFileValidationUri,omitzero"`
}

func (v Settings) Validate() error {
	if uri, set := v.AudioFileValidationURI.Value(); set && !httpsPattern.MatchString(uri) {
		return ErrInput
	}
	return nil
}

type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}
type UpdateResult struct {
	Before       *Configuration
	After        *Configuration
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "voicemail validation update (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
