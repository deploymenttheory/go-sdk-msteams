// Package assignmentmigration models direct-to-group policy assignment migration
// jobs, recommendations and conflicts. Migration and cleanup are explicit actions.
package assignmentmigration

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid assignment migration input")

// StateName is open: the pinned enum names do not prove a complete server state
// machine, terminality or tenant convergence. Future names remain representable.
type StateName string

const (
	NotStarted             StateName = "NotStarted"
	PopulationInProgress   StateName = "PopulationInProgress"
	GPACreationInProgress  StateName = "GpaCreationInProgress"
	WaitingForGPAProcessed StateName = "WaitingForGpaProcessed"
	CleanUpInProgress      StateName = "CleanUpInProgress"
	Completed              StateName = "Completed"
	Failed                 StateName = "Failed"
	ConnectionLost         StateName = "ConnectionLost"
)

// State retains string, Int32 or null enum data without a closed-name decoder.
// Zero state is omitted data; it does not mean NotStarted or a successful job.
type State struct{ raw json.RawMessage }

func (s *State) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		s.raw = slices.Clone(trimmed)
		return nil
	}
	var name string
	if json.Unmarshal(trimmed, &name) == nil {
		s.raw = slices.Clone(trimmed)
		return nil
	}
	var number int32
	if json.Unmarshal(trimmed, &number) == nil {
		s.raw = slices.Clone(trimmed)
		return nil
	}
	return client.ErrDecode
}
func (s State) MarshalJSON() ([]byte, error) {
	if s.raw == nil {
		return []byte("null"), nil
	}
	return slices.Clone(s.raw), nil
}
func (s State) Text() (StateName, bool) {
	var name string
	if len(s.raw) == 0 || bytes.Equal(s.raw, []byte("null")) || json.Unmarshal(s.raw, &name) != nil {
		return "", false
	}
	return StateName(name), true
}
func (s State) Number() (int32, bool) {
	var value int32
	if len(s.raw) == 0 || bytes.Equal(s.raw, []byte("null")) || json.Unmarshal(s.raw, &value) != nil {
		return 0, false
	}
	return value, true
}
func (s State) RawJSON() json.RawMessage { return slices.Clone(s.raw) }
func (State) String() string             { return "assignment migration state (redacted)" }
func (s State) GoString() string         { return s.String() }

// Job preserves received presence/future data rather than materializing the
// pinned CLR's absent GUID/DateTime/count defaults. Timestamp text stays intact.
type Job struct {
	MigrationEventID           *string                    `json:"migrationEventId"`
	TenantID                   *string                    `json:"tenantId"`
	GroupID                    *string                    `json:"groupId"`
	GroupName                  *string                    `json:"groupName"`
	PolicyType                 *string                    `json:"policyType"`
	PolicyInstance             *string                    `json:"policyInstance"`
	State                      State                      `json:"state"`
	StartTime                  *string                    `json:"startTime"`
	CompletionTime             *string                    `json:"completionTime"`
	DirectAssignmentsProcessed *int32                     `json:"directAssignmentsProcessed"`
	RawFields                  map[string]json.RawMessage `json:"-"`
}

func (v *Job) UnmarshalJSON(data []byte) error {
	type snapshot Job
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Job(value)
	return nil
}
func (v Job) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Job
	return json.Marshal(snapshot(v))
}
func (Job) String() string     { return "assignment migration job (redacted)" }
func (v Job) GoString() string { return v.String() }

type PolicyTypeRecommendation struct {
	PolicyType             *string                    `json:"policyType"`
	DirectAssignmentsCount *int32                     `json:"daCount"`
	RawFields              map[string]json.RawMessage `json:"-"`
}
type PolicyNameRecommendation struct {
	PolicyName             *string                    `json:"policyName"`
	DirectAssignmentsCount *int32                     `json:"daCount"`
	RawFields              map[string]json.RawMessage `json:"-"`
}
type ConflictDetail struct {
	UserID           *string                    `json:"userId"`
	PolicyInstance   *string                    `json:"policyInstance"`
	AssignmentSource *string                    `json:"assignmentSource"`
	RawFields        map[string]json.RawMessage `json:"-"`
}
type Conflict struct {
	TotalCountUsersWithConflict                 *int32                     `json:"totalCountUsersWithConflict"`
	UsersWithConflictDueDirectAssignmentsCount  *int32                     `json:"usersWithConflictDueDirectAssignmentsCount"`
	UsersWithConflictDueDirectAssignmentsDetail []*ConflictDetail          `json:"usersWithConflictDueDirectAssignmentsDetail"`
	UsersWithConflictDueGroupAssignmentsCount   *int32                     `json:"usersWithConflictDueGroupAssignmentsCount"`
	UsersWithConflictDueGroupAssignmentsDetail  []*ConflictDetail          `json:"usersWithConflictDueGroupAssignmentsDetail"`
	RawFields                                   map[string]json.RawMessage `json:"-"`
}

// RecommendationOptions distinguishes the unbound "500" default from an
// explicitly supplied empty/arbitrary threshold string. No numeric range or
// pagination control is inferred from the command's string parameter.
type RecommendationOptions struct{ GroupThreshold *string }

// CreateRequest is the pinned lowercase migration body. GroupID is a bound GUID;
// AllowNonEmptyGroup is always an explicit query Boolean, never a body field.
type CreateRequest struct {
	Authority          string `json:"authority"`
	PolicyType         string `json:"policytype"`
	PolicyName         string `json:"policyname"`
	GroupID            string `json:"groupid"`
	AllowNonEmptyGroup bool   `json:"-"`
}

func (v *PolicyTypeRecommendation) UnmarshalJSON(data []byte) error {
	type snapshot PolicyTypeRecommendation
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = PolicyTypeRecommendation(value)
	return nil
}
func (v PolicyTypeRecommendation) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot PolicyTypeRecommendation
	return json.Marshal(snapshot(v))
}
func (PolicyTypeRecommendation) String() string     { return "policy type recommendation (redacted)" }
func (v PolicyTypeRecommendation) GoString() string { return v.String() }

func (v *PolicyNameRecommendation) UnmarshalJSON(data []byte) error {
	type snapshot PolicyNameRecommendation
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = PolicyNameRecommendation(value)
	return nil
}
func (v PolicyNameRecommendation) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot PolicyNameRecommendation
	return json.Marshal(snapshot(v))
}
func (PolicyNameRecommendation) String() string     { return "policy name recommendation (redacted)" }
func (v PolicyNameRecommendation) GoString() string { return v.String() }

func (v *ConflictDetail) UnmarshalJSON(data []byte) error {
	type snapshot ConflictDetail
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = ConflictDetail(value)
	return nil
}
func (v ConflictDetail) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot ConflictDetail
	return json.Marshal(snapshot(v))
}
func (ConflictDetail) String() string     { return "assignment conflict detail (redacted)" }
func (v ConflictDetail) GoString() string { return v.String() }

func (v *Conflict) UnmarshalJSON(data []byte) error {
	type snapshot Conflict
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Conflict(value)
	return nil
}
func (v Conflict) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Conflict
	return json.Marshal(snapshot(v))
}
func (Conflict) String() string     { return "assignment conflict (redacted)" }
func (v Conflict) GoString() string { return v.String() }
