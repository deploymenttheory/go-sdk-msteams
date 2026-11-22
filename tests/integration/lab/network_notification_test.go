package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func roamingRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{&v.Identity, map[string]any{"Description": v.Description, "AllowIPVideo": v.AllowIPVideo, "MediaBitRateKb": v.MediaBitRateKb}, v.RawFields}, r, nil
	}
}
func notificationRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "SuggestedFeedsEnabledType": v.SuggestedFeedsEnabledType, "TrendingFeedsEnabledType": v.TrendingFeedsEnabledType}, v.RawFields}, r, nil
	}
}
func TestAccNetworkRoamingPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "roaming", "teams_network_roaming_policy", roamingRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			out = append(out, &compiledScalarView{new(v.Identity), map[string]any{"Description": v.Description, "AllowIPVideo": v.AllowIPVideo, "MediaBitRateKb": v.MediaBitRateKb}, v.RawFields})
		}
		return out, r, e
	})
}
func TestAccNotificationAndFeedsPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "notification", "teams_notification_and_feeds_policy", notificationRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "SuggestedFeedsEnabledType": v.SuggestedFeedsEnabledType, "TrendingFeedsEnabledType": v.TrendingFeedsEnabledType}, v.RawFields})
			}
		}
		return out, r, e
	})
}
