// Package teamsapppermissionpolicy manages Teams application permission policies.
package teamsapppermissionpolicy

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// CatalogApp represents the common fields and setter rules of the three source
// catalog-entry types. A missing Id is distinct from an explicit null setter.
type CatalogApp struct {
	Identity  *string                    `json:"Identity"`
	ID        *string                    `json:"Id"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

// Policy is a received snapshot. RawJSON retains repeated properties and wire
// order; RawFields exposes the last occurrence of each exact key for inspection.
// Mutation requests are separate and do not serialize this raw snapshot.
type Policy struct {
	Identity               *string                    `json:"Identity"`
	Description            *string                    `json:"Description"`
	DefaultCatalogAppsType *string                    `json:"DefaultCatalogAppsType"`
	GlobalCatalogAppsType  *string                    `json:"GlobalCatalogAppsType"`
	PrivateCatalogAppsType *string                    `json:"PrivateCatalogAppsType"`
	DefaultCatalogApps     []*CatalogApp              `json:"DefaultCatalogApps"`
	GlobalCatalogApps      []*CatalogApp              `json:"GlobalCatalogApps"`
	PrivateCatalogApps     []*CatalogApp              `json:"PrivateCatalogApps"`
	RawFields              map[string]json.RawMessage `json:"-"`
	RawJSON                json.RawMessage            `json:"-"`
}

func object(raw []byte, visit func(string, json.RawMessage) error) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return nil, client.ErrDecode
	}
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return nil, client.ErrDecode
		}
		name, ok := key.(string)
		if !ok {
			return nil, client.ErrDecode
		}
		var body json.RawMessage
		if d.Decode(&body) != nil || visit(name, body) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}
func (v *CatalogApp) UnmarshalJSON(raw []byte) error {
	var value CatalogApp
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			if json.Unmarshal(body, &value.Identity) != nil {
				return client.ErrDecode
			}
		case "id":
			if json.Unmarshal(body, &value.ID) != nil || value.ID == nil {
				return client.ErrDecode
			}
		}
		return nil
	})
	if e != nil {
		return e
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v CatalogApp) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain CatalogApp
	return json.Marshal(plain(v))
}
func (v *Policy) UnmarshalJSON(raw []byte) error {
	value := Policy{DefaultCatalogApps: []*CatalogApp{}, GlobalCatalogApps: []*CatalogApp{}, PrivateCatalogApps: []*CatalogApp{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		var scalar **string
		var list *[]*CatalogApp
		switch strings.ToLower(key) {
		case "identity":
			scalar = &value.Identity
		case "description":
			scalar = &value.Description
		case "defaultcatalogappstype":
			scalar = &value.DefaultCatalogAppsType
		case "globalcatalogappstype":
			scalar = &value.GlobalCatalogAppsType
		case "privatecatalogappstype":
			scalar = &value.PrivateCatalogAppsType
		case "defaultcatalogapps":
			list = &value.DefaultCatalogApps
		case "globalcatalogapps":
			list = &value.GlobalCatalogApps
		case "privatecatalogapps":
			list = &value.PrivateCatalogApps
		}
		if scalar != nil && json.Unmarshal(body, scalar) != nil {
			return client.ErrDecode
		}
		if list != nil && !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			var items []*CatalogApp
			if json.Unmarshal(body, &items) != nil {
				return client.ErrDecode
			}
			*list = append(*list, items...)
		}
		return nil
	})
	if e != nil {
		return e
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Policy
	value := plain(v)
	if value.DefaultCatalogApps == nil {
		value.DefaultCatalogApps = []*CatalogApp{}
	}
	if value.GlobalCatalogApps == nil {
		value.GlobalCatalogApps = []*CatalogApp{}
	}
	if value.PrivateCatalogApps == nil {
		value.PrivateCatalogApps = []*CatalogApp{}
	}
	return json.Marshal(value)
}
func (Policy) String() string         { return "Teams app permission policy (redacted)" }
func (v Policy) GoString() string     { return v.String() }
func (CatalogApp) String() string     { return "catalog app (redacted)" }
func (v CatalogApp) GoString() string { return v.String() }
