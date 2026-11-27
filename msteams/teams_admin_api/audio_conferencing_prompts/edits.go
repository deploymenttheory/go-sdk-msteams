package audioconferencingprompts

import "encoding/json"

// NewPrompt applies the pinned parameterized constructor's Id validation.
func NewPrompt(id, kind, locale string) (*Prompt, error) {
	if !validID(id) {
		return nil, ErrInput
	}
	return &Prompt{ID: new(id), Type: new(kind), Locale: new(locale)}, nil
}

// NewPromptPackage accepts empty/future default locales and validates its Id.
func NewPromptPackage(id, defaultLocale string) (*PromptPackage, error) {
	if !validID(id) {
		return nil, ErrInput
	}
	value := emptyPackage()
	value.ID, value.DefaultLocale = new(id), new(defaultLocale)
	return &value, nil
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
func clonePrompt(v *Prompt) *Prompt {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.Name = cloneString(v.Name)
	x.AudioPrompt = cloneString(v.AudioPrompt)
	x.TextPrompt = cloneString(v.TextPrompt)
	x.Type = cloneString(v.Type)
	x.Locale = cloneString(v.Locale)
	x.RawFields, x.RawJSON = cloneRaw(v.RawFields, v.RawJSON)
	return &x
}
func copyItems[T any](v []*T, clone func(*T) *T) []*T {
	out := make([]*T, len(v))
	for i, x := range v {
		out[i] = clone(x)
	}
	return out
}
func clonePackage(v *PromptPackage) *PromptPackage {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.Name = cloneString(v.Name)
	x.DefaultLocale = cloneString(v.DefaultLocale)
	x.InboundStartRecordingPrompt = copyItems(v.InboundStartRecordingPrompt, cloneString)
	x.InboundEndRecordingPrompt = copyItems(v.InboundEndRecordingPrompt, cloneString)
	x.OutboundStartRecordingPrompt = copyItems(v.OutboundStartRecordingPrompt, cloneString)
	x.OutboundEndRecordingPrompt = copyItems(v.OutboundEndRecordingPrompt, cloneString)
	x.MeetingStartRecordingPrompt = copyItems(v.MeetingStartRecordingPrompt, cloneString)
	x.MeetingEndRecordingPrompt = copyItems(v.MeetingEndRecordingPrompt, cloneString)
	x.RawFields, x.RawJSON = cloneRaw(v.RawFields, v.RawJSON)
	return &x
}
func sameString(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

type edit[T any] struct {
	replacement, add, remove []*T
	delta, valid             bool
	clone                    func(*T) *T
	equal                    func(*T, *T) (bool, error)
	validItem                func(*T) bool
}

func (e *edit[T]) Validate() error {
	if e == nil || !e.valid || e.clone == nil || e.equal == nil || e.validItem == nil {
		return ErrInput
	}
	for _, items := range [][]*T{e.replacement, e.add, e.remove} {
		for _, v := range items {
			if !e.validItem(v) {
				return ErrInput
			}
		}
	}
	return nil
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
			same, err := e.equal(remove, item)
			if err != nil {
				return nil, err
			}
			if same {
				out = append(out[:i], out[i+1:]...)
				break
			}
		}
	}
	return append(out, copyItems(e.add, e.clone)...), nil
}

// PromptEdit owns inputs and compares all six source key fields.
// A nil request field is unbound; ClearPrompts is an explicitly bound clear.
type PromptEdit struct{ edit[Prompt] }

// PackageEdit preserves the source prohibition on comparing package removals.
// A non-null Remove from an empty list succeeds; comparing an existing entry fails.
type PackageEdit struct{ edit[PromptPackage] }

func (e *PromptEdit) Validate() error {
	if e == nil {
		return ErrInput
	}
	return e.edit.Validate()
}
func (e *PromptEdit) Apply(current []*Prompt) ([]*Prompt, error) {
	if e == nil {
		return nil, ErrInput
	}
	return e.edit.Apply(current)
}
func (e *PackageEdit) Validate() error {
	if e == nil {
		return ErrInput
	}
	return e.edit.Validate()
}
func (e *PackageEdit) Apply(current []*PromptPackage) ([]*PromptPackage, error) {
	if e == nil {
		return nil, ErrInput
	}
	return e.edit.Apply(current)
}
func prompts() edit[Prompt] {
	return edit[Prompt]{valid: true, clone: clonePrompt,
		validItem: func(v *Prompt) bool { return v == nil || v.ID == nil || validID(*v.ID) },
		equal: func(a, b *Prompt) (bool, error) {
			if a == nil || b == nil {
				return false, ErrInput
			}
			return sameString(a.ID, b.ID) && sameString(a.Name, b.Name) && sameString(a.AudioPrompt, b.AudioPrompt) && sameString(a.TextPrompt, b.TextPrompt) && sameString(a.Type, b.Type) && sameString(a.Locale, b.Locale), nil
		}}
}
func packages() edit[PromptPackage] {
	return edit[PromptPackage]{valid: true, clone: clonePackage,
		validItem: func(v *PromptPackage) bool { return v == nil || v.ID == nil || validID(*v.ID) },
		equal:     func(_, _ *PromptPackage) (bool, error) { return false, ErrInput }}
}
func ReplacePrompts(v ...*Prompt) *PromptEdit {
	e := prompts()
	e.replacement = copyItems(v, e.clone)
	return &PromptEdit{e}
}
func ChangePrompts(add, remove []*Prompt) *PromptEdit {
	e := prompts()
	e.delta = true
	e.add = copyItems(add, e.clone)
	e.remove = copyItems(remove, e.clone)
	return &PromptEdit{e}
}
func ClearPrompts() *PromptEdit { return ReplacePrompts() }
func ReplacePackages(v ...*PromptPackage) *PackageEdit {
	e := packages()
	e.replacement = copyItems(v, e.clone)
	return &PackageEdit{e}
}
func ChangePackages(add, remove []*PromptPackage) *PackageEdit {
	e := packages()
	e.delta = true
	e.add = copyItems(add, e.clone)
	e.remove = copyItems(remove, e.clone)
	return &PackageEdit{e}
}
func ClearPackages() *PackageEdit      { return ReplacePackages() }
func (PromptEdit) String() string      { return "prompt edit (redacted)" }
func (v PromptEdit) GoString() string  { return v.String() }
func (PackageEdit) String() string     { return "prompt package edit (redacted)" }
func (v PackageEdit) GoString() string { return v.String() }
