package calltemplates_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	ct "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleTemplates_Create() {
	create := func(ctx context.Context, cfg *auth.Config) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		name, description, sales := "Sales routing", "Routes the Sales tag", "Sales"
		target, kind := "tel:+15555550100", "ExternalPstn"
		template, _, err := sdk.TeamsAdminAPI.CallTemplates.Tags.Create(ctx, &ct.TagsTemplate{
			Name: &name, Description: &description,
			Tags: []ct.Tag{{TagName: &sales, TagDetails: &ct.TagTarget{ID: &target, Type: &kind}}},
		})
		if err != nil || template.ID == nil {
			return "", err
		}
		return *template.ID, nil
	}
	_ = create
}

// UpdateByID stores the body as sent, so read the template, change it and
// send it back whole.
func ExampleTemplates_UpdateByID() {
	restrict := func(ctx context.Context, sdk *msteams.Client, id string) error {
		history := sdk.TeamsAdminAPI.CallTemplates.SharedCallHistory
		template, _, err := history.GetByID(ctx, id)
		if err != nil {
			return err
		}
		missed := ct.HistoryAuthorizedUsersOnly
		template.IncomingMissedCalls = &missed
		_, _, err = history.UpdateByID(ctx, id, template)
		return err
	}
	_ = restrict
}
