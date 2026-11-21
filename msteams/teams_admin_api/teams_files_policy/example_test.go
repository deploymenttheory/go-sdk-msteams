package teamsfilespolicy_test

import (
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

func ExampleSettings() {
	input := resource.Settings{NativeFileEntryPoints: policy.Set(""), SPChannelFilesTab: policy.Null[string]()}
	fmt.Println(input.NativeFileEntryPoints.IsSet(), input.SPChannelFilesTab.IsNull(), input.DefaultFileUploadAppID.IsSet())
	// Output: true true false
}
