package teamseducationassignmentsapppolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
)

func ExampleSettings() {
	settings := resource.Settings{MakeCodeEnabledType: policy.Set(""), TurnItInAPIKey: policy.Null[string]()}
	value, bound := settings.MakeCodeEnabledType.Value()
	fmt.Println(bound, value == "", settings.TurnItInAPIKey.IsNull(), settings.ParentDigestEnabledType.IsSet())
	// Output: true true true false
}
