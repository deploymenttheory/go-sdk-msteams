package schedules_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/schedules/service_test.go"
	jsonType     = "application/json; charset=utf-8"
	weeklyID     = "00000000-0000-4000-8000-14ff58a505f4"
	fixedID      = "00000000-0000-4000-8000-2705eaabb01b"
	probe        = "00000000-0000-4000-8000-000000000001"
	collection   = "/Teams.VoiceApps/schedules"
	holidays     = "/Teams.VoiceApps/auto-attendants/" + probe + "/holidays"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	contentType := jsonType
	if reply == nil {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "schedules"
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("schedule contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(commands ...string) []contracttest.Mapping {
	out := make([]contracttest.Mapping, 0, len(commands))
	for _, command := range commands {
		set := "__AllParameterSets"
		if command == "New-CsOnlineSchedule" {
			set = "WeeklyRecurrentScheduleParamSet"
		}
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: set})
	}
	return out
}

// apiError is the normalized error for a fixture with a top-level code.
func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &envelope)
	return &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{envelope.Code}}
}

func ptr[T any](v T) *T { return &v }

func TestUnitScheduleLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	createdWeekly, createdFixed := readFile(t, "testdata/create_weekly_201.json"), readFile(t, "testdata/create_fixed_201.json")
	got, updated, deleted := readFile(t, "testdata/get_weekly.json"), readFile(t, "testdata/update_weekly.json"), readFile(t, "testdata/delete.json")
	missing := readFile(t, "testdata/delete_missing_404.json")
	weeklyBody := json.RawMessage(`{"Name":"Fixture name","WeeklyRecurrentSchedule":{"MondayHours":[{"Start":"09:00:00","End":"17:30:00"}],"IsComplemented":true}}`)
	fixedBody := json.RawMessage(`{"Name":"Fixture name","FixedSchedule":{"DateTimeRanges":[{"Start":"2027-01-01T10:00:00","End":"2027-01-03T00:00:00"}]}}`)
	updateBody := json.RawMessage(`{"Name":"Fixture name","Type":"WeeklyRecurrence","WeeklyRecurrentSchedule":{"MondayHours":[{"Start":"09:00:00","End":"17:30:00"}],` +
		`"FridayHours":[{"Start":"08:00:00","End":"12:00:00"}],"IsComplemented":false}}`)
	run(t, contracttest.Case{Name: "actual-lifecycle", Source: "msteams/teams_admin_api/schedules/testdata/create_weekly_201.json",
		Mappings: append(mapping("Get-CsOnlineSchedule", "Set-CsOnlineSchedule", "Remove-CsOnlineSchedule", "New-CsOnlineSchedule"),
			contracttest.Mapping{Command: "New-CsOnlineSchedule", ParameterSet: "FixedScheduleParamSet"}),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", collection, "New-CsOnlineSchedule", weeklyBody, 201, createdWeekly),
			wire("POST", collection, "New-CsOnlineSchedule", fixedBody, 201, createdFixed),
			wire("GET", collection+"/"+weeklyID, "Get-CsOnlineSchedule", nil, 200, got),
			wire("PUT", collection+"/"+weeklyID, "Set-CsOnlineSchedule", updateBody, 200, updated),
			wire("DELETE", collection+"/"+weeklyID, "Remove-CsOnlineSchedule", nil, 200, deleted),
			wire("DELETE", collection+"/"+weeklyID, "Remove-CsOnlineSchedule", nil, 404, missing)},
		Want: contracttest.Expected{Status: 404, Body: missing, APIError: apiError(404, missing)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.Schedules
			value, response, err := s.Create(ctx, sched.CreateRequest{Name: "Fixture name", WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{
				MondayHours: []sched.TimeRange{{Start: "09:00:00", End: "17:30:00"}}, IsComplemented: ptr(true)}})
			if err != nil || response.StatusCode() != 201 || *value.ID != weeklyID || *value.Type != sched.TypeWeeklyRecurrence || value.FixedSchedule != nil ||
				value.WeeklyRecurrentSchedule.TuesdayHours == nil || value.WeeklyRecurrentSchedule.RecurrenceRange != nil || len(value.RawFields) != 6 {
				t.Fatal("actual weekly create receipt was changed")
			}
			fixed, _, err := s.Create(ctx, sched.CreateRequest{Name: "Fixture name", FixedSchedule: &sched.FixedSchedule{
				DateTimeRanges: []sched.DateTimeRange{{Start: "2027-01-01T10:00:00", End: "2027-01-03T00:00:00"}}}})
			if err != nil || *fixed.ID != fixedID || *fixed.Type != sched.TypeFixed || fixed.WeeklyRecurrentSchedule != nil || len(fixed.AssociatedConfigurationIDs) != 0 {
				t.Fatal("actual fixed create receipt was changed")
			}
			value, _, err = s.GetByID(ctx, weeklyID)
			if err != nil || !*value.WeeklyRecurrentSchedule.IsComplemented {
				t.Fatal("actual read was changed")
			}
			value.WeeklyRecurrentSchedule.IsComplemented = ptr(false)
			value.WeeklyRecurrentSchedule.FridayHours = []sched.TimeRange{{Start: "08:00:00", End: "12:00:00"}}
			value, _, err = s.UpdateByID(ctx, weeklyID, value)
			if err != nil || len(value.WeeklyRecurrentSchedule.FridayHours) != 1 || *value.WeeklyRecurrentSchedule.IsComplemented {
				t.Fatal("actual update reply was changed")
			}
			if response, err := s.DeleteByID(ctx, weeklyID); err != nil || response.StatusCode() != 200 {
				t.Fatal("actual delete reply was changed")
			}
			response, err = s.DeleteByID(ctx, weeklyID)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitScheduleRangeDraftsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	timeRange, dateRange, invalid := readFile(t, "testdata/time_range.json"), readFile(t, "testdata/date_time_range.json"), readFile(t, "testdata/time_range_invalid_400.json")
	run(t, contracttest.Case{Name: "actual-drafts", Source: "msteams/teams_admin_api/schedules/testdata/time_range.json",
		Mappings: mapping("New-CsOnlineTimeRange", "New-CsOnlineDateTimeRange"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", collection+"/schedules/time-ranges/draft", "New-CsOnlineTimeRange", json.RawMessage(`{"Start":"9:00","End":"17:30"}`), 200, timeRange),
			wire("POST", collection+"/schedules/date-time-ranges/draft", "New-CsOnlineDateTimeRange", json.RawMessage(`{"Start":"1/1/2027 10:00","End":"3/1/2027"}`), 200, dateRange),
			wire("POST", collection+"/schedules/date-time-ranges/draft", "New-CsOnlineDateTimeRange", json.RawMessage(`{"Start":"1/1/2027"}`), 200, dateRange),
			wire("POST", collection+"/schedules/time-ranges/draft", "New-CsOnlineTimeRange", json.RawMessage(`{"Start":"9:00","End":"25:00"}`), 400, invalid)},
		Want: contracttest.Expected{Status: 400, Body: invalid, APIError: apiError(400, invalid)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.Schedules
			value, _, err := s.NewTimeRange(ctx, "9:00", "17:30")
			if err != nil || *value != (sched.TimeRange{Start: "09:00:00", End: "17:30:00"}) {
				t.Fatal("time range draft was not decoded")
			}
			dates, _, err := s.NewDateTimeRange(ctx, "1/1/2027 10:00", ptr("3/1/2027"))
			if err != nil || *dates != (sched.DateTimeRange{Start: "01/01/2027 10:00", End: "03/01/2027 00:00"}) {
				t.Fatal("date-time range draft text was not preserved")
			}
			if _, _, err := s.NewDateTimeRange(ctx, "1/1/2027", nil); err != nil {
				t.Fatal("omitted end was sent")
			}
			value, response, err := s.NewTimeRange(ctx, "9:00", "25:00")
			if value != nil {
				t.Fatal("rejected draft returned a value")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitScheduleHolidaysUseAcceptanceAndModelResponses(t *testing.T) {
	t.Parallel()
	missing := readFile(t, "testdata/holidays_missing_404.json")
	content := []byte("Name,StartDateTime1,EndDateTime1\nNew Year,1/1/2027 00:00,2/1/2027 00:00\n")
	encoded, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	records := []byte(`{"HolidayVisualizationRecords":[{"CallFlow":{"Id":"x"},"Year":2027,"Name":"New Year","DateTimeRanges":[{"Start":"2027-01-01T00:00:00","End":"2027-01-02T00:00:00"}],"Extra":1}],"Diagnostics":[]}`)
	exported := []byte(`{"ExportHolidaysResult":{"SerializedHolidayRecords":` + string(encoded) + `},"Diagnostics":[]}`)
	imported := []byte(`{"ImportAutoAttendantHolidaysResult":{"ImportHolidayStatusRecords":[{"Name":"New Year","DataTimeRanges":[{"Start":"2027-01-01T00:00:00","End":"2027-01-02T00:00:00"}],"Succeeded":true,"FailureReason":null}],"AutoAttendant":null},"Diagnostics":[]}`)
	run(t, contracttest.Case{Name: "actual-holidays-missing", Source: "msteams/teams_admin_api/schedules/testdata/holidays_missing_404.json", Mappings: mapping("Get-CsAutoAttendantHolidays"),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", holidays+"?ResponseType=0&Years=2027", "Get-CsAutoAttendantHolidays", nil, 404, missing)},
		Want:  contracttest.Expected{Status: 404, Body: missing, APIError: apiError(404, missing)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.Schedules.ListHolidays(ctx, probe, &sched.HolidayOptions{Years: []string{"2027"}})
			return contracttest.Observation{Value: values, Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "holiday-shapes", Mappings: mapping("Get-CsAutoAttendantHolidays", "Export-CsAutoAttendantHolidays", "Import-CsAutoAttendantHolidays"),
		Steps: []contracttest.Exchange{discovery(t),
			// The module escapes later array items before escaping the whole value.
			wire("GET", holidays+"?Names=a+b%2Cc%252Cd&ResponseType=0&Years=2026%2C2027", "Get-CsAutoAttendantHolidays", nil, 200, records),
			wire("GET", holidays+"?ResponseType=1", "Export-CsAutoAttendantHolidays", nil, 200, exported),
			wire("PUT", holidays, "Import-CsAutoAttendantHolidays", json.RawMessage(`{"SerializedHolidayRecords":`+string(encoded)+`}`), 200, imported)},
		Want: contracttest.Expected{Status: 200, Body: imported},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.Schedules
			values, _, err := s.ListHolidays(ctx, probe, &sched.HolidayOptions{Years: []string{"2026", "2027"}, Names: []string{"a b", "c,d"}})
			if err != nil || len(values) != 1 || *values[0].Year != 2027 || len(values[0].DateTimeRanges) != 1 || string(values[0].CallFlow) != `{"Id":"x"}` || values[0].RawFields["Extra"] == nil {
				t.Fatal("holiday records were not decoded")
			}
			bytesOut, _, err := s.GetHolidaysContent(ctx, probe)
			if err != nil || !bytes.Equal(bytesOut, content) {
				t.Fatal("exported holidays were not decoded to file bytes")
			}
			result, response, err := s.ImportHolidays(ctx, probe, content)
			if err != nil || len(result.Statuses) != 1 || !*result.Statuses[0].Succeeded || result.Statuses[0].FailureReason != nil || len(result.Statuses[0].DateTimeRanges) != 1 {
				t.Fatal("import statuses were not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitScheduleListAndBodyOptions(t *testing.T) {
	t.Parallel()
	list := []byte(`{"Schedules":[{"Id":"a","Name":"n","Type":"Future","Unknown":true}],"Diagnostics":[]}`)
	created := readFile(t, "testdata/create_fixed_201.json")
	recurrence := json.RawMessage(`{"Name":"n","Type":"WeeklyRecurrence","WeeklyRecurrentSchedule":{"SundayHours":[{"Start":"00:00:00","End":"1.00:00:00"}],` +
		`"RecurrenceRange":{"Type":"EndDate","Start":"2027-01-01T00:00:00","End":"2027-12-31T00:00:00"}}}`)
	run(t, contracttest.Case{Name: "list-and-bodies", Mappings: mapping("Get-CsOnlineSchedule", "Set-CsOnlineSchedule"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", collection, "Get-CsOnlineSchedule", nil, 200, list),
			wire("GET", collection+"?Descending=False&ExcludeContent=True&First=5&NameFilter=msteams+sdk&Skip=0", "Get-CsOnlineSchedule", nil, 200, list),
			wire("POST", collection, "New-CsOnlineSchedule", json.RawMessage(`{"Name":"n","FixedSchedule":{"DateTimeRanges":[]}}`), 201, created),
			wire("PUT", collection+"/a", "Set-CsOnlineSchedule", recurrence, 200, created)},
		Want: contracttest.Expected{Status: 200, Body: created},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.Schedules
			values, _, err := s.List(ctx, nil)
			if err != nil || len(values) != 1 || *values[0].Type != "Future" || values[0].WeeklyRecurrentSchedule != nil || values[0].RawFields["Unknown"] == nil {
				t.Fatal("list lost unknown values")
			}
			if _, _, err := s.List(ctx, &sched.ListOptions{First: ptr(int32(5)), Skip: ptr(int32(0)), ExcludeContent: ptr(true), Descending: ptr(false), NameFilter: ptr("msteams sdk")}); err != nil {
				t.Fatal("list options were not sent")
			}
			if _, _, err := s.Create(ctx, sched.CreateRequest{Name: "n", FixedSchedule: &sched.FixedSchedule{}}); err != nil {
				t.Fatal("empty fixed schedule was not sent as an empty array")
			}
			value, response, err := s.UpdateByID(ctx, "a", &sched.Schedule{ID: ptr("ignored"), Name: ptr("n"), Type: ptr(sched.TypeWeeklyRecurrence),
				AssociatedConfigurationIDs: []string{"ignored"}, WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{
					SundayHours:     []sched.TimeRange{{Start: "00:00:00", End: "1.00:00:00"}},
					RecurrenceRange: &sched.RecurrenceRange{Type: ptr("EndDate"), Start: ptr("2027-01-01T00:00:00"), End: ptr("2027-12-31T00:00:00")}}})
			if err != nil || value == nil {
				t.Fatal("update with recurrence range failed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

type invocation func(context.Context, *sched.Service) (any, *resty.Response, error)

func TestUnitScheduleMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		path, command, body string
		invoke              invocation
	}{
		"list-missing": {collection, "Get-CsOnlineSchedule", `{}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) { return s.List(ctx, nil) }},
		"list-null-item": {collection, "Get-CsOnlineSchedule", `{"Schedules":[null]}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.List(ctx, nil)
		}},
		"get-null": {collection + "/a", "Get-CsOnlineSchedule", `{"Schedule":null}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.GetByID(ctx, "a")
		}},
		"get-bad-type": {collection + "/a", "Get-CsOnlineSchedule", `{"Schedule":{"Id":1}}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.GetByID(ctx, "a")
		}},
		"holidays-missing": {holidays + "?ResponseType=0", "Get-CsAutoAttendantHolidays", `{}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.ListHolidays(ctx, probe, nil)
		}},
		"export-null": {holidays + "?ResponseType=1", "Export-CsAutoAttendantHolidays", `{"ExportHolidaysResult":{"SerializedHolidayRecords":null}}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.GetHolidaysContent(ctx, probe)
		}},
		"export-bad-base64": {holidays + "?ResponseType=1", "Export-CsAutoAttendantHolidays", `{"ExportHolidaysResult":{"SerializedHolidayRecords":"!!"}}`, func(ctx context.Context, s *sched.Service) (any, *resty.Response, error) {
			return s.GetHolidaysContent(ctx, probe)
		}},
	} {
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire("GET", c.path, c.command, nil, 200, []byte(c.body))},
			Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.Schedules)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "undeclared-status", Steps: []contracttest.Exchange{discovery(t), wire("DELETE", collection+"/a", "Remove-CsOnlineSchedule", nil, 202, nil)},
		Want: contracttest.Expected{Status: 202, Error: sched.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.Schedules.DeleteByID(ctx, "a")
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitScheduleInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	hours := []sched.TimeRange{{Start: "09:00", End: "17:00"}}
	for name, call := range map[string]func(context.Context, *sched.Service) error{
		"get-blank":     func(ctx context.Context, s *sched.Service) error { _, _, e := s.GetByID(ctx, " "); return e },
		"get-traversal": func(ctx context.Context, s *sched.Service) error { _, _, e := s.GetByID(ctx, ".."); return e },
		"delete-slash":  func(ctx context.Context, s *sched.Service) error { _, e := s.DeleteByID(ctx, "a/b"); return e },
		"create-no-name": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.Create(ctx, sched.CreateRequest{FixedSchedule: &sched.FixedSchedule{}})
			return e
		},
		"create-no-type": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.Create(ctx, sched.CreateRequest{Name: "n"})
			return e
		},
		"create-both-types": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.Create(ctx, sched.CreateRequest{Name: "n", FixedSchedule: &sched.FixedSchedule{}, WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{MondayHours: hours}})
			return e
		},
		"create-blank-range": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.Create(ctx, sched.CreateRequest{Name: "n", FixedSchedule: &sched.FixedSchedule{DateTimeRanges: []sched.DateTimeRange{{Start: "x"}}}})
			return e
		},
		"create-blank-hours": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.Create(ctx, sched.CreateRequest{Name: "n", WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{SundayHours: []sched.TimeRange{{End: "x"}}}})
			return e
		},
		"update-nil": func(ctx context.Context, s *sched.Service) error { _, _, e := s.UpdateByID(ctx, "a", nil); return e },
		"update-no-name": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.UpdateByID(ctx, "a", &sched.Schedule{})
			return e
		},
		"time-blank-end": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.NewTimeRange(ctx, "9:00", "")
			return e
		},
		"date-blank-end": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.NewDateTimeRange(ctx, "1/1/2027", ptr(" "))
			return e
		},
		"holidays-blank": func(ctx context.Context, s *sched.Service) error { _, _, e := s.ListHolidays(ctx, "", nil); return e },
		"holidays-control": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.ListHolidays(ctx, "a", &sched.HolidayOptions{Names: []string{"a\nb"}})
			return e
		},
		"export-slash": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.GetHolidaysContent(ctx, "a/b")
			return e
		},
		"import-empty": func(ctx context.Context, s *sched.Service) error {
			_, _, e := s.ImportHolidays(ctx, "a", nil)
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: sched.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.Schedules)}
			}})
	}
}

func TestUnitScheduleNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *sched.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := sched.NewService(nil).NewTimeRange(ctx, "9:00", "17:00"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
