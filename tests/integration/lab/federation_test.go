package lab_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

// TestAccExternalAccessPolicyReadOnly reads the Global external access policy and its filtered list.
func TestAccExternalAccessPolicyReadOnly(t *testing.T) {
	labkit.RunFederationGlobal(t, "external-policy", func(ctx context.Context, sdk *msteams.Client) (*string, map[string]json.RawMessage, []map[string]json.RawMessage, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(ctx, "Global")
		labkit.WriteResponseFixture(t, "external_access_policy", "get_global", r, nil)
		if e != nil || v == nil {
			return nil, nil, nil, r, e
		}
		values, listResponse, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.ListFiltered(ctx, "Global")
		labkit.WriteResponseFixture(t, "external_access_policy", "list", listResponse, nil)
		if e != nil {
			return nil, nil, nil, listResponse, e
		}
		var listed []map[string]json.RawMessage
		for _, v := range values {
			if v != nil {
				listed = append(listed, v.RawFields)
			}
		}
		return v.Identity, v.RawFields, listed, r, nil
	})
}

// TestAccACSFederationConfigurationReadOnly reads the Global ACS federation configuration and its filtered list.
func TestAccACSFederationConfigurationReadOnly(t *testing.T) {
	labkit.RunFederationGlobal(t, "acs", func(ctx context.Context, sdk *msteams.Client) (*string, map[string]json.RawMessage, []map[string]json.RawMessage, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.GetByIdentity(ctx, "Global")
		labkit.WriteResponseFixture(t, "acs_federation_configuration", "get_global", r, nil)
		if e != nil || v == nil {
			return nil, nil, nil, r, e
		}
		values, listResponse, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.ListFiltered(ctx, "Global")
		labkit.WriteResponseFixture(t, "acs_federation_configuration", "list", listResponse, nil)
		if e != nil {
			return nil, nil, nil, listResponse, e
		}
		var listed []map[string]json.RawMessage
		for _, v := range values {
			if v != nil {
				listed = append(listed, v.RawFields)
			}
		}
		return v.Identity, v.RawFields, listed, r, nil
	})
}

// TestAccTeamsExternalAccessConfigurationReadOnly reads the Global Teams external access configuration and its filtered list.
func TestAccTeamsExternalAccessConfigurationReadOnly(t *testing.T) {
	labkit.RunFederationGlobal(t, "external-users", func(ctx context.Context, sdk *msteams.Client) (*string, map[string]json.RawMessage, []map[string]json.RawMessage, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.GetByIdentity(ctx, "Global")
		labkit.WriteResponseFixture(t, "teams_external_access_configuration", "get_global", r, nil)
		if e != nil || v == nil {
			return nil, nil, nil, r, e
		}
		values, listResponse, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.ListFiltered(ctx, "Global")
		labkit.WriteResponseFixture(t, "teams_external_access_configuration", "list", listResponse, nil)
		if e != nil {
			return nil, nil, nil, listResponse, e
		}
		var listed []map[string]json.RawMessage
		for _, v := range values {
			if v != nil {
				listed = append(listed, v.RawFields)
			}
		}
		return v.Identity, v.RawFields, listed, r, nil
	})
}

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
