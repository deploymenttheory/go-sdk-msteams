package teamsupdatemanagementpolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func updateManagementRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "AllowManagedUpdates": v.AllowManagedUpdates, "AllowPreview": v.AllowPreview, "UpdateDayOfWeek": v.UpdateDayOfWeek, "UpdateTime": v.UpdateTime, "UpdateTimeOfDay": v.UpdateTimeOfDay, "AllowPublicPreview": v.AllowPublicPreview, "UseNewTeamsClient": v.UseNewTeamsClient, "BlockLegacyAuthorization": v.BlockLegacyAuthorization, "DisabledInProductMessages": v.DisabledInProductMessages}, Fields: v.RawFields}, r, nil
	}
}

func TestAccUpdateManagementPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "update-management", "teams_update_management_policy", updateManagementRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "AllowManagedUpdates": v.AllowManagedUpdates, "AllowPreview": v.AllowPreview, "UpdateDayOfWeek": v.UpdateDayOfWeek, "UpdateTime": v.UpdateTime, "UpdateTimeOfDay": v.UpdateTimeOfDay, "AllowPublicPreview": v.AllowPublicPreview, "UseNewTeamsClient": v.UseNewTeamsClient, "BlockLegacyAuthorization": v.BlockLegacyAuthorization, "DisabledInProductMessages": v.DisabledInProductMessages}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
