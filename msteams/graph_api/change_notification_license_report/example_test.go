package licensereport_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	report "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/change_notification_license_report"
)

// The application supplies a delegated client with each documented permission
// and role. The closure compiles without acquiring credentials or running reads.
func ExampleService_Get() {
	read := func(ctx context.Context, sdk *msteams.Client) (*report.ReportResult, error) {
		result, err := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period30Days, report.GetOptions{MaxUsers: 100, Licenses: report.ListOptions{MaxPages: 5, MaxItems: 1000}})
		if errors.Is(err, report.ErrPartial) {
			// Retain the result and inspect per-user errors. Unknown is not false.
			return result, err
		}
		return result, err
	}
	_ = read
}

// Activity can be read independently when full license composition is unavailable.
func ExampleService_ListActivity() {
	read := func(ctx context.Context, sdk *msteams.Client) error {
		_, err := sdk.GraphAPI.ChangeNotificationLicenseReports.ListActivity(ctx, report.Period7Days, report.ListOptions{MaxPages: 2, MaxItems: 100})
		return err
	}
	_ = read
}

// Pinned projection is an explicit local compatibility choice, never a license
// lookup or proof of effective authorization. An empty source dictionary omits rows.
func ExampleProjectPinnedRows() {
	rows, err := report.ProjectPinnedRows(nil, nil, map[string][]*report.PinnedLicense{})
	_ = rows
	_ = err
}
