package teamsguestmeetingconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_guest_meeting_configuration resources these acceptance tests change.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"guest-meeting-configuration": guestMeetingConfigurationCleanup{sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations},
	}
}
