package teamsremotelogcollectionconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// Count the devices with a pending log collection request. No I/O happens
// during go test.
func ExampleService_GetByIdentity() {
	pending := func(ctx context.Context, sdk *msteams.Client) (int, error) {
		global, _, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.GetByIdentity(ctx, "Global")
		if err != nil {
			return 0, err
		}
		return len(global.Devices), nil
	}
	_ = pending
}
