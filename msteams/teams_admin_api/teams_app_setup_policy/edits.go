package teamsappsetuppolicy

import "encoding/json"

func NewApp(id string) *App                          { return &App{ID: new(id)} }
func NewPinnedApp(id string, order int64) *PinnedApp { return &PinnedApp{ID: new(id), Order: order} }
func NewCustomizationApp(id, customizationID string) *CustomizationApp {
	return &CustomizationApp{ID: new(id), AdditionalCustomizationID: new(customizationID)}
}
func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneRaw(fields map[string]json.RawMessage, raw json.RawMessage) (map[string]json.RawMessage, json.RawMessage) {
	var out map[string]json.RawMessage
	if fields != nil {
		out = make(map[string]json.RawMessage, len(fields))
		for k, v := range fields {
			out[k] = append(json.RawMessage(nil), v...)
		}
	}
	return out, append(json.RawMessage(nil), raw...)
}
func cloneApp(v *App) *App {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.RawFields, x.RawJSON = cloneRaw(v.RawFields, v.RawJSON)
	return &x
}
func clonePin(v *PinnedApp) *PinnedApp {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.RawFields, x.RawJSON = cloneRaw(v.RawFields, v.RawJSON)
	return &x
}
func cloneCustomization(v *CustomizationApp) *CustomizationApp {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.AdditionalCustomizationID = cloneString(v.AdditionalCustomizationID)
	x.RawFields, x.RawJSON = cloneRaw(v.RawFields, v.RawJSON)
	return &x
}
func sameString(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

// Typed edits share ownership/order mechanics, while each public edit type fixes
// the source-specific equality rule. A nil request edit is unbound.
type edit[T any] struct {
	replacement, add, remove []*T
	delta, valid             bool
	clone                    func(*T) *T
	equal                    func(*T, *T) bool
}

func (e *edit[T]) Validate() error {
	if e == nil || !e.valid || e.clone == nil || e.equal == nil {
		return ErrInput
	}
	return nil
}
func copyItems[T any](v []*T, clone func(*T) *T) []*T {
	out := make([]*T, len(v))
	for i, x := range v {
		out[i] = clone(x)
	}
	return out
}
func (e *edit[T]) Apply(current []*T) ([]*T, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	if !e.delta {
		return copyItems(e.replacement, e.clone), nil
	}
	if len(e.add) == 0 && len(e.remove) == 0 {
		return []*T{}, nil
	}
	out := copyItems(current, e.clone)
	for _, remove := range e.remove {
		if remove == nil {
			return nil, ErrInput
		}
		for i, item := range out {
			if item == nil {
				return nil, ErrInput
			}
			if e.equal(remove, item) {
				out = append(out[:i], out[i+1:]...)
				break
			}
		}
	}
	return append(out, copyItems(e.add, e.clone)...), nil
}

type AppEdit struct{ edit[App] }
type PinEdit struct{ edit[PinnedApp] }
type CustomizationEdit struct{ edit[CustomizationApp] }

func (e *AppEdit) Validate() error {
	if e == nil {
		return ErrInput
	}
	return e.edit.Validate()
}
func (e *AppEdit) Apply(current []*App) ([]*App, error) {
	if e == nil {
		return nil, ErrInput
	}
	return e.edit.Apply(current)
}
func (e *PinEdit) Validate() error {
	if e == nil {
		return ErrInput
	}
	return e.edit.Validate()
}
func (e *PinEdit) Apply(current []*PinnedApp) ([]*PinnedApp, error) {
	if e == nil {
		return nil, ErrInput
	}
	return e.edit.Apply(current)
}
func (e *CustomizationEdit) Validate() error {
	if e == nil {
		return ErrInput
	}
	return e.edit.Validate()
}
func (e *CustomizationEdit) Apply(current []*CustomizationApp) ([]*CustomizationApp, error) {
	if e == nil {
		return nil, ErrInput
	}
	return e.edit.Apply(current)
}
func apps() edit[App] {
	return edit[App]{valid: true, clone: cloneApp, equal: func(a, b *App) bool { return sameString(a.ID, b.ID) }}
}
func pins() edit[PinnedApp] {
	return edit[PinnedApp]{valid: true, clone: clonePin, equal: func(a, b *PinnedApp) bool { return sameString(a.ID, b.ID) && a.Order == b.Order }}
}
func customizations() edit[CustomizationApp] {
	return edit[CustomizationApp]{valid: true, clone: cloneCustomization, equal: func(a, b *CustomizationApp) bool {
		return sameString(a.ID, b.ID) && sameString(a.AdditionalCustomizationID, b.AdditionalCustomizationID)
	}}
}
func ReplaceApps(v ...*App) *AppEdit {
	e := apps()
	e.replacement = copyItems(v, e.clone)
	return &AppEdit{e}
}
func ChangeApps(add, remove []*App) *AppEdit {
	e := apps()
	e.delta = true
	e.add = copyItems(add, e.clone)
	e.remove = copyItems(remove, e.clone)
	return &AppEdit{e}
}
func ClearApps() *AppEdit { return ReplaceApps() }
func ReplacePins(v ...*PinnedApp) *PinEdit {
	e := pins()
	e.replacement = copyItems(v, e.clone)
	return &PinEdit{e}
}
func ChangePins(add, remove []*PinnedApp) *PinEdit {
	e := pins()
	e.delta = true
	e.add = copyItems(add, e.clone)
	e.remove = copyItems(remove, e.clone)
	return &PinEdit{e}
}
func ClearPins() *PinEdit { return ReplacePins() }
func ReplaceCustomizations(v ...*CustomizationApp) *CustomizationEdit {
	e := customizations()
	e.replacement = copyItems(v, e.clone)
	return &CustomizationEdit{e}
}
func ChangeCustomizations(add, remove []*CustomizationApp) *CustomizationEdit {
	e := customizations()
	e.delta = true
	e.add = copyItems(add, e.clone)
	e.remove = copyItems(remove, e.clone)
	return &CustomizationEdit{e}
}
func ClearCustomizations() *CustomizationEdit { return ReplaceCustomizations() }
func (AppEdit) String() string                { return "app edit (redacted)" }
func (v AppEdit) GoString() string            { return v.String() }
func (PinEdit) String() string                { return "pin edit (redacted)" }
func (v PinEdit) GoString() string            { return v.String() }
func (CustomizationEdit) String() string      { return "customization edit (redacted)" }
func (v CustomizationEdit) GoString() string  { return v.String() }
