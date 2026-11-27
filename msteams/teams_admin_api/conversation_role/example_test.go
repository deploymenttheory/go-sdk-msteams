package conversationrole_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleProjectCreateByIdentity() {
	var sdk *msteams.Client
	run := func(ctx context.Context) error {
		request, e := roles.ProjectCreateByIdentity("Global/00000000-0000-4000-8000-000000000001", roles.CreateParameters{Permissions: &roles.Permissions{}})
		if e != nil {
			return e
		}
		projection := request.Projection()
		// Inspect Body and OmittedPaths before choosing to send the source-exact body.
		// Nested permission flags are not qualified by this compatibility operation.
		_ = projection
		_, _, e = sdk.TeamsAdminAPI.ConversationRoles.Create(ctx, *request)
		return e
	}
	_ = run // An example does not select credentials or create a role.
}
func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	run := func(ctx context.Context) error {
		identity := "Global/00000000-0000-4000-8000-000000000001"
		service := sdk.TeamsAdminAPI.ConversationRoles
		request, _, e := service.PrepareUpdateByIdentity(ctx, identity, roles.UpdateParameters{Name: policy.Null[string]()})
		if e != nil {
			return e
		}
		_ = request.Projection() // Inspect owned body and any source omissions first.
		result, _, e := service.UpdateByIdentity(ctx, identity, *request)
		_ = result // Keep preparation/action receipts when reconciling a partial failure.
		return e
	}
	_ = run
}

// Global comparison is optional source presentation data, never creation defaults.
func ExampleService_CreateWithGlobalComparison() {
	create := func(ctx context.Context, sdk *msteams.Client, request roles.CreateRequest) error {
		result, _, err := sdk.TeamsAdminAPI.ConversationRoles.CreateWithGlobalComparison(ctx, request)
		if result != nil {
			_ = result.ComparisonErr  // Keep a warning/denial separately from action error.
			_ = result.CreateResponse // Retain the action receipt for reconciliation.
		}
		return err
	}
	_ = create // Compiles without selecting credentials or creating a role.
}
