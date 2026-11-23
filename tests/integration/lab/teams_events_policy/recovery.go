package teamseventspolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_events_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"events-policy": eventsPolicyCleanup{sdk.TeamsAdminAPI.TeamsEventsPolicies},
	}
}
