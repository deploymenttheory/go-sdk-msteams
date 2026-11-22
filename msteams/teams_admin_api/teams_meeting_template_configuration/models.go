// Package teamsmeetingtemplateconfiguration reads the pinned meeting-template configuration.
package teamsmeetingtemplateconfiguration

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Configuration retains nullable typed fields and the complete owned response snapshot.
type Configuration struct {
	Identity              *string                    `json:"Identity"`
	Description           *string                    `json:"Description"`
	TeamsMeetingTemplates []*Template                `json:"TeamsMeetingTemplates"`
	RawFields             map[string]json.RawMessage `json:"-"`
	RawJSON               json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	value := Configuration{TeamsMeetingTemplates: []*Template{}}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "description":
			return json.Unmarshal(body, &value.Description)
		case "teamsmeetingtemplates":
			if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				var items []*Template
				if json.Unmarshal(body, &items) != nil {
					return client.ErrDecode
				}
				value.TeamsMeetingTemplates = append(value.TeamsMeetingTemplates, items...)
			}
			return nil
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
func (Configuration) String() string     { return "meeting template configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// Template retains nullable typed fields and the complete owned response snapshot.
type Template struct {
	Identity            *string                    `json:"Identity"`
	Description         *string                    `json:"Description"`
	Name                *string                    `json:"Name"`
	Category            *string                    `json:"Category"`
	TeamsMeetingOptions []*Option                  `json:"TeamsMeetingOptions"`
	RawFields           map[string]json.RawMessage `json:"-"`
	RawJSON             json.RawMessage            `json:"-"`
}

func (v *Template) UnmarshalJSON(raw []byte) error {
	value := Template{TeamsMeetingOptions: []*Option{}}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "description":
			return json.Unmarshal(body, &value.Description)
		case "name":
			if json.Unmarshal(body, &value.Name) != nil || value.Name == nil {
				return client.ErrDecode
			}
			return nil
		case "category":
			return json.Unmarshal(body, &value.Category)
		case "teamsmeetingoptions":
			if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				var items []*Option
				if json.Unmarshal(body, &items) != nil {
					return client.ErrDecode
				}
				value.TeamsMeetingOptions = append(value.TeamsMeetingOptions, items...)
			}
			return nil
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
func (v Template) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Template
	return json.Marshal(plain(v))
}
func (Template) String() string     { return "meeting template (redacted)" }
func (v Template) GoString() string { return v.String() }

// Option retains nullable typed fields and the complete owned response snapshot.
type Option struct {
	Identity  *string                    `json:"Identity"`
	IsLocked  *bool                      `json:"IsLocked"`
	IsHidden  *bool                      `json:"IsHidden"`
	Value     *string                    `json:"Value"`
	Name      *string                    `json:"Name"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

func (v *Option) UnmarshalJSON(raw []byte) error {
	value := Option{}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "islocked":
			return json.Unmarshal(body, &value.IsLocked)
		case "ishidden":
			return json.Unmarshal(body, &value.IsHidden)
		case "value":
			return json.Unmarshal(body, &value.Value)
		case "name":
			if json.Unmarshal(body, &value.Name) != nil || value.Name == nil {
				return client.ErrDecode
			}
			return nil
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
func (v Option) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Option
	return json.Marshal(plain(v))
}
func (Option) String() string     { return "meeting template option (redacted)" }
func (v Option) GoString() string { return v.String() }
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
