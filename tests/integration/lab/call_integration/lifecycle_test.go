package callintegration

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	aaacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/auto_attendants"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	schedacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/schedules"
)

func guid(t *testing.T) string {
	t.Helper()
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		t.Fatal("call flow identity unavailable")
	}
	value[6] = (value[6] & 15) | 64
	value[8] = (value[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[:4], value[4:6], value[6:8], value[8:10], value[10:])
}

type integrationObservation struct {
	Success               bool   `json:"success"`
	ScheduleName          string `json:"schedule_name"`
	AttendantName         string `json:"attendant_name"`
	ScheduleID            string `json:"schedule_id"`
	AssociationScheduleID string `json:"association_schedule_id"`
	AssociationFlowID     string `json:"association_flow_id"`
	AssociationType       string `json:"association_type"`
	Enabled               bool   `json:"enabled"`
	Greeting              string `json:"greeting"`
	HolidayCount          int    `json:"holiday_count"`
	HolidayFlowID         string `json:"holiday_flow_id"`
}

func compareIntegration(t *testing.T, ctx context.Context, scheduleName, scheduleID, attendantName, attendantID, flowID, holidayFlowID, greeting string) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		t.Fatal("integrated acceptance requires the pinned PowerShell oracle")
	}
	command := labkit.OracleCommand(ctx, "read-call-integration.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath", "MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_SCHEDULE_FIXTURE="+scheduleName, "MSTEAMS_LAB_SCHEDULE_ID="+scheduleID, "MSTEAMS_LAB_AUTO_ATTENDANT_FIXTURE="+attendantName, "MSTEAMS_LAB_AUTO_ATTENDANT_ID="+attendantID)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned integrated CALL read failed; no output retained")
	}
	var observed integrationObservation
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success || observed.ScheduleName != scheduleName || observed.AttendantName != attendantName || observed.ScheduleID != scheduleID || observed.AssociationScheduleID != scheduleID || observed.AssociationFlowID != flowID || observed.AssociationType != aa.AssociationAfterHours || !observed.Enabled || observed.Greeting != greeting || observed.HolidayCount != 1 || observed.HolidayFlowID != holidayFlowID {
		t.Fatal("PowerShell and Go integrated CALL observations differ")
	}
	t.Log("PowerShell 8.1.0 / Go integrated CALL parity: passed")
}

