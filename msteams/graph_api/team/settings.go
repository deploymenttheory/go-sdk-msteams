// Package team provides native Graph team models and resource operations.
package team

// Settings is the native Graph settings payload. Nil containers/properties are
// omitted. Non-nil empty containers, explicit false and empty strings survive.
// PowerShell compositions may deliberately supply all five empty containers;
// native callers are not forced to send those defaults.
type Settings struct {
	MemberSettings    *MemberSettings    `json:"memberSettings,omitempty"`
	GuestSettings     *GuestSettings     `json:"guestSettings,omitempty"`
	MessagingSettings *MessagingSettings `json:"messagingSettings,omitempty"`
	FunSettings       *FunSettings       `json:"funSettings,omitempty"`
	DiscoverySettings *DiscoverySettings `json:"discoverySettings,omitempty"`
}

type MemberSettings struct {
	AllowCreateUpdateChannels         *bool `json:"allowCreateUpdateChannels,omitempty"`
	AllowDeleteChannels               *bool `json:"allowDeleteChannels,omitempty"`
	AllowAddRemoveApps                *bool `json:"allowAddRemoveApps,omitempty"`
	AllowCreateUpdateRemoveTabs       *bool `json:"allowCreateUpdateRemoveTabs,omitempty"`
	AllowCreateUpdateRemoveConnectors *bool `json:"allowCreateUpdateRemoveConnectors,omitempty"`
	AllowCreatePrivateChannels        *bool `json:"allowCreatePrivateChannels,omitempty"`
}

type GuestSettings struct {
	AllowCreateUpdateChannels *bool `json:"allowCreateUpdateChannels,omitempty"`
	AllowDeleteChannels       *bool `json:"allowDeleteChannels,omitempty"`
}

type MessagingSettings struct {
	AllowUserEditMessages    *bool `json:"allowUserEditMessages,omitempty"`
	AllowUserDeleteMessages  *bool `json:"allowUserDeleteMessages,omitempty"`
	AllowOwnerDeleteMessages *bool `json:"allowOwnerDeleteMessages,omitempty"`
	AllowTeamMentions        *bool `json:"allowTeamMentions,omitempty"`
	AllowChannelMentions     *bool `json:"allowChannelMentions,omitempty"`
}

// GiphyContentRating uses the documented Graph string representation. It does
// not reproduce the pinned module's numeric enum-serializer discrepancy.
// Keeping the underlying type open also preserves unknown future read values.
type GiphyContentRating string

const (
	GiphyModerate GiphyContentRating = "moderate"
	GiphyStrict   GiphyContentRating = "strict"
)

type FunSettings struct {
	AllowGiphy            *bool               `json:"allowGiphy,omitempty"`
	AllowStickersAndMemes *bool               `json:"allowStickersAndMemes,omitempty"`
	AllowCustomMemes      *bool               `json:"allowCustomMemes,omitempty"`
	GiphyContentRating    *GiphyContentRating `json:"giphyContentRating,omitempty"`
}

type DiscoverySettings struct {
	ShowInTeamsSearchAndSuggestions *bool `json:"showInTeamsSearchAndSuggestions,omitempty"`
}
