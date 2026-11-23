package labkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func ConfigurationReadEnabled(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid configuration lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("configuration reads require explicit read-only enablement")
	} else if err != nil {
		t.Fatal("invalid configuration lab enablement")
	}
	return cfg
}

func TenantConfigurationOracle(t *testing.T, ctx context.Context, kind, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": kind})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := OracleCommand(ctx, "read-messaging-migration-configuration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned tenant configuration read process failed")
	}
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		t.Fatal("pinned tenant configuration projection unavailable")
	}
	names := []string{"EnableVideoMessageCaptions", "EnableInOrganizationChatControl", "CustomEmojis", "Storyline", "Communities", "MessagingNotes", "FileTypeCheck", "UrlReputationCheck", "ContentBasedPhishingCheck", "ReportIncorrectSecurityDetections", "FileTypeCustomizationSetting", "FileTypeBlockList"}
	switch kind {
	case "upgrade-configuration":
		names = []string{"DownloadTeams", "SfBMeetingJoinUx", "BlockLegacyAuthorization"}
	case "upgrade-policy":
		names = []string{"Description", "Mode", "NotifySfbUsers", "Action"}
	case "tenant-abuse":
		names = []string{"LowSeatLimit", "TrialTenantValidation", "SkipValidation", "CreateThreadThresholdPerSeat", "AddMemberThresholdPerSeat", "SendMessageThresholdPerSeat", "CreateThreadThresholdForTrialTenants", "AddMemberThresholdForTrialTenants", "SendMessageThresholdForTrialTenants", "BlockedTenants"}
	case "update-management":
		names = []string{"Description", "AllowManagedUpdates", "AllowPreview", "UpdateDayOfWeek", "UpdateTime", "UpdateTimeOfDay", "AllowPublicPreview", "UseNewTeamsClient", "BlockLegacyAuthorization", "DisabledInProductMessages"}
	case "sip-devices":
		names = []string{"BulkSignIn"}
	case "template-permission":
		names = []string{"Description", "HiddenTemplates"}
	case "shifts":
		names = []string{"ShiftNoticeFrequency", "ShiftNoticeMessageType", "ShiftNoticeMessageCustom", "AccessType", "AccessGracePeriodMinutes", "EnableScheduleOwnerPermissions"}
	case "shifts-app":
		names = []string{"AllowTimeClockLocationDetection"}
	case "migration":
		names = []string{"EnableLegacyClientInterop"}
	case "mobility":
		names = []string{"Description", "IPAudioMobileMode", "IPVideoMobileMode", "MobileDialerPreference", "LinksInTeams"}
	case "room-video":
		names = []string{"Description", "Enabled", "AreaCode", "ReceiveExternalCalls", "ReceiveInternalCalls", "PlaceExternalCalls", "PlaceInternalCalls"}
	case "shared-calling":
		names = []string{"Description", "ResourceAccount", "EmergencyNumbers"}
	case "personal":
		names = []string{"PersonalAttendant", "CallScreening", "CalendarBookings", "InboundInternalCalls", "InboundFederatedCalls", "InboundPSTNCalls", "AutomaticTranscription", "AutomaticRecording"}
	case "recording-rollout":
		names = []string{"MeetingRecordingOwnership"}
	case "roaming":
		names = []string{"Description", "AllowIPVideo", "MediaBitRateKb"}
	case "notification":
		names = []string{"Description", "SuggestedFeedsEnabledType", "TrendingFeedsEnabledType"}
	case "multi-tenant":
		names = []string{"CopilotFromHomeTenant"}
	}
	if len(result.Settings) != len(names) {
		t.Fatal("incomplete configuration getter projection")
	}
	for _, name := range names {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("missing configuration getter")
		}
	}
	return result.Settings
}

type CompiledScalarView struct {
	Identity *string
	Settings map[string]any
	Fields   map[string]json.RawMessage
}

type CompiledScalarRead func(context.Context) (*CompiledScalarView, *resty.Response, error)

