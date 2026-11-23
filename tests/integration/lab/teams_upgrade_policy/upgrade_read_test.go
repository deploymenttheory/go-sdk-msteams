package teamsupgradepolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func TestAccUpgradePolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "upgrade-policy", "teams_upgrade_policy", func(sdk *msteams.Client) labkit.CompiledScalarRead {
		return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
			v, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.GetByIdentity(ctx, "Global")
			if e != nil || v == nil {
				return nil, r, e
			}
			return &labkit.CompiledScalarView{Identity: &v.Identity, Settings: map[string]any{"Description": v.Description, "Mode": v.Mode, "NotifySfbUsers": v.NotifySfbUsers, "Action": v.Action}, Fields: v.RawFields}, r, nil
		}
	}, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			out = append(out, &labkit.CompiledScalarView{Identity: &v.Identity, Settings: map[string]any{"Description": v.Description, "Mode": v.Mode, "NotifySfbUsers": v.NotifySfbUsers, "Action": v.Action}, Fields: v.RawFields})
		}
		return out, r, e
	})
}
