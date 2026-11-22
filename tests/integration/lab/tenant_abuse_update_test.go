package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func tenantAbuseRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"LowSeatLimit": v.LowSeatLimit, "TrialTenantValidation": v.TrialTenantValidation, "SkipValidation": v.SkipValidation, "CreateThreadThresholdPerSeat": v.CreateThreadThresholdPerSeat, "AddMemberThresholdPerSeat": v.AddMemberThresholdPerSeat, "SendMessageThresholdPerSeat": v.SendMessageThresholdPerSeat, "CreateThreadThresholdForTrialTenants": v.CreateThreadThresholdForTrialTenants, "AddMemberThresholdForTrialTenants": v.AddMemberThresholdForTrialTenants, "SendMessageThresholdForTrialTenants": v.SendMessageThresholdForTrialTenants, "BlockedTenants": v.BlockedTenants}, v.RawFields}, r, nil
	}
}
func TestAccTenantAbuseConfigurationReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "tenant-abuse", "teams_tenant_abuse_configuration", tenantAbuseRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"LowSeatLimit": v.LowSeatLimit, "TrialTenantValidation": v.TrialTenantValidation, "SkipValidation": v.SkipValidation, "CreateThreadThresholdPerSeat": v.CreateThreadThresholdPerSeat, "AddMemberThresholdPerSeat": v.AddMemberThresholdPerSeat, "SendMessageThresholdPerSeat": v.SendMessageThresholdPerSeat, "CreateThreadThresholdForTrialTenants": v.CreateThreadThresholdForTrialTenants, "AddMemberThresholdForTrialTenants": v.AddMemberThresholdForTrialTenants, "SendMessageThresholdForTrialTenants": v.SendMessageThresholdForTrialTenants, "BlockedTenants": v.BlockedTenants}, v.RawFields})
			}
		}
		return out, r, e
	})
}

func updateManagementRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "AllowManagedUpdates": v.AllowManagedUpdates, "AllowPreview": v.AllowPreview, "UpdateDayOfWeek": v.UpdateDayOfWeek, "UpdateTime": v.UpdateTime, "UpdateTimeOfDay": v.UpdateTimeOfDay, "AllowPublicPreview": v.AllowPublicPreview, "UseNewTeamsClient": v.UseNewTeamsClient, "BlockLegacyAuthorization": v.BlockLegacyAuthorization, "DisabledInProductMessages": v.DisabledInProductMessages}, v.RawFields}, r, nil
	}
}
func TestAccUpdateManagementPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "update-management", "teams_update_management_policy", updateManagementRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "AllowManagedUpdates": v.AllowManagedUpdates, "AllowPreview": v.AllowPreview, "UpdateDayOfWeek": v.UpdateDayOfWeek, "UpdateTime": v.UpdateTime, "UpdateTimeOfDay": v.UpdateTimeOfDay, "AllowPublicPreview": v.AllowPublicPreview, "UseNewTeamsClient": v.UseNewTeamsClient, "BlockLegacyAuthorization": v.BlockLegacyAuthorization, "DisabledInProductMessages": v.DisabledInProductMessages}, v.RawFields})
			}
		}
		return out, r, e
	})
}
