// Package grouprelationship manages backing-group member/owner references,
// independently of native Teams conversation memberships.
package grouprelationship

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// DirectoryObjectID is an Entra object identity, never a conversationMember ID.
type DirectoryObjectID string

// ReferenceKind selects a known Graph object collection, not a reference URL.
type ReferenceKind string

const (
	DirectoryObjects  ReferenceKind = "directoryObjects"
	Users             ReferenceKind = "users"
	ServicePrincipals ReferenceKind = "servicePrincipals"
)

// Reference supplies a directory ID and optional object kind. With no Kind,
// member references use directoryObjects and owner references use users. Explicit
// DirectoryObjects can preserve the pinned owner's alternate binding form.
type Reference struct {
	ID   DirectoryObjectID
	Kind ReferenceKind
}

// DirectoryObject preserves limited and future directory representations.
// Null userType/UPN/name is unknown, not evidence of ordinary-user membership.
type DirectoryObject struct {
	ID                DirectoryObjectID          `json:"id"`
	Type              *string                    `json:"@odata.type,omitempty"`
	DisplayName       *string                    `json:"displayName,omitempty"`
	UserPrincipalName *string                    `json:"userPrincipalName,omitempty"`
	UserType          *string                    `json:"userType,omitempty"`
	Mail              *string                    `json:"mail,omitempty"`
	AppID             *string                    `json:"appId,omitempty"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (d *DirectoryObject) UnmarshalJSON(data []byte) error {
	type snapshot DirectoryObject
	var value snapshot
	if json.Unmarshal(data, &value) != nil || value.ID == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*d = DirectoryObject(value)
	return nil
}

// ListOptions maps the relationship collection's supported query. Advanced
// query headers/count are explicit; expressions are not parsed or rewritten.
type ListOptions struct {
	Filter              string
	Search              string
	Select              []string
	Expand              []string
	Top                 int
	Count               *bool
	EventualConsistency bool
}

// ListPage is one direct relationship page, not transitive membership. A native
// next link is opaque metadata; its presence means the page is not the full list.
type ListPage struct {
	Items     []DirectoryObject
	NextLink  string
	RawFields map[string]json.RawMessage
}
