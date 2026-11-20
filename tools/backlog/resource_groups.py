"""Explicit domain grouping of the pinned command candidates for ticket ownership."""
import re

# Area | resource family | exact PowerShell nouns (case-insensitive).
GROUPS = '''
CORE|teams|Team TeamArchivedState TeamPicture
CORE|team_members|TeamUser
CORE|channels|TeamChannel TeamAllChannel TeamIncomingChannel
CORE|channel_members|TeamChannelUser
CORE|shared_channels|AssociatedTeam SharedWithTeam SharedWithTeamUser
CORE|team_templates|CsTeamTemplate CsTeamTemplateList CsTeamsHiddenTemplate CsTeamsHiddenMeetingTemplate
CORE|targeting|TeamTargetingHierarchy TeamTargetingHierarchyStatus CsTeamsTargetingPolicy
CORE|operations|Operation
CORE|deployment_migration|CsBatchTeamsDeployment CsBatchTeamsDeploymentStatus TenantPrivateChannelMigrationStatus MultiGeoRegion
POL|policy_assignments|CsGroupPolicyAssignment CsGroupPolicyAssignments CsUserPolicyAssignment CsBatchPolicyAssignmentOperation CsBatchOperationDefinition CsBatchOperationStatus CsRefreshGroupUsers CsReprocessBatchOperation CsReprocessGroupPolicyAssignment GPAGroupMembers GPAUserMembership
POL|policy_packages|CsPolicyPackage CsCustomPolicyPackage CsUserPolicyPackage CsUserPolicyPackageRecommendation CsGroupPolicyPackageAssignment CsBatchPolicyPackageAssignmentOperation
POL|assignment_migration|ClearDirectToGroupAssignmentMigration StartDirectToGroupAssignmentMigration DirectToGroupAssignmentsMigrationStatus GroupAssignmentRecommendationsPerPolicyName GroupAssignmentRecommendationsPerPolicyType GroupPolicyAssignmentConflict
POL|meeting_branding|CsTeamsMeetingBrandingPolicy CsTeamsMeetingBrandingTheme CsTeamsMeetingBackgroundImage CsTeamsNdiAssuranceSlate
POL|recording_messages|CsTeamsRecordingAndTranscriptionCustomMessage CsTeamsRecordingAndTranscriptionLocalizationCustomMessage
POL|federation|CsTenantFederationConfiguration CsTeamsAcsFederationConfiguration CsTeamsExternalAccessConfiguration CsExternalAccessPolicy CsEdgeAllowAllKnownDomains CsEdgeAllowList CsEdgeDomainPattern
VOICE|online_users|CsOnlineUser CsUser CsOnlineVoiceUser CsOnlineSipDomain
VOICE|calling_settings|CsUserCallingSettings CsUserCallingDelegate CsOnlineVoicemailUserSettings CsPersonalAttendantSettings
VOICE|calling_line_identity|CsCallingLineIdentity
VOICE|phone_assignments|CsPhoneNumberAssignment CsPhoneNumberAssignmentBlock CsPhoneNumberPolicyAssignment CsPhoneNumberSmsActivation
VOICE|phone_tags|CsPhoneNumberTag CsPhoneNumberTenantConfiguration CsTag
VOICE|number_inventory|CsOnlineTelephoneNumber CsOnlineTelephoneNumberCountry CsOnlineTelephoneNumberType CsHybridTelephoneNumber CsAcquiredPhoneNumber CsExportAcquiredPhoneNumberStatus
VOICE|number_orders|CsOnlineTelephoneNumberOrder CsOnlineTelephoneNumberReleaseOrder CsOnlineDirectRoutingTelephoneNumberUploadOrder CsPhoneNumberUsageChangeOrder
VOICE|bulk_number_orders|CsPhoneNumberBulkUpdateDrNumberAcquiredCapabilitiesOrder CsPhoneNumberBulkUpdateLocationIdOrder CsPhoneNumberBulkUpdateNetworkSiteIdOrder CsPhoneNumberBulkUpdateReverseNumberLookupOrder CsPhoneNumberBulkUpdateTagsOrder
VOICE|pstn_gateways|CsOnlinePSTNGateway
VOICE|voice_routes|CsOnlineVoiceRoute CsOnlinePstnUsage CsOnlineVoiceRoutingPolicy
VOICE|dial_plans|CsTenantDialPlan CsEffectiveTenantDialPlan CsVoiceNormalizationRule CsTeamsTranslationRule
VOICE|blocked_numbers|CsTenantBlockedCallingNumbers CsInboundBlockedNumberPattern CsInboundExemptNumberPattern CsTeamsUnassignedNumberTreatment
VOICE|emergency_addresses|CsOnlineLisCivicAddress CsOnlineLisLocation CsOnlineEnhancedEmergencyServiceDisclaimer
VOICE|emergency_network|CsOnlineLisPort CsOnlineLisSubnet CsOnlineLisSwitch CsOnlineLisWirelessAccessPoint
VOICE|tenant_network|CsTenantNetworkRegion CsTenantNetworkSite CsTenantNetworkSubnet CsTenantTrustedIPAddress CsTenantNetworkConfiguration
VOICE|emergency_policies|CsLocationPolicy CsTeamsEmergencyCallingPolicy CsTeamsEmergencyCallRoutingPolicy CsTeamsCarrierEmergencyCallRoutingPolicy CsTeamsEmergencyNumber CsTeamsEmergencyCallingExtendedNotification
VOICE|conferencing|CsOnlineDialInConferencingBridge CsOnlineDialInConferencingLanguagesSupported CsOnlineDialInConferencingServiceNumber CsOnlineDialInConferencingTenantSettings CsOnlineDialInConferencingTenantConfiguration CsOnlineDialInConferencingUser CsOnlineDialInConferencingUserDefaultNumber CsOnlineDialinConferencingPolicy CsOnlineDialOutPolicy CsDialoutPolicy CsOnlineAudioConferencingRoutingPolicy
VOICE|survivability|CsTeamsSurvivableBranchAppliance CsTeamsSurvivableBranchAppliancePolicy
VOICE|video_interop|CsVideoInteropServiceProvider CsTeamsVideoInteropServicePolicy
CALL|call_queues|CsCallQueue
CALL|auto_attendants|CsAutoAttendant CsAutoAttendantStatus CsAutoAttendantSupportedLanguage CsAutoAttendantSupportedTimeZone CsAutoAttendantTenantInformation
CALL|attendant_builders|CsAutoAttendantCallableEntity CsAutoAttendantCallFlow CsAutoAttendantCallHandlingAssociation CsAutoAttendantDialScope CsAutoAttendantMenu CsAutoAttendantMenuOption CsAutoAttendantPrompt
CALL|schedules|CsOnlineSchedule CsOnlineDateTimeRange CsOnlineTimeRange CsAutoAttendantHolidays
CALL|audio_prompts|CsOnlineAudioFile CsCustomPrompt CsCustomPromptPackage
CALL|resource_accounts|CsOnlineApplicationInstance CsOnlineApplicationInstanceAssociation CsOnlineApplicationInstanceAssociationStatus CsOnlineVoiceApplicationInstance
CALL|mainline_attendants|CsMainlineAttendantAppointmentBookingFlow CsMainlineAttendantFlow CsMainlineAttendantQuestionAnswerFlow CsMainlineAttendantSpamDetectionTemplate CsMainlineAttendantSupportedLanguages CsMainlineAttendantSupportedVoices CsMainlineAttendantTenantInformation
CALL|call_templates|CsAutoRecordingTemplate CsComplianceRecordingForCallQueueTemplate CsSharedCallHistoryTemplate CsSharedCallQueueHistoryTemplate CsSharedVoicemailTriageSettingsTemplate CsTagsTemplate
CALL|compliance_applications|CsTeamsComplianceRecordingApplication CsTeamsComplianceRecordingPairedApplication CsTeamsComplianceRecordingPolicy
EXT|app_catalog|TeamsApp TeamsArtifacts CsTeamsSettingsCustomApp
EXT|unified_apps|AllM365TeamsApps M365TeamsApp M365UnifiedCustomPendingApp M365UnifiedCustomPendingApps M365UnifiedTenantSettings
EXT|shifts_connections|CsTeamsShiftsConnection CsTeamsShiftsConnectionConnector CsTeamsShiftsConnectionInstance CsTeamsShiftsConnectionValidate
EXT|shifts_sync|CsTeamsShiftsConnectionBatchTeamMap CsTeamsShiftsConnectionTeamMap CsTeamsShiftsConnectionOperation CsTeamsShiftsConnectionSyncResult CsTeamsShiftsConnectionWfmTeam CsTeamsShiftsConnectionWfmUser CsTeamsShiftsConnectionErrorReport CsTeamsShiftsScheduleRecord
EXT|devices|CsSDGBulkSignInRequestsSummary CsSdgBulkSignInRequest CsSdgBulkSignInRequestStatus CsSDGDeviceTransferRequest CsSdgDeviceTaggingRequest CsTeamsRemoteLogCollectionDevice CsTeamsRemoteLogCollectionConfiguration
EXT|ai_agents|CsAgent CsAiAgents AIGeneratedKnowledge AIGeneratedKnowledgeContainer
EXT|tenant_identity|CsAadTenant CsAadUser CsCloudTenant CsCloudUser CsTenant CsOnlineDirectoryTenant CsGroup CsHostingProvider CsTenantLicensingConfiguration CsUserLicenseGracePeriod
EXT|migration|CsExMeetingMigration CsMeetingMigrationStatus CsMeetingMigrationTransactionHistory CsMoveTenantServiceInstanceTaskStatus CsAvsTenantPartition CsTenantCrossRegion CsTenantServiceInstance NgtProvInstance NgtProvInstanceFailOverStatus
EXT|provisioning_diagnostics|CsBusinessVoiceDirectoryDiagnosticData CsLoginInfo CsMasObjectChangelog CsMasVersionedSchemaData CsUserProvHistory CsUssUserSettings CsCustomHandlerCallBackNgtprov CsCustomHandlerNgtprov CsDirectObjectSync CsDirectoryObjectSync CsGenericNgtProvCommand CsMsodsSync CsRehomeuser CsTenantUserBackfill CsInternalHelper
EXT|configuration_diagnostics|CsConfiguration CsPublishPolicySchemaDefaults CsNotifyCache CsCacheOperation
EXT|cloud_call_data|CsCloudCallDataConnection
'''

