// Package teamsadminapi composes Teams administration resource services for a client.
package teamsadminapi

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	acsfederationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	applicationaccesspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
	applicationmeetingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
	assignmentmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
	audioconferencingprompts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	audioprompts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
	autoattendants "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	batchdeployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
	callqueues "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
	calltemplates "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	cloudcalldata "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/cloud_call_data"
	compliancerecording "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/compliance_recording"
	configurationdiagnostics "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	customappsettings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/custom_app_settings"
	externalaccesspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	mainlineattendants "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/mainline_attendants"
	meetingbranding "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/migration"
	onlinevoicemailpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	voicemailvalidation "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	policyassignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	policypackages "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
	privacyconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
	privatechannelmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/private_channel_migration"
	provisioningdiagnostics "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/provisioning_diagnostics"
	recordingmessages "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/region"
	resourceaccounts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/resource_accounts"
	schedules "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	shiftsconnections "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/shifts_connections"
	sipdevicesgateway "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/sip_devices_gateway"
	targetinghierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
	targetingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
	teamtemplate "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
	teamsaipolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ai_policy"
	teamsapppermissionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
	teamsappsetuppolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
	teamsaudioconferencingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
	teamsbyodanddeskspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_byod_and_desks_policy"
	teamscallholdpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
	teamscallparkpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
	teamscallingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
	teamschannelspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
	teamsclientconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
	teamscortanapolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
	teamscustombannertext "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
	teamseducationassignmentsapppolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
	teamseducationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
	teamsenhancedencryptionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
	teamseventspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
	teamsexternalaccessconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
	teamsfeedbackpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
	teamsfilespolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
	teamsfirstpartymeetingtemplateconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
	teamsguestcallingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_calling_configuration"
	teamsguestmeetingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
	teamsguestmessagingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
	teamsipphonepolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ip_phone_policy"
	teamsmedialoggingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
	teamsmeetingbroadcastconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
	teamsmeetingbroadcastpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_policy"
	teamsmeetingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
	teamsmeetingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_policy"
	teamsmeetingtemplateconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_configuration"
	teamsmeetingtemplatepermissionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
	teamsmessagingconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	teamsmessagingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_policy"
	teamsmigrationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
	teamsmobilitypolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_mobility_policy"
	teamsmultitenantorganizationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_multi_tenant_organization_configuration"
	teamsnetworkroamingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_network_roaming_policy"
	teamsnotificationandfeedspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_notification_and_feeds_policy"
	teamspersonalattendantpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_personal_attendant_policy"
	teamsrecordingrolloutpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_recording_roll_out_policy"
	teamsremotelogcollectionconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_configuration"
	teamsremotelogcollectiondevice "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_device"
	teamsroomvideoteleconferencingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_room_video_tele_conferencing_policy"
	teamssharedcallingroutingpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shared_calling_routing_policy"
	teamsshiftsapppolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_app_policy"
	teamsshiftspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_policy"
	teamssipdevicesconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_sip_devices_configuration"
	teamstemplatepermissionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
	teamstenantabuseconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
	teamsupdatemanagementpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
	teamsupgradeconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_configuration"
	teamsupgradepolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_policy"
	teamsvdipolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_vdi_policy"
	teamsvirtualappointmentspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_virtual_appointments_policy"
	teamsvoiceapplicationspolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
	teamsworkloadpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
	teamsworklocationdetectionpolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_location_detection_policy"
	tenantfederationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
	tenantidentity "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
	tenantmigrationconfiguration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_migration_configuration"
	unifiedapps "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/unified_apps"
)

