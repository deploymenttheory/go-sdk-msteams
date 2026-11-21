package teamseventspolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

func ExampleSettings() {
	input := resource.Settings{UseMicrosoftECDN: policy.Set(false), AllowEventIntegrations: policy.Null[bool](), AllowWebinars: policy.Set("Enabled"), Description: policy.Null[string]()}
	fmt.Println(input.UseMicrosoftECDN.IsSet(), input.Description.IsNull(), input.Registration.IsSet())
	// Output: true true false
}
