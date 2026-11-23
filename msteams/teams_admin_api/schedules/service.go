// Package schedules manages auto attendant schedules (fixed date-time ranges
// and weekly recurrent hours), the service-validated range builders, and auto
// attendant holiday import/export.
package schedules

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	voiceApps = "Teams.VoiceApps"
	schedules = "schedules"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func validText(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsFunc(value, unicode.IsControl)
}

func validPart(value string) bool {
	return validText(value) && value != "." && value != ".." && !strings.ContainsAny(value, `/\`)
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// envelope decodes a required, non-null property of a voice apps response.
func envelope(raw []byte, name string, out any) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value, ok := fields[name]
	if !ok || string(value) == "null" || json.Unmarshal(value, out) != nil {
		return client.ErrDecode
	}
	return nil
}

func decodeSchedule(response *resty.Response) (*Schedule, error) {
	var value Schedule
	if envelope(response.Body(), "Schedule", &value) != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}

func validRanges[T DateTimeRange | TimeRange](values []T) bool {
	for _, value := range values {
		var start, end string
		switch v := any(value).(type) {
		case DateTimeRange:
			start, end = v.Start, v.End
		case TimeRange:
			start, end = v.Start, v.End
		}
		if !validText(start) || !validText(end) {
			return false
		}
	}
	return true
}

func validWeekly(w *WeeklyRecurrentSchedule) bool {
	for _, day := range [][]TimeRange{w.MondayHours, w.TuesdayHours, w.WednesdayHours, w.ThursdayHours, w.FridayHours, w.SaturdayHours, w.SundayHours} {
		if !validRanges(day) {
			return false
		}
	}
	return true
}

// fixedBody always sends the array, as -FixedSchedule without ranges does.
func fixedBody(f *FixedSchedule) map[string]any {
	ranges := f.DateTimeRanges
	if ranges == nil {
		ranges = []DateTimeRange{}
	}
	return map[string]any{"DateTimeRanges": ranges}
}

// weeklyBody omits empty days, as the module does, and nil optional fields.
func weeklyBody(w *WeeklyRecurrentSchedule) map[string]any {
	body := map[string]any{}
	for name, day := range map[string][]TimeRange{"MondayHours": w.MondayHours, "TuesdayHours": w.TuesdayHours, "WednesdayHours": w.WednesdayHours,
		"ThursdayHours": w.ThursdayHours, "FridayHours": w.FridayHours, "SaturdayHours": w.SaturdayHours, "SundayHours": w.SundayHours} {
		if len(day) > 0 {
			body[name] = day
		}
	}
	if w.IsComplemented != nil {
		body["IsComplemented"] = *w.IsComplemented
	}
	if w.RecurrenceRange != nil {
		body["RecurrenceRange"] = w.RecurrenceRange
	}
	return body
}

// List reads schedules (Get-CsOnlineSchedule without Id). Nil options send no
// query, as the module does.
func (s *Service) List(ctx context.Context, options *ListOptions) ([]*Schedule, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if options != nil {
		if options.First != nil {
			query.Set("First", strconv.FormatInt(int64(*options.First), 10))
		}
		if options.Skip != nil {
			query.Set("Skip", strconv.FormatInt(int64(*options.Skip), 10))
		}
		for name, value := range map[string]*bool{"ExcludeContent": options.ExcludeContent, "Descending": options.Descending} {
			if value != nil {
				query.Set(name, map[bool]string{true: "True", false: "False"}[*value])
			}
		}
		for name, value := range map[string]*string{"NameFilter": options.NameFilter, "TypeFilter": options.TypeFilter, "SortBy": options.SortBy} {
			if value != nil {
				query.Set(name, *value)
			}
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, schedules}, Query: query,
		Header: headers("Get-CsOnlineSchedule"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*Schedule
	if envelope(response.Body(), "Schedules", &values) != nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// GetByID reads one schedule (Get-CsOnlineSchedule -Id). A missing schedule
// is an HTTP 404 API error.
func (s *Service) GetByID(ctx context.Context, id string) (*Schedule, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, schedules, id},
		Header: headers("Get-CsOnlineSchedule"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeSchedule(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// Create creates a schedule (New-CsOnlineSchedule). It is never replayed
// automatically; reconcile a lost response with List and the name.
func (s *Service) Create(ctx context.Context, request CreateRequest) (*Schedule, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Name) || (request.FixedSchedule == nil) == (request.WeeklyRecurrentSchedule == nil) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Name": request.Name}
	if request.FixedSchedule != nil {
		if !validRanges(request.FixedSchedule.DateTimeRanges) {
			return nil, nil, ErrInput
		}
		body["FixedSchedule"] = fixedBody(request.FixedSchedule)
	} else {
		if !validWeekly(request.WeeklyRecurrentSchedule) {
			return nil, nil, ErrInput
		}
		body["WeeklyRecurrentSchedule"] = weeklyBody(request.WeeklyRecurrentSchedule)
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, schedules},
		Header: headers("New-CsOnlineSchedule"), Body: body}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeSchedule(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// UpdateByID replaces a schedule (Set-CsOnlineSchedule -Instance). Like the
// module it sends Name, Type and the populated schedule body; ID and
// AssociatedConfigurationIDs are not sent. Pass a schedule read with GetByID.
func (s *Service) UpdateByID(ctx context.Context, id string, schedule *Schedule) (*Schedule, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) || schedule == nil || schedule.Name == nil || !validText(*schedule.Name) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Name": *schedule.Name}
	if schedule.Type != nil {
		body["Type"] = *schedule.Type
	}
	if schedule.FixedSchedule != nil {
		if !validRanges(schedule.FixedSchedule.DateTimeRanges) {
			return nil, nil, ErrInput
		}
		body["FixedSchedule"] = fixedBody(schedule.FixedSchedule)
	}
	if schedule.WeeklyRecurrentSchedule != nil {
		if !validWeekly(schedule.WeeklyRecurrentSchedule) {
			return nil, nil, ErrInput
		}
		body["WeeklyRecurrentSchedule"] = weeklyBody(schedule.WeeklyRecurrentSchedule)
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: []string{voiceApps, schedules, id},
		Header: headers("Set-CsOnlineSchedule"), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeSchedule(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// DeleteByID removes a schedule (Remove-CsOnlineSchedule). A missing schedule
// is an HTTP 404 API error.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{voiceApps, schedules, id},
		Header: headers("Remove-CsOnlineSchedule")}, http.StatusOK, http.StatusNoContent)
}

// NewDateTimeRange asks the service to validate and normalize a date-time range
// (New-CsOnlineDateTimeRange). It is a draft: nothing is stored. Accepted
// formats are d/M/yyyy H:mm, d/M/yyyy and yyyy-MM-ddTHH:mm:ss; a nil end lets
// the service choose the end (the start of the following day). The returned
// strings are the service's text unchanged.
func (s *Service) NewDateTimeRange(ctx context.Context, start string, end *string) (*DateTimeRange, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(start) || (end != nil && !validText(*end)) {
		return nil, nil, ErrInput
	}
	body := map[string]string{"Start": start}
	if end != nil {
		body["End"] = *end
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, schedules, schedules, "date-time-ranges", "draft"},
		Header: headers("New-CsOnlineDateTimeRange"), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value DateTimeRange
	if envelope(response.Body(), "DateTimeRange", &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// NewTimeRange asks the service to validate and normalize a time of day range
// (New-CsOnlineTimeRange), for example 9:00–17:30. It is a draft: nothing is
// stored. The service returns hh:mm:ss text, with 1.00:00:00 for end of day.
func (s *Service) NewTimeRange(ctx context.Context, start, end string) (*TimeRange, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(start) || !validText(end) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, schedules, schedules, "time-ranges", "draft"},
		Header: headers("New-CsOnlineTimeRange"), Body: map[string]string{"Start": start, "End": end}}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value TimeRange
	if envelope(response.Body(), "TimeRange", &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// escapeDataString is .NET Uri.EscapeDataString: only RFC 3986 unreserved
// characters are left as-is.
func escapeDataString(value string) string {
	var out strings.Builder
	for _, b := range []byte(value) {
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~' {
			out.WriteByte(b)
		} else {
			fmt.Fprintf(&out, "%%%02X", b)
		}
	}
	return out.String()
}

// joinArray reproduces the generated client's query aggregation: the first
// item verbatim, each later item escaped, joined by commas.
func joinArray(values []string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		if i == 0 {
			parts[i] = value
		} else {
			parts[i] = escapeDataString(value)
		}
	}
	return strings.Join(parts, ",")
}

func holidayQuery(responseType int, options *HolidayOptions) (url.Values, bool) {
	query := url.Values{"ResponseType": {strconv.Itoa(responseType)}}
	if options == nil {
		return query, true
	}
	for name, values := range map[string][]string{"Years": options.Years, "Names": options.Names} {
		for _, value := range values {
			if strings.ContainsFunc(value, unicode.IsControl) {
				return nil, false
			}
		}
		if len(values) > 0 {
			query.Set(name, joinArray(values))
		}
	}
	return query, true
}

// ListHolidays reads the holiday call handling of an auto attendant
// (Get-CsAutoAttendantHolidays). A missing auto attendant is an HTTP 404.
func (s *Service) ListHolidays(ctx context.Context, autoAttendantID string, options *HolidayOptions) ([]*HolidayRecord, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query, ok := holidayQuery(0, options)
	if !ok || !validPart(autoAttendantID) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, "auto-attendants", autoAttendantID, "holidays"}, Query: query,
		Header: headers("Get-CsAutoAttendantHolidays"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*HolidayRecord
	if envelope(response.Body(), "HolidayVisualizationRecords", &values) != nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// GetHolidaysContent exports an auto attendant's holidays as the serialized
// file bytes (Export-CsAutoAttendantHolidays), suitable for ImportHolidays.
func (s *Service) GetHolidaysContent(ctx context.Context, autoAttendantID string) ([]byte, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(autoAttendantID) {
		return nil, nil, ErrInput
	}
	query, _ := holidayQuery(1, nil)
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, "auto-attendants", autoAttendantID, "holidays"}, Query: query,
		Header: headers("Export-CsAutoAttendantHolidays"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var result struct {
		Records *string `json:"SerializedHolidayRecords"`
	}
	if envelope(response.Body(), "ExportHolidaysResult", &result) != nil || result.Records == nil {
		return nil, response, client.ErrDecode
	}
	content, err := base64.StdEncoding.DecodeString(*result.Records)
	if err != nil {
		return nil, response, client.ErrDecode
	}
	return content, response, nil
}

// ImportHolidays replaces an auto attendant's holidays from serialized file
// bytes (Import-CsAutoAttendantHolidays). Per-holiday outcomes are in the
// result; it is never replayed automatically.
func (s *Service) ImportHolidays(ctx context.Context, autoAttendantID string, content []byte) (*HolidayImportResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(autoAttendantID) || len(content) == 0 {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: []string{voiceApps, "auto-attendants", autoAttendantID, "holidays"},
		Header: headers("Import-CsAutoAttendantHolidays"), Body: map[string]string{"SerializedHolidayRecords": base64.StdEncoding.EncodeToString(content)}}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value HolidayImportResult
	if envelope(response.Body(), "ImportAutoAttendantHolidaysResult", &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}
