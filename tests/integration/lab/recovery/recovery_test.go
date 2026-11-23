// Package recoveryacc recovers owned lab resources from a retained journal.
// One journal can hold entries from several service partitions, and
// Journal.Recover refuses unknown kinds, so every service's handlers are
// merged here.
package recoveryacc

import (
	"context"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	applicationaccesspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/application_access_policy"
	audiopromptsacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/audio_prompts"
	callqueuesacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/call_queues"
	externalaccesspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	meetingbrandingacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/meeting_branding"
	onlinevoicemailpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/online_voicemail_policy"
	schedulesacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/schedules"
	teamtemplateacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/team_template"
	teamsapppermissionpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_app_permission_policy"
	teamsappsetuppolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_app_setup_policy"
	teamsaudioconferencingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_audio_conferencing_policy"
	teamscallholdpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_call_hold_policy"
	teamscallparkpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_call_park_policy"
	teamscallingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_calling_policy"
	teamschannelspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_channels_policy"
	teamscortanapolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_cortana_policy"
	teamsenhancedencryptionpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_enhanced_encryption_policy"
	teamseventspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_events_policy"
	teamsfeedbackpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_feedback_policy"
	teamsfilespolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_files_policy"
	teamsmeetingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_meeting_policy"
	teamsmeetingtemplatepermissionpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_meeting_template_permission_policy"
	teamsmessagingconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_messaging_configuration"
	teamsmessagingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_messaging_policy"
	teamsmigrationconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_migration_configuration"
	teamsmobilitypolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_mobility_policy"
	teamsmultitenantorganizationconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_multi_tenant_organization_configuration"
	teamsnetworkroamingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_network_roaming_policy"
	teamsnotificationandfeedspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_notification_and_feeds_policy"
	teamspersonalattendantpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_personal_attendant_policy"
	teamsrecordingrolloutpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_recording_roll_out_policy"
	teamsroomvideoteleconferencingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_room_video_tele_conferencing_policy"
	teamssharedcallingroutingpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_shared_calling_routing_policy"
	teamsshiftsapppolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_shifts_app_policy"
	teamsshiftspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_shifts_policy"
	teamssipdevicesconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_sip_devices_configuration"
	teamstemplatepermissionpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_template_permission_policy"
	teamstenantabuseconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_tenant_abuse_configuration"
	teamsupdatemanagementpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_update_management_policy"
	teamsupgradeconfigurationacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_upgrade_configuration"
	teamsvdipolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_vdi_policy"
	teamsvirtualappointmentspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_virtual_appointments_policy"
	teamsvoiceapplicationspolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_voice_applications_policy"
	teamsworkloadpolicyacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/teams_work_load_policy"
)

// handlerSets lists every service's recovery handlers.
func handlerSets(sdk *msteams.Client) []map[string]lab.Handler {
	return []map[string]lab.Handler{
		applicationaccesspolicyacc.RecoveryHandlers(sdk),
		audiopromptsacc.RecoveryHandlers(sdk),
		callqueuesacc.RecoveryHandlers(sdk),
		externalaccesspolicyacc.RecoveryHandlers(sdk),
		meetingbrandingacc.RecoveryHandlers(sdk),
		schedulesacc.RecoveryHandlers(sdk),
		onlinevoicemailpolicyacc.RecoveryHandlers(sdk),
		teamtemplateacc.RecoveryHandlers(sdk),
		teamsapppermissionpolicyacc.RecoveryHandlers(sdk),
		teamsappsetuppolicyacc.RecoveryHandlers(sdk),
		teamsaudioconferencingpolicyacc.RecoveryHandlers(sdk),
		teamscallholdpolicyacc.RecoveryHandlers(sdk),
		teamscallparkpolicyacc.RecoveryHandlers(sdk),
		teamscallingpolicyacc.RecoveryHandlers(sdk),
		teamschannelspolicyacc.RecoveryHandlers(sdk),
		teamscortanapolicyacc.RecoveryHandlers(sdk),
		teamsenhancedencryptionpolicyacc.RecoveryHandlers(sdk),
		teamseventspolicyacc.RecoveryHandlers(sdk),
		teamsfeedbackpolicyacc.RecoveryHandlers(sdk),
		teamsfilespolicyacc.RecoveryHandlers(sdk),
		teamsmeetingpolicyacc.RecoveryHandlers(sdk),
		teamsmeetingtemplatepermissionpolicyacc.RecoveryHandlers(sdk),
		teamsmessagingconfigurationacc.RecoveryHandlers(sdk),
		teamsmessagingpolicyacc.RecoveryHandlers(sdk),
		teamsmigrationconfigurationacc.RecoveryHandlers(sdk),
		teamsmobilitypolicyacc.RecoveryHandlers(sdk),
		teamsmultitenantorganizationconfigurationacc.RecoveryHandlers(sdk),
		teamsnetworkroamingpolicyacc.RecoveryHandlers(sdk),
		teamsnotificationandfeedspolicyacc.RecoveryHandlers(sdk),
		teamspersonalattendantpolicyacc.RecoveryHandlers(sdk),
		teamsrecordingrolloutpolicyacc.RecoveryHandlers(sdk),
		teamsroomvideoteleconferencingpolicyacc.RecoveryHandlers(sdk),
		teamssharedcallingroutingpolicyacc.RecoveryHandlers(sdk),
		teamsshiftsapppolicyacc.RecoveryHandlers(sdk),
		teamsshiftspolicyacc.RecoveryHandlers(sdk),
		teamssipdevicesconfigurationacc.RecoveryHandlers(sdk),
		teamstemplatepermissionpolicyacc.RecoveryHandlers(sdk),
		teamstenantabuseconfigurationacc.RecoveryHandlers(sdk),
		teamsupdatemanagementpolicyacc.RecoveryHandlers(sdk),
		teamsupgradeconfigurationacc.RecoveryHandlers(sdk),
		teamsvdipolicyacc.RecoveryHandlers(sdk),
		teamsvirtualappointmentspolicyacc.RecoveryHandlers(sdk),
		teamsvoiceapplicationspolicyacc.RecoveryHandlers(sdk),
		teamsworkloadpolicyacc.RecoveryHandlers(sdk),
	}
}

