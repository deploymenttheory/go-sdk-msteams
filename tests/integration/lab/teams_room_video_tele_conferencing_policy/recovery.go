package teamsroomvideoteleconferencingpolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_room_video_tele_conferencing_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"room-video-policy": roomVideoPolicyCleanup{sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies},
	}
}
