package teamsupgradeconfigurationacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func TestAccUpgradeConfigurationReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "upgrade-configuration", "teams_upgrade_configuration", func(sdk *msteams.Client) labkit.CompiledScalarRead {
		return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
			v, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.GetByIdentity(ctx, "Global")
			if e != nil || v == nil {
				return nil, r, e
			}
			return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"DownloadTeams": v.DownloadTeams, "SfBMeetingJoinUx": v.SfBMeetingJoinUx, "BlockLegacyAuthorization": v.BlockLegacyAuthorization}, Fields: v.RawFields}, r, nil
		}
	}, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"DownloadTeams": v.DownloadTeams, "SfBMeetingJoinUx": v.SfBMeetingJoinUx, "BlockLegacyAuthorization": v.BlockLegacyAuthorization}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
