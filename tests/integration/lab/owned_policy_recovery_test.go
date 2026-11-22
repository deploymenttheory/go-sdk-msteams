package lab_test

import (
	"context"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

// TestAccRecoverOwnedPolicies uses an explicitly selected, identity-bound journal.
// The handlers verify SDK fixture ownership and independent deletion/absence.
func TestAccRecoverOwnedPolicies(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit owned-policy recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for recovery")
	}
	sdk, binding, _ := labSDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	handlers := map[string]lab.Handler{
		"tenant-abuse-configuration": tenantAbuseConfigurationCleanup{sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations},
		"update-management-policy":   updateManagementPolicyCleanup{sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies},
		"template-permission-policy": templatePermissionPolicyCleanup{sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies},
		"sip-devices-configuration":  sipDevicesConfigurationCleanup{sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations},
		"shifts-policy":              shiftsPolicyCleanup{sdk.TeamsAdminAPI.TeamsShiftsPolicies},
		"shifts-app-configuration":   configurationRestore{shiftsAppSingleton(sdk)},
		"room-video-policy":          roomVideoPolicyCleanup{sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies},
		"shared-calling-policy":      sharedCallingPolicyCleanup{sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies},
		"personal-policy":            personalPolicyCleanup{sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies},
		"recording-rollout-policy":   recordingRollOutPolicyCleanup{sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies},
		"notification-policy":        notificationPolicyCleanup{sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies},
		"roaming-policy":             roamingPolicyCleanup{sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies},
		"multi-tenant-configuration": multiTenantConfigurationCleanup{sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations},
		"mobility-policy":            mobilityPolicyCleanup{sdk.TeamsAdminAPI.TeamsMobilityPolicies},
		"messaging-configuration":    configurationRestore{messagingSingleton(sdk)},
		"migration-configuration":    configurationRestore{migrationSingleton(sdk)},
		"team-template":              templateCleanup{sdk.TeamsAdminAPI.TeamTemplates},
		"branding-policy":            brandingPolicyCleanup{sdk.TeamsAdminAPI.MeetingBrandingPolicies},
		"app-permission-policy":      appPermissionPolicyCleanup{sdk.TeamsAdminAPI.TeamsAppPermissionPolicies},
		"app-setup-policy":           appSetupPolicyCleanup{sdk.TeamsAdminAPI.TeamsAppSetupPolicies},
		"voicemail-policy":           voicemailPolicyCleanup{sdk.TeamsAdminAPI.OnlineVoicemailPolicies},
		"audio-policy":               audioPolicyCleanup{sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies},
		"feedback-policy":            feedbackPolicyCleanup{sdk.TeamsAdminAPI.TeamsFeedbackPolicies},
		"files-policy":               filesPolicyCleanup{sdk.TeamsAdminAPI.TeamsFilesPolicies},
		"application-access-policy":  policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies},
		"external-access-policy":     externalPolicyCleanup{sdk.TeamsAdminAPI.ExternalAccessPolicies},
		"call-hold-policy":           callHoldPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallHoldPolicies},
		"call-park-policy":           callParkPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallParkPolicies},
		"channels-policy":            channelsPolicyCleanup{sdk.TeamsAdminAPI.TeamsChannelsPolicies},
		"cortana-policy":             cortanaPolicyCleanup{sdk.TeamsAdminAPI.TeamsCortanaPolicies},
		"template-baseline-policy":   meetingTemplatePolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies},
		"meeting-template-policy":    meetingTemplatePolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies},
		"meeting-policy":             meetingPolicyCleanup{sdk.TeamsAdminAPI.TeamsMeetingPolicies},
		"messaging-policy":           messagingPolicyCleanup{sdk.TeamsAdminAPI.TeamsMessagingPolicies},
		"events-policy":              eventsPolicyCleanup{sdk.TeamsAdminAPI.TeamsEventsPolicies},
		"enhanced-encryption-policy": enhancedEncryptionPolicyCleanup{sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies},
		"calling-policy":             callingPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallingPolicies},
	}
	if os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE") != "" {
		fixture := loadAssignmentUserFixture(t, binding)
		handlers["app-access-user-assignment"] = userAssignmentCleanup{sdk, fixture.UserID}
	}
	if os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE") != "" {
		fixture := loadAssignmentGroupFixture(t, binding)
		handler := groupAssignmentCleanup{sdk, fixture.GroupID}
		handlers["app-access-group-assignment"] = handler
		handlers["group-app-access-policy"] = groupOwnedPolicyCleanup{policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, handler}
	}
	_, err = journal.Recover(ctx, handlers)
	if err != nil {
		t.Fatal("owned-policy recovery incomplete; private journal retained")
	}
	t.Log("owned-policy recovery verified")
}
