package teamsmeetingtemplatepermissionpolicyacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_meeting_template_permission_policy resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"template-baseline-policy": meetingTemplatePolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies},
		"meeting-template-policy":  meetingTemplatePolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies},
	}
}
