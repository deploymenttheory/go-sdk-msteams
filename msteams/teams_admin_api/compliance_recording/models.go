// Package compliancerecording manages Teams compliance recording policies,
// the recording applications attached to them and policy grants.
package compliancerecording

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ErrInput identifies a request rejected locally before credentials or I/O.
var ErrInput = policy.ErrInput

// ErrResponse identifies an undeclared successful grant status.
var ErrResponse = errors.New("unexpected compliance recording assignment response")

// guidPattern is the model setter rule for application and paired IDs.
var guidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func validID(id string) bool { return guidPattern.MatchString(id) }

// decodeSnapshot unmarshals raw into value (a plain alias) and keeps the
// complete service object.
func decodeSnapshot(raw []byte, value any, fields *map[string]json.RawMessage, snapshot *json.RawMessage) error {
	if json.Unmarshal(raw, fields) != nil || *fields == nil || json.Unmarshal(raw, value) != nil {
		return client.ErrDecode
	}
	*snapshot = append(json.RawMessage(nil), raw...)
	return nil
}

// PairedApplication is a secondary recorder paired with an application. ID is
// the paired resource account (application instance) object ID.
type PairedApplication struct {
	Identity  *string                    `json:"Identity,omitempty"`
	ID        *string                    `json:"Id,omitempty"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *PairedApplication) UnmarshalJSON(raw []byte) error {
	type plain PairedApplication
	var value plain
	var snapshot json.RawMessage
	if err := decodeSnapshot(raw, &value, &value.RawFields, &snapshot); err != nil {
		return err
	}
	*v = PairedApplication(value)
	return nil
}

// NewPairedApplication is New-CsTeamsComplianceRecordingPairedApplication. It
// builds the element locally and performs no I/O; id must be a GUID.
func NewPairedApplication(id string) (*PairedApplication, error) {
	if !validID(id) {
		return nil, ErrInput
	}
	return &PairedApplication{ID: &id}, nil
}

// Application is a compliance recording application: the recorder bot's
// resource account (ID) and when recording is required. Applications live
// under a policy with identity "{policy}/{id}". Absent and null fields stay nil;
// RawFields and RawJSON keep the received object.
type Application struct {
	Identity                              *string                    `json:"Identity,omitempty"`
	ComplianceRecordingPairedApplications []*PairedApplication       `json:"ComplianceRecordingPairedApplications,omitempty"`
	ID                                    *string                    `json:"Id,omitempty"`
	RequiredBeforeMeetingJoin             *bool                      `json:"RequiredBeforeMeetingJoin,omitempty"`
	RequiredBeforeCallEstablishment       *bool                      `json:"RequiredBeforeCallEstablishment,omitempty"`
	RequiredDuringMeeting                 *bool                      `json:"RequiredDuringMeeting,omitempty"`
	RequiredDuringCall                    *bool                      `json:"RequiredDuringCall,omitempty"`
	ConcurrentInvitationCount             *uint32                    `json:"ConcurrentInvitationCount,omitempty"`
	RawFields                             map[string]json.RawMessage `json:"-"`
	RawJSON                               json.RawMessage            `json:"-"`
}

func (v *Application) UnmarshalJSON(raw []byte) error {
	type plain Application
	var value plain
	if err := decodeSnapshot(raw, &value, &value.RawFields, &value.RawJSON); err != nil {
		return err
	}
	*v = Application(value)
	return nil
}

func (Application) String() string     { return "compliance recording application (redacted)" }
func (v Application) GoString() string { return v.String() }

// Policy is a compliance recording policy. ComplianceRecordingApplications
// holds the applications attached to it.
type Policy struct {
	Identity                                            *string                    `json:"Identity"`
	ComplianceRecordingApplications                     []*Application             `json:"ComplianceRecordingApplications"`
	Enabled                                             *bool                      `json:"Enabled"`
	WarnUserOnRemoval                                   *bool                      `json:"WarnUserOnRemoval"`
	DisableComplianceRecordingAudioNotificationForCalls *bool                      `json:"DisableComplianceRecordingAudioNotificationForCalls"`
	Description                                         *string                    `json:"Description"`
	RecordReroutedCalls                                 *bool                      `json:"RecordReroutedCalls"`
	CustomPromptsEnabled                                *bool                      `json:"CustomPromptsEnabled"`
	CustomPromptsPackageID                              *string                    `json:"CustomPromptsPackageId"`
	CustomBanner                                        *string                    `json:"CustomBanner"`
	RawFields                                           map[string]json.RawMessage `json:"-"`
	RawJSON                                             json.RawMessage            `json:"-"`
}

func (v *Policy) UnmarshalJSON(raw []byte) error {
	type plain Policy
	var value plain
	if err := decodeSnapshot(raw, &value, &value.RawFields, &value.RawJSON); err != nil {
		return err
	}
	*v = Policy(value)
	return nil
}

func (Policy) String() string     { return "compliance recording policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// PolicySettings are the bound policy parameters. Unset omits, policy.Null
// sends null. A set application list replaces the whole list.
type PolicySettings struct {
	ComplianceRecordingApplications                     policy.Field[[]*Application] `json:"ComplianceRecordingApplications,omitzero"`
	Enabled                                             policy.Field[bool]           `json:"Enabled,omitzero"`
	WarnUserOnRemoval                                   policy.Field[bool]           `json:"WarnUserOnRemoval,omitzero"`
	DisableComplianceRecordingAudioNotificationForCalls policy.Field[bool]           `json:"DisableComplianceRecordingAudioNotificationForCalls,omitzero"`
	Description                                         policy.Field[string]         `json:"Description,omitzero"`
	RecordReroutedCalls                                 policy.Field[bool]           `json:"RecordReroutedCalls,omitzero"`
	CustomPromptsEnabled                                policy.Field[bool]           `json:"CustomPromptsEnabled,omitzero"`
	CustomPromptsPackageID                              policy.Field[string]         `json:"CustomPromptsPackageId,omitzero"`
	CustomBanner                                        policy.Field[string]         `json:"CustomBanner,omitzero"`
}

// ApplicationSettings are the bound application parameters. The module's
// -Priority parameter is accepted but never copied into the request, so it has
// no equivalent here. A set paired list replaces the whole list.
type ApplicationSettings struct {
	ComplianceRecordingPairedApplications policy.Field[[]*PairedApplication] `json:"ComplianceRecordingPairedApplications,omitzero"`
	RequiredBeforeMeetingJoin             policy.Field[bool]                 `json:"RequiredBeforeMeetingJoin,omitzero"`
	RequiredBeforeCallEstablishment       policy.Field[bool]                 `json:"RequiredBeforeCallEstablishment,omitzero"`
	RequiredDuringMeeting                 policy.Field[bool]                 `json:"RequiredDuringMeeting,omitzero"`
	RequiredDuringCall                    policy.Field[bool]                 `json:"RequiredDuringCall,omitzero"`
	ConcurrentInvitationCount             policy.Field[uint32]               `json:"ConcurrentInvitationCount,omitzero"`
}

// ListEdit is the module's PSListModifier for a list parameter, applied to the
// value read before the update. Replace (when non-nil) sets the list; Add
// appends; Remove deletes elements whose Id matches, case-insensitively.
type ListEdit[T any] struct {
	Replace []*T
	Add     []*T
	Remove  []string
}

// PolicyUpdate is a Set-CsTeamsComplianceRecordingPolicy composition.
type PolicyUpdate struct {
	PolicySettings
	Applications *ListEdit[Application]
	policy.Controls
}

// ApplicationUpdate is a Set-CsTeamsComplianceRecordingApplication composition.
type ApplicationUpdate struct {
	ApplicationSettings
	PairedApplications *ListEdit[PairedApplication]
	policy.Controls
}

// UpdateResult keeps the preparation read and acceptance on partial failure.
type UpdateResult[T any] struct {
	Before, After              *T
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult[T]) String() string     { return "compliance recording update result (redacted)" }
func (v UpdateResult[T]) GoString() string { return v.String() }

// GroupAssignment is an accepted group grant.
type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

// MarshalJSON returns the received object when there is one.
func (v GroupAssignment) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain GroupAssignment
	return json.Marshal(plain(v))
}

func (GroupAssignment) String() string     { return "compliance recording group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }
