package teamsmessagingpolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_policy"
)

func ExampleSettings() {
	input := resource.Settings{AllowGiphy: policy.Set(false), Description: policy.Null[string]()}
	fmt.Println(input.AllowGiphy.IsSet(), input.Description.IsNull())
	// Output: true true
}
