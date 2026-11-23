package mainlineattendantsacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// mainline_attendants resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	service := sdk.TeamsAdminAPI.MainlineAttendants
	return map[string]lab.Handler{
		spamKind:    spamCleanup(service),
		bookingKind: bookingCleanup(service),
		qaKind:      qaCleanup(service),
	}
}
