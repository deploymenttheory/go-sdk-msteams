package policy

import (
	"encoding/json"
	"slices"
)

// StringList is an owned full-array request value, including nullable entries.
// Nil/empty input serializes []; omitting the field is distinct from clearing it.
type StringList struct{ values []*string }

func Strings(values ...string) *StringList { return NullableStrings(stringPointers(values)...) }

// NullableStrings preserves null entries and copies each pointed-to value.
func NullableStrings(values ...*string) *StringList {
	return &StringList{values: cloneStringPointers(values)}
}
func (v StringList) MarshalJSON() ([]byte, error) {
	if v.values == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(v.values)
}
func (StringList) String() string     { return "policy string list (redacted)" }
func (v StringList) GoString() string { return v.String() }

// ListEdit applies the compiled cmdlet's local PSListModifier behavior. Its
// result is a full replacement array; it never emits server Add/Remove operators.
// A nonnil empty edit clears the list, as the pinned cmdlet's IsEmpty branch does.
// Nil *ListEdit means the property was not bound.
type ListEdit struct {
	replace []*string
	add     []*string
	remove  []*string
	delta   bool
	valid   bool
}

func ReplaceStrings(values ...string) *ListEdit {
	return ReplaceNullableStrings(stringPointers(values)...)
}
func ChangeStrings(add, remove []string) *ListEdit {
	return ChangeNullableStrings(stringPointers(add), stringPointers(remove))
}

// ReplaceNullableStrings replaces the list, preserving null entries and ownership.
func ReplaceNullableStrings(values ...*string) *ListEdit {
	return &ListEdit{replace: cloneStringPointers(values), valid: true}
}

// ChangeNullableStrings removes first ordinal/null matches, then appends additions.
// Inputs and their pointed-to strings are copied.
func ChangeNullableStrings(add, remove []*string) *ListEdit {
	return &ListEdit{add: cloneStringPointers(add), remove: cloneStringPointers(remove), delta: true, valid: true}
}
func (v *ListEdit) Validate() error {
	if v == nil || !v.valid {
		return ErrInput
	}
	return nil
}

// Apply copies the current list, removes the first ordinal match per Remove
// entry, then appends Add values without deduplication. It is not atomic with
// the preceding resource read and provides no concurrency/ETag guarantee.
func (v *ListEdit) Apply(current []string) (*StringList, error) {
	return v.ApplyNullable(stringPointers(current))
}

// ApplyNullable preserves null entries and owns copies of all returned strings.
// Like Apply, it does not mutate the preparation snapshot.
func (v *ListEdit) ApplyNullable(current []*string) (*StringList, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if !v.delta {
		return NullableStrings(v.replace...), nil
	}
	if len(v.add) == 0 && len(v.remove) == 0 {
		return NullableStrings(), nil
	}
	values := cloneStringPointers(current)
	for _, item := range v.remove {
		if i := slices.IndexFunc(values, func(value *string) bool {
			return value == nil && item == nil || value != nil && item != nil && *value == *item
		}); i >= 0 {
			values = slices.Delete(values, i, i+1)
		}
	}
	values = append(values, v.add...)
	return NullableStrings(values...), nil
}

func stringPointers(values []string) []*string {
	result := make([]*string, len(values))
	for i, value := range values {
		result[i] = new(value)
	}
	return result
}

func cloneStringPointers(values []*string) []*string {
	result := make([]*string, len(values))
	for i, value := range values {
		if value != nil {
			result[i] = new(*value)
		}
	}
	return result
}
func (ListEdit) String() string     { return "policy list edit (redacted)" }
func (v ListEdit) GoString() string { return v.String() }

// Controls distinguishes unbound Force from a bound false/true switch for
// compiled CRUD bodies. It does not bypass authentication or authorization.
type Controls struct{ Force *bool }

// UserAssignment, TenantAssignment and GroupAssignment preserve the pinned
// per-scope PolicyName null/empty distinction. Group Rank is a bound Int32;
// backend range and clear semantics remain separately qualified.
type UserAssignment struct {
	Identity   string
	PolicyName *string
}
type TenantAssignment struct {
	PolicyName *string
	Force      bool
}
type GroupAssignment struct {
	ID         string
	PolicyName *string
	Rank       int32
}

// Assignment selects exactly one scope. Its zero/mixed value fails locally,
// corresponding to the compiled Grant's otherwise unhandled Identity set.
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupAssignment
}
