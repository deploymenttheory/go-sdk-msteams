package teamsbyodanddeskspolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_byod_and_desks_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"byod-and-desks-policy": byodAndDesksPolicyCleanup{sdk.TeamsAdminAPI.TeamsBYODAndDesksPolicies},
	}
}
