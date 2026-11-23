package teamsnetworkroamingpolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func roamingRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: &v.Identity, Settings: map[string]any{"Description": v.Description, "AllowIPVideo": v.AllowIPVideo, "MediaBitRateKb": v.MediaBitRateKb}, Fields: v.RawFields}, r, nil
	}
}

func TestAccNetworkRoamingPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "roaming", "teams_network_roaming_policy", roamingRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			out = append(out, &labkit.CompiledScalarView{Identity: new(v.Identity), Settings: map[string]any{"Description": v.Description, "AllowIPVideo": v.AllowIPVideo, "MediaBitRateKb": v.MediaBitRateKb}, Fields: v.RawFields})
		}
		return out, r, e
	})
}
