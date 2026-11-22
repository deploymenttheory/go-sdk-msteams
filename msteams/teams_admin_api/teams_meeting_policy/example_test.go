package teamsmeetingpolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_policy"
)

func ExampleSettings() {
	input := resource.Settings{AllowIPVideo: policy.Set(false), Description: policy.Null[string]()}
	fmt.Println(input.AllowIPVideo.IsSet(), input.Description.IsNull())
	// Output: true true
}
