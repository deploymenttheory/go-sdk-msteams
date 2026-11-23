package teamsmessagingconfigurationacc

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func messagingSingleton(sdk *msteams.Client) labkit.SingletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
	return labkit.SingletonConfiguration{Kind: "messaging", Field: "EnableVideoMessageCaptions", Get: func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return messagingConfigurationProjection(v), r, nil
	}, Set: func(ctx context.Context, value bool) (*resty.Response, error) {
		_, r, e := s.UpdateByIdentity(ctx, "Global", messaging.UpdateRequest{ScalarSettings: messaging.ScalarSettings{EnableVideoMessageCaptions: policy.Set(value)}})
		return r, e
	}}
}
