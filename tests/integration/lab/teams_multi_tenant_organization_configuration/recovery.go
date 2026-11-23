package teamsmultitenantorganizationconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_multi_tenant_organization_configuration resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"multi-tenant-configuration": multiTenantConfigurationCleanup{sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations},
	}
}
