package team

import (
	"encoding/json"
	"time"
)

// Visibility remains open to future Graph values on reads.
type Visibility string

const (
	Public  Visibility = "public"
	Private Visibility = "private"
)

// Team is a native Graph snapshot. List returns only basic properties; nil does
// not mean false or an empty string. RawFields retains the complete read object,
// including future properties, and is never used as a mutation payload.
type Team struct {
	Settings
	ID              string                     `json:"id"`
	DisplayName     *string                    `json:"displayName,omitempty"`
	Description     *string                    `json:"description,omitempty"`
	Visibility      *Visibility                `json:"visibility,omitempty"`
	IsArchived      *bool                      `json:"isArchived,omitempty"`
	WebURL          *string                    `json:"webUrl,omitempty"`
	InternalID      *string                    `json:"internalId,omitempty"`
	Classification  *string                    `json:"classification,omitempty"`
	CreatedDateTime *time.Time                 `json:"createdDateTime,omitempty"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON requires an object with an identity, including for list items.
func (t *Team) UnmarshalJSON(data []byte) error {
	type snapshot Team
	var value snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.ID == "" {
		return errModel
	}
	if err := json.Unmarshal(data, &value.RawFields); err != nil {
		return err
	}
	*t = Team(value)
	return nil
}

// InitialMember binds an Entra user ID, not a Teams conversation-member ID.
// Owner selects the documented owner role; false creates an ordinary member.
type InitialMember struct {
	UserID string
	Owner  bool
}

// CreateRequest is the native asynchronous POST /teams request. TemplateID
// defaults to standard. Bindings use this client's configured Graph cloud.
// Application requests require at least one initial user. No group-first
// PowerShell composition or implicit owner lookup is performed here.
type CreateRequest struct {
	Settings
	DisplayName      string
	Description      *string
	Visibility       *Visibility
	FirstChannelName *string
	TemplateID       string
	Members          []InitialMember
}

// ArchiveOptions omits the SharePoint stage when its pointer is nil. Graph does
// not support supplying this parameter with application authentication, even
// when the explicit value is false.
type ArchiveOptions struct {
	ShouldSetSpoSiteReadOnlyForMembers *bool `json:"shouldSetSpoSiteReadOnlyForMembers,omitempty"`
}
