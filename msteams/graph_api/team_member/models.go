// Package teammember models native Graph conversation memberships in a team.
// Backing-group member/owner relationships use separate directory identities.
package teammember

import (
	"encoding/json"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// MembershipID is an opaque conversationMember identity returned by Graph.
// Never decode it or synthesize it from an Entra directory-user identifier.
type MembershipID string

// DirectoryUserID identifies an Entra user. It cannot address a native
// conversation membership, even when a read response contains both identities.
type DirectoryUserID string

// Role is a qualifier, not a primary membership category. An ordinary member's
// roles collection is empty; it does not contain the string "member". Unknown
// future read values remain representable.
type Role string

const (
	Owner Role = "owner"
	Guest Role = "guest"
)

// Member preserves the native resource and AAD subtype fields. Null or absent
// user fields do not imply that the member belongs to this client's tenant.
// RawFields retains the complete object, including discriminator, presence and
// future properties. It is never a mutation payload.
type Member struct {
	ID                          MembershipID               `json:"id"`
	Type                        *string                    `json:"@odata.type,omitempty"`
	Roles                       []Role                     `json:"roles,omitempty"`
	DisplayName                 *string                    `json:"displayName,omitempty"`
	UserID                      *DirectoryUserID           `json:"userId,omitempty"`
	TenantID                    *string                    `json:"tenantId,omitempty"`
	Email                       *string                    `json:"email,omitempty"`
	VisibleHistoryStartDateTime *time.Time                 `json:"visibleHistoryStartDateTime,omitempty"`
	OriginalSourceMembershipURL *string                    `json:"@microsoft.graph.originalSourceMembershipUrl,omitempty"`
	RawFields                   map[string]json.RawMessage `json:"-"`
}

func (m *Member) UnmarshalJSON(data []byte) error {
	type snapshot Member
	var value snapshot
	if json.Unmarshal(data, &value) != nil || value.ID == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*m = Member(value)
	return nil
}

// CreateRequest is a Go service input, not an API body. The resource service
// must build the user binding from its selected Graph cloud. Nil Roles omits
// the property; a pointer to an empty slice explicitly requests an ordinary
// member. Write validation belongs to the resource operation.
type CreateRequest struct {
	UserID DirectoryUserID
	Roles  *[]Role
}

// UpdateRequest contains only writable team-member properties. Nil Roles omits
// the property, an explicit [] demotes an owner, and ["owner"] promotes a member.
// A pointer to a nil slice represents JSON null, which is distinct from demotion
// and must be rejected by the resource operation. Guest is a server-assigned
// qualifier and cannot be set or changed through this patch.
type UpdateRequest struct {
	Roles *[]Role `json:"roles,omitempty"`
}
