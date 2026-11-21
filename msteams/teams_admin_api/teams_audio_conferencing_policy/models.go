// Package teamsaudioconferencingpolicy manages audio conferencing policies and assignments.
package teamsaudioconferencingpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/listchange"
)

var ErrInput = errors.New("invalid audio conferencing policy input")
var ErrResponse = errors.New("unexpected audio conferencing policy response")

// Policy retains arbitrary modern configuration fields independently of projections.
// Future shapes in a known field remain in RawFields/RawJSON if not projectable.
type Policy struct {
	Identity                  string                     `json:"Identity"`
	AllowTollFreeDialin       *bool                      `json:"AllowTollFreeDialin"`
	MeetingInvitePhoneNumbers []*string                  `json:"MeetingInvitePhoneNumbers"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var x struct {
		Identity                  string
		AllowTollFreeDialin       *bool
		MeetingInvitePhoneNumbers []*string
	}
	// Modern XdsConfiguration/JSON objects do not use the Core policy setters.
	// Project each known property independently while retaining the full payload.
	if v, ok := fields["Identity"]; ok {
		var value string
		if json.Unmarshal(v, &value) == nil {
			x.Identity = value
		}
	}
	if v, ok := fields["AllowTollFreeDialin"]; ok {
		var value *bool
		if json.Unmarshal(v, &value) == nil {
			x.AllowTollFreeDialin = value
		}
	}
	if v, ok := fields["MeetingInvitePhoneNumbers"]; ok {
		var value []*string
		if json.Unmarshal(v, &value) == nil {
			x.MeetingInvitePhoneNumbers = value
		}
	}
	*p = Policy{Identity: x.Identity, AllowTollFreeDialin: x.AllowTollFreeDialin, MeetingInvitePhoneNumbers: x.MeetingInvitePhoneNumbers, RawFields: fields, RawJSON: append(json.RawMessage(nil), raw...)}
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(p))
}
func (Policy) String() string     { return "audio conferencing policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// ListChange owns modern replacement values or server-side Add/Remove operators.
// A nil request field omits the list; the zero ListChange value is invalid.
type ListChange struct{ change *listchange.Change }

func ReplacePhoneNumbers(values ...string) *ListChange {
	return &ListChange{change: listchange.ReplaceStrings(values)}
}
func ChangePhoneNumbers(add, remove []string) *ListChange {
	return &ListChange{change: listchange.DeltaStrings(add, remove)}
}

// ReplacePhoneValues preserves source-compatible strings/legacy object entries.
// Metadata cleanup is structural compatibility, not backend acceptance evidence.
func ReplacePhoneValues(values []json.RawMessage) (*ListChange, error) {
	v, e := listchange.ReplaceValues(values)
	if e != nil {
		return nil, ErrInput
	}
	return &ListChange{change: v}, nil
}
func ChangePhoneValues(add, remove []json.RawMessage) (*ListChange, error) {
	v, e := listchange.DeltaValues(add, remove)
	if e != nil {
		return nil, ErrInput
	}
	return &ListChange{change: v}, nil
}
func (v ListChange) MarshalJSON() ([]byte, error) {
	if v.change == nil {
		return nil, ErrInput
	}
	raw, e := v.change.MarshalJSON()
	if e != nil {
		return nil, ErrInput
	}
	return raw, nil
}
func (ListChange) String() string     { return "phone list change (redacted)" }
func (v ListChange) GoString() string { return v.String() }

// Settings preserves unbound versus bound false and omitted/replacement/delta lists.
// Boolean arguments are non-null in the exported modern command's parameter type.
type Settings struct {
	AllowTollFreeDialin       *bool
	MeetingInvitePhoneNumbers *ListChange
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(identity string, input Settings) (map[string]any, error) {
	out := map[string]any{"Identity": identity}
	if input.AllowTollFreeDialin != nil {
		out["AllowTollFreeDialin"] = *input.AllowTollFreeDialin
	}
	if input.MeetingInvitePhoneNumbers != nil {
		raw, e := input.MeetingInvitePhoneNumbers.MarshalJSON()
		if e != nil {
			return nil, e
		}
		out["MeetingInvitePhoneNumbers"] = json.RawMessage(raw)
	}
	return out, nil
}

// GroupAssignment retains acceptance and the complete generated JSON response.
type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

func (v GroupAssignment) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain GroupAssignment
	return json.Marshal(plain(v))
}

func (GroupAssignment) String() string     { return "audio policy group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct{ PolicyName string }
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}

// Assignment selects exactly one scope. Group Rank zero is omitted as in the
// PowerShell composition; AssignToGroup can explicitly send native priority zero.
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
