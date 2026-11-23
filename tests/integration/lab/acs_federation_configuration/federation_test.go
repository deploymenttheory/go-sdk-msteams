package acsfederationconfigurationacc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

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
