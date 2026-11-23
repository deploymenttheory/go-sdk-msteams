package teamsartifact_test

import (
	"context"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	teamsartifact "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/teams_artifact"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test. Failures are per drive or search and do not stop the scan.
func ExampleService_List() {
	recentRecordings := func(ctx context.Context, sdk *msteams.Client, mail string) ([]teamsartifact.Artifact, []teamsartifact.Failure, error) {
		kind, since := teamsartifact.RecordingTranscript, time.Now().AddDate(0, 0, -30)
		result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, teamsartifact.Options{OneDrive: mail, Type: &kind, Start: &since})
		return result.Artifacts, result.Failures, err
	}
	_ = recentRecordings
}
