package teamscustombannertext_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func ExampleSettings() {
	settings := resource.Settings{Text: policy.Set(""), Description: policy.Null[string]()}
	text, bound := settings.Text.Value()
	fmt.Println(bound, text == "", settings.Description.IsNull())

}

func Example() {
	var sdk *msteams.Client
	ctx := context.Background()
	banners := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
	created, _, err := banners.Create(ctx, resource.CreateRequest{Text: "Lab banner"})
	if err != nil {

		return
	}
	if _, _, err = banners.GetByID(ctx, created.ID); err != nil {
		return
	}
	if _, _, err = banners.UpdateByID(ctx, created.ID, resource.UpdateRequest{Settings: resource.Settings{Text: policy.Set("Updated lab banner")}}); err != nil {
		return
	}
	_, _, _ = banners.DeleteByID(ctx, created.ID, policy.Controls{})
}
