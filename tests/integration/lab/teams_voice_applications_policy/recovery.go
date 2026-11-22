package teamsvoiceapplicationspolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_voice_applications_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"voice-applications-policy": voiceApplicationsPolicyCleanup{sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies},
	}
}
