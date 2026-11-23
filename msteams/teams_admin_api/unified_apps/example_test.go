package unifiedapps_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	ua "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/unified_apps"
)

func ExampleService_List() {
	blocked := func(ctx context.Context, sdk *msteams.Client) ([]string, error) {
		apps, _, err := sdk.TeamsAdminAPI.UnifiedApps.List(ctx)
		if err != nil {
			return nil, err
		}
		var ids []string
		for _, app := range apps {
			if app.ID != nil && app.IsBlocked != nil && *app.IsBlocked {
				ids = append(ids, *app.ID)
			}
		}
		return ids, nil
	}
	_ = blocked
}

func ExampleService_Update() {
	allowGroup := func(ctx context.Context, sdk *msteams.Client, appID, groupID string) error {
		_, err := sdk.TeamsAdminAPI.UnifiedApps.Update(ctx, appID, ua.AppUpdate{
			AssignmentType: ua.TypeUsersAndGroups,
			Operation:      ua.OperationAdd,
			Groups:         []string{groupID},
		})
		return err
	}
	_ = allowGroup
}

func ExampleService_ReviewCustomPending() {
	publish := func(ctx context.Context, sdk *msteams.Client, appID string) error {
		_, err := sdk.TeamsAdminAPI.UnifiedApps.ReviewCustomPending(ctx, appID, ua.ReviewPublished)
		return err
	}
	_ = publish
}
