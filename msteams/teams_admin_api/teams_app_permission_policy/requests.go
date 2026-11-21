package teamsapppermissionpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings keeps unknown/empty/null catalog mode values and nullable text.
type ScalarSettings struct {
	Description            policy.Field[string] `json:"Description,omitzero"`
	DefaultCatalogAppsType policy.Field[string] `json:"DefaultCatalogAppsType,omitzero"`
	GlobalCatalogAppsType  policy.Field[string] `json:"GlobalCatalogAppsType,omitzero"`
	PrivateCatalogAppsType policy.Field[string] `json:"PrivateCatalogAppsType,omitzero"`
}

// Settings is a sparse native write with full replacement arrays. Nil slices
// become []; explicit null array fields are not emitted by the source DTO.
type Settings struct {
	ScalarSettings
	DefaultCatalogApps policy.Field[[]*CatalogApp]
	GlobalCatalogApps  policy.Field[[]*CatalogApp]
	PrivateCatalogApps policy.Field[[]*CatalogApp]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest explicitly applies local edits to the preceding read snapshot.
type UpdateRequest struct {
	ScalarSettings
	DefaultCatalogApps *CatalogEdit
	GlobalCatalogApps  *CatalogEdit
	PrivateCatalogApps *CatalogEdit
	policy.Controls
}
type CreateRequest UpdateRequest
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "app permission update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

func fields(input Settings) (map[string]json.RawMessage, error) {
	body, e := policyrp.Fields(input.ScalarSettings)
	if e != nil {
		return nil, e
	}
	for _, c := range []struct {
		name  string
		value policy.Field[[]*CatalogApp]
	}{{"DefaultCatalogApps", input.DefaultCatalogApps}, {"GlobalCatalogApps", input.GlobalCatalogApps}, {"PrivateCatalogApps", input.PrivateCatalogApps}} {
		if !c.value.IsSet() {
			continue
		}
		if c.value.IsNull() {
			return nil, ErrInput
		}
		items, _ := c.value.Value()
		raw, e := json.Marshal(appDTO(items))
		if e != nil {
			return nil, ErrInput
		}
		body[c.name] = raw
	}
	return body, nil
}
func validateEdits(input UpdateRequest) error {
	for _, edit := range []*CatalogEdit{input.DefaultCatalogApps, input.GlobalCatalogApps, input.PrivateCatalogApps} {
		if edit != nil {
			if e := edit.Validate(); e != nil {
				return e
			}
		}
	}
	return nil
}
func resolve(input UpdateRequest, before *Policy) (Settings, error) {
	if e := validateEdits(input); e != nil {
		return Settings{}, e
	}
	if before == nil {
		before = &Policy{}
	}
	out := Settings{ScalarSettings: input.ScalarSettings}
	for _, c := range []struct {
		edit    *CatalogEdit
		current []*CatalogApp
		field   *policy.Field[[]*CatalogApp]
	}{{input.DefaultCatalogApps, before.DefaultCatalogApps, &out.DefaultCatalogApps}, {input.GlobalCatalogApps, before.GlobalCatalogApps, &out.GlobalCatalogApps}, {input.PrivateCatalogApps, before.PrivateCatalogApps, &out.PrivateCatalogApps}} {
		if c.edit != nil {
			v, e := c.edit.Apply(c.current)
			if e != nil {
				return Settings{}, e
			}
			*c.field = policy.Set(v)
		}
	}
	return out, nil
}
