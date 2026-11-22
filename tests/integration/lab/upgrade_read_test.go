package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func TestAccUpgradeConfigurationReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "upgrade-configuration", "teams_upgrade_configuration", func(sdk *msteams.Client) compiledScalarRead {
		return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
			v, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.GetByIdentity(ctx, "Global")
			if e != nil || v == nil {
				return nil, r, e
			}
			return &compiledScalarView{v.Identity, map[string]any{"DownloadTeams": v.DownloadTeams, "SfBMeetingJoinUx": v.SfBMeetingJoinUx, "BlockLegacyAuthorization": v.BlockLegacyAuthorization}, v.RawFields}, r, nil
		}
	}, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"DownloadTeams": v.DownloadTeams, "SfBMeetingJoinUx": v.SfBMeetingJoinUx, "BlockLegacyAuthorization": v.BlockLegacyAuthorization}, v.RawFields})
			}
		}
		return out, r, e
	})
}
func TestAccUpgradePolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "upgrade-policy", "teams_upgrade_policy", func(sdk *msteams.Client) compiledScalarRead {
		return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
			v, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.GetByIdentity(ctx, "Global")
			if e != nil || v == nil {
				return nil, r, e
			}
			return &compiledScalarView{&v.Identity, map[string]any{"Description": v.Description, "Mode": v.Mode, "NotifySfbUsers": v.NotifySfbUsers, "Action": v.Action}, v.RawFields}, r, nil
		}
	}, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			out = append(out, &compiledScalarView{&v.Identity, map[string]any{"Description": v.Description, "Mode": v.Mode, "NotifySfbUsers": v.NotifySfbUsers, "Action": v.Action}, v.RawFields})
		}
		return out, r, e
	})
}
