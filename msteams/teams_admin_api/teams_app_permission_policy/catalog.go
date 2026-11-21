package teamsapppermissionpolicy

import "encoding/json"

// NewCatalogApp is a pure local constructor. The pinned setters accept empty,
// whitespace and future non-GUID IDs; Go strings cannot express a null argument.
func NewCatalogApp(id string) *CatalogApp { return &CatalogApp{ID: new(id)} }
func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneApp(v *CatalogApp) *CatalogApp {
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
func cloneApps(v []*CatalogApp) []*CatalogApp {
	out := make([]*CatalogApp, len(v))
	for i, x := range v {
		out[i] = cloneApp(x)
	}
	return out
}

// CatalogEdit applies the source's local list-modifier rules. A nil request edit
// is unbound; ClearCatalogApps expresses a bound null or empty modifier.
type CatalogEdit struct {
	replacement, add, remove []*CatalogApp
	delta, valid             bool
}

func ReplaceCatalogApps(v ...*CatalogApp) *CatalogEdit {
	return &CatalogEdit{replacement: cloneApps(v), valid: true}
}
func ClearCatalogApps() *CatalogEdit { return ReplaceCatalogApps() }
func ChangeCatalogApps(add, remove []*CatalogApp) *CatalogEdit {
	return &CatalogEdit{add: cloneApps(add), remove: cloneApps(remove), delta: true, valid: true}
}
func (e *CatalogEdit) Validate() error {
	if e == nil || !e.valid {
		return ErrInput
	}
	return nil
}
func equalID(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

// Apply owns its result, removes the first ordinal Id match per Remove entry,
// then appends Add without deduplication. Identity and future fields are not keys.
func (e *CatalogEdit) Apply(current []*CatalogApp) ([]*CatalogApp, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	if !e.delta {
		return cloneApps(e.replacement), nil
	}
	if len(e.add) == 0 && len(e.remove) == 0 {
		return []*CatalogApp{}, nil
	}
	out := cloneApps(current)
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
	return append(out, cloneApps(e.add)...), nil
}
func (CatalogEdit) String() string     { return "catalog edit (redacted)" }
func (v CatalogEdit) GoString() string { return v.String() }
func appDTO(v []*CatalogApp) []map[string]*string {
	out := make([]map[string]*string, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]*string{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
