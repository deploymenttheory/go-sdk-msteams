// Package teamsmessagingpolicy manages compiled Teams messaging policy.
package teamsmessagingpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// Strings, Booleans and numeric settings retain source nullability.
type Policy struct {
	Identity                                      *string                    `json:"Identity"`
	Description                                   *string                    `json:"Description"`
	AllowUrlPreviews                              *bool                      `json:"AllowUrlPreviews"`
	AllowOwnerDeleteMessage                       *bool                      `json:"AllowOwnerDeleteMessage"`
	AllowUserEditMessage                          *bool                      `json:"AllowUserEditMessage"`
	AllowUserDeleteMessage                        *bool                      `json:"AllowUserDeleteMessage"`
	UsersCanDeleteBotMessages                     *bool                      `json:"UsersCanDeleteBotMessages"`
	AllowUserDeleteChat                           *bool                      `json:"AllowUserDeleteChat"`
	AllowUserChat                                 *bool                      `json:"AllowUserChat"`
	AllowRemoveUser                               *bool                      `json:"AllowRemoveUser"`
	AllowGiphy                                    *bool                      `json:"AllowGiphy"`
	GiphyRatingType                               *string                    `json:"GiphyRatingType"`
	AllowGiphyDisplay                             *bool                      `json:"AllowGiphyDisplay"`
	AllowPasteInternetImage                       *bool                      `json:"AllowPasteInternetImage"`
	AllowMemes                                    *bool                      `json:"AllowMemes"`
	AllowImmersiveReader                          *bool                      `json:"AllowImmersiveReader"`
	AllowStickers                                 *bool                      `json:"AllowStickers"`
	AllowUserTranslation                          *bool                      `json:"AllowUserTranslation"`
	ReadReceiptsEnabledType                       *string                    `json:"ReadReceiptsEnabledType"`
	AllowPriorityMessages                         *bool                      `json:"AllowPriorityMessages"`
	AllowSmartReply                               *bool                      `json:"AllowSmartReply"`
	AllowSmartCompose                             *bool                      `json:"AllowSmartCompose"`
	ChannelsInChatListEnabledType                 *string                    `json:"ChannelsInChatListEnabledType"`
	AudioMessageEnabledType                       *string                    `json:"AudioMessageEnabledType"`
	ChatPermissionRole                            *string                    `json:"ChatPermissionRole"`
	AllowFullChatPermissionUserToDeleteAnyMessage *bool                      `json:"AllowFullChatPermissionUserToDeleteAnyMessage"`
	AllowFluidCollaborate                         *bool                      `json:"AllowFluidCollaborate"`
	AllowVideoMessages                            *bool                      `json:"AllowVideoMessages"`
	AllowCommunicationComplianceEndUserReporting  *bool                      `json:"AllowCommunicationComplianceEndUserReporting"`
	AllowChatWithGroup                            *bool                      `json:"AllowChatWithGroup"`
	AllowSecurityEndUserReporting                 *bool                      `json:"AllowSecurityEndUserReporting"`
	InOrganizationChatControl                     *string                    `json:"InOrganizationChatControl"`
	AllowGroupChatJoinLinks                       *bool                      `json:"AllowGroupChatJoinLinks"`
	CreateCustomEmojis                            *bool                      `json:"CreateCustomEmojis"`
	UseB2BInvitesToAddExternalUsers               *bool                      `json:"UseB2BInvitesToAddExternalUsers"`
	AllowProactiveSummaries                       *bool                      `json:"AllowProactiveSummaries"`
	DeleteCustomEmojis                            *bool                      `json:"DeleteCustomEmojis"`
	AutoShareFilesInExternalChats                 *string                    `json:"AutoShareFilesInExternalChats"`
	DesignerForBackgroundsAndImages               *string                    `json:"DesignerForBackgroundsAndImages"`
	AllowCustomGroupChatAvatars                   *bool                      `json:"AllowCustomGroupChatAvatars"`
	RawFields                                     map[string]json.RawMessage `json:"-"`
	RawJSON                                       json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "messaging policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// Source strings remain opaque; only evidenced setter constraints are checked.
type Settings struct {
	Description                                   policy.Field[string] `json:"Description,omitzero"`
	AllowUrlPreviews                              policy.Field[bool]   `json:"AllowUrlPreviews,omitzero"`
	AllowOwnerDeleteMessage                       policy.Field[bool]   `json:"AllowOwnerDeleteMessage,omitzero"`
	AllowUserEditMessage                          policy.Field[bool]   `json:"AllowUserEditMessage,omitzero"`
	AllowUserDeleteMessage                        policy.Field[bool]   `json:"AllowUserDeleteMessage,omitzero"`
	UsersCanDeleteBotMessages                     policy.Field[bool]   `json:"UsersCanDeleteBotMessages,omitzero"`
	AllowUserDeleteChat                           policy.Field[bool]   `json:"AllowUserDeleteChat,omitzero"`
	AllowUserChat                                 policy.Field[bool]   `json:"AllowUserChat,omitzero"`
	AllowRemoveUser                               policy.Field[bool]   `json:"AllowRemoveUser,omitzero"`
	AllowGiphy                                    policy.Field[bool]   `json:"AllowGiphy,omitzero"`
	GiphyRatingType                               policy.Field[string] `json:"GiphyRatingType,omitzero"`
	AllowGiphyDisplay                             policy.Field[bool]   `json:"AllowGiphyDisplay,omitzero"`
	AllowPasteInternetImage                       policy.Field[bool]   `json:"AllowPasteInternetImage,omitzero"`
	AllowMemes                                    policy.Field[bool]   `json:"AllowMemes,omitzero"`
	AllowImmersiveReader                          policy.Field[bool]   `json:"AllowImmersiveReader,omitzero"`
	AllowStickers                                 policy.Field[bool]   `json:"AllowStickers,omitzero"`
	AllowUserTranslation                          policy.Field[bool]   `json:"AllowUserTranslation,omitzero"`
	ReadReceiptsEnabledType                       policy.Field[string] `json:"ReadReceiptsEnabledType,omitzero"`
	AllowPriorityMessages                         policy.Field[bool]   `json:"AllowPriorityMessages,omitzero"`
	AllowSmartReply                               policy.Field[bool]   `json:"AllowSmartReply,omitzero"`
	AllowSmartCompose                             policy.Field[bool]   `json:"AllowSmartCompose,omitzero"`
	ChannelsInChatListEnabledType                 policy.Field[string] `json:"ChannelsInChatListEnabledType,omitzero"`
	AudioMessageEnabledType                       policy.Field[string] `json:"AudioMessageEnabledType,omitzero"`
	ChatPermissionRole                            policy.Field[string] `json:"ChatPermissionRole,omitzero"`
	AllowFullChatPermissionUserToDeleteAnyMessage policy.Field[bool]   `json:"AllowFullChatPermissionUserToDeleteAnyMessage,omitzero"`
	AllowFluidCollaborate                         policy.Field[bool]   `json:"AllowFluidCollaborate,omitzero"`
	AllowVideoMessages                            policy.Field[bool]   `json:"AllowVideoMessages,omitzero"`
	AllowCommunicationComplianceEndUserReporting  policy.Field[bool]   `json:"AllowCommunicationComplianceEndUserReporting,omitzero"`
	AllowChatWithGroup                            policy.Field[bool]   `json:"AllowChatWithGroup,omitzero"`
	AllowSecurityEndUserReporting                 policy.Field[bool]   `json:"AllowSecurityEndUserReporting,omitzero"`
	InOrganizationChatControl                     policy.Field[string] `json:"InOrganizationChatControl,omitzero"`
	AllowGroupChatJoinLinks                       policy.Field[bool]   `json:"AllowGroupChatJoinLinks,omitzero"`
	CreateCustomEmojis                            policy.Field[bool]   `json:"CreateCustomEmojis,omitzero"`
	UseB2BInvitesToAddExternalUsers               policy.Field[bool]   `json:"UseB2BInvitesToAddExternalUsers,omitzero"`
	AllowProactiveSummaries                       policy.Field[bool]   `json:"AllowProactiveSummaries,omitzero"`
	DeleteCustomEmojis                            policy.Field[bool]   `json:"DeleteCustomEmojis,omitzero"`
	AutoShareFilesInExternalChats                 policy.Field[string] `json:"AutoShareFilesInExternalChats,omitzero"`
	DesignerForBackgroundsAndImages               policy.Field[string] `json:"DesignerForBackgroundsAndImages,omitzero"`
	AllowCustomGroupChatAvatars                   policy.Field[bool]   `json:"AllowCustomGroupChatAvatars,omitzero"`
}

// CreateRequest permits no settings, matching the compiled New processing.
type CreateRequest struct {
	Settings
	policy.Controls
}

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "messaging policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "messaging policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
