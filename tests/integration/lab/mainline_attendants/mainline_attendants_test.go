package mainlineattendantsacc

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
	ma "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/mainline_attendants"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "mainline_attendants"
	// A reserved fixture-form ID that cannot name a real flow or template.
	probeID = "00000000-0000-4000-8000-000000000001"
	// The booking link is never dereferenced by the service on create.
	bookingLink = "https://outlook.office.com/book/msteams-sdk-lifecycle@example.invalid/"
	// The knowledge base page does not resolve, so indexing fails quickly and
	// nothing tenant-specific is indexed.
	knowledgeBaseURL = "https://www.example.com/msteams-sdk-faq"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

type tenantObservation struct {
	Success         bool   `json:"success"`
	DefaultLanguage string `json:"default_language"`
	DefaultVoice    string `json:"default_voice"`
	DefaultTimeZone string `json:"default_time_zone"`
	LanguageCount   int    `json:"language_count"`
	VoiceCount      int    `json:"voice_count"`
}

type ownedObservation struct {
	Success            bool   `json:"success"`
	SpamName           string `json:"spam_name"`
	SpamDescription    string `json:"spam_description"`
	SpamEnabled        bool   `json:"spam_enabled"`
	SpamAction         string `json:"spam_action"`
	BookingName        string `json:"booking_name"`
	BookingDescription string `json:"booking_description"`
	BookingPlatform    string `json:"booking_platform"`
	QAName             string `json:"qa_name"`
	QADescription      string `json:"qa_description"`
}

func powerShellMainline(t *testing.T, ctx context.Context, out any, extra ...string) {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-mainline-attendants.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		command.Env = append(command.Env, key+"="+os.Getenv(key))
	}
	command.Env = append(append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1"), extra...)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell mainline attendant read failed; no output retained")
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(out) != nil {
		t.Fatal("pinned PowerShell mainline attendant projection unavailable")
	}
}

