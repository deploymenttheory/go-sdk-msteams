package lab_test

import (
	"context"
	"testing"

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
func templatePermissionRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, templatePermissionProjection(v), v.RawFields}, r, nil
	}
}
func TestAccTemplatePermissionPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "template-permission", "teams_template_permission_policy", templatePermissionRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, templatePermissionProjection(v), v.RawFields})
			}
		}
		return out, r, e
	})
}
func sipDevicesRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"BulkSignIn": v.BulkSignIn}, v.RawFields}, r, nil
	}
}
func TestAccSipDevicesConfigurationReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "sip-devices", "teams_sip_devices_configuration", sipDevicesRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"BulkSignIn": v.BulkSignIn}, v.RawFields})
			}
		}
		return out, r, e
	})
}
