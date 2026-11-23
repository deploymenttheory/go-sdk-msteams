package teamstemplatepermissionpolicy_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Policy, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}

func ExampleNewHiddenTemplate() {
	value := resource.NewHiddenTemplate("template-id")
	fmt.Println(*value.ID)
	// Output: template-id
}
