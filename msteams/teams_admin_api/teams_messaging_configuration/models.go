// Package teamsmessagingconfiguration manages compiled Teams messaging configuration.
package teamsmessagingconfiguration

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Storyline is the source Int32 enum. Unknown numeric values remain representable.
type Storyline int32

const (
	StorylineDisabled Storyline = iota
	StorylineEnabled
)

func (s *Storyline) UnmarshalJSON(raw []byte) error {
	var number int32
	if json.Unmarshal(raw, &number) == nil {
		*s = Storyline(number)
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return client.ErrDecode
	}
	switch strings.ToLower(text) {
	case "disabled":
		*s = StorylineDisabled
		return nil
	case "enabled":
		*s = StorylineEnabled
		return nil
	}
	value, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return client.ErrDecode
	}
	*s = Storyline(value)
	return nil
}

// Configuration exposes nullable getters and an owned complete response snapshot.
type Configuration struct {
	Identity                          *string                    `json:"Identity"`
	EnableVideoMessageCaptions        *bool                      `json:"EnableVideoMessageCaptions"`
	EnableInOrganizationChatControl   *bool                      `json:"EnableInOrganizationChatControl"`
	CustomEmojis                      *bool                      `json:"CustomEmojis"`
	Storyline                         *Storyline                 `json:"Storyline"`
	Communities                       *string                    `json:"Communities"`
	MessagingNotes                    *string                    `json:"MessagingNotes"`
	FileTypeCheck                     *string                    `json:"FileTypeCheck"`
	UrlReputationCheck                *string                    `json:"UrlReputationCheck"`
	ContentBasedPhishingCheck         *string                    `json:"ContentBasedPhishingCheck"`
	ReportIncorrectSecurityDetections *string                    `json:"ReportIncorrectSecurityDetections"`
	FileTypeCustomizationSetting      *string                    `json:"FileTypeCustomizationSetting"`
	FileTypeBlockList                 []*string                  `json:"FileTypeBlockList"`
	RawFields                         map[string]json.RawMessage `json:"-"`
	RawJSON                           json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Identity                          policyrp.ReadString `json:"Identity"`
		EnableVideoMessageCaptions        policyrp.ReadBool   `json:"EnableVideoMessageCaptions"`
		EnableInOrganizationChatControl   policyrp.ReadBool   `json:"EnableInOrganizationChatControl"`
		CustomEmojis                      policyrp.ReadBool   `json:"CustomEmojis"`
		Storyline                         readStoryline       `json:"Storyline"`
		Communities                       policyrp.ReadString `json:"Communities"`
		MessagingNotes                    policyrp.ReadString `json:"MessagingNotes"`
		FileTypeCheck                     policyrp.ReadString `json:"FileTypeCheck"`
		UrlReputationCheck                policyrp.ReadString `json:"UrlReputationCheck"`
		ContentBasedPhishingCheck         policyrp.ReadString `json:"ContentBasedPhishingCheck"`
		ReportIncorrectSecurityDetections policyrp.ReadString `json:"ReportIncorrectSecurityDetections"`
		FileTypeCustomizationSetting      policyrp.ReadString `json:"FileTypeCustomizationSetting"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	value := Configuration{
		Identity:                          wire.Identity.Value,
		EnableVideoMessageCaptions:        wire.EnableVideoMessageCaptions.Value,
		EnableInOrganizationChatControl:   wire.EnableInOrganizationChatControl.Value,
		CustomEmojis:                      wire.CustomEmojis.Value,
		Storyline:                         wire.Storyline.Value,
		Communities:                       wire.Communities.Value,
		MessagingNotes:                    wire.MessagingNotes.Value,
		FileTypeCheck:                     wire.FileTypeCheck.Value,
		UrlReputationCheck:                wire.UrlReputationCheck.Value,
		ContentBasedPhishingCheck:         wire.ContentBasedPhishingCheck.Value,
		ReportIncorrectSecurityDetections: wire.ReportIncorrectSecurityDetections.Value,
		FileTypeCustomizationSetting:      wire.FileTypeCustomizationSetting.Value,
		FileTypeBlockList:                 []*string{},
	}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		if strings.EqualFold(key, "FileTypeBlockList") && !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			// JsonConvert ignores scalar assignments to this getter-only list. Objects
			// still reject, while arrays populate the existing collection in wire order.
			token := bytes.TrimSpace(body)
			if len(token) > 0 && token[0] != '[' && token[0] != '{' {
				return nil
			}
			var items []policyrp.ReadString
			if json.Unmarshal(body, &items) != nil {
				return client.ErrDecode
			}
			for _, item := range items {
				value.FileTypeBlockList = append(value.FileTypeBlockList, item.Value)
			}
		}
		return nil
	})
	if err != nil {
		return client.ErrDecode
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	return json.Marshal(plain(v))
}
func (Configuration) String() string     { return "messaging configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }
func object(raw []byte, visit func(string, json.RawMessage) error) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return nil, client.ErrDecode
	}
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return nil, client.ErrDecode
		}
		name, ok := key.(string)
		if !ok {
			return nil, client.ErrDecode
		}
		var body json.RawMessage
		if d.Decode(&body) != nil || visit(name, body) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}
