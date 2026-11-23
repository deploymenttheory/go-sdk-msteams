package schedules

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Schedule types as the service spells them.
const (
	TypeFixed            = "Fixed"
	TypeWeeklyRecurrence = "WeeklyRecurrence"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("schedule input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected schedule response")

func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// DateTimeRange is a fixed-schedule interval. The service reads and returns
// yyyy-MM-ddTHH:mm:ss without an offset, in the auto attendant's time zone.
type DateTimeRange struct {
	Start string `json:"Start"`
	End   string `json:"End"`
}

// TimeRange is a time of day interval such as 09:00:00–17:00:00. End may be
// 1.00:00:00 for the end of the day.
type TimeRange struct {
	Start string `json:"Start"`
	End   string `json:"End"`
}

// FixedSchedule holds the date-time ranges of a Fixed schedule.
type FixedSchedule struct {
	DateTimeRanges []DateTimeRange `json:"DateTimeRanges"`
}

// RecurrenceRange bounds a weekly schedule. Start and End keep the service text.
type RecurrenceRange struct {
	Type                *string `json:"Type,omitempty"`
	Start               *string `json:"Start,omitempty"`
	End                 *string `json:"End,omitempty"`
	NumberOfOccurrences *int32  `json:"NumberOfOccurrences,omitempty"`
}

// WeeklyRecurrentSchedule holds the open hours of each day. IsComplemented
// inverts the schedule (the hours outside the listed ranges).
type WeeklyRecurrentSchedule struct {
	MondayHours     []TimeRange      `json:"MondayHours"`
	TuesdayHours    []TimeRange      `json:"TuesdayHours"`
	WednesdayHours  []TimeRange      `json:"WednesdayHours"`
	ThursdayHours   []TimeRange      `json:"ThursdayHours"`
	FridayHours     []TimeRange      `json:"FridayHours"`
	SaturdayHours   []TimeRange      `json:"SaturdayHours"`
	SundayHours     []TimeRange      `json:"SundayHours"`
	IsComplemented  *bool            `json:"IsComplemented"`
	RecurrenceRange *RecurrenceRange `json:"RecurrenceRange"`
}

// Schedule is an auto attendant schedule. Absent and null fields stay nil;
// RawFields keeps the complete service object.
type Schedule struct {
	ID                         *string                    `json:"Id"`
	Name                       *string                    `json:"Name"`
	Type                       *string                    `json:"Type"`
	WeeklyRecurrentSchedule    *WeeklyRecurrentSchedule   `json:"WeeklyRecurrentSchedule"`
	FixedSchedule              *FixedSchedule             `json:"FixedSchedule"`
	AssociatedConfigurationIDs []string                   `json:"AssociatedConfigurationIds"`
	RawFields                  map[string]json.RawMessage `json:"-"`
}

func (v *Schedule) UnmarshalJSON(raw []byte) error {
	type plain Schedule
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Schedule(value)
	return nil
}

// ListOptions are the generated client's optional list queries; the module's
// Get-CsOnlineSchedule sends none of them. Nil fields are omitted.
type ListOptions struct {
	First          *int32
	Skip           *int32
	ExcludeContent *bool
	NameFilter     *string
	TypeFilter     *string
	SortBy         *string
	Descending     *bool
}

// CreateRequest is New-CsOnlineSchedule. Exactly one of FixedSchedule and
// WeeklyRecurrentSchedule must be set, as the module requires -FixedSchedule
// or -WeeklyRecurrentSchedule.
type CreateRequest struct {
	Name                    string
	FixedSchedule           *FixedSchedule
	WeeklyRecurrentSchedule *WeeklyRecurrentSchedule
}

// HolidayOptions filters Get-CsAutoAttendantHolidays. Empty slices are omitted.
type HolidayOptions struct {
	Years []string
	Names []string
}

// HolidayRecord is one holiday call handling entry of an auto attendant.
// CallFlow is kept as raw JSON; its model belongs to auto attendants.
type HolidayRecord struct {
	Name           *string                    `json:"Name"`
	Year           *int32                     `json:"Year"`
	DateTimeRanges []DateTimeRange            `json:"DateTimeRanges"`
	CallFlow       json.RawMessage            `json:"CallFlow"`
	RawFields      map[string]json.RawMessage `json:"-"`
}

func (v *HolidayRecord) UnmarshalJSON(raw []byte) error {
	type plain HolidayRecord
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = HolidayRecord(value)
	return nil
}

// HolidayImportStatus reports one imported holiday. The service spells the
// ranges property DataTimeRanges.
type HolidayImportStatus struct {
	Name           *string         `json:"Name"`
	DateTimeRanges []DateTimeRange `json:"DataTimeRanges"`
	Succeeded      *bool           `json:"Succeeded"`
	FailureReason  *string         `json:"FailureReason"`
}

// HolidayImportResult is the Import-CsAutoAttendantHolidays outcome.
// AutoAttendant is kept as raw JSON; its model belongs to auto attendants.
type HolidayImportResult struct {
	Statuses      []*HolidayImportStatus `json:"ImportHolidayStatusRecords"`
	AutoAttendant json.RawMessage        `json:"AutoAttendant"`
}
