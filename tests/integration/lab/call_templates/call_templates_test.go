package calltemplatesacc

import (
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
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ct "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "call_templates"
	missingID      = "00000000-0000-4000-8000-000000000001"
	// botID is a syntactically valid, unassigned object ID. The service does
	// not resolve the bot when it stores a compliance recording template.
	botID = "00000000-0000-4000-8000-0000000000b0"
)

func ptr[T any](v T) *T { return &v }

func same[T comparable](value *T, want T) bool { return value != nil && *value == want }

// TestAccCallTemplateReads lists every template type and reads a missing
// template. It writes nothing.
func TestAccCallTemplateReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.CallTemplates
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, _, err := s.AutoRecording.List(ctx); err != nil {
		t.Fatal("auto recording template list failed")
	}
	if _, _, err := s.ComplianceRecording.List(ctx); err != nil {
		t.Fatal("compliance recording template list failed")
	}
	if _, _, err := s.SharedCallHistory.List(ctx); err != nil {
		t.Fatal("shared call history template list failed")
	}
	if _, _, err := s.SharedVoicemailTriageSettings.List(ctx); err != nil {
		t.Fatal("voicemail triage template list failed")
	}
	if _, _, err := s.Tags.List(ctx); err != nil {
		t.Fatal("tags template list failed")
	}
	_, response, err := s.Tags.GetByID(ctx, missingID)
	labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, nil)
	var apiErr *client.APIError
	if !labkit.IsHTTP(err, 404) || !errors.As(err, &apiErr) || len(apiErr.Codes) == 0 || apiErr.Codes[0] != "100004" {
		t.Fatal("missing template read was not a 100004 not-found error")
	}
	if _, _, err := s.AutoRecording.GetByID(ctx, missingID); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing auto recording template read was not a 404")
	}
	t.Log("call template reads: passed")
}

// lifecycle describes one template type's owned create/update/delete run.
type lifecycle[T any] struct {
	kind    string // journal kind
	label   string // fixture label and testdata prefix
	oracle  string // read-call-template.ps1 kind
	service func(*msteams.Client) *ct.Templates[T]
	build   func(name string) *T
	edit    func(*T)
	created func(*T) bool
	updated func(*T) bool
	// project maps a Go value to the PowerShell oracle projection.
	project func(*T) map[string]any
}

func runLifecycle[T any](t *testing.T, l lifecycle[T]) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("call template lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := l.service(sdk)
	cleanup := RecoveryHandlers(sdk)[l.kind].(templateCleanup[T])
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private call template recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{l.kind: cleanup}); err != nil {
			t.Error("call template cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private call template journal close failed")
		}
	})
	fixture, err := journal.FixtureName(l.label)
	if err != nil {
		t.Fatal("call template fixture name unavailable")
	}
	entry, err := journal.PlanCreate(l.kind, fixture)
	if err != nil {
		t.Fatal("call template create was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{fixture: "fixture-template"}

	created, response, err := service.Create(ctx, l.build(fixture))
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_create", response, replacements)
	if err != nil || response.StatusCode() != 201 || !cleanup.owned(created, fixture) || cleanup.id(created) == nil ||
		!labkit.DirectoryID.MatchString(*cleanup.id(created)) || !l.created(created) {
		t.Fatal("live call template create failed; recovery journal retained")
	}
	id := *cleanup.id(created)
	if err := journal.Applied(entry, id); err != nil {
		t.Fatal("created call template ID was not journaled")
	}
	value, response, err := service.GetByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_get", response, replacements)
	if err != nil || !cleanup.owned(value, fixture) || *cleanup.id(value) != id || !l.created(value) {
		t.Fatal("owned call template read-back failed")
	}
	t.Log("call template create/read-back: passed")

	l.edit(value)
	updated, response, err := service.UpdateByID(ctx, id, value)
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_update", response, replacements)
	if err != nil || !cleanup.owned(updated, fixture) || !l.updated(updated) {
		t.Fatal("owned call template update failed")
	}
	value, _, err = service.GetByID(ctx, id)
	if err != nil || !cleanup.owned(value, fixture) || !l.updated(value) {
		t.Fatal("owned call template update read-back failed")
	}
	compareOracle(t, ctx, l.oracle, id, fixture, l.project(value))
	t.Log("call template update/read-back: passed")

	listed, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_list", response, replacements)
	if err != nil {
		t.Fatal("call template list failed")
	}
	found := 0
	for _, item := range listed {
		if cleanup.owned(item, fixture) && same(cleanup.id(item), id) && l.updated(item) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned call template missing or ambiguous in list")
	}
	t.Log("call template list: passed")

	response, err = service.DeleteByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_delete", response, replacements)
	if err != nil || response.StatusCode() != 200 {
		t.Fatal("live call template delete failed")
	}
	if err := cleanup.absent(ctx, id); err != nil {
		t.Fatal("call template absence was not observed")
	}
	response, err = service.DeleteByID(ctx, id)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("repeated call template delete was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, l.label+"_delete_missing", response, replacements)
	t.Log("call template delete/absence: passed")
}

// compareOracle reads the owned template with the pinned module and compares
// its projection with the Go read-back.
func compareOracle(t *testing.T, ctx context.Context, kind, id, fixture string, want map[string]any) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	command := labkit.OracleCommand(ctx, "read-call-template.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"),
		"MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"),
		"MSTEAMS_LAB_CALL_TEMPLATE_KIND="+kind, "MSTEAMS_LAB_CALL_TEMPLATE_ID="+id, "MSTEAMS_LAB_CALL_TEMPLATE_FIXTURE="+fixture)
	var output labkit.OracleBuffer
	command.Stdout = &output
	command.Stderr = io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell call template read failed; no output retained")
	}
	var observed struct {
		Success    bool           `json:"success"`
		Projection map[string]any `json:"projection"`
	}
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success {
		t.Fatalf("pinned PowerShell could not read the owned call template (oracle failure stage: %s)", labkit.ReadOracleFailureStage(output.Data))
	}
	if !labkit.SameProjection(observed.Projection, want) {
		t.Fatal("PowerShell and Go call template projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go call template parity: passed")
}

