// Package audioconferencingprompts manages audio conferencing custom prompts.
package audioconferencingprompts

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Configuration preserves raw wire presence independently of typed write DTOs.
type Configuration struct {
	Identity  *string                    `json:"Identity"`
	Prompts   []*Prompt                  `json:"Prompts"`
	Packages  []*PromptPackage           `json:"Packages"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

// Prompt retains all six source comparison fields and optional metadata.
type Prompt struct {
	Identity    *string                    `json:"Identity"`
	ID          *string                    `json:"Id"`
	Name        *string                    `json:"Name"`
	AudioPrompt *string                    `json:"AudioPrompt"`
	TextPrompt  *string                    `json:"TextPrompt"`
	Type        *string                    `json:"Type"`
	Locale      *string                    `json:"Locale"`
	RawFields   map[string]json.RawMessage `json:"-"`
	RawJSON     json.RawMessage            `json:"-"`
}

// PromptPackage contains six readonly source arrays of nullable references.
// These reference strings do not inherit the package Id GUID constraint.
type PromptPackage struct {
	Identity                     *string                    `json:"Identity"`
	ID                           *string                    `json:"Id"`
	Name                         *string                    `json:"Name"`
	DefaultLocale                *string                    `json:"DefaultLocale"`
	InboundStartRecordingPrompt  []*string                  `json:"InboundStartRecordingPrompt"`
	InboundEndRecordingPrompt    []*string                  `json:"InboundEndRecordingPrompt"`
	OutboundStartRecordingPrompt []*string                  `json:"OutboundStartRecordingPrompt"`
	OutboundEndRecordingPrompt   []*string                  `json:"OutboundEndRecordingPrompt"`
	MeetingStartRecordingPrompt  []*string                  `json:"MeetingStartRecordingPrompt"`
	MeetingEndRecordingPrompt    []*string                  `json:"MeetingEndRecordingPrompt"`
	RawFields                    map[string]json.RawMessage `json:"-"`
	RawJSON                      json.RawMessage            `json:"-"`
}

var modelID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// The pinned .NET dollar anchor permits one final LF, but not CRLF.
func validID(v string) bool         { return modelID.MatchString(strings.TrimSuffix(v, "\n")) }
func isNull(v json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(v), []byte("null")) }

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
		if d.Decode(&body) != nil || visit(strings.ToLower(name), body) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}
func text(body json.RawMessage, out **string, required, id bool) error {
	if json.Unmarshal(body, out) != nil || required && *out == nil || id && *out != nil && !validID(**out) {
		return client.ErrDecode
	}
	return nil
}
func appendList[T any](body json.RawMessage, list *[]*T) error {
	if isNull(body) {
		return nil
	}
	var items []*T
	if json.Unmarshal(body, &items) != nil {
		return client.ErrDecode
	}
	*list = append(*list, items...)
	return nil
}
func (v *Configuration) UnmarshalJSON(raw []byte) error {
	x := Configuration{Prompts: []*Prompt{}, Packages: []*PromptPackage{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch key {
		case "identity":
			return text(body, &x.Identity, false, false)
		case "prompts":
			return appendList(body, &x.Prompts)
		case "packages":
			return appendList(body, &x.Packages)
		}
		return nil
	})
	if e != nil {
		return e
	}
	x.RawFields = fields
	x.RawJSON = append(json.RawMessage(nil), raw...)
	*v = x
	return nil
}
func (v *Prompt) UnmarshalJSON(raw []byte) error {
	var x Prompt
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch key {
		case "identity":
			return text(body, &x.Identity, false, false)
		case "id":
			return text(body, &x.ID, true, true)
		case "name":
			return text(body, &x.Name, false, false)
		case "audioprompt":
			return text(body, &x.AudioPrompt, false, false)
		case "textprompt":
			return text(body, &x.TextPrompt, false, false)
		case "type":
			return text(body, &x.Type, true, false)
		case "locale":
			return text(body, &x.Locale, true, false)
		}
		return nil
	})
	if e != nil {
		return e
	}
	x.RawFields = fields
	x.RawJSON = append(json.RawMessage(nil), raw...)
	*v = x
	return nil
}
func (v *PromptPackage) UnmarshalJSON(raw []byte) error {
	x := PromptPackage{}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch key {
		case "identity":
			return text(body, &x.Identity, false, false)
		case "id":
			return text(body, &x.ID, true, true)
		case "name":
			return text(body, &x.Name, false, false)
		case "defaultlocale":
			return text(body, &x.DefaultLocale, true, false)
		case "inboundstartrecordingprompt":
			return appendList(body, &x.InboundStartRecordingPrompt)
		case "inboundendrecordingprompt":
			return appendList(body, &x.InboundEndRecordingPrompt)
		case "outboundstartrecordingprompt":
			return appendList(body, &x.OutboundStartRecordingPrompt)
		case "outboundendrecordingprompt":
			return appendList(body, &x.OutboundEndRecordingPrompt)
		case "meetingstartrecordingprompt":
			return appendList(body, &x.MeetingStartRecordingPrompt)
		case "meetingendrecordingprompt":
			return appendList(body, &x.MeetingEndRecordingPrompt)
		}
		return nil
	})
	if e != nil {
		return e
	}
	x.RawFields = fields
	x.RawJSON = append(json.RawMessage(nil), raw...)
	*v = x
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	x := plain(v)
	if x.Prompts == nil {
		x.Prompts = []*Prompt{}
	}
	if x.Packages == nil {
		x.Packages = []*PromptPackage{}
	}
	return json.Marshal(x)
}
func (v Prompt) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Prompt
	return json.Marshal(plain(v))
}
func (v PromptPackage) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain PromptPackage
	x := plain(v)
	if x.InboundStartRecordingPrompt == nil {
		x.InboundStartRecordingPrompt = []*string{}
	}
	if x.InboundEndRecordingPrompt == nil {
		x.InboundEndRecordingPrompt = []*string{}
	}
	if x.OutboundStartRecordingPrompt == nil {
		x.OutboundStartRecordingPrompt = []*string{}
	}
	if x.OutboundEndRecordingPrompt == nil {
		x.OutboundEndRecordingPrompt = []*string{}
	}
	if x.MeetingStartRecordingPrompt == nil {
		x.MeetingStartRecordingPrompt = []*string{}
	}
	if x.MeetingEndRecordingPrompt == nil {
		x.MeetingEndRecordingPrompt = []*string{}
	}
	return json.Marshal(x)
}
func (Configuration) String() string     { return "audio conferencing prompts configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }
func (Prompt) String() string            { return "audio conferencing prompt (redacted)" }
func (v Prompt) GoString() string        { return v.String() }
func (PromptPackage) String() string     { return "audio conferencing prompt package (redacted)" }
func (v PromptPackage) GoString() string { return v.String() }
