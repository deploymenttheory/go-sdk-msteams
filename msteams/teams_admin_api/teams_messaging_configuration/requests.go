package teamsmessagingconfiguration

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes omission, explicit null and bound zero values.
type ScalarSettings struct {
	EnableVideoMessageCaptions        policy.Field[bool]      `json:"EnableVideoMessageCaptions,omitzero"`
	EnableInOrganizationChatControl   policy.Field[bool]      `json:"EnableInOrganizationChatControl,omitzero"`
	CustomEmojis                      policy.Field[bool]      `json:"CustomEmojis,omitzero"`
	Storyline                         policy.Field[Storyline] `json:"Storyline,omitzero"`
	Communities                       policy.Field[string]    `json:"Communities,omitzero"`
	MessagingNotes                    policy.Field[string]    `json:"MessagingNotes,omitzero"`
	FileTypeCheck                     policy.Field[string]    `json:"FileTypeCheck,omitzero"`
	UrlReputationCheck                policy.Field[string]    `json:"UrlReputationCheck,omitzero"`
	ContentBasedPhishingCheck         policy.Field[string]    `json:"ContentBasedPhishingCheck,omitzero"`
	ReportIncorrectSecurityDetections policy.Field[string]    `json:"ReportIncorrectSecurityDetections,omitzero"`
	FileTypeCustomizationSetting      policy.Field[string]    `json:"FileTypeCustomizationSetting,omitzero"`
}

// Settings sends a complete replacement array when FileTypeBlockList is nonnil.
type Settings struct {
	ScalarSettings
	FileTypeBlockList *policy.StringList `json:"FileTypeBlockList,omitempty"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	ScalarSettings
	FileTypeBlockList *policy.ListEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "messaging configuration update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
