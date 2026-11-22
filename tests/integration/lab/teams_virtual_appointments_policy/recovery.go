package teamsvirtualappointmentspolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_virtual_appointments_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"virtual-appointments-policy": virtualAppointmentsPolicyCleanup{sdk.TeamsAdminAPI.TeamsVirtualAppointmentsPolicies},
	}
}
