package teamtemplate

import (
	"bytes"
	"encoding/json"
)

// Definition is the flat request model accepted by template create/update.
// Pointer strings/booleans retain omitted versus explicitly empty/false values.
// Nil slices are omitted; non-nil empty slices serialize as []. Null/empty object
// entries are skipped. Unknown JSON Body keys are not write passthrough fields.
// Struct construction is sparse. JSON decoding uses the pinned Body constructor
// behavior, including five default settings objects; see ParseDefinition.
type Definition struct {
	TemplateID                  *string            `json:"templateId,omitempty"`
	Classification              *string            `json:"classification,omitempty"`
	DisplayName                 *string            `json:"displayName,omitempty"`
	Description                 *string            `json:"description,omitempty"`
	OwnerUserObjectID           *string            `json:"ownerUserObjectId,omitempty"`
	Specialization              *string            `json:"specialization,omitempty"`
	Visibility                  *string            `json:"visibility,omitempty"`
	Channels                    []*Channel         `json:"channels,omitempty"`
	MemberSettings              *MemberSettings    `json:"memberSettings,omitempty"`
	GuestSettings               *GuestSettings     `json:"guestSettings,omitempty"`
	MessagingSettings           *MessagingSettings `json:"messagingSettings,omitempty"`
	FunSettings                 *FunSettings       `json:"funSettings,omitempty"`
	DiscoverySettings           *DiscoverySettings `json:"discoverySettings,omitempty"`
	IsMembershipLimitedToOwners *bool              `json:"isMembershipLimitedToOwners,omitempty"`
	Apps                        []*App             `json:"apps,omitempty"`
	Icon                        *string            `json:"icon,omitempty"`
	ShortDescription            *string            `json:"shortDescription,omitempty"`
	Categories                  []string           `json:"categories,omitempty"`
	PublishedBy                 *string            `json:"publishedBy,omitempty"`
	URI                         *string            `json:"uri,omitempty"`
}

// NewDefinition supplies the two required expanded-set strings. Body input can
// instead use a Definition literal or ParseDefinition and is server-validated.
// Mandatory strings are checked for empty values, not invented whitespace rules.
func NewDefinition(displayName, shortDescription string) (Definition, error) {
	if displayName == "" || shortDescription == "" {
		return Definition{}, ErrInput
	}
	return Definition{DisplayName: &displayName, ShortDescription: &shortDescription}, nil
}

// MarshalJSON keeps explicitly empty collections while dropping null object
// items. It does not initialize settings or mutate any caller-owned value.
func (d Definition) MarshalJSON() ([]byte, error) {
	type plain Definition
	body := struct {
		plain
		Channels   *[]json.RawMessage `json:"channels,omitempty"`
		Apps       *[]json.RawMessage `json:"apps,omitempty"`
		Categories *[]string          `json:"categories,omitempty"`
	}{plain: plain(d)}
	if d.Channels != nil {
		items, err := encodedObjects(d.Channels)
		if err != nil {
			return nil, err
		}
		body.Channels = &items
	}
	if d.Apps != nil {
		items, err := encodedObjects(d.Apps)
		if err != nil {
			return nil, err
		}
		body.Apps = &items
	}
	if d.Categories != nil {
		body.Categories = &d.Categories
	}
	return json.Marshal(body)
}

// ParseDefinition follows the generated JSON Body model: root keys win over
// teamDefinition, then extendedProperties fills still-missing keys. A root null
// also masks a nested value. Missing/null settings become false-default objects;
// known properties of another JSON type are ignored, as in the pinned parser.
// The result writes known flat fields only. Response snapshots retain unknown
// fields separately in Template/TemplateSummary.RawFields.
func ParseDefinition(raw []byte) (*Definition, error) {
	var definition Definition
	if err := definition.UnmarshalJSON(raw); err != nil {
		return nil, err
	}
	return &definition, nil
}

func (d *Definition) UnmarshalJSON(raw []byte) error {
	fields, ok := object(raw)
	if d == nil || !ok {
		return ErrInput
	}
	fields = flatten(fields)
	value := Definition{
		TemplateID: stringField(fields, "templateId"), Classification: stringField(fields, "classification"),
		DisplayName: stringField(fields, "displayName"), Description: stringField(fields, "description"),
		OwnerUserObjectID: stringField(fields, "ownerUserObjectId"), Specialization: stringField(fields, "specialization"),
		Visibility: stringField(fields, "visibility"), Icon: stringField(fields, "icon"),
		ShortDescription: stringField(fields, "shortDescription"), PublishedBy: stringField(fields, "publishedBy"),
		URI: stringField(fields, "uri"), IsMembershipLimitedToOwners: boolField(fields, "isMembershipLimitedToOwners"),
		Channels: objectItems(fields["channels"], channelFromFields),
		Apps:     objectItems(fields["apps"], func(item map[string]json.RawMessage) *App { return &App{ID: stringField(item, "id")} }),
	}
	if items, ok := array(fields["categories"]); ok {
		value.Categories = []string{}
		for _, item := range items {
			if text := scalar[string](item); text != nil {
				value.Categories = append(value.Categories, *text)
			}
		}
	}
	value.readSettings(fields)
	*d = value
	return nil
}

