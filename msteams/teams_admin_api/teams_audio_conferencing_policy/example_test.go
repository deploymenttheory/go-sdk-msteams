package teamsaudioconferencingpolicy_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

func ExampleChangePhoneNumbers() {
	change := resource.ChangePhoneNumbers([]string{"+440000000001"}, nil)
	fmt.Println(change)
	// Output: phone list change (redacted)
}
func ExampleService_Create() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.Create(context.Background(), resource.CreateRequest{Identity: "Tag:Fixture", Settings: resource.Settings{AllowTollFreeDialin: new(false), MeetingInvitePhoneNumbers: resource.ReplacePhoneNumbers("+440000000001")}})
}
func ExampleService_UpdateGlobal() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// This is a sparse PUT with server Add/Remove, without a preparatory GET.
	_, _ = sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.UpdateGlobal(context.Background(), resource.Settings{MeetingInvitePhoneNumbers: resource.ChangePhoneNumbers([]string{"+440000000001"}, nil)})
}
