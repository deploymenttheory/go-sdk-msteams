package teamsshiftsapppolicyacc

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	shiftsApp "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_app_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func shiftsAppSingleton(sdk *msteams.Client) labkit.SingletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies
	return labkit.SingletonConfiguration{Kind: "shifts-app", Field: "AllowTimeClockLocationDetection", Get: func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, r, nil
	}, Set: func(ctx context.Context, value bool) (*resty.Response, error) {
		return s.UpdateGlobal(ctx, shiftsApp.UpdateRequest{AllowTimeClockLocationDetection: &value})
	}}
}