// Services is the typed Teams administration collection. Resource fields are added
// with their evidenced implementations; no placeholder management methods are exposed.
type Services struct {
	TeamsVoiceApplicationsPolicies   *teamsvoiceapplicationspolicy.Service
	TeamsWorkLoadPolicies            *teamsworkloadpolicy.Service
	TeamsVDIPolicies                 *teamsvdipolicy.Service
	TeamsVirtualAppointmentsPolicies *teamsvirtualappointmentspolicy.Service

	TeamsUpgradeConfigurations                 *teamsupgradeconfiguration.Service
	TeamsUpgradePolicies                       *teamsupgradepolicy.Service
	TeamsMediaLoggingPolicies                  *teamsmedialoggingpolicy.Service
	TeamsWorkLocationDetectionPolicies         *teamsworklocationdetectionpolicy.Service
	TeamsTenantAbuseConfigurations             *teamstenantabuseconfiguration.Service
	TeamsUpdateManagementPolicies              *teamsupdatemanagementpolicy.Service
	TeamsTemplatePermissionPolicies            *teamstemplatepermissionpolicy.Service
	TeamsSipDevicesConfigurations              *teamssipdevicesconfiguration.Service
	TeamsRoomVideoTeleConferencingPolicies     *teamsroomvideoteleconferencingpolicy.Service
	TeamsMeetingBroadcastPolicies              *teamsmeetingbroadcastpolicy.Service
	TeamsSharedCallingRoutingPolicies          *teamssharedcallingroutingpolicy.Service
	TeamsPersonalAttendantPolicies             *teamspersonalattendantpolicy.Service
	TeamsRecordingRollOutPolicies              *teamsrecordingrolloutpolicy.Service
	TeamsShiftsPolicies                        *teamsshiftspolicy.Service
	TeamsShiftsAppPolicies                     *teamsshiftsapppolicy.Service
	TeamsNetworkRoamingPolicies                *teamsnetworkroamingpolicy.Service
	TeamsNotificationAndFeedsPolicies          *teamsnotificationandfeedspolicy.Service
	TeamsMobilityPolicies                      *teamsmobilitypolicy.Service
	TeamsMultiTenantOrganizationConfigurations *teamsmultitenantorganizationconfiguration.Service
	TeamsMigrationConfigurations               *teamsmigrationconfiguration.Service
	TeamsGuestCallingConfigurations            *teamsguestcallingconfiguration.Service
	TeamsMessagingConfigurations               *teamsmessagingconfiguration.Service
	TeamsMeetingBroadcastConfigurations        *teamsmeetingbroadcastconfiguration.Service
	TeamsMeetingTemplatePermissionPolicies     *teamsmeetingtemplatepermissionpolicy.Service
	TeamsMeetingTemplateConfigurations         *teamsmeetingtemplateconfiguration.Service
	TeamsMessagingPolicies                     *teamsmessagingpolicy.Service
	TeamsGuestMeetingConfigurations            *teamsguestmeetingconfiguration.Service
	TeamsMeetingPolicies                       *teamsmeetingpolicy.Service
	TeamsMeetingConfigurations                 *teamsmeetingconfiguration.Service
	TeamsEventsPolicies                        *teamseventspolicy.Service
	TeamsFilesPolicies                         *teamsfilespolicy.Service
	TeamsIPPhonePolicies                       *teamsipphonepolicy.Service
	TeamsGuestMessagingConfigurations          *teamsguestmessagingconfiguration.Service
	TeamsEnhancedEncryptionPolicies            *teamsenhancedencryptionpolicy.Service
	TeamsEducationConfigurations               *teamseducationconfiguration.Service
	TeamsEducationAssignmentsAppPolicies       *teamseducationassignmentsapppolicy.Service
	TeamsCustomBannerTexts                     *teamscustombannertext.Service
	TeamsCortanaPolicies                       *teamscortanapolicy.Service
	TeamsClientConfigurations                  *teamsclientconfiguration.Service
	TeamsChannelsPolicies                      *teamschannelspolicy.Service
	TeamsFeedbackPolicies                      *teamsfeedbackpolicy.Service
	TeamsCallingPolicies                       *teamscallingpolicy.Service
	TeamsCallParkPolicies                      *teamscallparkpolicy.Service
	TeamsCallHoldPolicies                      *teamscallholdpolicy.Service
	TeamsAudioConferencingPolicies             *teamsaudioconferencingpolicy.Service
	TeamsBYODAndDesksPolicies                  *teamsbyodanddeskspolicy.Service
	AudioConferencingCustomPrompts             *audioconferencingprompts.Service
	TeamsAppSetupPolicies                      *teamsappsetuppolicy.Service
	TeamsAppPermissionPolicies                 *teamsapppermissionpolicy.Service
	TeamsAIPolicies                            *teamsaipolicy.Service
	RecordingMessages                          *recordingmessages.Service
	PrivacyConfigurations                      *privacyconfiguration.Service
	PolicyPackages                             *policypackages.Service
	PolicyAssignments                          *policyassignments.Service
	OnlineVoicemailValidationConfigurations    *voicemailvalidation.Service
	OnlineVoicemailPolicies                    *onlinevoicemailpolicy.Service
	MeetingBrandingPolicies                    *meetingbranding.Service
	AssignmentMigrations                       *assignmentmigration.Service
	TeamsExternalAccessConfigurations          *teamsexternalaccessconfiguration.Service
	ACSFederationConfigurations                *acsfederationconfiguration.Service
	ExternalAccessPolicies                     *externalaccesspolicy.Service
	scope                                      *client.Scope
	ApplicationAccessPolicies                  *applicationaccesspolicy.Service
	ApplicationMeetingConfigurations           *applicationmeetingconfiguration.Service
	TenantFederationConfigurations             *tenantfederationconfiguration.Service
	TenantMigrationConfigurations              *tenantmigrationconfiguration.Service
	TeamTemplates                              *teamtemplate.Service
	AudioFiles                                 *audioprompts.Service
	CloudCallData                              *cloudcalldata.Service
	ConfigurationDiagnostics                   *configurationdiagnostics.Service
	ResourceAccounts                           *resourceaccounts.Service
	Schedules                                  *schedules.Service
	CallQueues                                 *callqueues.Service
	TenantIdentity                             *tenantidentity.Service
	CallTemplates                              *calltemplates.Service
	ComplianceRecording                        *compliancerecording.Service
	ShiftsConnections                          *shiftsconnections.Service
	MainlineAttendants                         *mainlineattendants.Service
	AIAgents                                   *aiagents.Service
	ProvisioningDiagnostics                    *provisioningdiagnostics.Service
	CustomAppSettings                          *customappsettings.Service
	SIPDevicesGateway                          *sipdevicesgateway.Service
	TeamsRemoteLogCollectionConfigurations     *teamsremotelogcollectionconfiguration.Service
	TeamsRemoteLogCollectionDevices            *teamsremotelogcollectiondevice.Service
	Migration                                  *migration.Service
	AutoAttendants                             *autoattendants.Service
	TargetingPolicies                          *targetingpolicy.Service
	TargetingHierarchy                         *targetinghierarchy.Service
	BatchDeployments                           *batchdeployment.Service
	PrivateChannelMigration                    *privatechannelmigration.Service
	Regions                                    *region.Service
	UnifiedApps                                *unifiedapps.Service

	TeamsFirstPartyMeetingTemplateConfigurations *teamsfirstpartymeetingtemplateconfiguration.Service
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
	knowledgeScope, err := client.NewScope(client.AIGeneratedKnowledge, session)
	if err != nil {
		return nil, err
	}
	appCatalogScope, err := client.NewScope(client.AppCatalog, session)
	if err != nil {
		return nil, err
	}
	return &Services{
		TeamsVoiceApplicationsPolicies:   teamsvoiceapplicationspolicy.NewService(scope),
		TeamsWorkLoadPolicies:            teamsworkloadpolicy.NewService(backend),
		TeamsVDIPolicies:                 teamsvdipolicy.NewService(scope),
		TeamsVirtualAppointmentsPolicies: teamsvirtualappointmentspolicy.NewService(scope),

		TeamsUpgradeConfigurations:                 teamsupgradeconfiguration.NewService(scope),
		TeamsUpgradePolicies:                       teamsupgradepolicy.NewService(backend),
		TeamsMediaLoggingPolicies:                  teamsmedialoggingpolicy.NewService(backend),
		TeamsWorkLocationDetectionPolicies:         teamsworklocationdetectionpolicy.NewService(scope),
		TeamsTenantAbuseConfigurations:             teamstenantabuseconfiguration.NewService(scope),
		TeamsUpdateManagementPolicies:              teamsupdatemanagementpolicy.NewService(scope),
		TeamsTemplatePermissionPolicies:            teamstemplatepermissionpolicy.NewService(scope),
		TeamsSipDevicesConfigurations:              teamssipdevicesconfiguration.NewService(scope),
		TeamsRoomVideoTeleConferencingPolicies:     teamsroomvideoteleconferencingpolicy.NewService(backend),
		TeamsMeetingBroadcastPolicies:              teamsmeetingbroadcastpolicy.NewService(backend),
		TeamsSharedCallingRoutingPolicies:          teamssharedcallingroutingpolicy.NewService(scope),
		TeamsPersonalAttendantPolicies:             teamspersonalattendantpolicy.NewService(scope),
		TeamsRecordingRollOutPolicies:              teamsrecordingrolloutpolicy.NewService(scope),
		TeamsShiftsPolicies:                        teamsshiftspolicy.NewService(scope),
		TeamsShiftsAppPolicies:                     teamsshiftsapppolicy.NewService(backend),
		TeamsNetworkRoamingPolicies:                teamsnetworkroamingpolicy.NewService(backend),
		TeamsNotificationAndFeedsPolicies:          teamsnotificationandfeedspolicy.NewService(scope),
		TeamsMobilityPolicies:                      teamsmobilitypolicy.NewService(scope),
		TeamsMultiTenantOrganizationConfigurations: teamsmultitenantorganizationconfiguration.NewService(scope),
		TeamsMigrationConfigurations:               teamsmigrationconfiguration.NewService(backend),
		TeamsGuestCallingConfigurations:            teamsguestcallingconfiguration.NewService(backend),
		TeamsMessagingConfigurations:               teamsmessagingconfiguration.NewService(scope),
		TeamsMeetingBroadcastConfigurations:        teamsmeetingbroadcastconfiguration.NewService(backend),
		TeamsMeetingTemplatePermissionPolicies:     teamsmeetingtemplatepermissionpolicy.NewService(scope),
		TeamsMeetingTemplateConfigurations:         teamsmeetingtemplateconfiguration.NewService(scope),
		TeamsMessagingPolicies:                     teamsmessagingpolicy.NewService(scope),
		TeamsGuestMeetingConfigurations:            teamsguestmeetingconfiguration.NewService(scope),
		TeamsMeetingPolicies:                       teamsmeetingpolicy.NewService(scope),
		TeamsMeetingConfigurations:                 teamsmeetingconfiguration.NewService(scope),
		TeamsEventsPolicies:                        teamseventspolicy.NewService(scope),
		TeamsFilesPolicies:                         teamsfilespolicy.NewService(scope),
		TeamsIPPhonePolicies:                       teamsipphonepolicy.NewService(backend),
		TeamsGuestMessagingConfigurations:          teamsguestmessagingconfiguration.NewService(backend),
		TeamsEnhancedEncryptionPolicies:            teamsenhancedencryptionpolicy.NewService(backend),
		TeamsEducationConfigurations:               teamseducationconfiguration.NewService(scope),
		TeamsEducationAssignmentsAppPolicies:       teamseducationassignmentsapppolicy.NewService(scope),
		TeamsCustomBannerTexts:                     teamscustombannertext.NewService(scope),
		TeamsCortanaPolicies:                       teamscortanapolicy.NewService(backend),
		TeamsClientConfigurations:                  teamsclientconfiguration.NewService(scope),
		TeamsChannelsPolicies:                      teamschannelspolicy.NewService(scope),
		TeamsFeedbackPolicies:                      teamsfeedbackpolicy.NewService(scope),
		TeamsCallingPolicies:                       teamscallingpolicy.NewService(scope),
		TeamsCallParkPolicies:                      teamscallparkpolicy.NewService(backend),
		TeamsCallHoldPolicies:                      teamscallholdpolicy.NewService(scope),
		TeamsAudioConferencingPolicies:             teamsaudioconferencingpolicy.NewService(backend),
		TeamsBYODAndDesksPolicies:                  teamsbyodanddeskspolicy.NewService(scope),
		AudioConferencingCustomPrompts:             audioconferencingprompts.NewService(scope),
		TeamsAppSetupPolicies:                      teamsappsetuppolicy.NewService(scope),
		TeamsAppPermissionPolicies:                 teamsapppermissionpolicy.NewService(scope),
		TeamsAIPolicies:                            teamsaipolicy.NewService(scope),
		RecordingMessages:                          recordingmessages.NewService(scope),
		PrivacyConfigurations:                      privacyconfiguration.NewService(scope),
		PolicyPackages:                             policypackages.NewService(scope),
		PolicyAssignments:                          policyassignments.NewService(scope),
		scope:                                      scope,
		OnlineVoicemailValidationConfigurations:    voicemailvalidation.NewService(scope),
		OnlineVoicemailPolicies:                    onlinevoicemailpolicy.NewService(scope),
		MeetingBrandingPolicies:                    meetingbranding.NewService(scope),
		AssignmentMigrations:                       assignmentmigration.NewService(scope),
		TeamsExternalAccessConfigurations:          teamsexternalaccessconfiguration.NewService(scope),
		ACSFederationConfigurations:                acsfederationconfiguration.NewService(scope),
		ExternalAccessPolicies:                     externalaccesspolicy.NewService(scope),
		ApplicationAccessPolicies:                  applicationaccesspolicy.NewService(backend),
		ApplicationMeetingConfigurations:           applicationmeetingconfiguration.NewService(backend),
		TenantFederationConfigurations:             tenantfederationconfiguration.NewService(backend),
		TenantMigrationConfigurations:              tenantmigrationconfiguration.NewService(backend),
		TeamTemplates:                              teamtemplate.NewService(backend),
		AudioFiles:                                 audioprompts.NewService(backend),
		CloudCallData:                              cloudcalldata.NewService(backend),
		ConfigurationDiagnostics:                   configurationdiagnostics.NewService(backend, scope),
		ResourceAccounts:                           resourceaccounts.NewService(backend),
		Schedules:                                  schedules.NewService(backend),
		CallQueues:                                 callqueues.NewService(backend),
		TenantIdentity:                             tenantidentity.NewService(backend),
		CallTemplates:                              calltemplates.NewService(backend),
		ComplianceRecording:                        compliancerecording.NewService(scope),
		ShiftsConnections:                          shiftsconnections.NewService(backend),
		MainlineAttendants:                         mainlineattendants.NewService(backend),
		AIAgents:                                   aiagents.NewService(backend, knowledgeScope),
		ProvisioningDiagnostics:                    provisioningdiagnostics.NewService(backend),
		CustomAppSettings:                          customappsettings.NewService(backend),
		SIPDevicesGateway:                          sipdevicesgateway.NewService(backend),
		TeamsRemoteLogCollectionConfigurations:     teamsremotelogcollectionconfiguration.NewService(scope),
		TeamsRemoteLogCollectionDevices:            teamsremotelogcollectiondevice.NewService(scope),
		Migration:                                  migration.NewService(backend),
		AutoAttendants:                             autoattendants.NewService(backend),
		TargetingPolicies:                          targetingpolicy.NewService(backend),
		TargetingHierarchy:                         hierarchy,
		BatchDeployments:                           batchdeployment.NewService(backend),
		PrivateChannelMigration:                    privatechannelmigration.NewService(migrationScope),
		Regions:                                    region.NewService(regionScope),
		UnifiedApps:                                unifiedapps.NewService(appCatalogScope),

		TeamsFirstPartyMeetingTemplateConfigurations: teamsfirstpartymeetingtemplateconfiguration.NewService(scope),
	}, nil
}

// Scope returns the backend dependency shared by this collection's resources.
func (s *Services) Scope() *client.Scope { return s.scope }
