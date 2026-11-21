package audioconferencingprompts_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
)

func ExampleNewPrompt() {
	prompt, err := resource.NewPrompt("11111111-1111-1111-1111-111111111111", "FutureType", "xx-future")
	if err != nil {
		return
	}
	fmt.Println(*prompt.Type)
	// Output: FutureType
}

func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	prompt, err := resource.NewPrompt("11111111-1111-1111-1111-111111111111", "FutureType", "xx-future")
	if err != nil {
		return
	}
	// Add preserves existing entries and duplicates; use ClearPrompts to clear.
	_, _, _ = sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(context.Background(), "Global", resource.UpdateRequest{Prompts: resource.ChangePrompts([]*resource.Prompt{prompt}, nil)})
}

func ExampleClearPackages() {
	// Source package removal comparisons fail on nonempty collections.
	// Clearing or replacing the list is an explicit operation.
	edit := resource.ClearPackages()
	_, _ = edit.Apply(nil)
}
