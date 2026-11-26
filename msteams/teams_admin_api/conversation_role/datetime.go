package conversationrole

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// DateTimeKind selects the CLR DateTime representation used by the pinned provider.
type DateTimeKind uint8

const (
	DateTimeUnspecified DateTimeKind = iota
	DateTimeUTC
	DateTimeLocal
)

// DateTime preserves timezone-free wall time, UTC or host-local time at CLR's
// 100-nanosecond precision. Its zero value is the unspecified CLR minimum.
type DateTime struct {
	value time.Time
	kind  DateTimeKind
}

var dateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,7})?(?:Z|[+-]\d{2}:\d{2})?$`)

// NewDateTime selects kind explicitly. Unspecified retains the supplied wall
// clock; UTC and Local convert the instant. Finer than 100ns input is rejected.
func NewDateTime(value time.Time, kind DateTimeKind) (DateTime, error) {
	if kind > DateTimeLocal || value.Nanosecond()%100 != 0 {
		return DateTime{}, ErrInput
	}
	switch kind {
	case DateTimeUTC:
		value = value.UTC()
	case DateTimeLocal:
		value = value.In(time.Local)
	case DateTimeUnspecified:
		value = time.Date(value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), time.UTC)
	}
	if value.Year() < 1 || value.Year() > 9999 {
		return DateTime{}, ErrInput
	}
	return DateTime{value: value, kind: kind}, nil
}

// ParseDateTime accepts the provider's ISO form, including an unspecified clock.
// Offset input becomes host-local, matching the pinned Json.NET DateTime parser.
func ParseDateTime(value string) (DateTime, error) {
	if !dateTimePattern.MatchString(value) {
		return DateTime{}, ErrInput
	}
	kind := DateTimeUnspecified
	layout := "2006-01-02T15:04:05"
	if strings.HasSuffix(value, "Z") {
		kind = DateTimeUTC
		layout = time.RFC3339Nano
	} else if len(value) > 19 && (value[len(value)-6] == '+' || value[len(value)-6] == '-') {
		kind = DateTimeLocal
		layout = time.RFC3339Nano
	}
	parsed, e := time.Parse(layout, value)
	if e != nil {
		return DateTime{}, ErrInput
	}
	return NewDateTime(parsed, kind)
}
func (v DateTime) Kind() DateTimeKind { return v.kind }

// Time returns the stored clock. Check Kind before treating unspecified as an instant.
func (v DateTime) Time() time.Time { return v.value }
func (v DateTime) MarshalJSON() ([]byte, error) {
	value, e := NewDateTime(v.value, v.kind)
	if e != nil {
		return nil, e
	}
	layout := "2006-01-02T15:04:05.9999999"
	switch value.kind {
	case DateTimeUTC:
		layout += "Z"
	case DateTimeLocal:
		layout += "-07:00"
	}
	return json.Marshal(value.value.Format(layout))
}
func (v *DateTime) UnmarshalJSON(raw []byte) error {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return client.ErrDecode
	}
	value, e := ParseDateTime(text)
	if e != nil {
		return client.ErrDecode
	}
	*v = value
	return nil
}
func (DateTime) String() string     { return "conversation DateTime (redacted)" }
func (v DateTime) GoString() string { return v.String() }