// TestAccComplianceRecordingTemplateLifecycle creates, updates and deletes one
// owned compliance recording for call queue template.
func TestAccComplianceRecordingTemplateLifecycle(t *testing.T) {
	runLifecycle(t, lifecycle[ct.ComplianceRecordingTemplate]{
		kind: "compliance-recording-template", label: "compliance", oracle: "compliance",
		service: func(sdk *msteams.Client) *ct.Templates[ct.ComplianceRecordingTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.ComplianceRecording
		},
		build: func(name string) *ct.ComplianceRecordingTemplate {
			return &ct.ComplianceRecordingTemplate{Name: &name, Description: ptr("SDK lab non-default initial description"), BotID: ptr(botID),
				RequiredDuringCall: ptr(true), RequiredBeforeCall: ptr(false), ConcurrentInvitationCount: ptr(int32(1))}
		},
		edit: func(v *ct.ComplianceRecordingTemplate) {
			v.Description, v.RequiredDuringCall, v.RequiredBeforeCall = ptr("SDK lab non-default updated description"), ptr(false), ptr(true)
			v.ConcurrentInvitationCount = ptr(int32(2))
		},
		created: func(v *ct.ComplianceRecordingTemplate) bool {
			return same(v.BotID, botID) && same(v.RequiredDuringCall, true) && same(v.RequiredBeforeCall, false) && same(v.ConcurrentInvitationCount, int32(1))
		},
		updated: func(v *ct.ComplianceRecordingTemplate) bool {
			return same(v.Description, "SDK lab non-default updated description") && same(v.BotID, botID) &&
				same(v.RequiredDuringCall, false) && same(v.RequiredBeforeCall, true) && same(v.ConcurrentInvitationCount, int32(2))
		},
		// Get-* reports the stored flags. (Only New-*'s returned object swaps
		// RequiredDuringCall and RequiredBeforeCall; see the contract notes.)
		project: func(v *ct.ComplianceRecordingTemplate) map[string]any {
			return map[string]any{"name": *v.Name, "description": *v.Description, "bot_id": *v.BotID,
				"required_during_call": *v.RequiredDuringCall, "required_before_call": *v.RequiredBeforeCall,
				"concurrent_invitation_count": *v.ConcurrentInvitationCount}
		},
	})
}