// TestAccScheduleAutoAttendantComposition links only newly owned schedules and
// attendant. It assigns no account, number, user, group or policy and needs no
// calling licence. Reverse-order journal recovery deletes the dependent first.
func TestAccScheduleAutoAttendantComposition(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("CALL integration requires explicit mutation opt-in and private journaling")
	} else if err != nil {
		t.Fatal(err)
	}
	if !labkit.OracleEnabled(t) {
		t.Fatal("CALL integration requires the pinned oracle")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private integrated CALL journal unavailable")
	}
	handlers := aaacc.RecoveryHandlers(sdk)
	for kind, handler := range schedacc.RecoveryHandlers(sdk) {
		handlers[kind] = handler
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		results, err := journal.Recover(ctx, handlers)
		if err != nil {
			t.Error("integrated CALL cleanup incomplete; private journal retained")
		}
		for _, result := range results {
			if result.Outcome != "cleaned" {
				t.Error("integrated CALL resource cleanup unresolved")
			}
		}
		if err := journal.Close(); err != nil {
			t.Error("integrated CALL journal close failed")
		}
		if !t.Failed() {
			t.Log("integrated CALL cleanup: passed")
		}
	})
	scheduleName, err := journal.FixtureName("weekly")
	if err != nil {
		t.Fatal("owned schedule name unavailable")
	}
	attendantName, err := journal.FixtureName("attendant")
	if err != nil {
		t.Fatal("owned attendant name unavailable")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	scheduleEntry, err := journal.PlanCreate("schedule", scheduleName)
	if err != nil {
		t.Fatal("schedule create was not journaled")
	}
	schedule, response, err := sdk.TeamsAdminAPI.Schedules.Create(ctx, sched.CreateRequest{Name: scheduleName, WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{MondayHours: []sched.TimeRange{{Start: "09:00:00", End: "17:00:00"}}, IsComplemented: ptr(true)}})
	if err == nil && schedule != nil && schedule.ID != nil {
		err = journal.Applied(scheduleEntry, *schedule.ID)
	}
	if err != nil || schedule == nil || schedule.ID == nil {
		t.Fatal("integrated schedule create failed; journal retained")
	}
	replacements := map[string]string{scheduleName: "msteams-sdk-fixture-weekly", attendantName: "msteams-sdk-fixture-attendant", *schedule.ID: scheduleID}
	labkit.WriteResponseFixture(t, "schedules", "integration_create", response, replacements)
	fixedName, err := journal.FixtureName("fixed")
	if err != nil {
		t.Fatal("owned fixed schedule name unavailable")
	}
	fixedEntry, err := journal.PlanCreate("schedule", fixedName)
	if err != nil {
		t.Fatal("fixed schedule create was not journaled")
	}
	fixed, response, err := sdk.TeamsAdminAPI.Schedules.Create(ctx, sched.CreateRequest{Name: fixedName, FixedSchedule: &sched.FixedSchedule{DateTimeRanges: []sched.DateTimeRange{{Start: "2027-01-01T00:00:00", End: "2027-01-02T00:00:00"}}}})
	if err == nil && fixed != nil && fixed.ID != nil {
		err = journal.Applied(fixedEntry, *fixed.ID)
	}
	if err != nil || fixed == nil || fixed.ID == nil {
		t.Fatal("integrated fixed schedule create failed; journal retained")
	}
	replacements[fixedName] = "msteams-sdk-fixture-fixed"
	replacements[*fixed.ID] = "00000000-0000-4000-8000-000000000005"
	labkit.WriteResponseFixture(t, "schedules", "integration_fixed_create", response, replacements)
	defaultID, afterHoursID, holidayID := guid(t), guid(t), guid(t)
	replacements[defaultID], replacements[afterHoursID] = flowID, attendantID
	replacements[holidayID] = "00000000-0000-4000-8000-000000000006"
	associations, err := draftAssociations(ctx, sdk.TeamsAdminAPI.AutoAttendants, *schedule.ID, *fixed.ID, afterHoursID, holidayID)
	if err != nil {
		t.Fatal("owned schedule association drafts failed")
	}

	disconnect := aa.MenuOption{Action: ptr(aa.ActionDisconnectCall), DtmfResponse: ptr(aa.DtmfAutomatic)}
	makeFlow := func(id, name, greeting string) aa.CallFlow {
		return aa.CallFlow{ID: &id, Name: &name, Greetings: []aa.Prompt{{ActiveType: ptr(aa.PromptTextToSpeech), TextToSpeechPrompt: &greeting}}, Menu: &aa.Menu{Name: ptr(name + " menu"), MenuOptions: []aa.MenuOption{disconnect}}}
	}
	defaultFlow := makeFlow(defaultID, "Default", "SDK integration greeting")
	afterHours := makeFlow(afterHoursID, "After hours", "SDK integration after hours")
	holiday := makeFlow(holidayID, "Holiday", "SDK integration after hours")
	attendantEntry, err := journal.PlanCreate("auto-attendant", attendantName)
	if err != nil {
		t.Fatal("attendant create was not journaled")
	}
	attendant, response, err := sdk.TeamsAdminAPI.AutoAttendants.Create(ctx, aa.Settings{Name: &attendantName, LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"), DefaultCallFlow: &defaultFlow, CallFlows: &[]aa.CallFlow{afterHours, holiday}, CallHandlingAssociations: &associations})
	if err == nil && attendant != nil && attendant.ID != nil {
		err = journal.Applied(attendantEntry, *attendant.ID)
	}
	if err != nil || attendant == nil || attendant.ID == nil {
		handlers["auto-attendant"] = rejectedAttendantCleanup{service: sdk.TeamsAdminAPI.AutoAttendants, fallback: handlers["auto-attendant"], entryID: attendantEntry, fixture: attendantName, rejected: confirmedCreateRejection(response, err)}
		logCreateFailure(t, response, err)
		t.Fatal("integrated attendant create failed; journal retained")
	}
	replacements[*attendant.ID] = "00000000-0000-4000-8000-000000000004"
	labkit.WriteResponseFixture(t, "auto_attendants", "integration_create", response, replacements)
	t.Log("integrated CALL create: passed")
	check := func(expectedGreeting string) *aa.AutoAttendant {
		value, response, err := sdk.TeamsAdminAPI.AutoAttendants.GetByID(ctx, *attendant.ID)
		if err != nil || value == nil || value.Name == nil || *value.Name != attendantName || len(value.CallHandlingAssociations) != 2 || value.DefaultCallFlow == nil || len(value.DefaultCallFlow.Greetings) != 1 || value.DefaultCallFlow.Greetings[0].TextToSpeechPrompt == nil || *value.DefaultCallFlow.Greetings[0].TextToSpeechPrompt != expectedGreeting {
			t.Fatal("integrated identity/read-back mismatch")
		}
		for _, link := range value.CallHandlingAssociations {
			if link.Type == nil || link.ScheduleID == nil || link.CallFlowID == nil {
				t.Fatal("integrated schedule association incomplete")
			}
			switch *link.Type {
			case aa.AssociationAfterHours:
				if *link.ScheduleID != *schedule.ID || *link.CallFlowID != afterHoursID {
					t.Fatal("after-hours schedule identity changed")
				}
			case aa.AssociationHoliday:
				if *link.ScheduleID != *fixed.ID || *link.CallFlowID != holidayID {
					t.Fatal("holiday schedule identity changed")
				}
			default:
				t.Fatal("unexpected integrated schedule association")
			}
		}
		labkit.WriteResponseFixture(t, "auto_attendants", "integration_read", response, replacements)
		compareIntegration(t, ctx, scheduleName, *schedule.ID, attendantName, *attendant.ID, afterHoursID, holidayID, expectedGreeting)
		return value
	}
	holidays, response, err := sdk.TeamsAdminAPI.Schedules.ListHolidays(ctx, *attendant.ID, &sched.HolidayOptions{Years: []string{"2027"}})
	invalid := err != nil || len(holidays) != 1 || holidays[0] == nil || len(holidays[0].DateTimeRanges) != 1 || !holidayRangeMatches(holidays[0].DateTimeRanges[0])
	if invalid {
		logHolidayFailure(t, response, holidays, err)
	}
	if response != nil && response.StatusCode() == http.StatusOK {
		labkit.WriteResponseFixtureWhen(t, os.Getenv("MSTEAMS_LAB_CAPTURE_HOLIDAY_FIXTURE") == "1", "schedules", "integration_holidays", response, replacements)
	}
	if invalid {
		t.Fatal("owned holiday schedule visualization failed")
	}
	var holidayFlow aa.CallFlow
	if json.Unmarshal(holidays[0].CallFlow, &holidayFlow) != nil || holidayFlow.ID == nil || *holidayFlow.ID != holidayID {
		t.Fatal("holiday visualization lost the attendant call-flow identity")
	}
	exported, _, err := sdk.TeamsAdminAPI.Schedules.GetHolidaysContent(ctx, *attendant.ID)
	if err != nil || len(exported) == 0 {
		t.Fatal("owned holiday schedule export failed")
	}
	// Serialized holiday bytes may contain private identifiers. Keep them only
	// in memory; do not pass opaque base64 through the JSON fixture writer.
	defer func() { clear(exported) }()
	t.Log("integrated CALL populated holiday read and nonempty export receipt: passed")
	value := check("SDK integration greeting")
	t.Log("integrated CALL identity read-back: passed")
	update := value.Settings()
	flow := *update.DefaultCallFlow
	flow.Greetings = []aa.Prompt{{ActiveType: ptr(aa.PromptTextToSpeech), TextToSpeechPrompt: ptr("SDK integration updated greeting")}}
	update.DefaultCallFlow = &flow
	if _, _, err := sdk.TeamsAdminAPI.AutoAttendants.UpdateByID(ctx, *attendant.ID, update); err != nil {
		t.Fatal("integrated attendant replacement failed")
	}
	check("SDK integration updated greeting")
	t.Log("integrated CALL update/read-back: passed")
}
