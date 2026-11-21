// Package listchange implements the pinned modern PolicyUtils list wire shapes.
// Compiled policy-provider list edits use a separate read/modify/write contract.
package listchange

import (
	"encoding/json"
	"errors"
	"slices"
)

var ErrInput = errors.New("invalid modern list change")

// Change owns replacement values or Add/Remove operators. Resource packages
// wrap this type to retain their public types and local error identities.
type Change struct {
	replacement any
	add         any
	remove      any
	delta       bool
}

func ReplaceStrings(values []string) *Change {
	return &Change{replacement: append([]string{}, values...)}
}

func DeltaStrings(add, remove []string) *Change {
	v := &Change{delta: true}
	if len(add) > 0 {
		v.add = slices.Clone(add)
	}
	if len(remove) > 0 {
		v.remove = slices.Clone(remove)
	}
	return v
}

func ReplaceObjects(values []json.RawMessage) (*Change, error) {
	v, err := items(values, true)
	if err != nil {
		return nil, err
	}
	return &Change{replacement: v}, nil
}

func DeltaObjects(add, remove []json.RawMessage) (*Change, error) {
	return deltaValues(add, remove, true)
}

// ReplaceValues accepts string/object items; legacy metadata is removed only
// from object items. It preserves mixed arrays supported by the pinned helper.
func ReplaceValues(values []json.RawMessage) (*Change, error) {
	v, err := items(values, false)
	if err != nil {
		return nil, err
	}
	return &Change{replacement: v}, nil
}

func DeltaValues(add, remove []json.RawMessage) (*Change, error) {
	return deltaValues(add, remove, false)
}

func deltaValues(add, remove []json.RawMessage, objectsOnly bool) (*Change, error) {
	a, err := items(add, objectsOnly)
	if err != nil {
		return nil, err
	}
	r, err := items(remove, objectsOnly)
	if err != nil {
		return nil, err
	}
	v := &Change{delta: true}
	if len(a) > 0 {
		v.add = a
	}
	if len(r) > 0 {
		v.remove = r
	}
	return v, nil
}

func items(values []json.RawMessage, objectsOnly bool) ([]json.RawMessage, error) {
	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		if !objectsOnly {
			var text *string
			if json.Unmarshal(value, &text) == nil && text != nil {
				encoded, err := json.Marshal(*text)
				if err != nil {
					return nil, ErrInput
				}
				result = append(result, encoded)
				continue
			}
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(value, &fields) != nil || fields == nil {
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

func (v *Change) MarshalJSON() ([]byte, error) {
	if v == nil || !v.delta && v.replacement == nil {
		return nil, ErrInput
	}
	if !v.delta {
		return json.Marshal(v.replacement)
	}
	body := map[string]any{}
	if v.add != nil {
		body["Add"] = v.add
	}
	if v.remove != nil {
		body["Remove"] = v.remove
	}
	return json.Marshal(body)
}
