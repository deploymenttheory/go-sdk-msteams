package callintegration

import (
	"context"
	"testing"
	"time"

	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func logHolidayFailure(t *testing.T, response *resty.Response, values []*sched.HolidayRecord, err error) {
	t.Helper()
	labkit.LogAccessResponse(t, response, err)
	t.Log("integrated holiday visualization category: " + holidayCategory(values, err))
}

// The pinned Online.Models.DateTimeRange.ParseFrom uses this exact invariant
// reply format. Fixed-schedule request strings use a different ISO format.
func holidayRangeMatches(value sched.DateTimeRange) bool {
	start, err := time.Parse("02/01/2006 15:04", value.Start)
	if err != nil {
		return false
	}
	end, err := time.Parse("02/01/2006 15:04", value.End)
	return err == nil && start.Format("2006-01-02 15:04") == "2027-01-01 00:00" && end.Format("2006-01-02 15:04") == "2027-01-02 00:00"
}

// Categories reveal count/format classes only, never raw dates, names or IDs.
// This diagnostic does not relax the populated holiday boundary assertions.
func holidayCategory(values []*sched.HolidayRecord, err error) string {
	if err != nil {
		return "response-error"
	}
	if len(values) == 0 {
		return "zero-records"
	}
	if len(values) != 1 {
		return "multiple-records"
	}
	if values[0] == nil {
		return "missing-record"
	}
	ranges := values[0].DateTimeRanges
	if len(ranges) == 0 {
		return "zero-ranges"
	}
	if len(ranges) != 1 {
		return "multiple-ranges"
	}
	start, end := ranges[0].Start, ranges[0].End
	if start == "" || end == "" {
		return "empty-date-values"
	}
	if holidayRangeMatches(ranges[0]) {
		return "matched"
	}
	a, aerr := time.Parse("2006-01-02T15:04:05", start)
	b, berr := time.Parse("2006-01-02T15:04:05", end)
	if aerr == nil && berr == nil {
		if a.Format("2006-01-02 15:04:05") == "2027-01-01 00:00:00" && b.Format("2006-01-02 15:04:05") == "2027-01-02 00:00:00" {
			return "equivalent-iso-format"
		}
		return "different-iso-boundary"
	}
	for _, format := range []struct{ layout, category string }{{time.RFC3339Nano, "offset-date-format"}, {"02/01/2006 15:04", "day-first-date-format"}, {time.DateOnly, "date-only-format"}} {
		_, aerr = time.Parse(format.layout, start)
		_, berr = time.Parse(format.layout, end)
		if aerr == nil && berr == nil {
			return format.category
		}
	}
	return "unrecognized-date-format"
}

func TestUnitHolidayDiagnosticsPublishBoundedCountAndDateCategories(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, start, end, want string }{
		{"exact", "01/01/2027 00:00", "02/01/2027 00:00", "matched"},
		{"precision", "2027-01-01T00:00:00.000", "2027-01-02T00:00:00.000", "equivalent-iso-format"},
		{"different-boundary", "2027-01-03T00:00:00", "2027-01-04T00:00:00", "different-iso-boundary"},
		{"offset", "2027-01-01T00:00:00Z", "2027-01-02T00:00:00Z", "offset-date-format"},
		{"day-first", "03/01/2027 00:00", "04/01/2027 00:00", "day-first-date-format"},
		{"date-only", "2027-01-01", "2027-01-02", "date-only-format"},
		{"private", "private-tenant-value", "private-resource-value", "unrecognized-date-format"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := []*sched.HolidayRecord{{DateTimeRanges: []sched.DateTimeRange{{Start: tc.start, End: tc.end}}}}
			if holidayCategory(values, nil) != tc.want {
				t.Fatal("bounded holiday format category changed")
			}
		})
	}
	for _, tc := range []struct {
		values []*sched.HolidayRecord
		err    error
		want   string
	}{
		{nil, context.DeadlineExceeded, "response-error"}, {nil, nil, "zero-records"},
		{[]*sched.HolidayRecord{nil, nil}, nil, "multiple-records"}, {[]*sched.HolidayRecord{nil}, nil, "missing-record"},
		{[]*sched.HolidayRecord{{}}, nil, "zero-ranges"},
		{[]*sched.HolidayRecord{{DateTimeRanges: []sched.DateTimeRange{{}, {}}}}, nil, "multiple-ranges"},
		{[]*sched.HolidayRecord{{DateTimeRanges: []sched.DateTimeRange{{}}}}, nil, "empty-date-values"},
	} {
		if holidayCategory(tc.values, tc.err) != tc.want {
			t.Fatal("bounded holiday count/error category changed")
		}
	}
}

func TestUnitHolidayRangeRetainsPinnedFormatAndExactPopulatedBoundaries(t *testing.T) {
	t.Parallel()
	for _, value := range []sched.DateTimeRange{
		{Start: "01/01/2027 00:00", End: "01/01/2027 00:00"},
		{Start: "01/01/2027 00:00", End: "31/12/2026 00:00"},
		{Start: "32/01/2027 00:00", End: "02/01/2027 00:00"},
		{Start: "02/01/2027 00:00", End: "03/01/2027 00:00"},
		{Start: "01/01/2027 00:01", End: "02/01/2027 00:00"},
		{Start: "2027-01-01T00:00:00", End: "2027-01-02T00:00:00"},
	} {
		if holidayRangeMatches(value) {
			t.Fatal("holiday assertion accepted a wrong boundary or unobserved format")
		}
	}
	if !holidayRangeMatches(sched.DateTimeRange{Start: "01/01/2027 00:00", End: "02/01/2027 00:00"}) {
		t.Fatal("holiday assertion rejected the pinned reply format")
	}
}
