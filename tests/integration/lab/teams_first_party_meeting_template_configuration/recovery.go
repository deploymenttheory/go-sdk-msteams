package teamsfirstpartymeetingtemplateconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the
// teams_first_party_meeting_template_configuration changes these acceptance tests make.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"first-party-template-config": firstPartyTemplateDescriptionRestore{sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations},
	}
}
