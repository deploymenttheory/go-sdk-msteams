package teamstemplatepermissionpolicy

import "encoding/json"

// NewHiddenTemplate is a pure local constructor. The pinned setters accept empty,
// whitespace and future non-GUID IDs; Go strings cannot express a null argument.
func NewHiddenTemplate(id string) *HiddenTemplate { return &HiddenTemplate{ID: new(id)} }
func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneTemplate(v *HiddenTemplate) *HiddenTemplate {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.ID = cloneString(v.ID)
	x.RawJSON = append(json.RawMessage(nil), v.RawJSON...)
	if v.RawFields != nil {
		x.RawFields = make(map[string]json.RawMessage, len(v.RawFields))
		for k, b := range v.RawFields {
			x.RawFields[k] = append(json.RawMessage(nil), b...)
		}
	}
	return &x
}
func cloneTemplates(v []*HiddenTemplate) []*HiddenTemplate {
	out := make([]*HiddenTemplate, len(v))
	for i, x := range v {
		out[i] = cloneTemplate(x)
	}
	return out
}

// HiddenTemplateEdit applies the source's local list-modifier rules. A nil request edit
// is unbound; ClearHiddenTemplates expresses a bound null or empty modifier.
type HiddenTemplateEdit struct {
	replacement, add, remove []*HiddenTemplate
	delta, valid             bool
}

func ReplaceHiddenTemplates(v ...*HiddenTemplate) *HiddenTemplateEdit {
	return &HiddenTemplateEdit{replacement: cloneTemplates(v), valid: true}
}
func ClearHiddenTemplates() *HiddenTemplateEdit { return ReplaceHiddenTemplates() }
func ChangeHiddenTemplates(add, remove []*HiddenTemplate) *HiddenTemplateEdit {
	return &HiddenTemplateEdit{add: cloneTemplates(add), remove: cloneTemplates(remove), delta: true, valid: true}
}
func (e *HiddenTemplateEdit) Validate() error {
	if e == nil || !e.valid {
		return ErrInput
	}
	return nil
}
func equalID(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

// Apply owns its result, removes the first ordinal Id match per Remove entry,
// then appends Add without deduplication. Identity and future fields are not keys.
func (e *HiddenTemplateEdit) Apply(current []*HiddenTemplate) ([]*HiddenTemplate, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	if !e.delta {
		return cloneTemplates(e.replacement), nil
	}
	if len(e.add) == 0 && len(e.remove) == 0 {
		return []*HiddenTemplate{}, nil
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
			if equalID(remove.ID, item.ID) {
				out = append(out[:i], out[i+1:]...)
				break
			}
		}
	}
	return append(out, cloneTemplates(e.add)...), nil
}
func (HiddenTemplateEdit) String() string     { return "hidden template edit (redacted)" }
func (v HiddenTemplateEdit) GoString() string { return v.String() }
func templateDTO(v []*HiddenTemplate) []map[string]*string {
	out := make([]map[string]*string, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]*string{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
