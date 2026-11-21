package teamsclientconfiguration_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

func ExampleSettings() {
	input := resource.Settings{AllowEmailIntoChannel: policy.Set(false), RestrictedSenderList: policy.Set(""), UseUnifiedDomain: policy.Null[string]()}
	fmt.Println(input.AllowEmailIntoChannel.IsSet(), input.RestrictedSenderList.IsSet(), input.UseUnifiedDomain.IsNull(), input.AllowGuestUser.IsSet())
	// Output: true true true false
}
