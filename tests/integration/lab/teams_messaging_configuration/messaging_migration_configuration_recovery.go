package teamsmessagingconfigurationacc

import (
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
)

func messagingConfigurationProjection(v *messaging.Configuration) map[string]any {
	return map[string]any{"EnableVideoMessageCaptions": v.EnableVideoMessageCaptions, "EnableInOrganizationChatControl": v.EnableInOrganizationChatControl, "CustomEmojis": v.CustomEmojis, "Storyline": v.Storyline, "Communities": v.Communities, "MessagingNotes": v.MessagingNotes, "FileTypeCheck": v.FileTypeCheck, "UrlReputationCheck": v.UrlReputationCheck, "ContentBasedPhishingCheck": v.ContentBasedPhishingCheck, "ReportIncorrectSecurityDetections": v.ReportIncorrectSecurityDetections, "FileTypeCustomizationSetting": v.FileTypeCustomizationSetting, "FileTypeBlockList": v.FileTypeBlockList}
}
