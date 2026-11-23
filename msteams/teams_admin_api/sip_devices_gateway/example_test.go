package sipdevicesgateway_test

import (
	"context"
	"os"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	sdg "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/sip_devices_gateway"
)

// Submit the rows of a New-CsSdgBulkSignInRequest CSV, then follow the batch.
// The example compiles a real workflow and performs no I/O during go test.
func ExampleService_CreateBulkSignInRequest() {
	submit := func(ctx context.Context, sdk *msteams.Client, csvFile *os.File, region string) ([]*sdg.BulkSignInItemStatus, error) {
		items, err := sdg.ReadBulkSignInItems(csvFile)
		if err != nil {
			return nil, err
		}
		batch, _, err := sdk.TeamsAdminAPI.SIPDevicesGateway.CreateBulkSignInRequest(ctx, region, items)
		if err != nil || batch.BatchID == nil {
			return nil, err
		}
		status, _, err := sdk.TeamsAdminAPI.SIPDevicesGateway.GetBulkSignInRequestStatus(ctx, *batch.BatchID)
		if err != nil {
			return nil, err
		}
		var failed []*sdg.BulkSignInItemStatus
		for _, item := range status.BatchItems {
			if item.Failed() {
				failed = append(failed, item)
			}
		}
		return failed, nil
	}
	_ = submit
}

// List the tenant's earlier bulk sign-in batches.
func ExampleService_ListBulkSignInRequests() {
	list := func(ctx context.Context, sdk *msteams.Client) (int, error) {
		summary, _, err := sdk.TeamsAdminAPI.SIPDevicesGateway.ListBulkSignInRequests(ctx)
		if err != nil {
			return 0, err
		}
		return len(summary.Requests), nil
	}
	_ = list
}
