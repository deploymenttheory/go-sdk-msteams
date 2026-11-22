package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
	"github.com/go-resty/resty/v2"
)

func templatePermissionProjection(v *permission.Policy) map[string]any {
	templates := make([]map[string]*string, len(v.HiddenTemplates))
	for i, item := range v.HiddenTemplates {
		if item != nil {
			templates[i] = map[string]*string{"Id": item.ID, "Identity": item.Identity}
		}
	}
	return map[string]any{"Description": v.Description, "HiddenTemplates": templates}
}
func templatePermissionRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: templatePermissionProjection(v), Fields: v.RawFields}, r, nil
	}
}
func TestAccTemplatePermissionPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "template-permission", "teams_template_permission_policy", templatePermissionRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: templatePermissionProjection(v), Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
func sipDevicesRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"BulkSignIn": v.BulkSignIn}, Fields: v.RawFields}, r, nil
	}
}
func TestAccSipDevicesConfigurationReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "sip-devices", "teams_sip_devices_configuration", sipDevicesRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"BulkSignIn": v.BulkSignIn}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