func RunCompiledScalarGlobal(t *testing.T, kind, family string, read func(*msteams.Client) CompiledScalarRead, list func(context.Context, *msteams.Client) ([]*CompiledScalarView, *resty.Response, error)) {
	t.Helper()
	cfg := ConfigurationReadEnabled(t)
	oracle := OracleEnabled(t)
	sdk, binding, _ := SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	v, response, err := read(sdk)(ctx)
	WriteResponseFixture(t, family, "get_global", response, nil)
	if err != nil {
		LogAccessResponse(t, response, err)
		t.Fatal("compiled scalar Global read failed")
	}
	if v == nil || v.Identity == nil || !strings.EqualFold(*v.Identity, "Global") {
		t.Fatal("compiled scalar Global identity differs")
	}
	values, response, err := list(ctx, sdk)
	WriteResponseFixture(t, family, "list", response, nil)
	if err != nil {
		LogAccessResponse(t, response, err)
		t.Fatal("compiled scalar Global list request failed")
	}
	if len(values) != 1 || values[0] == nil {
		t.Fatal("compiled scalar Global list selection differs")
	}
	if !SameConfigurationFields(v.Fields, values[0].Fields) {
		reason := ResponseComparisonFailure(v.Fields, values[0].Fields)
		t.Log(reason)
		LogMetadataMismatches(t, v.Fields, values[0].Fields)
		t.Fatal("compiled scalar Global read/list differ")
	}
	if !SameProjection(v.Settings, values[0].Settings) {
		t.Fatal("compiled scalar Global typed getters differ")
	}

	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !SameProjection(v.Settings, TenantConfigurationOracle(t, ctx, kind, tenant)) {
			t.Fatal("Go and PowerShell compiled scalar getters differ")
		}
	}
	t.Log("compiled scalar Global read/list and pinned getter comparison verified")
}

// These operations are serialized with every tenant mutation in CI. Recovery
// persists only the selected Boolean and restores it through the same public API.
type SingletonConfiguration struct {
	Kind, Field string
	Get         func(context.Context) (map[string]any, *resty.Response, error)
	Set         func(context.Context, bool) (*resty.Response, error)
}

type ConfigurationRestore struct{ SingletonConfiguration }

func (ConfigurationRestore) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (ConfigurationRestore) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func (h ConfigurationRestore) Restore(ctx context.Context, entry lab.Entry) error {
	var original struct {
		Value *bool `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&original) != nil || d.Decode(new(any)) != io.EOF || original.Value == nil {
		return lab.ErrUnresolved
	}
	// A previous restoration may have succeeded before its receipt was saved.
	current, _, err := h.Get(ctx)
	if err != nil {
		return err
	}
	if value, ok := current[h.Field].(*bool); ok && value != nil && *value == *original.Value {
		return nil
	}
	if _, err := h.Set(ctx, *original.Value); err != nil {
		return err
	}
	return ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		current, _, err := h.Get(ctx)
		value, ok := current[h.Field].(*bool)
		return err == nil && ok && value != nil && *value == *original.Value, err
	})
}

func RunSingletonUpdate(t *testing.T, build func(*msteams.Client) SingletonConfiguration) {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid configuration mutation enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("configuration updates require explicit mutation enablement and recovery directory")
	} else if err != nil {
		t.Fatal("invalid configuration mutation enablement")
	}
	oracle := OracleEnabled(t)
	sdk, binding, _ := SDKClient(t)
	s := build(sdk)
	fixturePackage := "teams_" + s.Kind + "_configuration"
	if s.Kind == "shifts-app" {
		fixturePackage = "teams_shifts_app_policy"
	}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private configuration recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{s.Kind + "-configuration": ConfigurationRestore{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("configuration recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.Get(ctx)
	if err != nil {
		LogAccessResponse(t, response, err)
		t.Fatal("configuration original read failed")
	}
	value, ok := before[s.Field].(*bool)
	if !ok || value == nil {
		t.Fatal("configuration original Boolean is unavailable; update refused")
	}
	compare := func(settings map[string]any) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !SameProjection(settings, TenantConfigurationOracle(t, ctx, s.Kind, tenant)) {
			t.Fatal("Go and PowerShell configuration update getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]bool{"value": *value})
	if err != nil {
		t.Fatal("configuration original Boolean could not be journaled")
	}
	entry, err := journal.PlanRestore(s.Kind+"-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("configuration restoration intent not journaled")
	}
	response, err = s.Set(ctx, !*value)
	WriteResponseFixture(t, fixturePackage, "update", response, nil)
	if err != nil {
		LogAccessResponse(t, response, err)
		t.Fatal("configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("configuration update receipt not journaled")
	}
	want := make(map[string]any, len(before))
	for k, v := range before {
		want[k] = v
	}
	changed := !*value
	want[s.Field] = &changed
	var updated map[string]any
	if err := ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, r, err := s.Get(ctx)
		matched := err == nil && SameProjection(want, read)
		if matched {
			updated = read
			WriteResponseFixture(t, fixturePackage, "update_readback", r, nil)
		}
		return matched, err
	}); err != nil {
		t.Fatal("configuration update or omitted-field preservation did not converge")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("configuration original Boolean restoration failed")
	}
	restored, response, err := s.Get(ctx)
	if err != nil || !SameProjection(before, restored) {
		t.Fatal("configuration original getter state was not independently restored")
	}
	WriteResponseFixture(t, fixturePackage, "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}
