// Package teamsguestmessagingconfiguration manages the generated guest messaging configuration API.
package teamsguestmessagingconfiguration

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid guest messaging configuration input")
var ErrResponse = errors.New("unexpected guest messaging configuration HTTP status")

// Configuration is one tenant guest messaging configuration. Every public
// setting is nullable so an absent or JSON null read value stays distinct.
type Configuration struct {
	Identity               *string `json:"Identity"`
	AllowUserEditMessage   *bool   `json:"AllowUserEditMessage"`
	AllowUserDeleteMessage *bool   `json:"AllowUserDeleteMessage"`
	// UsersCanDeleteBotMessages is read-only here: the service returns it, but
	// public 8.1.0 Set-CsTeamsGuestMessagingConfiguration has no parameter for it.
	UsersCanDeleteBotMessages *bool                      `json:"UsersCanDeleteBotMessages"`
	AllowUserDeleteChat       *bool                      `json:"AllowUserDeleteChat"`
	AllowUserChat             *bool                      `json:"AllowUserChat"`
	AllowGiphy                *bool                      `json:"AllowGiphy"`
	GiphyRatingType           *string                    `json:"GiphyRatingType"`
	AllowMemes                *bool                      `json:"AllowMemes"`
	AllowImmersiveReader      *bool                      `json:"AllowImmersiveReader"`
	AllowStickers             *bool                      `json:"AllowStickers"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(value)
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	return json.Marshal(plain(v))
}
func (Configuration) String() string     { return "guest messaging configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// UpdateRequest preserves omitted settings (nil) versus explicitly bound
// values, including false and the empty string. Public 8.1.0 binds every
// setting as a non-nullable System.Boolean or System.String, so there is no
// null request setting. GiphyRatingType carries no pinned validation set and is
// sent verbatim; the service decides which rating names it accepts.
type UpdateRequest struct {
	AllowUserEditMessage   *bool
	AllowUserDeleteMessage *bool
	AllowUserDeleteChat    *bool
	AllowUserChat          *bool
	AllowGiphy             *bool
	GiphyRatingType        *string
	AllowMemes             *bool
	AllowImmersiveReader   *bool
	AllowStickers          *bool
}

// body returns the sparse PUT document: Identity plus each bound setting.
func (r UpdateRequest) body(identity string) map[string]any {
	body := map[string]any{"Identity": identity}
	for name, value := range map[string]*bool{
		"AllowUserEditMessage":   r.AllowUserEditMessage,
		"AllowUserDeleteMessage": r.AllowUserDeleteMessage,
		"AllowUserDeleteChat":    r.AllowUserDeleteChat,
		"AllowUserChat":          r.AllowUserChat,
		"AllowGiphy":             r.AllowGiphy,
		"AllowMemes":             r.AllowMemes,
		"AllowImmersiveReader":   r.AllowImmersiveReader,
		"AllowStickers":          r.AllowStickers,
	} {
		if value != nil {
			body[name] = *value
		}
	}
	if r.GiphyRatingType != nil {
		body["GiphyRatingType"] = *r.GiphyRatingType
	}
	return body
}
