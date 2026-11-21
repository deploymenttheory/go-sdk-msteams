// Package group models the Microsoft 365 groups that back teams. Group lifecycle
// and teamification are separate operations; creating a group does not create a team.
package group

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Visibility preserves Graph's group strings, including future read values.
// These values differ in casing from the native team visibility enumeration.
type Visibility string

const (
	Public           Visibility = "Public"
	Private          Visibility = "Private"
	HiddenMembership Visibility = "HiddenMembership"
)

// Properties is the shared writable subset used by the pinned team commands.
// Nil omits a property; a non-nil empty string is retained as caller intent.
type Properties struct {
	DisplayName    *string     `json:"displayName,omitempty"`
	Description    *string     `json:"description,omitempty"`
	Visibility     *Visibility `json:"visibility,omitempty"`
	MailNickname   *string     `json:"mailNickname,omitempty"`
	Classification *string     `json:"classification,omitempty"`
}

// UpdateRequest patches backing-group properties without team settings or
// read-only fields. It preserves omitted and explicitly empty properties.
type UpdateRequest Properties

// Group is a native read snapshot. Omitted/null booleans remain unknown, and
// RawFields retains complete read data, including future extension properties.
// Neither provisioning flags nor group creation alone prove a team exists.
type Group struct {
	Properties
	ID                          string                     `json:"id"`
	GroupTypes                  []string                   `json:"groupTypes,omitempty"`
	MailEnabled                 *bool                      `json:"mailEnabled,omitempty"`
	SecurityEnabled             *bool                      `json:"securityEnabled,omitempty"`
	Mail                        *string                    `json:"mail,omitempty"`
	CreatedDateTime             *time.Time                 `json:"createdDateTime,omitempty"`
	ResourceProvisioningOptions []string                   `json:"resourceProvisioningOptions,omitempty"`
	RawFields                   map[string]json.RawMessage `json:"-"`
}

func (g *Group) UnmarshalJSON(data []byte) error {
	type snapshot Group
	var value snapshot
	if json.Unmarshal(data, &value) != nil || value.ID == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*g = Group(value)
	return nil
}

// CreateRequest is a Go service input, not a wire body. The resource operation
// must derive member/owner bindings from its selected Graph cloud and user
// directory IDs. No arbitrary reference URL is accepted. CreationOptions and
// ResourceBehaviorOptions are explicit inputs for the pinned PowerShell
// compositions; ordinary native construction does not add undocumented flags.
type CreateRequest struct {
	Properties
	GroupTypes              []string
	MailEnabled             bool
	SecurityEnabled         bool
	OwnerIDs                []string
	MemberIDs               []string
	CreationOptions         *[]string
	ResourceBehaviorOptions *[]string
	// Extensions contains explicitly supplied directory-extension properties,
	// including the captured education marker. Read snapshots are not resent.
	Extensions map[string]json.RawMessage
}

// NewMicrosoft365Group constructs the standard native Unified-group inputs
// without credentials or HTTP. Visibility remains omitted; the PowerShell
// composition must choose its separate Private/education defaults explicitly.
func NewMicrosoft365Group(displayName, mailNickname string) (CreateRequest, error) {
	if strings.TrimSpace(displayName) == "" || strings.TrimSpace(mailNickname) == "" {
		return CreateRequest{}, client.ErrRequest
	}
	return CreateRequest{
		Properties: Properties{DisplayName: &displayName, MailNickname: &mailNickname},
		GroupTypes: []string{"Unified"}, MailEnabled: true, SecurityEnabled: false,
	}, nil
}
