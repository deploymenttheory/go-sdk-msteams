package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func mobilityRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsMobilityPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "IPAudioMobileMode": v.IPAudioMobileMode, "IPVideoMobileMode": v.IPVideoMobileMode, "MobileDialerPreference": v.MobileDialerPreference, "LinksInTeams": v.LinksInTeams}, Fields: v.RawFields}, r, nil
	}
}
func multiTenantRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"CopilotFromHomeTenant": v.CopilotFromHomeTenant}, Fields: v.RawFields}, r, nil
	}
}

func TestAccMobilityPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "mobility", "teams_mobility_policy", mobilityRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsMobilityPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "IPAudioMobileMode": v.IPAudioMobileMode, "IPVideoMobileMode": v.IPVideoMobileMode, "MobileDialerPreference": v.MobileDialerPreference, "LinksInTeams": v.LinksInTeams}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
func TestAccMultiTenantOrganizationConfigurationReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "multi-tenant", "teams_multi_tenant_organization_configuration", multiTenantRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"CopilotFromHomeTenant": v.CopilotFromHomeTenant}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