// TestAccMainlineAttendantReads reads the tenant information and every list,
// and reads a nonexistent flow and template. It changes nothing.
func TestAccMainlineAttendantReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.MainlineAttendants
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	information, response, err := service.GetTenantInformation(ctx)
	if err != nil || information.DefaultLanguage == nil || information.DefaultVoice == nil || len(information.SupportedLanguages) == 0 || len(information.SupportedVoices) == 0 {
		t.Fatal("mainline attendant tenant information read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "tenant_information", response, replacements)
	t.Log("mainline attendant tenant information: passed")

	if _, response, err = service.ListFlows(ctx, nil); err != nil {
		t.Fatal("mainline attendant flow list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "flows", response, replacements)
	if _, response, err = service.ListAppointmentBookingFlows(ctx, nil); err != nil {
		t.Fatal("appointment booking flow list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "appointment_booking_flows", response, replacements)
	if _, response, err = service.ListQuestionAnswerFlows(ctx, nil); err != nil {
		t.Fatal("question answer flow list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flows", response, replacements)
	if _, response, err = service.ListSpamDetectionTemplates(ctx, &ma.ListOptions{First: new(int32(50)), SortBy: new("Name"), Descending: true}); err != nil {
		t.Fatal("spam detection template list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_templates", response, replacements)
	t.Log("mainline attendant lists: passed")

	if _, response, err = service.GetFlowByID(ctx, probeID); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing flow was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "flow_missing", response, replacements)
	if _, response, err = service.GetSpamDetectionTemplateByID(ctx, probeID); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing spam detection template was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_missing", response, replacements)
	t.Log("mainline attendant missing reads: passed")

	if labkit.OracleEnabled(t) {
		var observed tenantObservation
		powerShellMainline(t, ctx, &observed, "MSTEAMS_LAB_MAINLINE_MODE=tenant")
		if !observed.Success || observed.DefaultLanguage != *information.DefaultLanguage || observed.DefaultVoice != *information.DefaultVoice ||
			observed.DefaultTimeZone != text(information.DefaultTimeZone) || observed.LanguageCount != len(information.SupportedLanguages) ||
			observed.VoiceCount != len(information.SupportedVoices) {
			t.Fatal("PowerShell and Go mainline attendant tenant information differ")
		}
		t.Log("PowerShell 8.1.0 / Go mainline attendant tenant information parity: passed")
	}
}

// TestAccMainlineAttendantLifecycle creates, updates, reads, lists and deletes
// one owned spam detection template, Microsoft Bookings appointment booking
// flow and question and answer flow. Each create is journaled first.
func TestAccMainlineAttendantLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("mainline attendant lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.MainlineAttendants
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private mainline attendant recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, RecoveryHandlers(sdk)); err != nil {
			t.Error("mainline attendant cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private mainline attendant journal close failed")
		}
	})
	phase := func() (context.Context, context.CancelFunc) { return context.WithTimeout(t.Context(), cfg.Timeout) }
	replacements := tenantReplacements(t, binding)
	fixtures, fixtureIDs, entries := map[string]string{}, map[string]string{}, map[string]string{}
	for _, item := range []struct{ kind, label, id string }{
		{spamKind, spamLabel, "00000000-0000-4000-8000-0000000000b1"},
		{bookingKind, bookingLabel, "00000000-0000-4000-8000-0000000000b2"},
		{qaKind, qaLabel, "00000000-0000-4000-8000-0000000000b3"},
	} {
		fixture, err := journal.FixtureName(item.label)
		if err != nil {
			t.Fatal("mainline attendant fixture name unavailable")
		}
		fixtures[item.label] = fixture
		replacements[fixture] = "fixture-" + item.label
		fixtureIDs[item.label] = item.id
	}
	plan := func(kind, label string) {
		entry, err := journal.PlanCreate(kind, fixtures[label])
		if err != nil {
			t.Fatal("mainline attendant create was not journaled")
		}
		entries[label] = entry
	}
	applied := func(label string, id *string) string {
		if id == nil || !labkit.DirectoryID.MatchString(*id) {
			t.Fatal("created mainline attendant resource has no identity; recovery journal retained")
		}
		replacements[*id] = fixtureIDs[label]
		if err := journal.Applied(entries[label], *id); err != nil {
			t.Fatal("created mainline attendant identity was not journaled")
		}
		return *id
	}

	// Spam detection template.
	ctx, cancel := phase()
	defer cancel()
	plan(spamKind, spamLabel)
	template, response, err := service.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: fixtures[spamLabel], Description: "SDK lifecycle"})
	if err == nil && template != nil && template.Identity != nil {
		replacements[*template.Identity] = fixtureIDs[spamLabel]
	}
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_create", response, replacements)
	if err != nil || response.StatusCode() != 201 || text(template.Name) != fixtures[spamLabel] {
		t.Fatal("live spam detection template create failed; recovery journal retained")
	}
	spamID := applied(spamLabel, template.Identity)
	if template.EnableSpamDetection == nil || !*template.EnableSpamDetection || text(template.Action) != ma.SpamActionDisconnectCall {
		t.Fatal("spam detection template create lost the module defaults")
	}
	disabled, updatedText := false, "SDK lifecycle updated"
	template, response, err = service.UpdateSpamDetectionTemplateByID(ctx, spamID, ma.UpdateSpamDetectionTemplateRequest{Name: template.Name, Description: &updatedText,
		EnableSpamDetection: &disabled, Action: template.Action, RelatedConfigurationIDs: []ma.RelatedConfiguration{}})
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_update", response, replacements)
	if err != nil || text(template.Description) != updatedText || template.EnableSpamDetection == nil || *template.EnableSpamDetection {
		t.Fatal("spam detection template update failed")
	}
	template, response, err = service.GetSpamDetectionTemplateByID(ctx, spamID)
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_get", response, replacements)
	if err != nil || text(template.Name) != fixtures[spamLabel] || text(template.Description) != updatedText || *template.EnableSpamDetection {
		t.Fatal("spam detection template read-back failed")
	}
	templates, _, err := service.ListSpamDetectionTemplates(ctx, &ma.ListOptions{NameFilter: new(fixtures[spamLabel])})
	if err != nil || len(templates) != 1 || text(templates[0].Identity) != spamID {
		t.Fatal("owned spam detection template missing or ambiguous in list")
	}
	t.Log("spam detection template create/update/read-back: passed")
	cancel()

	// Appointment booking flow on Microsoft Bookings.
	ctx, cancel = phase()
	defer cancel()
	plan(bookingKind, bookingLabel)
	platform, link := ma.BookingPlatformMicrosoftBookings, bookingLink
	flow, response, err := service.CreateAppointmentBookingFlow(ctx, ma.CreateAppointmentBookingFlowRequest{Name: fixtures[bookingLabel], Description: "SDK lifecycle",
		AppointmentBookingPlatform: &platform, BookingLink: &link})
	if err == nil && flow != nil && flow.Identity != nil {
		replacements[*flow.Identity] = fixtureIDs[bookingLabel]
	}
	labkit.WriteResponseFixture(t, fixtureService, "appointment_booking_flow_create", response, replacements)
	if err != nil || response.StatusCode() != 201 || text(flow.Name) != fixtures[bookingLabel] {
		t.Fatal("live appointment booking flow create failed; recovery journal retained")
	}
	bookingID := applied(bookingLabel, flow.Identity)
	if text(flow.Type) != ma.FlowTypeAppointmentBooking || text(flow.AppointmentBookingPlatform) != platform {
		t.Fatal("appointment booking flow create lost its type or platform")
	}
	flow, response, err = service.UpdateAppointmentBookingFlowByID(ctx, bookingID, ma.UpdateAppointmentBookingFlowRequest{Name: flow.Name, Description: &updatedText,
		RelatedConfigurationIDs: []ma.RelatedConfiguration{}, AppointmentBookingPlatform: &platform, BookingLink: &link})
	labkit.WriteResponseFixture(t, fixtureService, "appointment_booking_flow_update", response, replacements)
	if err != nil || text(flow.Description) != updatedText {
		t.Fatal("appointment booking flow update failed")
	}
	flow, response, err = service.GetAppointmentBookingFlowByID(ctx, bookingID)
	labkit.WriteResponseFixture(t, fixtureService, "appointment_booking_flow_get", response, replacements)
	if err != nil || text(flow.Name) != fixtures[bookingLabel] || text(flow.Description) != updatedText {
		t.Fatal("appointment booking flow read-back failed")
	}
	t.Log("appointment booking flow create/update/read-back: passed")
	cancel()

	// Question and answer flow; the service indexes the knowledge base first.
	ctx, cancel = phase()
	defer cancel()
	plan(qaKind, qaLabel)
	flow, response, err = service.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: fixtures[qaLabel], Description: "SDK lifecycle",
		KnowledgeBase: ma.KnowledgeBase{URLs: []ma.KnowledgeBaseURL{{URL: knowledgeBaseURL, Description: "FAQ page"}}}})
	if err == nil && flow != nil && flow.Identity != nil {
		replacements[*flow.Identity] = fixtureIDs[qaLabel]
	}
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flow_create", response, replacements)
	if err != nil || response.StatusCode() != 201 || text(flow.Name) != fixtures[qaLabel] {
		t.Fatal("live question answer flow create failed; recovery journal retained")
	}
	qaID := applied(qaLabel, flow.Identity)
	if documents, err := flow.KnowledgeBaseDocuments(); err != nil || len(documents) != 1 || text(flow.Type) != ma.FlowTypeQuestionAnswer {
		t.Fatal("question answer flow create lost its knowledge base status")
	}
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		value, _, err := service.GetQuestionAnswerFlowByID(ctx, qaID)
		if err != nil {
			return false, err
		}
		indexing, err := value.Indexing()
		flow = value
		return !indexing, err
	}); err != nil {
		t.Fatal("question answer flow indexing did not finish")
	}
	cancel()
	ctx, cancel = phase()
	defer cancel()
	flow, response, err = service.UpdateQuestionAnswerFlowByID(ctx, qaID, ma.UpdateQuestionAnswerFlowRequest{Name: flow.Name, Description: &updatedText,
		RelatedConfigurationIDs: []ma.RelatedConfiguration{}, APIAuthenticationType: flow.APIAuthenticationType, KnowledgeBase: flow.KnowledgeBase})
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flow_update", response, replacements)
	if err != nil || text(flow.Description) != updatedText {
		t.Fatal("question answer flow update failed")
	}
	flow, response, err = service.GetFlowByID(ctx, qaID)
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flow_get", response, replacements)
	if err != nil || text(flow.Name) != fixtures[qaLabel] || text(flow.Description) != updatedText || text(flow.Type) != ma.FlowTypeQuestionAnswer {
		t.Fatal("question answer flow read-back failed")
	}
	t.Log("question answer flow create/indexing/update/read-back: passed")

	typed := ma.FlowTypeQuestionAnswer
	listed, response, err := service.ListFlows(ctx, &ma.FlowListOptions{Type: &typed, ListOptions: ma.ListOptions{NameFilter: new(fixtures[qaLabel])}})
	labkit.WriteResponseFixture(t, fixtureService, "flows_filtered", response, replacements)
	if err != nil || len(listed) != 1 || text(listed[0].Identity) != qaID {
		t.Fatal("owned question answer flow missing or ambiguous in filtered list")
	}
	bookings, _, err := service.ListAppointmentBookingFlows(ctx, &ma.ListOptions{NameFilter: new(fixtures[bookingLabel])})
	if err != nil || len(bookings) != 1 || text(bookings[0].Identity) != bookingID {
		t.Fatal("owned appointment booking flow missing or ambiguous in list")
	}
	t.Log("mainline attendant flow list/filter: passed")
	cancel()

	if labkit.OracleEnabled(t) {
		ctx, cancel = phase()
		defer cancel()
		var observed ownedObservation
		powerShellMainline(t, ctx, &observed, "MSTEAMS_LAB_MAINLINE_MODE=owned", "MSTEAMS_LAB_MAINLINE_SPAM_ID="+spamID,
			"MSTEAMS_LAB_MAINLINE_BOOKING_ID="+bookingID, "MSTEAMS_LAB_MAINLINE_QA_ID="+qaID)
		if !observed.Success || observed.SpamName != fixtures[spamLabel] || observed.SpamDescription != updatedText || observed.SpamEnabled ||
			observed.SpamAction != ma.SpamActionDisconnectCall || observed.BookingName != fixtures[bookingLabel] || observed.BookingDescription != updatedText ||
			observed.BookingPlatform != platform || observed.QAName != fixtures[qaLabel] || observed.QADescription != updatedText {
			t.Fatal("PowerShell and Go mainline attendant observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go mainline attendant parity: passed")
		cancel()
	}

	ctx, cancel = phase()
	defer cancel()
	handlers := RecoveryHandlers(sdk)
	response, err = service.DeleteSpamDetectionTemplateByID(ctx, spamID)
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_delete", response, replacements)
	if err != nil || handlers[spamKind].(ownedCleanup).absent(ctx, spamID) != nil {
		t.Fatal("spam detection template delete failed")
	}
	if response, err = service.DeleteSpamDetectionTemplateByID(ctx, spamID); !labkit.IsHTTP(err, 404) {
		t.Fatal("repeated spam detection template delete was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "spam_detection_template_delete", response, replacements)
	t.Log("spam detection template delete/absence: passed")

	response, err = service.DeleteAppointmentBookingFlowByID(ctx, bookingID)
	labkit.WriteResponseFixture(t, fixtureService, "appointment_booking_flow_delete", response, replacements)
	if err != nil || handlers[bookingKind].(ownedCleanup).absent(ctx, bookingID) != nil {
		t.Fatal("appointment booking flow delete failed")
	}
	t.Log("appointment booking flow delete/absence: passed")

	response, err = service.DeleteQuestionAnswerFlowByID(ctx, qaID)
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flow_delete", response, replacements)
	if err != nil || handlers[qaKind].(ownedCleanup).absent(ctx, qaID) != nil {
		t.Fatal("question answer flow delete failed")
	}
	if _, response, err = service.GetFlowByID(ctx, qaID); !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted question answer flow is still readable")
	}
	labkit.WriteResponseFixture(t, fixtureService, "question_answer_flow_get", response, replacements)
	t.Log("question answer flow delete/absence: passed")
}

// TestAccMainlineAttendantRecover is an explicit operator-selected recovery
// path for an interrupted run. OpenJournal checks the binding locally.
func TestAccMainlineAttendantRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for mainline attendant recovery")
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
		t.Fatal("mainline attendant recovery incomplete; private journal retained")
	}
	t.Log("owned mainline attendant recovery: passed")
}
