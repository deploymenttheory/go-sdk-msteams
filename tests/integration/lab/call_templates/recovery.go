package calltemplatesacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	ct "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// call_templates resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	s := sdk.TeamsAdminAPI.CallTemplates
	return map[string]lab.Handler{
		"auto-recording-template": templateCleanup[ct.AutoRecordingTemplate]{s.AutoRecording,
			func(v *ct.AutoRecordingTemplate) *string { return v.Name }, func(v *ct.AutoRecordingTemplate) *string { return v.ID }},
		"compliance-recording-template": templateCleanup[ct.ComplianceRecordingTemplate]{s.ComplianceRecording,
			func(v *ct.ComplianceRecordingTemplate) *string { return v.Name }, func(v *ct.ComplianceRecordingTemplate) *string { return v.ID }},
		"shared-call-history-template": templateCleanup[ct.SharedCallHistoryTemplate]{s.SharedCallHistory,
			func(v *ct.SharedCallHistoryTemplate) *string { return v.Name }, func(v *ct.SharedCallHistoryTemplate) *string { return v.ID }},
		"voicemail-triage-template": templateCleanup[ct.SharedVoicemailTriageSettingsTemplate]{s.SharedVoicemailTriageSettings,
			func(v *ct.SharedVoicemailTriageSettingsTemplate) *string { return v.Name },
			func(v *ct.SharedVoicemailTriageSettingsTemplate) *string { return v.ID }},
		"tags-template": templateCleanup[ct.TagsTemplate]{s.Tags,
			func(v *ct.TagsTemplate) *string { return v.Name }, func(v *ct.TagsTemplate) *string { return v.ID }},
	}
}
