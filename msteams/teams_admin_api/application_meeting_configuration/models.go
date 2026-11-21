// Package applicationmeetingconfiguration manages the tenant application-meeting
// allow-list through the native ConfigAPI configuration resource.
package applicationmeetingconfiguration

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	ErrInput    = errors.New("invalid application meeting configuration input")
	ErrResponse = errors.New("application meeting configuration returned an unexpected successful HTTP status")
)

// Configuration retains the complete received object, including omitted/null
// settings and future fields. AppIDs projects string-array responses; raw fields
// remain authoritative for the helper's legacy object-item compatibility shapes.
type Configuration struct {
	Identity  string
	AppIDs    []string
	RawFields map[string]json.RawMessage
}

func (v *Configuration) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	var identity string
	if json.Unmarshal(data, &fields) != nil || fields == nil || json.Unmarshal(fields["Identity"], &identity) != nil || strings.TrimSpace(identity) == "" {
		return client.ErrDecode
	}
	var ids []string
	if raw, ok := fields["AllowRemoveParticipantAppIds"]; ok {
		if json.Unmarshal(raw, &ids) != nil {
			ids = nil
		}
	}
	*v = Configuration{Identity: identity, AppIDs: ids, RawFields: fields}
	return nil
}

// MarshalJSON preserves the read snapshot; updates require a separate request.
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	return json.Marshal(struct {
		Identity string   `json:"Identity"`
		AppIDs   []string `json:"AllowRemoveParticipantAppIds"`
	}{v.Identity, v.AppIDs})
}
func (Configuration) String() string     { return "application meeting configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// AppIDChange is a replacement or Add/Remove delta. Constructors copy inputs.
// The zero value is invalid; a nil *AppIDChange omits the setting entirely.
type AppIDChange struct {
	replacement any
	add         any
	remove      any
	delta       bool
	valid       bool
}

// ReplaceAppIDs emits an array. Nil and empty inputs both explicitly clear it.
func ReplaceAppIDs(ids []string) *AppIDChange {
	return &AppIDChange{replacement: append([]string{}, ids...), valid: true}
}

// ChangeAppIDs emits capitalized operators; empty operators are omitted.
// Both empty produces {}, distinct from clearing the list with [].
func ChangeAppIDs(add, remove []string) *AppIDChange {
	v := &AppIDChange{delta: true, valid: true}
	if len(add) > 0 {
		v.add = slices.Clone(add)
	}
	if len(remove) > 0 {
		v.remove = slices.Clone(remove)
	}
	return v
}

// LegacyAppIDs models the pinned helper's PSObject compatibility branch. Each
// item must be an object; legacy serialization metadata is removed. This does
// not establish backend acceptance of an object as an application identifier.
func LegacyAppIDs(items []json.RawMessage) (*AppIDChange, error) {
	values, err := legacyItems(items)
	if err != nil {
		return nil, err
	}
	return &AppIDChange{replacement: values, valid: true}, nil
}

// ChangeLegacyAppIDs retains object-item Add/Remove shapes without legacy metadata.
func ChangeLegacyAppIDs(add, remove []json.RawMessage) (*AppIDChange, error) {
	a, err := legacyItems(add)
	if err != nil {
		return nil, err
	}
	r, err := legacyItems(remove)
	if err != nil {
		return nil, err
	}
	v := &AppIDChange{delta: true, valid: true}
	if len(a) > 0 {
		v.add = a
	}
	if len(r) > 0 {
		v.remove = r
	}
	return v, nil
}

func legacyItems(items []json.RawMessage) ([]json.RawMessage, error) {
	values := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var fields map[string]json.RawMessage
		if json.Unmarshal(item, &fields) != nil || fields == nil {
			return nil, ErrInput
		}
		for _, name := range []string{"XsAnyElements", "XsAnyAttributes", "DataSource", "Key", "ScopeClass", "Anchor", "Element", "TypedIdentity"} {
			delete(fields, name)
		}
		value, err := json.Marshal(fields)
		if err != nil {
			return nil, ErrInput
		}
		values = append(values, value)
	}
	return values, nil
}

func (v AppIDChange) MarshalJSON() ([]byte, error) {
	if !v.valid {
		return nil, ErrInput
	}
	if !v.delta {
		return json.Marshal(v.replacement)
	}
	fields := map[string]any{}
	if v.add != nil {
		fields["Add"] = v.add
	}
	if v.remove != nil {
		fields["Remove"] = v.remove
	}
	return json.Marshal(fields)
}

// UpdateRequest contains only bound settings. Nil omits the property rather
// than sending null; replacement and delta behavior are chosen explicitly.
type UpdateRequest struct {
	AppIDs *AppIDChange
}
