package licensereport

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
)

var ErrSourceShape = errors.New("license report data cannot represent the pinned projection")

// PinnedPlan preserves the source's nullable string ID, not a normalized Guid.
// Matching is ordinal and case-sensitive, even though Graph documents a Guid.
type PinnedPlan struct{ ID *string }

func (PinnedPlan) String() string     { return "pinned plan value (redacted)" }
func (v PinnedPlan) GoString() string { return v.String() }

// PinnedLicense preserves null versus empty nested collections and null entries.
// A nil ServicePlans collection throws in the source predicate; empty is false.
type PinnedLicense struct{ ServicePlans []*PinnedPlan }

func (PinnedLicense) String() string     { return "pinned license value (redacted)" }
func (v PinnedLicense) GoString() string { return v.String() }

// PinnedRow is the source's local output projection. Its string verdict is not
// a native entitlement/authorization result. Missing license keys omit rows;
// present-null collections yield an empty string rather than a false verdict.
type PinnedRow struct {
	UserPrincipalName      string
	DisplayName            string
	NumberOfMessagesSent   int64
	HasRequiredLicensePlan string
}

func (PinnedRow) String() string     { return "pinned license report row (redacted)" }
func (v PinnedRow) GoString() string { return v.String() }

// ProjectPinnedRows is a pure compatibility projection, never a Graph request.
// Native read/composition data remains available independently when source Int32
// limits, duplicate UPNs or malformed nested values cannot be projected exactly.
func ProjectPinnedRows(activity []Activity, users []DirectoryUser, licenses map[string][]*PinnedLicense) ([]PinnedRow, error) {
	if licenses == nil {
		return nil, ErrSourceShape
	}
	byUPN := make(map[string]DirectoryUser)
	for _, user := range users {
		if user.UserPrincipalName == nil {
			return nil, ErrSourceShape
		}
		key := *user.UserPrincipalName
		if _, exists := byUPN[key]; exists {
			return nil, ErrSourceShape
		}
		byUPN[key] = user
	}
	result := make([]PinnedRow, 0, len(activity))
	for _, a := range activity {
		if a.UserPrincipalName == nil {
			return result, ErrSourceShape
		}
		upn := *a.UserPrincipalName
		values, present := licenses[upn]
		if !present {
			continue
		}
		private, e := sourceCount(a, "privateChatMessageCount", a.PrivateChatMessageCount)
		if e != nil {
			return result, e
		}
		team, e := sourceCount(a, "teamChatMessageCount", a.TeamChatMessageCount)
		if e != nil {
			return result, e
		}
		// The source IL performs unchecked Int32 add, then widens to Int64.
		count := int64(int32(uint32(private) + uint32(team)))
		row := PinnedRow{UserPrincipalName: upn, DisplayName: "Not Found", NumberOfMessagesSent: count}
		if user, found := byUPN[upn]; found && user.DisplayName != nil {
			row.DisplayName = *user.DisplayName
		}
		if values != nil {
			match, e := pinnedMatch(values)
			if e != nil {
				return result, e
			}
			if match {
				row.HasRequiredLicensePlan = "True"
			} else {
				row.HasRequiredLicensePlan = "False"
			}
		}
		result = append(result, row)
	}
	return result, nil
}
func sourceCount(activity Activity, key string, value *json.Number) (int32, error) {
	if value == nil {
		if raw, exists := activity.RawFields[key]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return 0, ErrSourceShape
		}
		return 0, nil
	}
	n, e := integer(value)
	if e != nil || !n.IsInt64() {
		return 0, ErrSourceRange
	}
	v := n.Int64()
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, ErrSourceRange
	}
	return int32(v), nil
}
func pinnedMatch(licenses []*PinnedLicense) (bool, error) {
	for _, license := range licenses {
		if license == nil || license.ServicePlans == nil {
			return false, ErrSourceShape
		}
		for _, plan := range license.ServicePlans {
			if plan == nil {
				return false, ErrSourceShape
			}
			if plan.ID != nil && *plan.ID == SourcePlanID {
				return true, nil
			}
		}
	}
	return false, nil
}