SPECIAL = {
    'teams': 'Preserve group/team identity relationships, archive transitions and picture upload; observe asynchronous creation and resulting membership.',
    'channels': 'Distinguish standard/private/shared channel creation and visibility, including incoming channels and asynchronous provisioning.',
    'shared_channels': 'Preserve host/shared-with team relationships and membership visibility; test access denied and removal semantics.',
    'policy_assignments': 'Distinguish direct, group and batch assignment; retain precedence/rank, effective assignment and per-user batch outcomes.',
    'policy_packages': 'Preserve package composition and assignment semantics; explicitly test delegated-only restrictions where evidenced.',
    'assignment_migration': 'Model migration state/status and conflicts; do not infer completion from request acceptance or erase existing assignments.',
    'meeting_branding': 'Include 8.1.0 branding display/chat/reaction fields and asset upload behavior; preserve URI and image variants.',
    'phone_assignments': 'Preserve telephone number type, assignment/removal and block semantics; distinguish partial success from completion.',
    'number_orders': 'Include New-CsPhoneNumberUsageChangeOrder; preserve ordering, completion/release and polling semantics without unsafe automatic replay.',
    'bulk_number_orders': 'Preserve order identifiers, item-level failures, polling and recovery without repeating successful changes.',
    'pstn_gateways': 'Include the 8.1.0 InviteWithReplaces value and read-back; preserve gateway identity and SIP configuration semantics.',
    'dial_plans': 'Implement local normalization rule builders and effective-plan/test operations; preserve rule order and matching semantics.',
    'emergency_addresses': 'Preserve address validation and civic-address/location relationships; evidence country-specific fields and disclaimer handling.',
    'conferencing': 'Preserve bridge/service-number/user relationships, language choices and registration semantics; resolve missing runtime candidates before implementation.',
    'call_queues': 'Model nested agent/routing/overflow/timeout settings; Get-CsCallQueue no longer returns Statistics in the 8.1.0 release notes.',
    'auto_attendants': 'Use attendant builders, resource accounts, schedules and prompt contracts; preserve nested flows and update behaviour.',
    'attendant_builders': 'Implement native local constructors and validation; no network call is introduced for a local PowerShell builder.',
    'schedules': 'Preserve recurrence, time zones, date/time intervals and holiday import/export; verify daylight-saving and boundary cases against evidence.',
    'audio_prompts': 'Preserve audio bytes, content types, import/export identifiers and custom prompt composition; test malformed formats and streaming failures.',
    'resource_accounts': 'Preserve application identity, association state and synchronization; model authentication restrictions independently for each operation.',
    'app_catalog': 'Preserve package bytes, upload/update semantics, catalog identifiers and artifact retrieval; avoid treating upload acceptance as full propagation.',
    'shifts_connections': 'Preserve connector-specific configuration, validation and connection-instance lifecycles; test supported authentication modes per command.',
    'shifts_sync': 'Model asynchronous synchronization, mappings, reports and partial errors; distinguish external workforce IDs from Teams IDs.',
    'provisioning_diagnostics': 'Resolve public versus privileged/conditional support commands explicitly. Unsupported caller permissions are distinct from absent SDK implementation.',
    'configuration_diagnostics': 'Separate local cache effects from remote configuration operations; do not expose internal plumbing as invented REST operations.',
    'devices': 'Preserve device/tag/transfer/bulk-sign-in identities and statuses; test partial failure and avoid logging device sign-in secrets.',
    'ai_agents': 'Preserve agent/knowledge identifiers, deletion scope and version-specific parameters; distinguish local builders from remote mutations.',
}


