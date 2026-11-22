package teamsupdatemanagementpolicy_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
)

// Pure serialization vectors were checked against the pinned provider's loaded
// Json.NET serializer; these are caller inputs, not fabricated HTTP responses.
func TestUnitDateTimePreservesPinnedKindsAndTickPrecision(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, input, want string
		kind              resource.DateTimeKind
	}{
		{"unspecified", "2026-04-12T18:32:44", "2026-04-12T18:32:44", resource.DateTimeUnspecified},
		{"unspecified-ticks", "2026-04-12T18:32:44.1234567", "2026-04-12T18:32:44.1234567", resource.DateTimeUnspecified},
		{"utc", "2026-04-12T18:32:44Z", "2026-04-12T18:32:44Z", resource.DateTimeUTC},
		{"utc-ticks", "2026-04-12T18:32:44.1234567Z", "2026-04-12T18:32:44.1234567Z", resource.DateTimeUTC},
		{"trailing-zeroes", "2026-04-12T18:32:44.1234000Z", "2026-04-12T18:32:44.1234Z", resource.DateTimeUTC},
		{"minimum", "0001-01-01T00:00:00", "0001-01-01T00:00:00", resource.DateTimeUnspecified},
		{"maximum", "9999-12-31T23:59:59.9999999Z", "9999-12-31T23:59:59.9999999Z", resource.DateTimeUTC},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, e := resource.ParseDateTime(tc.input)
			if e != nil || value.Kind() != tc.kind {
				t.Fatal("source DateTime kind rejected or changed")
			}
			raw, e := json.Marshal(value)
			var text string
			if e != nil || json.Unmarshal(raw, &text) != nil || text != tc.want {
				t.Fatal("source DateTime representation or precision changed")
			}
			var roundtrip resource.DateTime
			if json.Unmarshal(raw, &roundtrip) != nil || roundtrip.Kind() != value.Kind() || !roundtrip.Time().Equal(value.Time()) {
				t.Fatal("DateTime roundtrip lost kind or ticks")
			}
		})
	}
}
func TestUnitOffsetDateTimeUsesHostLocalLikePinnedParser(t *testing.T) {
	t.Parallel()
	value, e := resource.ParseDateTime("2026-04-12T18:32:44.1234567+05:30")
	expected := time.Date(2026, 4, 12, 18, 32, 44, 123456700, time.FixedZone("input", 330*60)).In(time.Local)
	if e != nil || value.Kind() != resource.DateTimeLocal || !value.Time().Equal(expected) {
		t.Fatal("offset DateTime lost its instant or local kind")
	}
	raw, e := json.Marshal(value)
	var text string
	if e != nil || json.Unmarshal(raw, &text) != nil || text != expected.Format("2006-01-02T15:04:05.9999999-07:00") {
		t.Fatal("offset DateTime was not converted to host local representation")
	}
	wall, e := resource.NewDateTime(expected, resource.DateTimeUnspecified)
	if e != nil || wall.Time().Hour() != expected.Hour() || wall.Kind() != resource.DateTimeUnspecified {
		t.Fatal("explicit unspecified kind shifted the supplied wall clock")
	}
}
func TestUnitDateTimeRejectsUnrepresentableOrMalformedInput(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "0000-01-01T00:00:00", "10000-01-01T00:00:00", "2026-02-30T18:32:44", "2026-04-12T18:32:44.12345678Z", "2026-04-12", "2026-04-12T18:32:44junk"} {
		if _, e := resource.ParseDateTime(value); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid DateTime accepted")
		}
	}
	for _, raw := range []string{`null`, `false`, `123`, `"invalid"`} {
		var value resource.DateTime
		if e := json.Unmarshal([]byte(raw), &value); !errors.Is(e, client.ErrDecode) {
			t.Fatal("invalid DateTime JSON accepted")
		}
	}
	if _, e := resource.NewDateTime(time.Date(2026, 4, 12, 18, 32, 44, 1, time.UTC), resource.DateTimeUTC); !errors.Is(e, resource.ErrInput) {
		t.Fatal("sub-tick precision silently rounded")
	}
	if _, e := resource.NewDateTime(time.Now(), resource.DateTimeKind(8)); !errors.Is(e, resource.ErrInput) {
		t.Fatal("unknown DateTime kind accepted")
	}
}
