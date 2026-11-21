package teamseducationconfiguration_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

func ExampleSettings() {
	settings := resource.Settings{ParentGuardianPreferredContactMethod: policy.Set(""), UpdateParentInformation: policy.Null[string]()}
	value, bound := settings.ParentGuardianPreferredContactMethod.Value()
	fmt.Println(bound, value == "", settings.UpdateParentInformation.IsNull(), settings.EduGenerativeAIEnhancements.IsSet())
	// Output: true true true false
}