// App is a template app reference, not an app installation operation.
type App struct {
	ID *string `json:"id,omitempty"`
}

// Channel is a template channel definition, not a provisioned channel resource.
type Channel struct {
	ID                  *string `json:"id,omitempty"`
	DisplayName         *string `json:"displayName,omitempty"`
	Description         *string `json:"description,omitempty"`
	IsFavoriteByDefault *bool   `json:"isFavoriteByDefault,omitempty"`
	Tabs                []*Tab  `json:"tabs,omitempty"`
}

func (c Channel) MarshalJSON() ([]byte, error) {
	type plain Channel
	body := struct {
		plain
		Tabs *[]json.RawMessage `json:"tabs,omitempty"`
	}{plain: plain(c)}
	if c.Tabs != nil {
		items, err := encodedObjects(c.Tabs)
		if err != nil {
			return nil, err
		}
		body.Tabs = &items
	}
	return json.Marshal(body)
}

type Tab struct {
	ID             *string           `json:"id,omitempty"`
	TeamsAppID     *string           `json:"teamsAppId,omitempty"`
	Name           *string           `json:"name,omitempty"`
	SortOrderIndex *string           `json:"sortOrderIndex,omitempty"`
	MessageID      *string           `json:"messageId,omitempty"`
	WebURL         *string           `json:"webUrl,omitempty"`
	Key            *string           `json:"key,omitempty"`
	Configuration  *TabConfiguration `json:"configuration,omitempty"`
}

func (t Tab) MarshalJSON() ([]byte, error) {
	type plain Tab
	body := struct {
		plain
		Configuration *json.RawMessage `json:"configuration,omitempty"`
	}{plain: plain(t)}
	if t.Configuration != nil {
		raw, err := json.Marshal(t.Configuration)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(raw, []byte("{}")) {
			value := json.RawMessage(raw)
			body.Configuration = &value
		}
	}
	return json.Marshal(body)
}

type TabConfiguration struct {
	EntityID   *string `json:"entityId,omitempty"`
	ContentURL *string `json:"contentUrl,omitempty"`
	RemoveURL  *string `json:"removeUrl,omitempty"`
	WebsiteURL *string `json:"websiteUrl,omitempty"`
}

type DiscoverySettings struct {
	ShowInTeamsSearchAndSuggestions bool `json:"showInTeamsSearchAndSuggestions"`
}

type FunSettings struct {
	AllowGiphy            bool    `json:"allowGiphy"`
	AllowStickersAndMemes bool    `json:"allowStickersAndMemes"`
	AllowCustomMemes      bool    `json:"allowCustomMemes"`
	GiphyContentRating    *string `json:"giphyContentRating,omitempty"`
}

type GuestSettings struct {
	AllowCreateUpdateChannels bool `json:"allowCreateUpdateChannels"`
	AllowDeleteChannels       bool `json:"allowDeleteChannels"`
}

type MemberSettings struct {
	AllowCreateUpdateChannels         bool `json:"allowCreateUpdateChannels"`
	AllowDeleteChannels               bool `json:"allowDeleteChannels"`
	AllowAddRemoveApps                bool `json:"allowAddRemoveApps"`
	UploadCustomApp                   bool `json:"uploadCustomApp"`
	AllowCreateUpdateRemoveTabs       bool `json:"allowCreateUpdateRemoveTabs"`
	AllowCreateUpdateRemoveConnectors bool `json:"allowCreateUpdateRemoveConnectors"`
	AllowCreatePrivateChannels        bool `json:"allowCreatePrivateChannels"`
}

type MessagingSettings struct {
	AllowUserEditMessages    bool `json:"allowUserEditMessages"`
	AllowUserDeleteMessages  bool `json:"allowUserDeleteMessages"`
	AllowOwnerDeleteMessages bool `json:"allowOwnerDeleteMessages"`
	AllowTeamMentions        bool `json:"allowTeamMentions"`
	AllowChannelMentions     bool `json:"allowChannelMentions"`
}

