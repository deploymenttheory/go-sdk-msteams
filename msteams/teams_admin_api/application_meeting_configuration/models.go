// Package applicationmeetingconfiguration manages the tenant application-meeting
// allow-list through the native ConfigAPI configuration resource.
package applicationmeetingconfiguration

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/listchange"

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

// AppIDChange describes replacement or an Add/Remove delta. Constructors copy input.
// The zero value is invalid; nil omits the setting entirely.
type AppIDChange struct{ change *listchange.Change }

// ReplaceAppIDs replaces the list; nil/empty emits an explicit [].
func ReplaceAppIDs(ids []string) *AppIDChange {
	return &AppIDChange{change: listchange.ReplaceStrings(ids)}
}

// ChangeAppIDs emits Add/Remove arrays, omitting empty operators. Both empty emits {}.
func ChangeAppIDs(add, remove []string) *AppIDChange {
	return &AppIDChange{change: listchange.DeltaStrings(add, remove)}
}

// LegacyAppIDs preserves object-item compatibility after removing legacy metadata.
// It does not establish backend acceptance of an object as an application ID.
func LegacyAppIDs(items []json.RawMessage) (*AppIDChange, error) {
	change, err := listchange.ReplaceObjects(items)
	if err != nil {
		return nil, ErrInput
	}
	return &AppIDChange{change: change}, nil
}

// ChangeLegacyAppIDs emits object-item Add/Remove deltas without legacy metadata.
func ChangeLegacyAppIDs(add, remove []json.RawMessage) (*AppIDChange, error) {
	change, err := listchange.DeltaObjects(add, remove)
	if err != nil {
		return nil, ErrInput
	}
	return &AppIDChange{change: change}, nil
}

// MarshalJSON retains replacement/delta and empty/omitted distinctions.
func (v AppIDChange) MarshalJSON() ([]byte, error) {
	if v.change == nil {
		return nil, ErrInput
	}
	return v.change.MarshalJSON()
}

// UpdateRequest contains only bound settings. Nil omits the property rather
// than sending null; replacement and delta behavior are chosen explicitly.
type UpdateRequest struct {
	AppIDs *AppIDChange
}
