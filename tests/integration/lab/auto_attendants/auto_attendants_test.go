package autoattendantsacc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "auto_attendants"
	// probe is a reserved fixture-form ID that names no attendant, schedule or group.
	probe = "00000000-0000-4000-8000-000000000001"
	// fixtureAttendant replaces the live attendant ID in captured fixtures.
	fixtureAttendant = "00000000-0000-4000-8000-000000000003"
	// fixtureDefaultFlow and fixtureAfterHoursFlow replace live call flow IDs.
	fixtureDefaultFlow    = "00000000-0000-4000-8000-000000000004"
	fixtureAfterHoursFlow = "00000000-0000-4000-8000-000000000005"
	// pstn is a reserved fictional North American number (555-0100..0199).
	pstn = "tel:+14255550100"
)

// oracleRun runs read-auto-attendant.ps1 with the selected mode and decodes
// its JSON report.
func oracleRun(t *testing.T, ctx context.Context, extra ...string) map[string]json.RawMessage {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-auto-attendant.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if v, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+v)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"),
		"MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	command.Env = append(command.Env, extra...)
	var output labkit.OracleBuffer
	command.Stdout, command.Stderr = &output, io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell auto attendant read failed; no output retained")
	}
	var observed map[string]json.RawMessage
	if json.Unmarshal(output.Data, &observed) != nil || string(observed["success"]) != "true" {
		t.Fatal("pinned PowerShell could not complete the auto attendant read")
	}
	return observed
}

func same(observed map[string]json.RawMessage, key string, want any) bool {
	raw, err := json.Marshal(want)
	return err == nil && string(observed[key]) == string(raw)
}

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

