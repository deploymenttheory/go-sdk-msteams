// Package channel manages Graph channel resources on the pinned beta routes.
package channel

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// MembershipType is open to future read values. Only the three known values
// can be selected for creation or a membership-type filter.
type MembershipType string

const (
	Standard MembershipType = "standard"
	Private  MembershipType = "private"
	Shared   MembershipType = "shared"
)

// Channel retains nullable properties, their original presence and future
// fields. Response fields are never implicitly copied to a mutation payload.
type Channel struct {
	ID                 string                     `json:"id"`
	ODataID            *string                    `json:"@odata.id,omitempty"`
	DisplayName        *string                    `json:"displayName,omitempty"`
	Description        *string                    `json:"description,omitempty"`
	MembershipType     *MembershipType            `json:"membershipType,omitempty"`
	TenantID           *string                    `json:"tenantId,omitempty"`
	CreatedDateTime    *time.Time                 `json:"createdDateTime,omitempty"`
	Email              *string                    `json:"email,omitempty"`
	WebURL             *string                    `json:"webUrl,omitempty"`
	IsArchived         *bool                      `json:"isArchived,omitempty"`
	LayoutType         *string                    `json:"layoutType,omitempty"`
	ModerationSettings json.RawMessage            `json:"moderationSettings,omitempty"`
	Summary            json.RawMessage            `json:"summary,omitempty"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (c *Channel) UnmarshalJSON(data []byte) error {
	type snapshot Channel
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*c = Channel(value)
	return nil
}

// HostTeamID reads the documented tenant-prefixed OData identity used by
// all/incoming channel projections. It never follows that URL or guesses that
// the recipient team is the host. Unknown/missing shapes return ErrProjection.
func (c Channel) HostTeamID() (string, error) {
	if c.ODataID == nil {
		return "", ErrProjection
	}
	u, err := url.Parse(*c.ODataID)
	if err != nil || !u.IsAbs() || u.Host == "" || u.User != nil || u.Opaque != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", ErrProjection
	}
	parts := strings.Split(strings.TrimPrefix(u.EscapedPath(), "/"), "/")
	if len(parts) != 7 || parts[0] != "beta" || parts[1] != "tenants" || parts[2] == "" || parts[3] != "teams" || parts[5] != "channels" {
		return "", ErrProjection
	}
	host, hostErr := url.PathUnescape(parts[4])
	id, idErr := url.PathUnescape(parts[6])
	if hostErr != nil || idErr != nil || !directoryID.MatchString(host) || id != c.ID {
		return "", ErrProjection
	}
	return host, nil
}

// InitialSharedTeam identifies a team in the documented sharedWithTeams input
// of channel creation. It is a write reference, not a sharing relationship snapshot.
type InitialSharedTeam struct {
	ID string `json:"id"`
}

// CreateRequest contains only the established channel creation properties.
// Empty MembershipType selects standard. OwnerUserID is an existing Entra
// directory UUID, not a membership ID or an unresolved UPN.
type CreateRequest struct {
	DisplayName    string
	Description    *string
	MembershipType MembershipType
	OwnerUserID    *string
	// SharedWithTeams is an optional initial-share collection for a shared
	// channel. Nil omits the field; an explicitly empty slice sends []. A
	// pointer to a nil slice is rejected rather than sending undocumented null.
	SharedWithTeams *[]InitialSharedTeam
}

// UpdateRequest preserves omission versus explicitly empty strings. Membership
// type and response-only fields cannot accidentally become update options.
type UpdateRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ReadOptions selects properties and optionally requests unknown enum members.
type ReadOptions struct {
	Select                    []string
	IncludeUnknownEnumMembers bool
}

// Collection distinguishes hosted, all and incoming channel resources.
type Collection string

const (
	Hosted   Collection = "channels"
	All      Collection = "allChannels"
	Incoming Collection = "incomingChannels"
)

// ListOptions selects an explicit channel collection. Zero selects hosted.
// MembershipType filtering applies to hosted/all, as in the pinned commands.
type ListOptions struct {
	ReadOptions
	Collection     Collection
	MembershipType *MembershipType
}

// ListPage is a single response. NextLink is untrusted opaque metadata and does
// not imply a complete list. Use a bounded pager to traverse it safely.
type ListPage struct {
	Items     []Channel
	NextLink  string
	RawFields map[string]json.RawMessage
}
