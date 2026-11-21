package teamsenhancedencryptionpolicy_test

import (
	"fmt"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

func ExampleSettings() {
	settings := resource.Settings{Description: new(""), CallingEndToEndEncryptionEnabledType: new("DisabledUserOverride")}
	fmt.Println(settings.Description != nil, settings.CallingEndToEndEncryptionEnabledType != nil, settings.MeetingEndToEndEncryption == nil)
	// Output: true true true
}