// mergeHandlers combines handler sets and rejects a kind registered twice.
func mergeHandlers(t *testing.T, sets ...map[string]lab.Handler) map[string]lab.Handler {
	t.Helper()
	handlers := map[string]lab.Handler{}
	for _, set := range sets {
		for kind, handler := range set {
			if _, duplicate := handlers[kind]; duplicate {
				t.Fatalf("recovery kind %q is registered by two services", kind)
			}
			handlers[kind] = handler
		}
	}
	return handlers
}

func TestUnitRecoveryRegistryKinds(t *testing.T) {
	t.Parallel()
	// Handler construction only wraps service clients; nothing is sent.
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ClientSecret, TenantID: "00000000-0000-0000-0000-000000000001",
		ClientID: "00000000-0000-0000-0000-000000000002", ClientSecret: "offline"})
	if err != nil {
		t.Fatal(err)
	}
	handlers := mergeHandlers(t, handlerSets(sdk)...)
	want := []string{
		"app-permission-policy",
		"app-setup-policy",
		"application-access-policy",
		"audio-file",
		"audio-policy",
		"branding-policy",
		"call-hold-policy",
		"call-park-policy",
		"call-queue",
		"calling-policy",
		"channels-policy",
		"cortana-policy",
		"enhanced-encryption-policy",
		"events-policy",
		"external-access-policy",
		"feedback-policy",
		"files-policy",
		"meeting-policy",
		"meeting-template-policy",
		"messaging-configuration",
		"messaging-policy",
		"migration-configuration",
		"mobility-policy",
		"multi-tenant-configuration",
		"notification-policy",
		"personal-policy",
		"recording-rollout-policy",
		"roaming-policy",
		"room-video-policy",
		"schedule",
		"shared-calling-policy",
		"shifts-app-configuration",
		"shifts-policy",
		"sip-devices-configuration",
		"team-template",
		"template-baseline-policy",
		"template-permission-policy",
		"tenant-abuse-configuration",
		"update-management-policy",
		"upgrade-configuration",
		"vdi-policy",
		"virtual-appointments-policy",
		"voice-applications-policy",
		"voicemail-policy",
		"work-load-policy",
	}
	if len(handlers) != len(want) {
		t.Fatalf("recovery registry has %d kinds, want %d", len(handlers), len(want))
	}
	for _, kind := range want {
		if handlers[kind] == nil {
			t.Errorf("recovery kind %q is not registered", kind)
		}
	}
}

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
	sdk, binding, _ := labkit.SDKClient(t)
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
	handlers := mergeHandlers(t, append(handlerSets(sdk), applicationaccesspolicyacc.FixtureRecoveryHandlers(t, sdk, binding))...)
	_, err = journal.Recover(ctx, handlers)
	if err != nil {
		t.Fatal("owned-policy recovery incomplete; private journal retained")
	}
	t.Log("owned-policy recovery verified")
}
