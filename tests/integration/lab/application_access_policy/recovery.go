package applicationaccesspolicyacc

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// application_access_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"application-access-policy": policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies},
	}
}

// FixtureRecoveryHandlers returns the handlers that restore the dedicated
// user and group fixtures. Each is registered only when its private fixture
// file is configured, because recovery must bind to that exact identity.
func FixtureRecoveryHandlers(t *testing.T, sdk *msteams.Client, binding string) map[string]lab.Handler {
	t.Helper()
	handlers := map[string]lab.Handler{}
	if os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE") != "" {
		fixture := loadAssignmentUserFixture(t, binding)
		handlers["app-access-user-assignment"] = userAssignmentCleanup{sdk, fixture.UserID}
	}
	if os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE") != "" {
		fixture := loadAssignmentGroupFixture(t, binding)
		handler := groupAssignmentCleanup{sdk, fixture.GroupID}
		handlers["app-access-group-assignment"] = handler
		handlers["group-app-access-policy"] = groupOwnedPolicyCleanup{policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, handler}
	}
	return handlers
}
