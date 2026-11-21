package teamscallingpolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

func ExampleSettings() {
	input := resource.Settings{AllowPrivateCalling: policy.Set(false), CallingSpendUserLimit: policy.Set(int64(0)), CallRecordingExpirationDays: policy.Set(int64(120)), Description: policy.Null[string]()}
	fmt.Println(input.AllowPrivateCalling.IsSet(), input.Description.IsNull(), input.Copilot.IsSet())
	// Output: true true false
}
