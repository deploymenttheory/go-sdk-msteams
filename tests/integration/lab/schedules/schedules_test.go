package schedulesacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "schedules"
	// A reserved fixture-form ID that cannot name a real auto attendant.
	probeAutoAttendant = "00000000-0000-4000-8000-000000000001"
)

type scheduleObservation struct {
	Success      bool     `json:"success"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Hours        []string `json:"hours"`
	Ranges       []string `json:"ranges"`
	Complemented bool     `json:"complemented"`
}

func powerShellSchedule(t *testing.T, ctx context.Context, fixture, id string) scheduleObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-schedule.ps1")
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_SCHEDULE_FIXTURE="+fixture, "MSTEAMS_LAB_SCHEDULE_ID="+id)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell schedule read failed; no output retained")
	}
	var observed scheduleObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell schedule projection unavailable")
	}
	return observed
}

// goHours renders weekly hours as the oracle does.
func goHours(w *sched.WeeklyRecurrentSchedule) []string {
	out := []string{}
	if w == nil {
		return out
	}
	for _, day := range []struct {
		name  string
		hours []sched.TimeRange
	}{{"MondayHours", w.MondayHours}, {"TuesdayHours", w.TuesdayHours}, {"WednesdayHours", w.WednesdayHours}, {"ThursdayHours", w.ThursdayHours},
		{"FridayHours", w.FridayHours}, {"SaturdayHours", w.SaturdayHours}, {"SundayHours", w.SundayHours}} {
		for _, r := range day.hours {
			out = append(out, day.name+":"+r.Start+"-"+r.End)
		}
	}
	return out
}

func goRanges(f *sched.FixedSchedule) []string {
	out := []string{}
	if f != nil {
		for _, r := range f.DateTimeRanges {
			out = append(out, r.Start+"/"+r.End)
		}
	}
	return out
}

func compareSchedule(t *testing.T, ctx context.Context, fixture string, value *sched.Schedule) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	observed := powerShellSchedule(t, ctx, fixture, *value.ID)
	complemented := value.WeeklyRecurrentSchedule != nil && value.WeeklyRecurrentSchedule.IsComplemented != nil && *value.WeeklyRecurrentSchedule.IsComplemented
	if observed.Name != fixture || observed.Type != *value.Type || !slices.Equal(observed.Hours, goHours(value.WeeklyRecurrentSchedule)) ||
		!slices.Equal(observed.Ranges, goRanges(value.FixedSchedule)) || observed.Complemented != complemented {
		t.Fatal("PowerShell and Go schedule observations differ")
	}
	t.Log("PowerShell 8.1.0 / Go schedule parity: passed")
}

// TestAccScheduleLifecycle validates ranges through the draft endpoints,
// creates one owned weekly and one owned fixed schedule, reads, lists, updates
// and deletes them, and probes holiday routes against a nonexistent auto
// attendant. It never touches a pre-existing schedule.
func TestAccScheduleLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("schedule lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.Schedules
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private schedule recovery journal unavailable")
	}
	cleanup := scheduleCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"schedule": cleanup}); err != nil {
			t.Error("schedule cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private schedule journal close failed")
		}
	})
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	timeRange, response, err := service.NewTimeRange(ctx, "9:00", "17:30")
	if err != nil || timeRange.Start != "09:00:00" || timeRange.End != "17:30:00" {
		t.Fatal("time range draft was not normalized")
	}
	labkit.WriteResponseFixture(t, fixtureService, "time_range", response, nil)
	end := "3/1/2027"
	dateRange, response, err := service.NewDateTimeRange(ctx, "1/1/2027 10:00", &end)
	if err != nil || dateRange.Start == "" || dateRange.End == "" {
		t.Fatal("date-time range draft failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "date_time_range", response, nil)
	if _, response, err = service.NewTimeRange(ctx, "9:00", "25:00"); !labkit.IsHTTP(err, 400) {
		t.Fatal("invalid time range was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "time_range_invalid", response, nil)
	t.Log("schedule range drafts: passed")

	created := map[string]string{}
	for _, label := range []string{"weekly", "fixed"} {
		fixture, err := journal.FixtureName(label)
		if err != nil {
			t.Fatal("schedule fixture name unavailable")
		}
		entry, err := journal.PlanCreate("schedule", fixture)
		if err != nil {
			t.Fatal("schedule create was not journaled")
		}
		request := sched.CreateRequest{Name: fixture}
		if label == "weekly" {
			complemented := true
			request.WeeklyRecurrentSchedule = &sched.WeeklyRecurrentSchedule{MondayHours: []sched.TimeRange{*timeRange}, IsComplemented: &complemented}
		} else {
			request.FixedSchedule = &sched.FixedSchedule{DateTimeRanges: []sched.DateTimeRange{{Start: "2027-01-01T10:00:00", End: "2027-01-03T00:00:00"}}}
		}
		value, response, err := service.Create(ctx, request)
		if err != nil || response.StatusCode() != 201 || !ownedSchedule(value, fixture) {
			t.Fatal("live schedule create failed; recovery journal retained")
		}
		if err := journal.Applied(entry, *value.ID); err != nil {
			t.Fatal("created schedule ID was not journaled")
		}
		labkit.WriteResponseFixture(t, fixtureService, "create_"+label, response, map[string]string{fixture: "fixture-schedule-" + label})
		created[label] = fixture
		compareSchedule(t, ctx, fixture, value)
		created[label+"-id"] = *value.ID
	}
	weekly, weeklyID := created["weekly"], created["weekly-id"]
	t.Log("schedule create: passed")

	value, response, err := service.GetByID(ctx, weeklyID)
	if err != nil || !ownedSchedule(value, weekly) || value.Type == nil || *value.Type != sched.TypeWeeklyRecurrence ||
		len(value.WeeklyRecurrentSchedule.MondayHours) != 1 || value.FixedSchedule != nil {
		t.Fatal("owned weekly schedule read-back failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_weekly", response, map[string]string{weekly: "fixture-schedule-weekly"})
	listed, _, err := service.List(ctx, nil)
	found := 0
	for _, item := range listed {
		if item.ID != nil && (*item.ID == weeklyID || *item.ID == created["fixed-id"]) {
			found++
		}
	}
	if err != nil || found != 2 {
		t.Fatal("owned schedules missing or ambiguous in list")
	}
	t.Log("schedule get/list: passed")

	notComplemented := false
	value.WeeklyRecurrentSchedule.IsComplemented = &notComplemented
	value.WeeklyRecurrentSchedule.FridayHours = []sched.TimeRange{{Start: "08:00:00", End: "12:00:00"}}
	updated, response, err := service.UpdateByID(ctx, weeklyID, value)
	if err != nil || !ownedSchedule(updated, weekly) || len(updated.WeeklyRecurrentSchedule.FridayHours) != 1 || *updated.WeeklyRecurrentSchedule.IsComplemented {
		t.Fatal("schedule update was not applied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "update_weekly", response, map[string]string{weekly: "fixture-schedule-weekly"})
	if value, _, err = service.GetByID(ctx, weeklyID); err != nil || len(value.WeeklyRecurrentSchedule.FridayHours) != 1 || *value.WeeklyRecurrentSchedule.IsComplemented {
		t.Fatal("schedule update read-back failed")
	}
	compareSchedule(t, ctx, weekly, value)
	t.Log("schedule update/read-back: passed")

	replacements := map[string]string{strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
	if _, response, err = service.ListHolidays(ctx, probeAutoAttendant, &sched.HolidayOptions{Years: []string{"2027"}}); !labkit.IsHTTP(err, 404) {
		t.Fatal("holidays of a missing auto attendant were not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "holidays_missing", response, replacements)
	if _, _, err = service.GetHolidaysContent(ctx, probeAutoAttendant); !labkit.IsHTTP(err, 404) {
		t.Fatal("holiday export of a missing auto attendant was not a 404")
	}
	if _, _, err = service.ImportHolidays(ctx, probeAutoAttendant, []byte("Name,StartDateTime1\n")); !labkit.IsHTTP(err, 404) {
		t.Fatal("holiday import to a missing auto attendant was not a 404")
	}
	t.Log("holiday routes: passed")

	for _, label := range []string{"weekly", "fixed"} {
		id := created[label+"-id"]
		response, err := service.DeleteByID(ctx, id)
		if err != nil || response.StatusCode() != 200 {
			t.Fatal("live schedule delete failed")
		}
		if label == "weekly" {
			labkit.WriteResponseFixture(t, fixtureService, "delete", response, nil)
		}
		if err := cleanup.absent(ctx, lab.Entry{Fixture: created[label], ResourceID: id}); err != nil {
			t.Fatal("schedule deletion was not observed")
		}
	}
	if response, err = service.DeleteByID(ctx, weeklyID); !labkit.IsHTTP(err, 404) {
		t.Fatal("repeated schedule delete was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "delete_missing", response, replacements)
	t.Log("schedule delete: passed")
}

// TestAccScheduleRecover is an explicit operator-selected recovery path for
// an interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccScheduleRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for schedule recovery")
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
		t.Fatal("schedule recovery incomplete; private journal retained")
	}
	t.Log("owned schedule recovery: passed")
}
