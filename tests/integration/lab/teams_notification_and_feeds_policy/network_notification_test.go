package teamsnotificationandfeedspolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func notificationRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "SuggestedFeedsEnabledType": v.SuggestedFeedsEnabledType, "TrendingFeedsEnabledType": v.TrendingFeedsEnabledType}, Fields: v.RawFields}, r, nil
	}
}

func TestAccNotificationAndFeedsPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "notification", "teams_notification_and_feeds_policy", notificationRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "SuggestedFeedsEnabledType": v.SuggestedFeedsEnabledType, "TrendingFeedsEnabledType": v.TrendingFeedsEnabledType}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
