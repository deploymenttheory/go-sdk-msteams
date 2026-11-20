package teamtemplate_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

func ExampleNewDefinition() {
	definition, err := template.NewDefinition("Project", "Project collaboration")
	if err != nil {
		return
	}
	fmt.Println(*definition.DisplayName, definition.FunSettings == nil)
	// Output: Project true
}

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config) (receipt *template.TemplateSummary, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		definition, err := template.NewDefinition("Project", "Project collaboration")
		if err != nil {
			return nil, err
		}
		receipt, _, err = sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, "en-US", definition)
		return receipt, err
	}
	_ = create
}