def snake(noun):
    return re.sub(r'(?<!^)(?=[A-Z][a-z])', '_', re.sub(r'^Cs', '', noun)).lower()


def classify(commands):
    lookup = {}
    for line in GROUPS.strip().splitlines():
        area, family, nouns = line.split('|')
        for noun in nouns.split():
            if noun.casefold() in lookup:
                raise ValueError('Duplicate noun ' + noun)
            lookup[noun.casefold()] = (area, family)
    groups = {}
    unmapped = []
    sessions = []
    excluded = []
    for cmd in commands:
        noun = cmd['noun']
        if noun.casefold() == 'csinternalpstelemetry':
            excluded.append(cmd)
            continue
        if noun in ('MicrosoftTeams', 'TeamsEnvironmentConfig', 'CsOCEContext', 'CsRegionContext'):
            sessions.append(cmd)
            continue
        key = lookup.get(noun.casefold())
        if key is None and (noun.endswith(('Policy', 'Configuration')) or noun == 'CsTeamsCustomBannerText'):
            key = ('POL', snake(noun))
        if key is None:
            unmapped.append(cmd['name'])
            continue
        groups.setdefault(key, []).append(cmd)
    if unmapped:
        raise ValueError('Unmapped commands: ' + ', '.join(unmapped))
    return groups, sessions, excluded
