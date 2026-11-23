package tenantfederationconfigurationacc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

// TestAccTenantFederationConfigurationReadOnly reads the Global tenant federation configuration and its filtered list.
func TestAccTenantFederationConfigurationReadOnly(t *testing.T) {
	labkit.RunFederationGlobal(t, "tenant", func(ctx context.Context, sdk *msteams.Client) (*string, map[string]json.RawMessage, []map[string]json.RawMessage, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.GetByIdentity(ctx, "Global")
		labkit.WriteResponseFixture(t, "tenant_federation_configuration", "get_global", r, nil)
		if e != nil || v == nil {
			return nil, nil, nil, r, e
		}
		values, listResponse, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.ListFiltered(ctx, "Global")
		labkit.WriteResponseFixture(t, "tenant_federation_configuration", "list", listResponse, nil)
		if e != nil {
			return nil, nil, nil, listResponse, e
		}
		var listed []map[string]json.RawMessage
		for _, v := range values {
			listed = append(listed, v.RawFields)
		}
		return &v.Identity, v.RawFields, listed, r, nil
	})
}
