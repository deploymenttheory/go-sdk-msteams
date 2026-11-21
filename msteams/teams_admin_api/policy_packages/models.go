// Package policypackages manages native custom packages and explicit assignments.
package policypackages

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	ErrInput    = errors.New("invalid policy package input")
	ErrResponse = errors.New("unexpected policy package response")
)

// Package covers the generated package, summary and recommendation projections.
// Policies keeps its open dictionary shape. RawFields preserves received missing,
// null and future data; snapshots do not invent defaults or local composition.
type Package struct {
	Name               *string                    `json:"Name"`
	Description        *string                    `json:"Description"`
	PackageType        *string                    `json:"PackageType"`
	RecommendationType *string                    `json:"RecommendationType"`
	Policies           map[string]json.RawMessage `json:"Policies"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

// UserErrors retains the known generated LegacyUser projection and future keys.
type UserErrors struct {
	LegacyUser []*string                  `json:"LegacyUser"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

// AssignmentResult preserves partial business failures independently of HTTP
// acceptance. OperationID is optional and opaque; no polling path is inferred.
// Typed fields are also retained for decoded non-success operation responses.
type AssignmentResult struct {
	HTTPAccepted            bool                       `json:"-"`
	ResponseMessage         *string                    `json:"ResponseMessage"`
	FailedAssignPolicies    []*string                  `json:"FailedAssignPolicies"`
	FailedCreatePolicies    []*string                  `json:"FailedCreatePolicies"`
	FailedUserIDs           []*string                  `json:"FailedUserIds"`
	FailedUserErrorMessages *UserErrors                `json:"FailedUserErrorMessages"`
	OperationID             *string                    `json:"OperationId"`
	RawFields               map[string]json.RawMessage `json:"-"`
}

// MessageResult retains custom-create/update business data and HTTP acceptance,
// including when a recognized successful response cannot be decoded.
type MessageResult struct {
	HTTPAccepted    bool                       `json:"-"`
	ResponseMessage *string                    `json:"ResponseMessage"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func snapshot(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}
func marshalSnapshot(value any, fields map[string]json.RawMessage) ([]byte, error) {
	if fields != nil {
		return json.Marshal(fields)
	}
	return json.Marshal(value)
}
func (v *Package) UnmarshalJSON(raw []byte) error {
	type plain Package
	var value plain
	if e := snapshot(raw, &value, &value.RawFields); e != nil {
		return e
	}
	*v = Package(value)
	return nil
}
func (v Package) MarshalJSON() ([]byte, error) {
	type plain Package
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *UserErrors) UnmarshalJSON(raw []byte) error {
	type plain UserErrors
	var value plain
	if e := snapshot(raw, &value, &value.RawFields); e != nil {
		return e
	}
	*v = UserErrors(value)
	return nil
}
func (v UserErrors) MarshalJSON() ([]byte, error) {
	type plain UserErrors
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *AssignmentResult) UnmarshalJSON(raw []byte) error {
	type plain AssignmentResult
	var value plain
	if e := snapshot(raw, &value, &value.RawFields); e != nil {
		return e
	}
	*v = AssignmentResult(value)
	return nil
}
func (v AssignmentResult) MarshalJSON() ([]byte, error) {
	type plain AssignmentResult
	return marshalSnapshot(plain(v), v.RawFields)
}
func (v *MessageResult) UnmarshalJSON(raw []byte) error {
	type plain MessageResult
	var value plain
	if e := snapshot(raw, &value, &value.RawFields); e != nil {
		return e
	}
	*v = MessageResult(value)
	return nil
}
func (v MessageResult) MarshalJSON() ([]byte, error) {
	type plain MessageResult
	return marshalSnapshot(plain(v), v.RawFields)
}

func (Package) String() string              { return "policy package snapshot (redacted)" }
func (v Package) GoString() string          { return v.String() }
func (UserErrors) String() string           { return "policy package user errors (redacted)" }
func (v UserErrors) GoString() string       { return v.String() }
func (AssignmentResult) String() string     { return "policy package assignment result (redacted)" }
func (v AssignmentResult) GoString() string { return v.String() }
func (MessageResult) String() string        { return "policy package operation result (redacted)" }
func (v MessageResult) GoString() string    { return v.String() }
