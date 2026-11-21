// Package teamsfeedbackpolicy manages feedback policies and scoped assignments.
package teamsfeedbackpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput
var ErrResponse = errors.New("unexpected feedback policy assignment response")

// Policy preserves nullable known fields and an owned complete received snapshot.
// Feedback modes and collection permissions remain independently nullable.
type Policy struct {
	Identity                  *string                    `json:"Identity"`
	UserInitiatedMode         *string                    `json:"UserInitiatedMode"`
	ReceiveSurveysMode        *string                    `json:"ReceiveSurveysMode"`
	AllowScreenshotCollection *bool                      `json:"AllowScreenshotCollection"`
	AllowEmailCollection      *bool                      `json:"AllowEmailCollection"`
	AllowLogCollection        *bool                      `json:"AllowLogCollection"`
	EnableFeatureSuggestions  *bool                      `json:"EnableFeatureSuggestions"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "feedback policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose string enums or cross-field restrictions.
type Settings struct {
	UserInitiatedMode         policy.Field[string] `json:"UserInitiatedMode,omitzero"`
	ReceiveSurveysMode        policy.Field[string] `json:"ReceiveSurveysMode,omitzero"`
	AllowScreenshotCollection policy.Field[bool]   `json:"AllowScreenshotCollection,omitzero"`
	AllowEmailCollection      policy.Field[bool]   `json:"AllowEmailCollection,omitzero"`
	AllowLogCollection        policy.Field[bool]   `json:"AllowLogCollection,omitzero"`
	EnableFeatureSuggestions  policy.Field[bool]   `json:"EnableFeatureSuggestions,omitzero"`
}
type CreateRequest struct {
	Settings
	policy.Controls
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance independently on failure.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "feedback policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

// GroupAssignment retains HTTP acceptance and future operation payload fields.
type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

func (p GroupAssignment) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot GroupAssignment
	return json.Marshal(snapshot(p))
}
func (GroupAssignment) String() string     { return "feedback policy group assignment (redacted)" }
func (p GroupAssignment) GoString() string { return p.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct{ PolicyName string }
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
