package meetingbranding_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func TestUnitExplicitEmptyFilterRejectsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "bound-empty-filter", Want: testutil.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.ListFiltered(ctx, "")
		return testutil.Observation{Value: value, Response: response, Err: err}
	}})
}
