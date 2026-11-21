// Package onlinevoicemailvalidationconfiguration models the compiled voicemail
// validation configuration. URI values are settings, never local file/network I/O.
package onlinevoicemailvalidationconfiguration

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
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

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var value snapshot
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	// Validate each setter occurrence, including case variants/duplicates, before
	// a later value can hide an invalid earlier assignment in RawFields.
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, err := d.Token(); err != nil {
		return client.ErrDecode
	}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return client.ErrDecode
		}
		var encoded json.RawMessage
		if d.Decode(&encoded) != nil {
			return client.ErrDecode
		}
		if name, ok := key.(string); ok && strings.EqualFold(name, "AudioFileValidationUri") && !bytes.Equal(bytes.TrimSpace(encoded), []byte("null")) {
			var uri string
			if json.Unmarshal(encoded, &uri) != nil || !httpsPattern.MatchString(uri) {
				return client.ErrDecode
			}
		}
	}
	*v = Configuration(value)
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
