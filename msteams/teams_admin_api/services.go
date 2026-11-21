// Package teamsadminapi composes Teams administration resource services for a client.
package teamsadminapi

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	acsfederationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	applicationaccesspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
	applicationmeetingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
	assignmentmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
	audioconferencingprompts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	batchdeployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
	externalaccesspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	meetingbranding "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	onlinevoicemailpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	voicemailvalidation "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	policyassignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	policypackages "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
	privacyconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
	privatechannelmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/private_channel_migration"
	recordingmessages "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/region"
	targetinghierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	targetingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
	teamtemplate "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
	teamsapppermissionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
	teamsappsetuppolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
	teamsaudioconferencingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
	teamscallholdpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
	teamscallparkpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
	teamscallingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
	teamsexternalaccessconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
	tenantfederationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

// Services is the typed Teams administration collection. Resource fields are added
// with their evidenced implementations; no placeholder management methods are exposed.
type Services struct {
	TeamsCallingPolicies                    *teamscallingpolicy.Service
	TeamsCallParkPolicies                   *teamscallparkpolicy.Service
	TeamsCallHoldPolicies                   *teamscallholdpolicy.Service
	TeamsAudioConferencingPolicies          *teamsaudioconferencingpolicy.Service
	AudioConferencingCustomPrompts          *audioconferencingprompts.Service
	TeamsAppSetupPolicies                   *teamsappsetuppolicy.Service
	TeamsAppPermissionPolicies              *teamsapppermissionpolicy.Service
	RecordingMessages                       *recordingmessages.Service
	PrivacyConfigurations                   *privacyconfiguration.Service
	PolicyPackages                          *policypackages.Service
	PolicyAssignments                       *policyassignments.Service
	OnlineVoicemailValidationConfigurations *voicemailvalidation.Service
	OnlineVoicemailPolicies                 *onlinevoicemailpolicy.Service
	MeetingBrandingPolicies                 *meetingbranding.Service
	AssignmentMigrations                    *assignmentmigration.Service
	TeamsExternalAccessConfigurations       *teamsexternalaccessconfiguration.Service
	ACSFederationConfigurations             *acsfederationconfiguration.Service
	ExternalAccessPolicies                  *externalaccesspolicy.Service
	scope                                   *client.Scope
	ApplicationAccessPolicies               *applicationaccesspolicy.Service
	ApplicationMeetingConfigurations        *applicationmeetingconfiguration.Service
	TenantFederationConfigurations          *tenantfederationconfiguration.Service
	TeamTemplates                           *teamtemplate.Service
	TargetingPolicies                       *targetingpolicy.Service
	TargetingHierarchy                      *targetinghierarchy.Service
	BatchDeployments                        *batchdeployment.Service
	PrivateChannelMigration                 *privatechannelmigration.Service
	Regions                                 *region.Service
}

// NewServices binds the collection to its client's session.
func NewServices(session *client.Session) (*Services, error) {
	scope, err := client.NewScope(client.TeamsAdmin, session)
	if err != nil {
		return nil, err
	}
	backend := configapi.New(scope)
	hierarchy, err := targetinghierarchy.NewService(session)
	if err != nil {
		return nil, err
	}
	migrationScope, err := client.NewScope(client.PrivateChannelMigration, session)
	if err != nil {
		return nil, err
	}
	regionScope, err := client.NewScope(client.Regions, session)
	if err != nil {
		return nil, err
	}
	return &Services{
		TeamsCallingPolicies:                    teamscallingpolicy.NewService(scope),
		TeamsCallParkPolicies:                   teamscallparkpolicy.NewService(backend),
		TeamsCallHoldPolicies:                   teamscallholdpolicy.NewService(scope),
		TeamsAudioConferencingPolicies:          teamsaudioconferencingpolicy.NewService(backend),
		AudioConferencingCustomPrompts:          audioconferencingprompts.NewService(scope),
		TeamsAppSetupPolicies:                   teamsappsetuppolicy.NewService(scope),
		TeamsAppPermissionPolicies:              teamsapppermissionpolicy.NewService(scope),
		RecordingMessages:                       recordingmessages.NewService(scope),
		PrivacyConfigurations:                   privacyconfiguration.NewService(scope),
		PolicyPackages:                          policypackages.NewService(scope),
		PolicyAssignments:                       policyassignments.NewService(scope),
		scope:                                   scope,
		OnlineVoicemailValidationConfigurations: voicemailvalidation.NewService(scope),
		OnlineVoicemailPolicies:                 onlinevoicemailpolicy.NewService(scope),
		MeetingBrandingPolicies:                 meetingbranding.NewService(scope),
		AssignmentMigrations:                    assignmentmigration.NewService(scope),
		TeamsExternalAccessConfigurations:       teamsexternalaccessconfiguration.NewService(scope),
		ACSFederationConfigurations:             acsfederationconfiguration.NewService(scope),
		ExternalAccessPolicies:                  externalaccesspolicy.NewService(scope),
		ApplicationAccessPolicies:               applicationaccesspolicy.NewService(backend),
		ApplicationMeetingConfigurations:        applicationmeetingconfiguration.NewService(backend),
		TenantFederationConfigurations:          tenantfederationconfiguration.NewService(backend),
		TeamTemplates:                           teamtemplate.NewService(backend),
		TargetingPolicies:                       targetingpolicy.NewService(backend),
		TargetingHierarchy:                      hierarchy,
		BatchDeployments:                        batchdeployment.NewService(backend),
		PrivateChannelMigration:                 privatechannelmigration.NewService(migrationScope),
		Regions:                                 region.NewService(regionScope),
	}, nil
}

// Scope returns the backend dependency shared by this collection's resources.
func (s *Services) Scope() *client.Scope { return s.scope }
