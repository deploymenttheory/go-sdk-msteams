package teamsaudioconferencingpolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_audio_conferencing_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"audio-policy": audioPolicyCleanup{sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies},
	}
}
