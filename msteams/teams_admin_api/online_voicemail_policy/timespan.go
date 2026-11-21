package onlinevoicemailpolicy

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

const ticksPerSecond uint64 = 10_000_000
const ticksPerDay = ticksPerSecond * 86_400

// TimeSpan stores signed .NET 100ns ticks, including the full Int64 range.
// It does not imply a permitted backend recording length or a service default.
type TimeSpan struct{ ticks int64 }

func TimeSpanFromTicks(ticks int64) TimeSpan { return TimeSpan{ticks: ticks} }

// TimeSpanFromDuration rejects sub-tick precision instead of silently truncating.
func TimeSpanFromDuration(value time.Duration) (TimeSpan, error) {
	if value%100 != 0 {
		return TimeSpan{}, ErrInput
	}
	return TimeSpanFromTicks(int64(value / 100)), nil
}
func (v TimeSpan) Ticks() int64 { return v.ticks }

// String is the invariant constant format emitted by pinned Newtonsoft.
func (v TimeSpan) String() string {
	negative := v.ticks < 0
	// Divide before changing sign: days and the sub-day remainder are both
	// safely negatable even when ticks is Int64 minimum.
	days := v.ticks / int64(ticksPerDay)
	remainder := v.ticks % int64(ticksPerDay)
	if negative {
		days, remainder = -days, -remainder
	}
	seconds, fraction := remainder/int64(ticksPerSecond), remainder%int64(ticksPerSecond)
	value := fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
	if days > 0 {
		value = fmt.Sprintf("%d.%s", days, value)
	}
	if fraction > 0 {
		value += fmt.Sprintf(".%07d", fraction)
	}
	if negative {
		value = "-" + value
	}
	return value
}
func (v TimeSpan) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }

var constantTimeSpan = regexp.MustCompile(`^(-)?(?:([0-9]+)\.)?([0-9]{2}):([0-9]{2}):([0-9]{2})(?:\.([0-9]{7}))?$`)

// ParseTimeSpanConstant accepts the constant wire form. PowerShell's broader
// input binding and locale-dependent parsing are separately qualified.
func ParseTimeSpanConstant(value string) (TimeSpan, error) {
	parts := constantTimeSpan.FindStringSubmatch(value)
	if parts == nil {
		return TimeSpan{}, ErrInput
	}
	negative := parts[1] != ""
	limit := uint64(math.MaxInt64)
	if negative {
		limit++
	}
	var values [5]uint64
	for i, field := range parts[2:] {
		if field == "" {
			continue
		}
		number, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return TimeSpan{}, ErrInput
		}
		values[i] = number
	}
	if values[1] > 23 || values[2] > 59 || values[3] > 59 {
		return TimeSpan{}, ErrInput
	}
	factors := [...]uint64{ticksPerDay, ticksPerSecond * 3600, ticksPerSecond * 60, ticksPerSecond, 1}
	var total uint64
	for i, number := range values {
		if number > (limit-total)/factors[i] {
			return TimeSpan{}, ErrInput
		}
		total += number * factors[i]
	}
	if negative && total == uint64(math.MaxInt64)+1 {
		return TimeSpanFromTicks(math.MinInt64), nil
	}
	if total > math.MaxInt64 {
		return TimeSpan{}, ErrInput
	}
	ticks := int64(total)
	if negative {
		ticks = -ticks
	}
	return TimeSpanFromTicks(ticks), nil
}

func (v *TimeSpan) UnmarshalJSON(raw []byte) error {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return client.ErrDecode
	}
	value, err := ParseTimeSpanConstant(text)
	if err != nil {
		return client.ErrDecode
	}
	*v = value
	return nil
}
