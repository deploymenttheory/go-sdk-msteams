package batchdeployment_test

import (
	"context"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	deployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
)

func ExampleService_Create() {
	submit := func(ctx context.Context, sdk *msteams.Client, teams, users io.Reader, recipients string) (*deployment.Submission, error) {
		archive, err := deployment.BuildArchive(ctx, teams, users, deployment.ArchiveOptions{MaxInputBytes: 8 << 20, MaxArchiveBytes: 16 << 20})
		if err != nil {
			return nil, err
		}
		value, _, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: recipients})
		return value, err
	}
	_ = submit
}
