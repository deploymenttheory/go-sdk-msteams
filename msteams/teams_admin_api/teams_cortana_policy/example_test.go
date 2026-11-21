package teamscortanapolicy_test

import (
	"fmt"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

func ExampleSettings() {
	input := resource.Settings{Description: new(""), AllowCortanaVoiceInvocation: new(false)}
	fmt.Println(input.Description != nil, input.AllowCortanaVoiceInvocation != nil, input.AllowCortanaAmbientListening == nil)
	// Output: true true true
}
