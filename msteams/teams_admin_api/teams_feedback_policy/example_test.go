package teamsfeedbackpolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func ExampleSettings() {
	input := resource.Settings{AllowLogCollection: policy.Set(false), AllowEmailCollection: policy.Null[bool](), ReceiveSurveysMode: policy.Set("Enabled"), UserInitiatedMode: policy.Null[string]()}
	fmt.Println(input.AllowLogCollection.IsSet(), input.UserInitiatedMode.IsNull(), input.EnableFeatureSuggestions.IsSet())
	// Output: true true false
}
