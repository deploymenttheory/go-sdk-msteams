package targetinghierarchy_test

import (
	"context"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
)

// Compile-only examples borrow the application's configured client and source.
// They perform no authentication, file opening or tenant mutation in ordinary tests.
func ExampleService_UploadCSV() {
	upload := func(ctx context.Context, sdk *msteams.Client, body io.ReadCloser) (hierarchy.UploadResult, error) {
		return sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, hierarchy.CSV{Filename: "hierarchy.csv", Body: body, MaxBytes: 16 << 20}, hierarchy.V2)
	}
	_ = upload
}

func ExampleService_GetStatus() {
	read := func(ctx context.Context, sdk *msteams.Client, requestID string) (*hierarchy.Status, error) {
		value, _, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(ctx, requestID, hierarchy.V2)
		return value, err
	}
	_ = read
}
