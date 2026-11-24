package teamsfirstpartymeetingtemplateconfiguration

import "encoding/json"

// NewTemplate is a pure local constructor matching the pinned name constructor:
// it sets Name and an empty option list. Go strings cannot express a null name.
func NewTemplate(name string) *Template {
	return &Template{Name: new(name), TeamsMeetingOptions: []*Option{}}
}

// NewOption is a pure local constructor for a named template option.
func NewOption(name string) *Option { return &Option{Name: new(name)} }

func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneBool(v *bool) *bool {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneRaw(raw json.RawMessage, fields map[string]json.RawMessage) (json.RawMessage, map[string]json.RawMessage) {
	out := append(json.RawMessage(nil), raw...)
	if fields == nil {
		return out, nil
	}
	copied := make(map[string]json.RawMessage, len(fields))
	for k, b := range fields {
		copied[k] = append(json.RawMessage(nil), b...)
	}
	return out, copied
}
func cloneOption(v *Option) *Option {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity, x.Name, x.Value = cloneString(v.Identity), cloneString(v.Name), cloneString(v.Value)
	x.IsLocked, x.IsHidden = cloneBool(v.IsLocked), cloneBool(v.IsHidden)
	x.RawJSON, x.RawFields = cloneRaw(v.RawJSON, v.RawFields)
	return &x
}
func cloneTemplate(v *Template) *Template {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity, x.Name = cloneString(v.Identity), cloneString(v.Name)
	x.Description, x.Category = cloneString(v.Description), cloneString(v.Category)
	if v.TeamsMeetingOptions != nil {
		x.TeamsMeetingOptions = make([]*Option, len(v.TeamsMeetingOptions))
		for i, o := range v.TeamsMeetingOptions {
			x.TeamsMeetingOptions[i] = cloneOption(o)
		}
	}
	x.RawJSON, x.RawFields = cloneRaw(v.RawJSON, v.RawFields)
	return &x
}
func cloneTemplates(v []*Template) []*Template {
	out := make([]*Template, len(v))
	for i, x := range v {
		out[i] = cloneTemplate(x)
	}
	return out
}

// TemplateEdit applies the pinned local list-modifier rules. A nil request edit
// is unbound; ClearTemplates expresses a bound null or empty modifier.
type TemplateEdit struct {
	replacement, add, remove []*Template
	delta, valid             bool
}

// ReplaceTemplates replaces the complete template list with owned copies.
func ReplaceTemplates(v ...*Template) *TemplateEdit {
	return &TemplateEdit{replacement: cloneTemplates(v), valid: true}
}

// ClearTemplates binds an empty template list.
func ClearTemplates() *TemplateEdit { return ReplaceTemplates() }

// ChangeTemplates removes the first ordinal Name match per Remove entry, then
// appends Add entries without deduplication.
func ChangeTemplates(add, remove []*Template) *TemplateEdit {
	return &TemplateEdit{add: cloneTemplates(add), remove: cloneTemplates(remove), delta: true, valid: true}
}
func (e *TemplateEdit) Validate() error {
	if e == nil || !e.valid {
		return ErrInput
	}
	return nil
}
func equalName(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

// Apply owns its result. Names compare ordinally (case-sensitive), like the
// pinned key predicate; Identity, options and unknown fields are not keys.
func (e *TemplateEdit) Apply(current []*Template) ([]*Template, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	if !e.delta {
		return cloneTemplates(e.replacement), nil
	}
	if len(e.add) == 0 && len(e.remove) == 0 {
		return []*Template{}, nil
	}
	out := cloneTemplates(current)
	for _, remove := range e.remove {
		if remove == nil {
			return nil, ErrInput
		}
		for i, item := range out {
			if item == nil {
				return nil, ErrInput
			}
			if equalName(remove.Name, item.Name) {
				out = append(out[:i], out[i+1:]...)
				break
			}
		}
	}
	return append(out, cloneTemplates(e.add)...), nil
}
func (TemplateEdit) String() string     { return "first-party meeting template edit (redacted)" }
func (v TemplateEdit) GoString() string { return v.String() }

// The compiled serializer writes every typed template and option property once
// the template list is bound; unknown response fields are not retained in writes.
type optionDTO struct {
	Identity *string `json:"Identity"`
	IsLocked *bool   `json:"IsLocked"`
	IsHidden *bool   `json:"IsHidden"`
	Value    *string `json:"Value"`
	Name     *string `json:"Name"`
}
type templateDTO struct {
	Identity            *string      `json:"Identity"`
	TeamsMeetingOptions []*optionDTO `json:"TeamsMeetingOptions"`
	Description         *string      `json:"Description"`
	Name                *string      `json:"Name"`
	Category            *string      `json:"Category"`
}

func templatesDTO(v []*Template) []*templateDTO {
	out := make([]*templateDTO, len(v))
	for i, x := range v {
		if x == nil {
			continue
		}
		options := make([]*optionDTO, len(x.TeamsMeetingOptions))
		for j, o := range x.TeamsMeetingOptions {
			if o != nil {
				options[j] = &optionDTO{Identity: o.Identity, IsLocked: o.IsLocked, IsHidden: o.IsHidden, Value: o.Value, Name: o.Name}
			}
		}
		out[i] = &templateDTO{Identity: x.Identity, TeamsMeetingOptions: options, Description: x.Description, Name: x.Name, Category: x.Category}
	}
	return out
}
