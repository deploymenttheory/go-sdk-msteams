package conversationrolessetting_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	settings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_roles_setting"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	run := func(ctx context.Context) error {
		service := sdk.TeamsAdminAPI.ConversationRoleSettings
		request, _, e := service.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{ConversationRoles: policy.Null[*settings.ListModifier]()})
		if e != nil {
			return e
		}
		_ = request.Projection() // Explicitly inspect this prepared full-list clear.
		result, _, e := service.UpdateByIdentity(ctx, "Global", *request)
		_ = result // Retain Before/read/action stages if the response is uncertain.
		return e
	}
	_ = run // No Global configuration mutation is executed by the example.
}