// TestAccSharedCallHistoryTemplateLifecycle creates, updates and deletes one
// owned shared call history template. The oracle also reads it through the
// deprecated Get-CsSharedCallQueueHistoryTemplate alias.
func TestAccSharedCallHistoryTemplateLifecycle(t *testing.T) {
	runLifecycle(t, lifecycle[ct.SharedCallHistoryTemplate]{
		kind: "shared-call-history-template", label: "history", oracle: "history",
		service: func(sdk *msteams.Client) *ct.Templates[ct.SharedCallHistoryTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.SharedCallHistory
		},
		build: func(name string) *ct.SharedCallHistoryTemplate {
			return &ct.SharedCallHistoryTemplate{Name: &name, Description: ptr("SDK lab non-default initial description"),
				IncomingMissedCalls: ptr(ct.HistoryAuthorizedUsersOnly), AnsweredAndOutboundCalls: ptr(ct.HistoryAuthorizedUsersAndAgents)}
		},
		edit: func(v *ct.SharedCallHistoryTemplate) {
			v.Description, v.IncomingMissedCalls = ptr("SDK lab non-default updated description"), ptr(ct.HistoryAuthorizedUsersAndAgents)
			v.AnsweredAndOutboundCalls, v.IncomingRedirectedCalls = ptr(ct.HistoryNone), ptr(ct.HistoryAuthorizedUsersAndGroupMembers)
		},
		created: func(v *ct.SharedCallHistoryTemplate) bool {
			return same(v.IncomingMissedCalls, ct.HistoryAuthorizedUsersOnly) && same(v.AnsweredAndOutboundCalls, ct.HistoryAuthorizedUsersAndAgents) &&
				same(v.IncomingRedirectedCalls, ct.HistoryNone)
		},
		updated: func(v *ct.SharedCallHistoryTemplate) bool {
			return same(v.Description, "SDK lab non-default updated description") && same(v.IncomingMissedCalls, ct.HistoryAuthorizedUsersAndAgents) &&
				same(v.AnsweredAndOutboundCalls, ct.HistoryNone) && same(v.IncomingRedirectedCalls, ct.HistoryAuthorizedUsersAndGroupMembers)
		},
		project: func(v *ct.SharedCallHistoryTemplate) map[string]any {
			return map[string]any{"name": *v.Name, "description": *v.Description, "incoming_missed_calls": *v.IncomingMissedCalls,
				"answered_and_outbound_calls": *v.AnsweredAndOutboundCalls, "incoming_redirected_calls": *v.IncomingRedirectedCalls, "alias_matches": true}
		},
	})
}

// TestAccVoicemailTriageTemplateLifecycle creates, updates and deletes one
// owned shared voicemail triage settings template.
func TestAccVoicemailTriageTemplateLifecycle(t *testing.T) {
	runLifecycle(t, lifecycle[ct.SharedVoicemailTriageSettingsTemplate]{
		kind: "voicemail-triage-template", label: "triage", oracle: "triage",
		service: func(sdk *msteams.Client) *ct.Templates[ct.SharedVoicemailTriageSettingsTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.SharedVoicemailTriageSettings
		},
		build: func(name string) *ct.SharedVoicemailTriageSettingsTemplate {
			return &ct.SharedVoicemailTriageSettingsTemplate{Name: &name, Description: ptr("SDK lab non-default initial description"),
				EnableUrgencyDetection: ptr(false), EnableCategoryDetection: ptr(true)}
		},
		edit: func(v *ct.SharedVoicemailTriageSettingsTemplate) {
			v.Description, v.EnableUrgencyDetection, v.EnableVoiceToTextSummary = ptr("SDK lab non-default updated description"), ptr(true), ptr(false)
		},
		// Omitted features default to enabled on create.
		created: func(v *ct.SharedVoicemailTriageSettingsTemplate) bool {
			return same(v.EnableUrgencyDetection, false) && same(v.EnableCategoryDetection, true) &&
				same(v.EnableCallToActionDetection, true) && same(v.EnableVoiceToTextSummary, true)
		},
		updated: func(v *ct.SharedVoicemailTriageSettingsTemplate) bool {
			return same(v.Description, "SDK lab non-default updated description") && same(v.EnableUrgencyDetection, true) &&
				same(v.EnableCategoryDetection, true) && same(v.EnableCallToActionDetection, true) && same(v.EnableVoiceToTextSummary, false)
		},
		project: func(v *ct.SharedVoicemailTriageSettingsTemplate) map[string]any {
			return map[string]any{"name": *v.Name, "description": *v.Description, "urgency": *v.EnableUrgencyDetection,
				"category": *v.EnableCategoryDetection, "call_to_action": *v.EnableCallToActionDetection, "summary": *v.EnableVoiceToTextSummary}
		},
	})
}

