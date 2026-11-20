// Package applicationaccesspolicy manages application-access policies and their assignments.
package applicationaccesspolicy

import (
	"encoding/json"
	"errors"
	"slices"
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
// Its zero value is invalid; an omitted *ListChange leaves AppIds absent.
type ListChange struct {
	replacement any
	add         any
	remove      any
	delta       bool
	valid       bool
}

// ReplaceAppIDs replaces the list; nil and empty both emit an explicit [].
func ReplaceAppIDs(ids []string) *ListChange {
	return &ListChange{replacement: append([]string{}, ids...), valid: true}
}

// ChangeAppIDs emits Add/Remove operators, omitting an empty operator. Both empty emits {}.
func ChangeAppIDs(add, remove []string) *ListChange {
	change := &ListChange{delta: true, valid: true}
	if len(add) > 0 {
		change.add = slices.Clone(add)
	}
	if len(remove) > 0 {
		change.remove = slices.Clone(remove)
	}
	return change
}

// LegacyAppIDs preserves the pinned helper's object-item compatibility branch.
// Each item must be a JSON object. Serialization metadata is removed, not sent.
// Local acceptance does not establish backend acceptance of an object as an app ID.
func LegacyAppIDs(items []json.RawMessage) (*ListChange, error) {
	result, err := legacyItems(items)
	if err != nil {
		return nil, err
	}
	return &ListChange{replacement: result, valid: true}, nil
}

// ChangeLegacyAppIDs emits object-item Add/Remove deltas without legacy metadata.
func ChangeLegacyAppIDs(add, remove []json.RawMessage) (*ListChange, error) {
	addItems, err := legacyItems(add)
	if err != nil {
		return nil, err
	}
	removeItems, err := legacyItems(remove)
	if err != nil {
		return nil, err
	}
	change := &ListChange{delta: true, valid: true}
	if len(addItems) > 0 {
		change.add = addItems
	}
	if len(removeItems) > 0 {
		change.remove = removeItems
	}
	return change, nil
}

func legacyItems(items []json.RawMessage) ([]json.RawMessage, error) {
	result := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var fields map[string]json.RawMessage
		if json.Unmarshal(item, &fields) != nil || fields == nil {
			return nil, ErrInput
		}
		for _, key := range []string{"XsAnyElements", "XsAnyAttributes", "DataSource", "Key", "ScopeClass", "Anchor", "Element", "TypedIdentity"} {
			delete(fields, key)
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			return nil, ErrInput
		}
		result = append(result, encoded)
	}
	return result, nil
}

// MarshalJSON retains replacement/delta and empty/omitted distinctions.
func (c ListChange) MarshalJSON() ([]byte, error) {
	if !c.valid {
		return nil, ErrInput
	}
	if !c.delta {
		return json.Marshal(c.replacement)
	}
	body := make(map[string]any)
	if c.add != nil {
		body["Add"] = c.add
	}
	if c.remove != nil {
		body["Remove"] = c.remove
	}
	return json.Marshal(body)
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
