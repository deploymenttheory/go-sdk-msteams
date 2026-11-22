package audioprompts_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_Create() {
	upload := func(ctx context.Context, cfg *auth.Config, content []byte) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		file, _, err := sdk.TeamsAdminAPI.AudioFiles.Create(ctx, audio.ImportRequest{
			ApplicationID: audio.OrgAutoAttendant,
			FileName:      "greeting.wav",
			Content:       content,
		})
		if err != nil || file.ID == nil {
			return "", err
		}
		return *file.ID, nil
	}
	_ = upload
}

// GetContentByID returns the stored (service-transcoded) audio bytes.
func ExampleService_GetContentByID() {
	export := func(ctx context.Context, sdk *msteams.Client, id string) ([]byte, error) {
		content, _, err := sdk.TeamsAdminAPI.AudioFiles.GetContentByID(ctx, audio.TenantGlobal, id)
		return content, err
	}
	_ = export
}
