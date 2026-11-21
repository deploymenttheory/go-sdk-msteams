// Package recordingmessages manages recording/transcription custom messages.
package recordingmessages

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Message preserves received presence and future fields independently of writes.
type Message struct {
	Identity      *string                    `json:"Identity"`
	ID            *string                    `json:"Id"`
	Description   *string                    `json:"Description"`
	Localizations []*Localization            `json:"RecordingAndTranscriptionLocalizationCustomMessage"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

// Localization is the complete known localization DTO, including nullable text.
type Localization struct {
	Identity                     *string                    `json:"Identity"`
	Language                     *string                    `json:"Language"`
	InitiatorImplicit            *string                    `json:"InitiatorImplicit"`
	ParticipantImplicit          *string                    `json:"ParticipantImplicit"`
	InitiatorExplicit            *string                    `json:"InitiatorExplicit"`
	ParticipantExplicitRequested *string                    `json:"ParticipantExplicitRequested"`
	ParticipantExplicitProvided  *string                    `json:"ParticipantExplicitProvided"`
	AgreementDialogue            *string                    `json:"AgreementDialogue"`
	RawFields                    map[string]json.RawMessage `json:"-"`
}

var messageID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// The source .NET regex accepts one final LF via its dollar anchor.
func validModelID(v string) bool { return messageID.MatchString(strings.TrimSuffix(v, "\n")) }

// inspectObject visits duplicates in wire order so a later value cannot bypass
// a setter failure. RawFields retains received key spelling and future values.
func inspectObject(raw []byte, check func(string, json.RawMessage) error) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, err := d.Token(); err != nil {
		return nil, client.ErrDecode
	}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return nil, client.ErrDecode
		}
		var value json.RawMessage
		if d.Decode(&value) != nil || check(key.(string), value) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}

func (v *Message) UnmarshalJSON(raw []byte) error {
	type plain Message
	var value plain
	if json.Unmarshal(raw, &value) != nil {
		return client.ErrDecode
	}
	fields, err := inspectObject(raw, func(key string, body json.RawMessage) error {
		if strings.EqualFold(key, "Id") {
			var id *string
			if json.Unmarshal(body, &id) != nil || id != nil && !validModelID(*id) {
				return client.ErrDecode
			}
		}
		if strings.EqualFold(key, "RecordingAndTranscriptionLocalizationCustomMessage") && bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			return client.ErrDecode
		}
		return nil
	})
	if err != nil {
		return err
	}
	value.RawFields = fields
	*v = Message(value)
	return nil
}
func (v Message) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type plain Message
	value := plain(v)
	if value.Localizations == nil {
		value.Localizations = []*Localization{}
	}
	return json.Marshal(value)
}
func (v *Localization) UnmarshalJSON(raw []byte) error {
	type plain Localization
	var value plain
	if json.Unmarshal(raw, &value) != nil {
		return client.ErrDecode
	}
	fields, err := inspectObject(raw, func(key string, body json.RawMessage) error {
		if strings.EqualFold(key, "Language") && bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			return client.ErrDecode
		}
		return nil
	})
	if err != nil {
		return err
	}
	value.RawFields = fields
	*v = Localization(value)
	return nil
}
func (v Localization) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type plain Localization
	return json.Marshal(plain(v))
}
func (Message) String() string          { return "recording message (redacted)" }
func (v Message) GoString() string      { return v.String() }
func (Localization) String() string     { return "recording localization (redacted)" }
func (v Localization) GoString() string { return v.String() }
