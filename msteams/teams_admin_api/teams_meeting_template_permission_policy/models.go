// Package teamsmeetingtemplatepermissionpolicy manages Teams meeting-template permission policies.
package teamsmeetingtemplatepermissionpolicy

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// HiddenTemplate follows the pinned Id setter. Missing Id remains distinct
// from an explicit null setter.
type HiddenTemplate struct {
	Identity  *string                    `json:"Identity"`
	ID        *string                    `json:"Id"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

// Policy is a received snapshot. RawJSON retains repeated properties and wire
// order; RawFields exposes the last occurrence of each exact key for inspection.
// Mutation requests are separate and do not serialize this raw snapshot.
type Policy struct {
	Identity                 *string                    `json:"Identity"`
	Description              *string                    `json:"Description"`
	DefaultMeetingTemplateId *string                    `json:"DefaultMeetingTemplateId"`
	HiddenMeetingTemplates   []*HiddenTemplate          `json:"HiddenMeetingTemplates"`
	RawFields                map[string]json.RawMessage `json:"-"`
	RawJSON                  json.RawMessage            `json:"-"`
}

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
func (v *HiddenTemplate) UnmarshalJSON(raw []byte) error {
	var value HiddenTemplate
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			if json.Unmarshal(body, &value.Identity) != nil {
				return client.ErrDecode
			}
		case "id":
			if json.Unmarshal(body, &value.ID) != nil || value.ID == nil {
				return client.ErrDecode
			}
		}
		return nil
	})
	if e != nil {
		return e
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v HiddenTemplate) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain HiddenTemplate
	return json.Marshal(plain(v))
}
func (v *Policy) UnmarshalJSON(raw []byte) error {
	value := Policy{HiddenMeetingTemplates: []*HiddenTemplate{}}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "description":
			return json.Unmarshal(body, &value.Description)
		case "defaultmeetingtemplateid":
			return json.Unmarshal(body, &value.DefaultMeetingTemplateId)
		case "hiddenmeetingtemplates":
			if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				var items []*HiddenTemplate
				if json.Unmarshal(body, &items) != nil {
					return client.ErrDecode
				}
				value.HiddenMeetingTemplates = append(value.HiddenMeetingTemplates, items...)
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
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Policy
	value := plain(v)
	if value.HiddenMeetingTemplates == nil {
		value.HiddenMeetingTemplates = []*HiddenTemplate{}
	}
	return json.Marshal(value)
}
func (Policy) String() string             { return "Teams meeting template permission policy (redacted)" }
func (v Policy) GoString() string         { return v.String() }
func (HiddenTemplate) String() string     { return "hidden meeting template (redacted)" }
func (v HiddenTemplate) GoString() string { return v.String() }
