package teamsremotelogcollectiondevice_test

import (
	"context"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	device "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_device"
)

// Request logs from one user's device for three days, as the module's help
// recommends, then remove the request. No I/O happens during go test.
func ExampleService_Create() {
	collect := func(ctx context.Context, sdk *msteams.Client, deviceID, userID string) error {
		devices := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices
		expiry := time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339)
		created, _, err := devices.Create(ctx, device.CreateRequest{DeviceID: deviceID, UserID: userID, ExpireAfter: expiry})
		if err != nil {
			return err
		}
		_, _, err = devices.DeleteByID(ctx, created.ID, policy.Controls{})
		return err
	}
	_ = collect
}

// Extend an existing request; UpdateByID reads it first, then PATCHes once.
func ExampleService_UpdateByID() {
	extend := func(ctx context.Context, sdk *msteams.Client, id string, until time.Time) (bool, error) {
		result, _, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.UpdateByID(ctx, id,
			device.UpdateRequest{Settings: device.Settings{ExpireAfter: policy.Set(until.UTC().Format(time.RFC3339))}})
		return result != nil && result.HTTPAccepted, err
	}
	_ = extend
}
