package teamsmeetingbroadcastpolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_meeting_broadcast_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"meeting-broadcast-policy": meetingBroadcastPolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies},
	}
}
