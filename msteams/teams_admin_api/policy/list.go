package policy

import (
	"encoding/json"
	"slices"
)

// StringList is an owned full-array request value. Nil/empty input serializes [];
// omitting the *StringList field is distinct from clearing that setting.
type StringList struct{ values []string }

func Strings(values ...string) *StringList { return &StringList{values: append([]string{}, values...)} }
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
	replace []string
	add     []string
	remove  []string
	delta   bool
	valid   bool
}

func ReplaceStrings(values ...string) *ListEdit {
	return &ListEdit{replace: append([]string{}, values...), valid: true}
}
func ChangeStrings(add, remove []string) *ListEdit {
	return &ListEdit{add: slices.Clone(add), remove: slices.Clone(remove), delta: true, valid: true}
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
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if !v.delta {
		return Strings(v.replace...), nil
	}
	if len(v.add) == 0 && len(v.remove) == 0 {
		return Strings(), nil
	}
	values := append([]string{}, current...)
	for _, item := range v.remove {
		if i := slices.Index(values, item); i >= 0 {
			values = slices.Delete(values, i, i+1)
		}
	}
	values = append(values, v.add...)
	return Strings(values...), nil
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
