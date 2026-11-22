package teamsshiftsapppolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func shiftsAppRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, Fields: v.RawFields}, r, nil
	}
}

func TestAccShiftsAppPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "shifts-app", "teams_shifts_app_policy", shiftsAppRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, Fields: v.RawFields})
		}
		return out, r, e
	})
}
