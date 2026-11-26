package teamsguestcallingconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_guest_calling_configuration resources these acceptance tests change.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"guest-calling-configuration": guestCallingConfigurationCleanup{sdk.TeamsAdminAPI.TeamsGuestCallingConfigurations},
	}
}
