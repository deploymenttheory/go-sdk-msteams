package teamschannelspolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
)

func ExampleSettings() {
	input := resource.Settings{AllowPrivateChannelCreation: policy.Set(false), AllowCreatePrivateChannel: policy.Set("FutureMode"), Description: policy.Null[string]()}
	fmt.Println(input.AllowPrivateChannelCreation.IsSet(), input.AllowCreatePrivateChannel.IsSet(), input.Description.IsNull(), input.AllowSharedChannelCreation.IsSet())
	// Output: true true true false
}