func newGUID(t *testing.T) string {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	b[6], b[8] = (b[6]&0x0f)|0x40, (b[8]&0x3f)|0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func greeting(text string) aa.Prompt {
	return aa.Prompt{ActiveType: new(aa.PromptTextToSpeech), TextToSpeechPrompt: &text}
}

// TestAccAutoAttendantReads reads the supported languages, time zones and
// tenant information, checks the missing-identity answers, and validates
// builder values on the draft routes, which store nothing.
func TestAccAutoAttendantReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AutoAttendants
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := tenantReplacements(t, binding)

	info, response, err := service.GetTenantInformation(ctx)
	if err != nil || info.DefaultLanguage == nil || info.DefaultTimeZone == nil || len(info.SupportedLanguages) == 0 || len(info.SupportedTimeZones) == 0 {
		t.Fatal("tenant information read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "information", response, replacements)
	languages, _, err := service.ListSupportedLanguages(ctx)
	if err != nil || len(languages) != len(info.SupportedLanguages) {
		t.Fatal("supported language list failed")
	}
	zones, _, err := service.ListSupportedTimeZones(ctx)
	if err != nil || len(zones) != len(info.SupportedTimeZones) {
		t.Fatal("supported time zone list failed")
	}
	language, response, err := service.GetSupportedLanguageByID(ctx, "en-US")
	if err != nil || language.ID == nil || *language.ID != "en-US" || len(language.Voices) == 0 {
		t.Fatal("supported language read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "language", response, replacements)
	if _, response, err = service.GetSupportedLanguageByID(ctx, "xx-XX"); !missing(err) {
		t.Fatal("unsupported language was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "language_missing", response, replacements)
	zone, response, err := service.GetSupportedTimeZoneByID(ctx, "UTC")
	if err != nil || zone.DisplayName == nil {
		t.Fatal("supported time zone read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "time_zone", response, replacements)
	if _, _, err = service.GetSupportedTimeZoneByID(ctx, "Nowhere"); !missing(err) {
		t.Fatal("unknown time zone was not a 404")
	}
	t.Log("auto attendant languages, time zones and tenant information: passed")

	if _, _, err := service.List(ctx, &aa.ListOptions{First: new(int32(5)), ExcludeContent: new(true), IncludeStatus: new(true)}); err != nil {
		t.Fatal("auto attendant list failed")
	}
	if _, response, err = service.GetByID(ctx, probe); !missing(err) {
		t.Fatal("missing auto attendant was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, replacements)
	if _, _, err = service.GetStatusByID(ctx, probe, &aa.StatusOptions{IncludeResources: []string{"AudioFile"}}); !missing(err) {
		t.Fatal("missing auto attendant status was not a 404")
	}
	if _, _, err = service.RefreshByID(ctx, probe); !missing(err) {
		t.Fatal("missing auto attendant refresh was not a 404")
	}
	if _, err = service.DeleteByID(ctx, probe); !missing(err) {
		t.Fatal("missing auto attendant delete was not a 404")
	}
	t.Log("auto attendant missing identity: passed")

	// Draft routes validate builder values and store nothing.
	prompt, response, err := service.DraftPrompt(ctx, greeting("Welcome to the SDK probe"))
	if err != nil || *prompt.ActiveType != aa.PromptTextToSpeech {
		t.Fatal("prompt draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_prompt", response, replacements)
	if _, response, err = service.DraftPrompt(ctx, aa.Prompt{ActiveType: new(aa.PromptNone)}); !invalid(err) {
		t.Fatal("empty prompt was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_prompt_invalid", response, replacements)
	target, response, err := service.DraftCallableEntity(ctx, aa.CallableEntity{ID: new(pstn), Type: new(aa.EntityExternalPstn)})
	if err != nil || *target.ID != pstn {
		t.Fatal("callable entity draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_callable_entity", response, replacements)
	option, response, err := service.DraftMenuOption(ctx, aa.MenuOption{Action: new(aa.ActionTransferCallToTarget), DtmfResponse: new(aa.DtmfTone1), CallTarget: target})
	if err != nil || *option.DtmfResponse != aa.DtmfTone1 || option.CallTarget == nil {
		t.Fatal("menu option draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_menu_option", response, replacements)
	menu, response, err := service.DraftMenu(ctx, aa.Menu{Name: new("Default menu"), Prompts: []aa.Prompt{*prompt}, MenuOptions: []aa.MenuOption{*option}, DirectorySearchMethod: new(aa.SearchNone)})
	if err != nil || len(menu.MenuOptions) != 1 {
		t.Fatal("menu draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_menu", response, replacements)
	flow, response, err := service.DraftCallFlow(ctx, aa.CallFlow{Name: new("Default call flow"), Greetings: []aa.Prompt{*prompt}, Menu: menu})
	if err != nil || flow.ID == nil || !attendantID.MatchString(*flow.ID) {
		t.Fatal("call flow draft did not assign an ID")
	}
	replacements[*flow.ID] = fixtureDefaultFlow
	labkit.WriteResponseFixture(t, fixtureService, "draft_call_flow", response, replacements)
	association, response, err := service.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: new(aa.AssociationAfterHours), ScheduleID: new(probe), CallFlowID: flow.ID})
	if err != nil || association.Enabled == nil || !*association.Enabled {
		t.Fatal("call handling association draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_call_handling_association", response, replacements)
	if _, response, err = service.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: new(aa.AssociationHoliday), ScheduleID: new(probe),
		CallFlowID: flow.ID, Enabled: new(false)}); !invalid(err) {
		t.Fatal("disabled holiday association was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_call_handling_association_invalid", response, replacements)
	if _, response, err = service.DraftGroupDialScope(ctx, []string{probe}); !missing(err) {
		t.Fatal("unknown dial scope group was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "draft_dial_scope_missing", response, replacements)
	t.Log("auto attendant builder drafts: passed")

	if labkit.OracleEnabled(t) {
		observed := oracleRun(t, ctx, "MSTEAMS_LAB_AUTO_ATTENDANT_MODE=reads")
		if !same(observed, "default_language", *info.DefaultLanguage) || !same(observed, "default_time_zone", *info.DefaultTimeZone) ||
			!same(observed, "language_count", len(languages)) || !same(observed, "time_zone_count", len(zones)) ||
			!same(observed, "en_us_voices", len(language.Voices)) || !same(observed, "utc_display_name", *zone.DisplayName) {
			t.Fatal("PowerShell and Go auto attendant reads differ")
		}
		t.Log("PowerShell 8.1.0 / Go auto attendant read parity: passed")
	}
}

// TestAccAutoAttendantLifecycle creates one uniquely named auto attendant
// with two call flows whose IDs the test chooses, reads and lists it,
// replaces it from a read-back, refreshes it and deletes it. It needs no
// resource account, phone number or licence and never touches a
// pre-existing attendant.
func TestAccAutoAttendantLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("auto attendant lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AutoAttendants
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private auto attendant recovery journal unavailable")
	}
	cleanup := attendantCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"auto-attendant": cleanup}); err != nil {
			t.Error("auto attendant cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private auto attendant journal close failed")
		}
	})
	fixture, err := journal.FixtureName("attendant")
	if err != nil {
		t.Fatal("auto attendant fixture name unavailable")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := tenantReplacements(t, binding)
	replacements[fixture] = "msteams-sdk-fixture-attendant"
	defaultFlow, afterHours := newGUID(t), newGUID(t)
	replacements[defaultFlow], replacements[afterHours] = fixtureDefaultFlow, fixtureAfterHoursFlow

	transfer := aa.MenuOption{Action: new(aa.ActionTransferCallToTarget), DtmfResponse: new(aa.DtmfTone1),
		CallTarget: &aa.CallableEntity{ID: new(pstn), Type: new(aa.EntityExternalPstn)}}
	disconnect := aa.MenuOption{Action: new(aa.ActionDisconnectCall), DtmfResponse: new(aa.DtmfAutomatic)}
	settings := aa.Settings{Name: &fixture, LanguageID: new("en-US"), TimeZoneID: new("UTC"),
		DefaultCallFlow: &aa.CallFlow{ID: &defaultFlow, Name: new("Default call flow"), Greetings: []aa.Prompt{greeting("Welcome to the SDK probe")},
			Menu: &aa.Menu{Name: new("Default menu"), Prompts: []aa.Prompt{greeting("Press 1 for the SDK probe line")}, MenuOptions: []aa.MenuOption{transfer},
				DirectorySearchMethod: new(aa.SearchNone)}},
		CallFlows: &[]aa.CallFlow{{ID: &afterHours, Name: new("After hours call flow"),
			Menu: &aa.Menu{Name: new("After hours menu"), MenuOptions: []aa.MenuOption{disconnect}}}}}
	entry, err := journal.PlanCreate("auto-attendant", fixture)
	if err != nil {
		t.Fatal("auto attendant create was not journaled")
	}
	created, response, err := service.Create(ctx, settings)
	if err == nil && created != nil && created.ID != nil {
		replacements[*created.ID] = fixtureAttendant
	}
	labkit.WriteResponseFixture(t, fixtureService, "create", response, replacements)
	if err != nil || !ownedAttendant(created, fixture) {
		t.Fatal("live auto attendant create failed; recovery journal retained")
	}
	id := *created.ID
	if err := journal.Applied(entry, id); err != nil {
		t.Fatal("created auto attendant ID was not journaled")
	}
	if created.DefaultCallFlow == nil || *created.DefaultCallFlow.ID != defaultFlow || len(created.CallFlows) != 1 || *created.CallFlows[0].ID != afterHours ||
		created.VoiceID == nil || len(created.ApplicationInstances) != 0 {
		t.Fatal("create receipt lost the chosen call flows")
	}
	t.Log("auto attendant create: passed")

	value, response, err := service.GetByID(ctx, id)
	if err != nil || !ownedAttendant(value, fixture) || *value.LanguageID != "en-US" || *value.TimeZoneID != "UTC" ||
		*value.DefaultCallFlow.Greetings[0].TextToSpeechPrompt != "Welcome to the SDK probe" {
		t.Fatal("owned auto attendant read-back failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get", response, replacements)
	statuses, response, err := service.GetStatusByID(ctx, id, nil)
	if err != nil || statuses == nil {
		t.Fatal("owned auto attendant status read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "status", response, replacements)
	t.Log("auto attendant get/status: passed")

	listed, response, err := service.List(ctx, &aa.ListOptions{NameFilter: &fixture, IncludeStatus: new(true)})
	if err != nil {
		t.Fatal("auto attendant list failed")
	}
	found := 0
	for _, item := range listed {
		if item.ID != nil && *item.ID == id && ownedAttendant(item, fixture) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned auto attendant missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	t.Log("auto attendant list: passed")

	// Set-CsAutoAttendant sends the whole attendant: change the greeting and
	// time zone on a body rebuilt from the read-back and prove the rest survive.
	update := value.Settings()
	update.TimeZoneID = new("GMT Standard Time")
	update.DefaultCallFlow.Greetings = []aa.Prompt{greeting("Updated SDK probe greeting")}
	updated, response, err := service.UpdateByID(ctx, id, update)
	labkit.WriteResponseFixture(t, fixtureService, "update", response, replacements)
	if err != nil || !ownedAttendant(updated, fixture) {
		t.Fatal("live auto attendant update failed")
	}
	value, response, err = service.GetByID(ctx, id)
	if err != nil || !ownedAttendant(value, fixture) || *value.TimeZoneID != "GMT Standard Time" || *value.LanguageID != "en-US" ||
		*value.DefaultCallFlow.Greetings[0].TextToSpeechPrompt != "Updated SDK probe greeting" || len(value.CallFlows) != 1 ||
		*value.CallFlows[0].ID != afterHours || len(value.DefaultCallFlow.Menu.MenuOptions) != 1 ||
		*value.DefaultCallFlow.Menu.MenuOptions[0].CallTarget.ID != pstn {
		t.Fatal("auto attendant update was not read back with unchanged settings preserved")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_updated", response, replacements)
	if labkit.OracleEnabled(t) {
		observed := oracleRun(t, ctx, "MSTEAMS_LAB_AUTO_ATTENDANT_MODE=attendant", "MSTEAMS_LAB_AUTO_ATTENDANT_FIXTURE="+fixture, "MSTEAMS_LAB_AUTO_ATTENDANT_ID="+id)
		if !same(observed, "name", fixture) || !same(observed, "language_id", *value.LanguageID) || !same(observed, "time_zone_id", *value.TimeZoneID) ||
			!same(observed, "voice_id", *value.VoiceID) || !same(observed, "default_call_flow_id", *value.DefaultCallFlow.ID) ||
			!same(observed, "greeting", *value.DefaultCallFlow.Greetings[0].TextToSpeechPrompt) ||
			!same(observed, "menu_option_targets", []string{pstn}) || !same(observed, "call_flow_ids", []string{afterHours}) {
			t.Fatal("PowerShell and Go auto attendant observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go auto attendant parity: passed")
	}
	t.Log("auto attendant update/read-back: passed")

	refreshed, response, err := service.RefreshByID(ctx, id)
	if err != nil || !ownedAttendant(refreshed, fixture) {
		t.Fatal("auto attendant refresh failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "refresh", response, replacements)
	t.Log("auto attendant refresh: passed")

	response, err = service.DeleteByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, "delete", response, replacements)
	if err != nil {
		t.Fatal("live auto attendant delete failed")
	}
	if err := cleanup.absent(ctx, lab.Entry{Fixture: fixture, ResourceID: id}); err != nil {
		t.Fatal("auto attendant deletion was not observed")
	}
	if response, err = service.DeleteByID(ctx, id); !missing(err) {
		t.Fatal("repeated auto attendant delete was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "delete_missing", response, replacements)
	if slices.ContainsFunc(listed, func(item *aa.AutoAttendant) bool { return item == nil }) {
		t.Fatal("nil auto attendant in list")
	}
	t.Log("auto attendant delete: passed")
}

// TestAccAutoAttendantRecover is an explicit operator-selected recovery path
// for an interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccAutoAttendantRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for auto attendant recovery")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("private recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, RecoveryHandlers(sdk)); err != nil {
		t.Fatal("auto attendant recovery incomplete; private journal retained")
	}
	t.Log("owned auto attendant recovery: passed")
}
