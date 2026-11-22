package teamstenantabuseconfigurationacc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccTenantAbuseConfigurationAuthorizationDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	if !labkit.OracleEnabled(t) {
		t.Fatal("PowerShell comparison is required for tenant access outcomes")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.GetByIdentity(ctx, "Global")
	labkit.LogAccessResponse(t, response, err)
	var api *client.APIError
	if value != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) || errors.Is(err, auth.ErrTenantCapabilityUnavailable) || !errors.As(err, &api) || !slices.Equal(api.ServiceErrorCodes, []int{40301}) {
		t.Fatal("expected tenant-abuse authorization denial changed; successful tenant-abuse reads need separate qualification")
	}
	labkit.WriteResponseFixture(t, "teams_tenant_abuse_configuration", "get_global", response, nil)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	labkit.AccessOutcomeOracle(t, ctx, tenant, "tenant-abuse", "Get-CsTeamsTenantAbuseConfiguration", "forbidden")
	t.Log("confirmed authorization denial: HTTP 403, PolicyRP 40301; PowerShell forbidden response matched")
}
