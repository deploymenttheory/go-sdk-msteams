package onlinevoicemailpolicy_test

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
)

func TestUnitPinnedVoicemailDurationModelObservations(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/voicemail-models.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Cases []struct {
			Ticks *int64
			Wire  json.RawMessage
		}
	}
	if json.Unmarshal(raw, &capture) != nil || len(capture.Cases) != 16 {
		t.Fatal("invalid pinned voicemail observations")
	}
	for _, c := range capture.Cases {
		var value resource.Policy
		if json.Unmarshal(c.Wire, &value) != nil {
			t.Fatal("pinned voicemail model decode failed")
		}
		if c.Ticks == nil {
			if value.MaximumRecordingLength != nil {
				t.Fatal("null duration acquired a value")
			}
		} else {
			if value.MaximumRecordingLength == nil || value.MaximumRecordingLength.Ticks() != *c.Ticks {
				t.Fatal("pinned duration tick precision lost")
			}
			wire, err := json.Marshal(resource.TimeSpanFromTicks(*c.Ticks))
			if err != nil || string(wire) != string(value.RawFields["MaximumRecordingLength"]) {
				t.Fatal("duration formatting differs from pinned Newtonsoft")
			}
		}
		encoded, err := json.Marshal(value)
		var before, after any
		if err != nil || json.Unmarshal(c.Wire, &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("pinned voicemail snapshot changed")
		}
	}
}

func TestUnitTimeSpanBoundsAndSubTickPrecision(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"10675199.02:48:05.4775808", "-10675199.02:48:05.4775809", "999999999999999999999999.00:00:00", "00:60:00", "24:00:00", "00:00:60", "PT90S", "90", "00:00:00.00000001", ""} {
		if _, err := resource.ParseTimeSpanConstant(value); err == nil {
			t.Fatal("invalid or overflowing duration accepted")
		}
	}
	for _, ticks := range []int64{math.MinInt64, -864000000001, -1, 0, 1, 864000000001, math.MaxInt64} {
		value := resource.TimeSpanFromTicks(ticks)
		parsed, err := resource.ParseTimeSpanConstant(value.String())
		if err != nil || parsed.Ticks() != ticks {
			t.Fatal("full-range duration round trip failed")
		}
	}
	for _, duration := range []time.Duration{time.Nanosecond, -time.Nanosecond, 101 * time.Nanosecond} {
		if _, err := resource.TimeSpanFromDuration(duration); err == nil {
			t.Fatal("sub-tick precision silently truncated")
		}
	}
	value, err := resource.TimeSpanFromDuration(90 * time.Second)
	if err != nil || value.String() != "00:01:30" {
		t.Fatal("Go duration conversion differs")
	}
}

func TestUnitVoicemailSnapshotsPreserveAbsenceNullAndFutureFields(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"Identity":"Fixture","Future":false}`, `{"Identity":"Fixture","MaximumRecordingLength":null,"ShareData":"FutureValue","EnableTranscription":false,"Future":{"value":[]}}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil || value.MaximumRecordingLength != nil || value.PrimarySystemPromptLanguage != nil || value.EnableVoicemailTriage != nil {
			t.Fatal("missing policy fields acquired client defaults")
		}
		encoded, err := json.Marshal(value)
		var before, after any
		if err != nil || json.Unmarshal([]byte(raw), &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("wire presence/future fields changed")
		}
	}
}
