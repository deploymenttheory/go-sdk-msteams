// Package conversationrole manages conditional compiled conversation-role contracts.
package conversationrole

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// ErrInput identifies a value rejected before a management write.
var ErrInput = policy.ErrInput

// Role retains typed properties and the complete owned response snapshot.
// Omitted CLR DateTime fields project to the unspecified minimum; RawFields
// separately retains omission. Explicit null for these value types is invalid.
type Role struct {
	Identity     *string                    `json:"Identity"`
	Permissions  *Permissions               `json:"Permissions"`
	ID           *string                    `json:"Id"`
	Name         *string                    `json:"Name"`
	IsAssignable *bool                      `json:"IsAssignable"`
	CreatedAt    DateTime                   `json:"CreatedAt"`
	UpdatedAt    DateTime                   `json:"UpdatedAt"`
	RawFields    map[string]json.RawMessage `json:"-"`
	RawJSON      json.RawMessage            `json:"-"`
}

var guidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func validModelID(value string) bool { return guidPattern.MatchString(strings.TrimSuffix(value, "\n")) }

func (v *Role) UnmarshalJSON(raw []byte) error {
	type snapshot Role
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	if next.ID != nil && !validModelID(*next.ID) {
		return client.ErrDecode
	}
	if bytes.Equal(bytes.TrimSpace(next.RawFields["Permissions"]), []byte("null")) {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Role(next)
	return nil
}
func (v Role) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Role
	return json.Marshal(snapshot(v))
}
func (Role) String() string     { return "conditional conversation role (redacted)" }
func (v Role) GoString() string { return v.String() }

