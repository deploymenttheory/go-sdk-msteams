// Package applicationaccesspolicy manages application-access policies and their assignments.
package applicationaccesspolicy

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/listchange"
)

// ErrInput identifies invalid local request construction without retaining input.
var ErrInput = errors.New("invalid application access policy input")

// Policy projects known properties and retains every received JSON field.
// RawFields remains authoritative for unknown and non-string AppIds shapes.
type Policy struct {
	Identity    string
	AppIDs      []string
	Description *string
	RawFields   map[string]json.RawMessage
}

// UnmarshalJSON retains unknown fields rather than closing the service schema.
func (p *Policy) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if fields == nil {
		return errors.New("policy must be an object")
	}
	*p = Policy{RawFields: fields}
	if value, ok := fields["Identity"]; ok {
		if err := json.Unmarshal(value, &p.Identity); err != nil {
			return err
		}
	}
	if value, ok := fields["Description"]; ok {
		if err := json.Unmarshal(value, &p.Description); err != nil {
			return err
		}
	}
	if value, ok := fields["AppIds"]; ok {
		var ids []string
		if json.Unmarshal(value, &ids) == nil {
			p.AppIDs = ids
		}
	}
	return nil
}

// MarshalJSON preserves the received object, including omission and unknown fields.
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawFields != nil {
		return json.Marshal(p.RawFields)
	}
	return json.Marshal(struct {
		Identity    string   `json:"Identity"`
		AppIDs      []string `json:"AppIds,omitempty"`
		Description *string  `json:"Description,omitempty"`
	}{Identity: p.Identity, AppIDs: p.AppIDs, Description: p.Description})
}

// ListChange describes replacement or an Add/Remove delta. Constructors copy input.
// The zero value is invalid; nil omits the setting entirely.
type ListChange struct{ change *listchange.Change }

// ReplaceAppIDs replaces the list; nil/empty emits an explicit [].
func ReplaceAppIDs(ids []string) *ListChange {
	return &ListChange{change: listchange.ReplaceStrings(ids)}
}

// ChangeAppIDs emits Add/Remove arrays, omitting empty operators. Both empty emits {}.
func ChangeAppIDs(add, remove []string) *ListChange {
	return &ListChange{change: listchange.DeltaStrings(add, remove)}
}

// LegacyAppIDs preserves object-item compatibility after removing legacy metadata.
// It does not establish backend acceptance of an object as an application ID.
func LegacyAppIDs(items []json.RawMessage) (*ListChange, error) {
	change, err := listchange.ReplaceObjects(items)
	if err != nil {
		return nil, ErrInput
	}
	return &ListChange{change: change}, nil
}

// ChangeLegacyAppIDs emits object-item Add/Remove deltas without legacy metadata.
func ChangeLegacyAppIDs(add, remove []json.RawMessage) (*ListChange, error) {
	change, err := listchange.DeltaObjects(add, remove)
	if err != nil {
		return nil, ErrInput
	}
	return &ListChange{change: change}, nil
}

// MarshalJSON retains replacement/delta and empty/omitted distinctions.
func (v ListChange) MarshalJSON() ([]byte, error) {
	if v.change == nil {
		return nil, ErrInput
	}
	return v.change.MarshalJSON()
}

// CreateRequest and UpdateRequest are write models, separate from Policy responses.
type CreateRequest struct {
	Identity    string
	AppIDs      *ListChange
	Description *string
}
type UpdateRequest struct {
	AppIDs      *ListChange
	Description *string
}

func writeBody(identity string, ids *ListChange, description *string) (map[string]any, error) {
	body := map[string]any{"Identity": identity}
	if ids != nil {
		encoded, err := ids.MarshalJSON()
		if err != nil {
			return nil, err
		}
		body["AppIds"] = json.RawMessage(slices.Clone(encoded))
	}
	if description != nil {
		body["Description"] = *description
	}
	return body, nil
}

// GroupAssignment distinguishes acceptance from completion; it never polls implicitly.
type GroupAssignment struct {
	OperationID string `json:"OperationId"`
	Accepted    bool   `json:"-"`
}