// TestAccTagsTemplateLifecycle creates, updates and deletes one owned IVR
// tags template whose tags route to fictional 555 PSTN numbers.
func TestAccTagsTemplateLifecycle(t *testing.T) {
	tag := func(name, number string) ct.Tag {
		return ct.Tag{TagName: ptr(name), TagDetails: &ct.TagTarget{ID: ptr("tel:" + number), Type: ptr("ExternalPstn")}}
	}
	tagsAre := func(v *ct.TagsTemplate, want ...ct.Tag) bool {
		if len(v.Tags) != len(want) {
			return false
		}
		for i, got := range v.Tags {
			if !same(got.TagName, *want[i].TagName) || got.TagDetails == nil || !same(got.TagDetails.ID, *want[i].TagDetails.ID) ||
				!same(got.TagDetails.Type, "ExternalPstn") {
				return false
			}
		}
		return true
	}
	sales, support, billing := tag("Sales", "+15555550100"), tag("Support", "+15555550101"), tag("Billing", "+15555550102")
	runLifecycle(t, lifecycle[ct.TagsTemplate]{
		kind: "tags-template", label: "tags", oracle: "tags",
		service: func(sdk *msteams.Client) *ct.Templates[ct.TagsTemplate] { return sdk.TeamsAdminAPI.CallTemplates.Tags },
		build: func(name string) *ct.TagsTemplate {
			return &ct.TagsTemplate{Name: &name, Description: ptr("SDK lab non-default initial description"), Tags: []ct.Tag{sales}}
		},
		edit: func(v *ct.TagsTemplate) {
			v.Description, v.Tags = ptr("SDK lab non-default updated description"), []ct.Tag{support, billing}
		},
		created: func(v *ct.TagsTemplate) bool { return tagsAre(v, sales) },
		updated: func(v *ct.TagsTemplate) bool {
			return same(v.Description, "SDK lab non-default updated description") && tagsAre(v, support, billing)
		},
		project: func(v *ct.TagsTemplate) map[string]any {
			tags := []any{}
			for _, item := range v.Tags {
				tags = append(tags, map[string]any{"name": *item.TagName, "target": *item.TagDetails.ID})
			}
			return map[string]any{"name": *v.Name, "description": *v.Description, "tags": tags}
		},
	})
}

// TestAccAutoRecordingTemplateCreateDenied records the lab's denial of auto
// recording templates: the service validates the SharePoint site with the
// caller's identity, which the app-only lab identity cannot satisfy. The
// create is journaled in case a future tenant accepts it.
func TestAccAutoRecordingTemplateCreateDenied(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("auto recording create requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	owner := userFixtureID(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.CallTemplates.AutoRecording
	cleanup := RecoveryHandlers(sdk)["auto-recording-template"]
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private call template recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"auto-recording-template": cleanup}); err != nil {
			t.Error("auto recording cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private call template journal close failed")
		}
	})
	fixture, err := journal.FixtureName("recording")
	if err != nil {
		t.Fatal("auto recording fixture name unavailable")
	}
	if _, err := journal.PlanCreate("auto-recording-template", fixture); err != nil {
		t.Fatal("auto recording create was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	_, response, err := service.Create(ctx, &ct.AutoRecordingTemplate{Name: &fixture, Description: ptr("SDK lab non-default initial description"),
		TranscriptionEnabled: ptr(false), RecordingEnabled: ptr(true), AgentViewPermission: ptr(ct.AgentViewPermissionNone),
		SharePointHostName: ptr("example.sharepoint.com"), SharePointSiteName: ptr(fixture), RecordingDocumentOwner: &owner})
	labkit.WriteResponseFixture(t, fixtureService, "recording_create", response, map[string]string{fixture: "fixture-template", owner: missingID})
	var apiErr *client.APIError
	if !labkit.IsHTTP(err, 401) || !errors.As(err, &apiErr) || len(apiErr.Codes) == 0 || apiErr.Codes[0] != "100012" ||
		!strings.Contains(string(response.Body()), "SharePointClient_Unauthorized") {
		t.Fatal("auto recording create was not the SharePoint authorization denial; journal recovery removes any created template")
	}
	t.Log("auto recording create denial: verified")

	values, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, fixtureService, "recording_list", response, nil)
	if err != nil {
		t.Fatal("auto recording template list failed")
	}
	for _, value := range values {
		if value.Name != nil && *value.Name == fixture {
			t.Fatal("denied auto recording template was listed")
		}
	}
	// The module checks the template exists before Set; the API itself
	// validates the body first, so an incomplete body is a 400.
	_, response, err = service.UpdateByID(ctx, missingID, &ct.AutoRecordingTemplate{Name: &fixture, TranscriptionEnabled: ptr(false), RecordingEnabled: ptr(false)})
	labkit.WriteResponseFixture(t, fixtureService, "recording_update_invalid", response, map[string]string{fixture: "fixture-template"})
	if !labkit.IsHTTP(err, 400) {
		t.Fatal("auto recording update without recording or transcription was not rejected")
	}
	if _, err := service.DeleteByID(ctx, missingID); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing auto recording template delete was not a 404")
	}
	t.Log("auto recording missing-template writes: passed")
}

func userFixtureID(t *testing.T) string {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private user fixture is required")
	}
	defer clear(raw)
	var fixture struct {
		UserID string `json:"user_id"`
	}
	if json.Unmarshal(raw, &fixture) != nil || !labkit.DirectoryID.MatchString(fixture.UserID) {
		t.Fatal("private user fixture has no directory object ID")
	}
	return fixture.UserID
}

// TestAccCallTemplateRecover is an explicit operator-selected recovery path
// for an interrupted run.
func TestAccCallTemplateRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for call template recovery")
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
		t.Fatal("call template recovery incomplete; private journal retained")
	}
	t.Log("owned call template recovery: passed")
}