func (d *Definition) readSettings(fields map[string]json.RawMessage) {
	discovery, _ := object(fields["discoverySettings"])
	d.DiscoverySettings = &DiscoverySettings{ShowInTeamsSearchAndSuggestions: boolean(discovery, "showInTeamsSearchAndSuggestions")}
	fun, _ := object(fields["funSettings"])
	d.FunSettings = &FunSettings{AllowGiphy: boolean(fun, "allowGiphy"), AllowStickersAndMemes: boolean(fun, "allowStickersAndMemes"),
		AllowCustomMemes: boolean(fun, "allowCustomMemes"), GiphyContentRating: stringField(fun, "giphyContentRating")}
	guest, _ := object(fields["guestSettings"])
	d.GuestSettings = &GuestSettings{AllowCreateUpdateChannels: boolean(guest, "allowCreateUpdateChannels"), AllowDeleteChannels: boolean(guest, "allowDeleteChannels")}
	member, _ := object(fields["memberSettings"])
	d.MemberSettings = &MemberSettings{AllowCreateUpdateChannels: boolean(member, "allowCreateUpdateChannels"), AllowDeleteChannels: boolean(member, "allowDeleteChannels"),
		AllowAddRemoveApps: boolean(member, "allowAddRemoveApps"), UploadCustomApp: boolean(member, "uploadCustomApp"),
		AllowCreateUpdateRemoveTabs: boolean(member, "allowCreateUpdateRemoveTabs"), AllowCreateUpdateRemoveConnectors: boolean(member, "allowCreateUpdateRemoveConnectors"),
		AllowCreatePrivateChannels: boolean(member, "allowCreatePrivateChannels")}
	messaging, _ := object(fields["messagingSettings"])
	d.MessagingSettings = &MessagingSettings{AllowUserEditMessages: boolean(messaging, "allowUserEditMessages"), AllowUserDeleteMessages: boolean(messaging, "allowUserDeleteMessages"),
		AllowOwnerDeleteMessages: boolean(messaging, "allowOwnerDeleteMessages"), AllowTeamMentions: boolean(messaging, "allowTeamMentions"), AllowChannelMentions: boolean(messaging, "allowChannelMentions")}
}

func channelFromFields(fields map[string]json.RawMessage) *Channel {
	return &Channel{ID: stringField(fields, "id"), DisplayName: stringField(fields, "displayName"), Description: stringField(fields, "description"),
		IsFavoriteByDefault: boolField(fields, "isFavoriteByDefault"), Tabs: objectItems(fields["tabs"], tabFromFields)}
}

func tabFromFields(fields map[string]json.RawMessage) *Tab {
	value := &Tab{ID: stringField(fields, "id"), TeamsAppID: stringField(fields, "teamsAppId"), Name: stringField(fields, "name"),
		SortOrderIndex: stringField(fields, "sortOrderIndex"), MessageID: stringField(fields, "messageId"), WebURL: stringField(fields, "webUrl"), Key: stringField(fields, "key")}
	if config, ok := object(fields["configuration"]); ok {
		value.Configuration = &TabConfiguration{EntityID: stringField(config, "entityId"), ContentURL: stringField(config, "contentUrl"),
			RemoveURL: stringField(config, "removeUrl"), WebsiteURL: stringField(config, "websiteUrl")}
	}
	return value
}

func object(raw []byte) (map[string]json.RawMessage, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, false
	}
	return fields, true
}

func array(raw []byte) ([]json.RawMessage, bool) {
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil || items == nil {
		return nil, false
	}
	return items, true
}

func flatten(fields map[string]json.RawMessage) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage, len(fields))
	for key, value := range fields {
		result[key] = value
	}
	for _, name := range []string{"teamDefinition", "extendedProperties"} {
		nested, _ := object(result[name])
		for key, value := range nested {
			if _, exists := result[key]; !exists {
				result[key] = value
			}
		}
		delete(result, name)
	}
	return result
}

func scalar[T string | bool | int64](raw []byte) *T {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil
	}
	var value T
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}

func stringField(fields map[string]json.RawMessage, name string) *string {
	return scalar[string](fields[name])
}
func boolField(fields map[string]json.RawMessage, name string) *bool {
	return scalar[bool](fields[name])
}
func boolean(fields map[string]json.RawMessage, name string) bool {
	value := boolField(fields, name)
	return value != nil && *value
}

func encodedObjects[T any](values []*T) ([]json.RawMessage, error) {
	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		// The pinned array writer drops null/empty-object model values, while
		// retaining an explicit array even when every item was dropped.
		if !bytes.Equal(raw, []byte("{}")) && !bytes.Equal(raw, []byte("null")) {
			result = append(result, raw)
		}
	}
	return result, nil
}

func objectItems[T any](raw []byte, parse func(map[string]json.RawMessage) *T) []*T {
	items, ok := array(raw)
	if !ok {
		return nil
	}
	result := make([]*T, 0, len(items))
	for _, item := range items {
		if fields, ok := object(item); ok {
			result = append(result, parse(fields))
		}
	}
	return result
}