// Permissions is the local permission DTO. The pinned management serializer
// projects Chat and Calling away; use an inspectable write projection to see it.
type Permissions struct {
	Identity  *string                    `json:"Identity"`
	Chat      *ChatPermissions           `json:"Chat"`
	Calling   *CallingPermissions        `json:"Calling"`
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

func (v *Permissions) UnmarshalJSON(raw []byte) error {
	type snapshot Permissions
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	for _, key := range []string{"Chat", "Calling"} {
		if bytes.Equal(bytes.TrimSpace(next.RawFields[key]), []byte("null")) {
			return client.ErrDecode
		}
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Permissions(next)
	return nil
}
func (v Permissions) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Permissions
	return json.Marshal(snapshot(v))
}
func (Permissions) String() string     { return "conversation permissions (redacted)" }
func (v Permissions) GoString() string { return v.String() }

// NewPermissions models the local permission constructor, without authentication
// or a claim that the resulting flags persist through a management write.
func NewPermissions(chat *ChatPermissions, calling *CallingPermissions) (*Permissions, error) {
	if chat == nil || calling == nil {
		return nil, ErrInput
	}
	return &Permissions{Chat: chat, Calling: calling}, nil
}

// ChatPermissionsValues supplies nullable flags to the pure local constructor.
type ChatPermissionsValues struct {
	AddParticipants            *bool `json:"AddParticipants"`
	RemoveSelf                 *bool `json:"RemoveSelf"`
	RemoveParticipants         *bool `json:"RemoveParticipants"`
	ListParticipants           *bool `json:"ListParticipants"`
	GetOwnDetails              *bool `json:"GetOwnDetails"`
	GetOthersDetails           *bool `json:"GetOthersDetails"`
	UpdateOwnDetails           *bool `json:"UpdateOwnDetails"`
	UpdateOthersDetails        *bool `json:"UpdateOthersDetails"`
	ShareHistory               *bool `json:"ShareHistory"`
	GetChatThreadProperties    *bool `json:"GetChatThreadProperties"`
	UpdateChatThreadProperties *bool `json:"UpdateChatThreadProperties"`
	DeleteChatThread           *bool `json:"DeleteChatThread"`
	SendMessage                *bool `json:"SendMessage"`
	ReceiveMessage             *bool `json:"ReceiveMessage"`
	EditOwnMessage             *bool `json:"EditOwnMessage"`
	EditOthersMessage          *bool `json:"EditOthersMessage"`
	DeleteOwnMessage           *bool `json:"DeleteOwnMessage"`
	DeleteOthersMessage        *bool `json:"DeleteOthersMessage"`
}

func (ChatPermissionsValues) String() string     { return "conversation permission values (redacted)" }
func (v ChatPermissionsValues) GoString() string { return v.String() }

// ChatPermissions retains null, false and future response properties.
type ChatPermissions struct {
	Identity *string `json:"Identity"`
	ChatPermissionsValues
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

func (v *ChatPermissions) UnmarshalJSON(raw []byte) error {
	type snapshot ChatPermissions
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = ChatPermissions(next)
	return nil
}
func (v ChatPermissions) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot ChatPermissions
	return json.Marshal(snapshot(v))
}
func (ChatPermissions) String() string     { return "conversation permissions (redacted)" }
func (v ChatPermissions) GoString() string { return v.String() }

// NewChatPermissions constructs local values only. Omission and explicit null both
// project to CLR nullable null; explicit false remains a bound scalar value.
func NewChatPermissions(values ChatPermissionsValues) *ChatPermissions {
	values.AddParticipants = copyBool(values.AddParticipants)
	values.RemoveSelf = copyBool(values.RemoveSelf)
	values.RemoveParticipants = copyBool(values.RemoveParticipants)
	values.ListParticipants = copyBool(values.ListParticipants)
	values.GetOwnDetails = copyBool(values.GetOwnDetails)
	values.GetOthersDetails = copyBool(values.GetOthersDetails)
	values.UpdateOwnDetails = copyBool(values.UpdateOwnDetails)
	values.UpdateOthersDetails = copyBool(values.UpdateOthersDetails)
	values.ShareHistory = copyBool(values.ShareHistory)
	values.GetChatThreadProperties = copyBool(values.GetChatThreadProperties)
	values.UpdateChatThreadProperties = copyBool(values.UpdateChatThreadProperties)
	values.DeleteChatThread = copyBool(values.DeleteChatThread)
	values.SendMessage = copyBool(values.SendMessage)
	values.ReceiveMessage = copyBool(values.ReceiveMessage)
	values.EditOwnMessage = copyBool(values.EditOwnMessage)
	values.EditOthersMessage = copyBool(values.EditOthersMessage)
	values.DeleteOwnMessage = copyBool(values.DeleteOwnMessage)
	values.DeleteOthersMessage = copyBool(values.DeleteOthersMessage)
	return &ChatPermissions{ChatPermissionsValues: values}
}

// CallingPermissionsValues supplies nullable flags to the pure local constructor.
type CallingPermissionsValues struct {
	AddParticipants       *bool `json:"AddParticipants"`
	AddPhoneNumbers       *bool `json:"AddPhoneNumbers"`
	RemoveParticipants    *bool `json:"RemoveParticipants"`
	SendVideo             *bool `json:"SendVideo"`
	RestrictOthersVideo   *bool `json:"RestrictOthersVideo"`
	SendAudio             *bool `json:"SendAudio"`
	RestrictOthersAudio   *bool `json:"RestrictOthersAudio"`
	ShareScreen           *bool `json:"ShareScreen"`
	MuteSelf              *bool `json:"MuteSelf"`
	UnmuteSelf            *bool `json:"UnmuteSelf"`
	MuteOthers            *bool `json:"MuteOthers"`
	SpotlightParticipants *bool `json:"SpotlightParticipants"`
	RemoveSpotlights      *bool `json:"RemoveSpotlights"`
}

func (CallingPermissionsValues) String() string     { return "conversation permission values (redacted)" }
func (v CallingPermissionsValues) GoString() string { return v.String() }

// CallingPermissions retains null, false and future response properties.
type CallingPermissions struct {
	Identity *string `json:"Identity"`
	CallingPermissionsValues
	RawFields map[string]json.RawMessage `json:"-"`
	RawJSON   json.RawMessage            `json:"-"`
}

func (v *CallingPermissions) UnmarshalJSON(raw []byte) error {
	type snapshot CallingPermissions
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = CallingPermissions(next)
	return nil
}
func (v CallingPermissions) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot CallingPermissions
	return json.Marshal(snapshot(v))
}
func (CallingPermissions) String() string     { return "conversation permissions (redacted)" }
func (v CallingPermissions) GoString() string { return v.String() }

// NewCallingPermissions constructs local values only. Omission and explicit null both
// project to CLR nullable null; explicit false remains a bound scalar value.
func NewCallingPermissions(values CallingPermissionsValues) *CallingPermissions {
	values.AddParticipants = copyBool(values.AddParticipants)
	values.AddPhoneNumbers = copyBool(values.AddPhoneNumbers)
	values.RemoveParticipants = copyBool(values.RemoveParticipants)
	values.SendVideo = copyBool(values.SendVideo)
	values.RestrictOthersVideo = copyBool(values.RestrictOthersVideo)
	values.SendAudio = copyBool(values.SendAudio)
	values.RestrictOthersAudio = copyBool(values.RestrictOthersAudio)
	values.ShareScreen = copyBool(values.ShareScreen)
	values.MuteSelf = copyBool(values.MuteSelf)
	values.UnmuteSelf = copyBool(values.UnmuteSelf)
	values.MuteOthers = copyBool(values.MuteOthers)
	values.SpotlightParticipants = copyBool(values.SpotlightParticipants)
	values.RemoveSpotlights = copyBool(values.RemoveSpotlights)
	return &CallingPermissions{CallingPermissionsValues: values}
}

func copyBool(v *bool) *bool {
	if v == nil {
		return nil
	}
	next := *v
	return &next
}
