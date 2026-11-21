// Package policy provides request values shared by compiled policy resources.
// Modern generated ConfigAPI list operators have a separate wire contract.
package policy

import (
	"bytes"
	"encoding/json"
	"errors"
)

var (
	ErrInput      = errors.New("invalid compiled policy input")
	ErrUnsetField = errors.New("unset policy field cannot be serialized directly")
)

// Field distinguishes omission, a bound value and explicit null. Resource
// request fields use json:",omitzero" so the zero value is omitted. Set stores
// a value copy; ownership of any nested mutable values remains with the caller.
type Field[T any] struct {
	value T
	set   bool
	null  bool
}

func Set[T any](value T) Field[T]   { return Field[T]{value: value, set: true} }
func Null[T any]() Field[T]         { return Field[T]{set: true, null: true} }
func (v Field[T]) IsZero() bool     { return !v.set }
func (v Field[T]) IsSet() bool      { return v.set }
func (v Field[T]) IsNull() bool     { return v.set && v.null }
func (v Field[T]) Value() (T, bool) { return v.value, v.set && !v.null }

func (v Field[T]) MarshalJSON() ([]byte, error) {
	if !v.set {
		return nil, ErrUnsetField
	}
	if v.null {
		return []byte("null"), nil
	}
	return json.Marshal(v.value)
}

func (v *Field[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*v = Null[T]()
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = Set(value)
	return nil
}
func (Field[T]) String() string     { return "policy field (redacted)" }
func (v Field[T]) GoString() string { return v.String() }
