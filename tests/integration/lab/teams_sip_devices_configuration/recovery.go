package teamssipdevicesconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_sip_devices_configuration resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"sip-devices-configuration": sipDevicesConfigurationCleanup{sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations},
	}
}
