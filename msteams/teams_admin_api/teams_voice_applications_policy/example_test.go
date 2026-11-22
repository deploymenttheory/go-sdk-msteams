package teamsvoiceapplicationspolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

func ExampleService_CreateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies.CreateByIdentity(context.Background(), "sdk-owned-policy", resource.CreateRequest{Settings: resource.Settings{Description: policy.Set("Administrative controls"), AllowCallQueueOptOutChange: policy.Set(true)}})
}
