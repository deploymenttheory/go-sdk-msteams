// Package sharedchannel manages native shared-with-team relationships.
package sharedchannel

import (
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// SharedWithTeam is the relationship snapshot, preserving optional/future data.
// ID is the shared-with team identifier, not the host team or channel ID.
type SharedWithTeam struct {
	ID          string                     `json:"id"`
	DisplayName *string                    `json:"displayName,omitempty"`
	TenantID    *string                    `json:"tenantId,omitempty"`
	IsHostTeam  *bool                      `json:"isHostTeam,omitempty"`
	Type        *string                    `json:"@odata.type,omitempty"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *SharedWithTeam) UnmarshalJSON(data []byte) error {
	type snapshot SharedWithTeam
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = SharedWithTeam(value)
	return nil
}

// ListPage is one relationship page. NextLink is untrusted opaque metadata.
type ListPage struct {
	Items     []SharedWithTeam
	NextLink  string
	RawFields map[string]json.RawMessage
}
