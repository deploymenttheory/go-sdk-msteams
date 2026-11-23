package teamscallholdpolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_call_hold_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"call-hold-policy": callHoldPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallHoldPolicies},
	}
}
