package externalaccesspolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// external_access_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"external-access-policy": externalPolicyCleanup{sdk.TeamsAdminAPI.ExternalAccessPolicies},
	}
}
