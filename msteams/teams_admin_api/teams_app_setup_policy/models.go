// Package teamsappsetuppolicy manages Teams application setup policies.
package teamsappsetuppolicy

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// App is the common Id/Identity shape used by presets, meeting presets and calling pins.
type App struct {
	Identity  *string                    `json:"Identity"`
	ID        *string                    `json:"Id"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

// PinnedApp uses signed Int64 Order for app-bar and message-bar entries.
type PinnedApp struct {
	Identity  *string                    `json:"Identity"`
	ID        *string                    `json:"Id"`
	Order     int64                      `json:"Order"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

// CustomizationApp has a second nullable/missing-aware source string setter.
type CustomizationApp struct {
	Identity                  *string                    `json:"Identity"`
	ID                        *string                    `json:"Id"`
	AdditionalCustomizationID *string                    `json:"AdditionalCustomizationId"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
}

// Policy retains raw presence, future fields and duplicate property ordering.
// Readonly source list properties ignore null and append repeated arrays.
type Policy struct {
	Identity                    *string                    `json:"Identity"`
	Description                 *string                    `json:"Description"`
	AllowSideLoading            *bool                      `json:"AllowSideLoading"`
	AllowUserPinning            *bool                      `json:"AllowUserPinning"`
	AppPresetList               []*App                     `json:"AppPresetList"`
	PinnedAppBarApps            []*PinnedApp               `json:"PinnedAppBarApps"`
	PinnedMessageBarApps        []*PinnedApp               `json:"PinnedMessageBarApps"`
	AppPresetMeetingList        []*App                     `json:"AppPresetMeetingList"`
	AdditionalCustomizationApps []*CustomizationApp        `json:"AdditionalCustomizationApps"`
	PinnedCallingBarApps        []*App                     `json:"PinnedCallingBarApps"`
	RawFields                   map[string]json.RawMessage `json:"-"`
	RawJSON                     json.RawMessage            `json:"-"`
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
		if d.Decode(&body) != nil || visit(strings.ToLower(name), body) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}
func isNull(raw json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }

func decodeOrder(raw json.RawMessage) (int64, error) {
	if isNull(raw) {
		return 0, client.ErrDecode
	}
	var value int64
	if json.Unmarshal(raw, &value) == nil {
		return value, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return 0, client.ErrDecode
	}
	value, e := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if e != nil {
		return 0, client.ErrDecode
	}
	return value, nil
}

type element struct {
	identity, id, customization *string
	order                       int64
	fields                      map[string]json.RawMessage
	raw                         json.RawMessage
}

func decodeElement(raw []byte, kind string) (element, error) {
	var v element
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch key {
		case "identity":
			if json.Unmarshal(body, &v.identity) != nil {
				return client.ErrDecode
			}
		case "id":
			if json.Unmarshal(body, &v.id) != nil || v.id == nil {
				return client.ErrDecode
			}
		case "order":
			if kind == "pin" {
				order, e := decodeOrder(body)
				if e != nil {
					return e
				}
				v.order = order
			}
		case "additionalcustomizationid":
			if kind == "customization" && (json.Unmarshal(body, &v.customization) != nil || v.customization == nil) {
				return client.ErrDecode
			}
		}
		return nil
	})
	if e != nil {
		return element{}, e
	}
	v.fields = fields
	v.raw = append(json.RawMessage(nil), raw...)
	return v, nil
}
func (v *App) UnmarshalJSON(raw []byte) error {
	x, e := decodeElement(raw, "app")
	if e != nil {
		return e
	}
	*v = App{Identity: x.identity, ID: x.id, RawFields: x.fields, RawJSON: x.raw}
	return nil
}
func (v *PinnedApp) UnmarshalJSON(raw []byte) error {
	x, e := decodeElement(raw, "pin")
	if e != nil {
		return e
	}
	*v = PinnedApp{Identity: x.identity, ID: x.id, Order: x.order, RawFields: x.fields, RawJSON: x.raw}
	return nil
}
func (v *CustomizationApp) UnmarshalJSON(raw []byte) error {
	x, e := decodeElement(raw, "customization")
	if e != nil {
		return e
	}
	*v = CustomizationApp{Identity: x.identity, ID: x.id, AdditionalCustomizationID: x.customization, RawFields: x.fields, RawJSON: x.raw}
	return nil
}
func (v App) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain App
	return json.Marshal(plain(v))
}
func (v PinnedApp) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain PinnedApp
	return json.Marshal(plain(v))
}
func (v CustomizationApp) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain CustomizationApp
	return json.Marshal(plain(v))
}
func appendList[T any](body json.RawMessage, list *[]*T) error {
	if isNull(body) {
		return nil
	}
	var items []*T
	if json.Unmarshal(body, &items) != nil {
		return client.ErrDecode
	}
	*list = append(*list, items...)
	return nil
}
func (v *Policy) UnmarshalJSON(raw []byte) error {
	value := Policy{AppPresetList: []*App{}, PinnedAppBarApps: []*PinnedApp{}, PinnedMessageBarApps: []*PinnedApp{}, AppPresetMeetingList: []*App{}, AdditionalCustomizationApps: []*CustomizationApp{}, PinnedCallingBarApps: []*App{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch key {
		case "identity":
			if json.Unmarshal(body, &value.Identity) != nil {
				return client.ErrDecode
			}
		case "description":
			if json.Unmarshal(body, &value.Description) != nil {
				return client.ErrDecode
			}
		case "allowsideloading":
			if json.Unmarshal(body, &value.AllowSideLoading) != nil {
				return client.ErrDecode
			}
		case "allowuserpinning":
			if json.Unmarshal(body, &value.AllowUserPinning) != nil {
				return client.ErrDecode
			}
		case "apppresetlist":
			return appendList(body, &value.AppPresetList)
		case "pinnedappbarapps":
			return appendList(body, &value.PinnedAppBarApps)
		case "pinnedmessagebarapps":
			return appendList(body, &value.PinnedMessageBarApps)
		case "apppresetmeetinglist":
			return appendList(body, &value.AppPresetMeetingList)
		case "additionalcustomizationapps":
			return appendList(body, &value.AdditionalCustomizationApps)
		case "pinnedcallingbarapps":
			return appendList(body, &value.PinnedCallingBarApps)
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
	x := plain(v)
	if x.AppPresetList == nil {
		x.AppPresetList = []*App{}
	}
	if x.PinnedAppBarApps == nil {
		x.PinnedAppBarApps = []*PinnedApp{}
	}
	if x.PinnedMessageBarApps == nil {
		x.PinnedMessageBarApps = []*PinnedApp{}
	}
	if x.AppPresetMeetingList == nil {
		x.AppPresetMeetingList = []*App{}
	}
	if x.AdditionalCustomizationApps == nil {
		x.AdditionalCustomizationApps = []*CustomizationApp{}
	}
	if x.PinnedCallingBarApps == nil {
		x.PinnedCallingBarApps = []*App{}
	}
	return json.Marshal(x)
}
func (Policy) String() string               { return "Teams app setup policy (redacted)" }
func (v Policy) GoString() string           { return v.String() }
func (App) String() string                  { return "setup app (redacted)" }
func (v App) GoString() string              { return v.String() }
func (PinnedApp) String() string            { return "setup pin (redacted)" }
func (v PinnedApp) GoString() string        { return v.String() }
func (CustomizationApp) String() string     { return "setup customization (redacted)" }
func (v CustomizationApp) GoString() string { return v.String() }
