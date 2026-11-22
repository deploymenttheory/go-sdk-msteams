// Package teamsvoiceapplicationspolicy manages compiled Teams voice applications policy.
package teamsvoiceapplicationspolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// Boolean and string settings remain nullable; source strings stay opaque.
type Policy struct {
	Identity                                            *string                    `json:"Identity"`
	AllowAutoAttendantAfterHoursGreetingChange          *bool                      `json:"AllowAutoAttendantAfterHoursGreetingChange"`
	AllowAutoAttendantAfterHoursRoutingChange           *bool                      `json:"AllowAutoAttendantAfterHoursRoutingChange"`
	AllowAutoAttendantBusinessHoursChange               *bool                      `json:"AllowAutoAttendantBusinessHoursChange"`
	AllowAutoAttendantBusinessHoursGreetingChange       *bool                      `json:"AllowAutoAttendantBusinessHoursGreetingChange"`
	AllowAutoAttendantBusinessHoursRoutingChange        *bool                      `json:"AllowAutoAttendantBusinessHoursRoutingChange"`
	AllowAutoAttendantHolidayGreetingChange             *bool                      `json:"AllowAutoAttendantHolidayGreetingChange"`
	AllowAutoAttendantHolidayRoutingChange              *bool                      `json:"AllowAutoAttendantHolidayRoutingChange"`
	AllowAutoAttendantHolidaysChange                    *bool                      `json:"AllowAutoAttendantHolidaysChange"`
	AllowAutoAttendantLanguageChange                    *bool                      `json:"AllowAutoAttendantLanguageChange"`
	AllowAutoAttendantTimeZoneChange                    *bool                      `json:"AllowAutoAttendantTimeZoneChange"`
	AllowCallQueueAgentOptChange                        *bool                      `json:"AllowCallQueueAgentOptChange"`
	AllowCallQueueConferenceModeChange                  *bool                      `json:"AllowCallQueueConferenceModeChange"`
	AllowCallQueueLanguageChange                        *bool                      `json:"AllowCallQueueLanguageChange"`
	AllowCallQueueMembershipChange                      *bool                      `json:"AllowCallQueueMembershipChange"`
	AllowCallQueueMusicOnHoldChange                     *bool                      `json:"AllowCallQueueMusicOnHoldChange"`
	AllowCallQueueNoAgentSharedVoicemailGreetingChange  *bool                      `json:"AllowCallQueueNoAgentSharedVoicemailGreetingChange"`
	AllowCallQueueNoAgentsRoutingChange                 *bool                      `json:"AllowCallQueueNoAgentsRoutingChange"`
	AllowCallQueueOptOutChange                          *bool                      `json:"AllowCallQueueOptOutChange"`
	AllowCallQueueOverflowRoutingChange                 *bool                      `json:"AllowCallQueueOverflowRoutingChange"`
	AllowCallQueueOverflowSharedVoicemailGreetingChange *bool                      `json:"AllowCallQueueOverflowSharedVoicemailGreetingChange"`
	AllowCallQueuePresenceBasedRoutingChange            *bool                      `json:"AllowCallQueuePresenceBasedRoutingChange"`
	AllowCallQueueRoutingMethodChange                   *bool                      `json:"AllowCallQueueRoutingMethodChange"`
	AllowCallQueueTimeoutRoutingChange                  *bool                      `json:"AllowCallQueueTimeoutRoutingChange"`
	AllowCallQueueTimeoutSharedVoicemailGreetingChange  *bool                      `json:"AllowCallQueueTimeoutSharedVoicemailGreetingChange"`
	AllowCallQueueWelcomeGreetingChange                 *bool                      `json:"AllowCallQueueWelcomeGreetingChange"`
	CallQueueAgentMonitorMode                           *string                    `json:"CallQueueAgentMonitorMode"`
	CallQueueAgentMonitorNotificationMode               *string                    `json:"CallQueueAgentMonitorNotificationMode"`
	Description                                         *string                    `json:"Description"`
	HistoricalAgentMetricsPermission                    *string                    `json:"HistoricalAgentMetricsPermission"`
	HistoricalAutoAttendantMetricsPermission            *string                    `json:"HistoricalAutoAttendantMetricsPermission"`
	HistoricalCallQueueMetricsPermission                *string                    `json:"HistoricalCallQueueMetricsPermission"`
	RealTimeAgentMetricsPermission                      *string                    `json:"RealTimeAgentMetricsPermission"`
	RealTimeAutoAttendantMetricsPermission              *string                    `json:"RealTimeAutoAttendantMetricsPermission"`
	RealTimeCallQueueMetricsPermission                  *string                    `json:"RealTimeCallQueueMetricsPermission"`
	RawFields                                           map[string]json.RawMessage `json:"-"`
	RawJSON                                             json.RawMessage            `json:"-"`
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
func (Policy) String() string     { return "voice applications policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	AllowAutoAttendantAfterHoursGreetingChange          policy.Field[bool]   `json:"AllowAutoAttendantAfterHoursGreetingChange,omitzero"`
	AllowAutoAttendantAfterHoursRoutingChange           policy.Field[bool]   `json:"AllowAutoAttendantAfterHoursRoutingChange,omitzero"`
	AllowAutoAttendantBusinessHoursChange               policy.Field[bool]   `json:"AllowAutoAttendantBusinessHoursChange,omitzero"`
	AllowAutoAttendantBusinessHoursGreetingChange       policy.Field[bool]   `json:"AllowAutoAttendantBusinessHoursGreetingChange,omitzero"`
	AllowAutoAttendantBusinessHoursRoutingChange        policy.Field[bool]   `json:"AllowAutoAttendantBusinessHoursRoutingChange,omitzero"`
	AllowAutoAttendantHolidayGreetingChange             policy.Field[bool]   `json:"AllowAutoAttendantHolidayGreetingChange,omitzero"`
	AllowAutoAttendantHolidayRoutingChange              policy.Field[bool]   `json:"AllowAutoAttendantHolidayRoutingChange,omitzero"`
	AllowAutoAttendantHolidaysChange                    policy.Field[bool]   `json:"AllowAutoAttendantHolidaysChange,omitzero"`
	AllowAutoAttendantLanguageChange                    policy.Field[bool]   `json:"AllowAutoAttendantLanguageChange,omitzero"`
	AllowAutoAttendantTimeZoneChange                    policy.Field[bool]   `json:"AllowAutoAttendantTimeZoneChange,omitzero"`
	AllowCallQueueAgentOptChange                        policy.Field[bool]   `json:"AllowCallQueueAgentOptChange,omitzero"`
	AllowCallQueueConferenceModeChange                  policy.Field[bool]   `json:"AllowCallQueueConferenceModeChange,omitzero"`
	AllowCallQueueLanguageChange                        policy.Field[bool]   `json:"AllowCallQueueLanguageChange,omitzero"`
	AllowCallQueueMembershipChange                      policy.Field[bool]   `json:"AllowCallQueueMembershipChange,omitzero"`
	AllowCallQueueMusicOnHoldChange                     policy.Field[bool]   `json:"AllowCallQueueMusicOnHoldChange,omitzero"`
	AllowCallQueueNoAgentSharedVoicemailGreetingChange  policy.Field[bool]   `json:"AllowCallQueueNoAgentSharedVoicemailGreetingChange,omitzero"`
	AllowCallQueueNoAgentsRoutingChange                 policy.Field[bool]   `json:"AllowCallQueueNoAgentsRoutingChange,omitzero"`
	AllowCallQueueOptOutChange                          policy.Field[bool]   `json:"AllowCallQueueOptOutChange,omitzero"`
	AllowCallQueueOverflowRoutingChange                 policy.Field[bool]   `json:"AllowCallQueueOverflowRoutingChange,omitzero"`
	AllowCallQueueOverflowSharedVoicemailGreetingChange policy.Field[bool]   `json:"AllowCallQueueOverflowSharedVoicemailGreetingChange,omitzero"`
	AllowCallQueuePresenceBasedRoutingChange            policy.Field[bool]   `json:"AllowCallQueuePresenceBasedRoutingChange,omitzero"`
	AllowCallQueueRoutingMethodChange                   policy.Field[bool]   `json:"AllowCallQueueRoutingMethodChange,omitzero"`
	AllowCallQueueTimeoutRoutingChange                  policy.Field[bool]   `json:"AllowCallQueueTimeoutRoutingChange,omitzero"`
	AllowCallQueueTimeoutSharedVoicemailGreetingChange  policy.Field[bool]   `json:"AllowCallQueueTimeoutSharedVoicemailGreetingChange,omitzero"`
	AllowCallQueueWelcomeGreetingChange                 policy.Field[bool]   `json:"AllowCallQueueWelcomeGreetingChange,omitzero"`
	CallQueueAgentMonitorMode                           policy.Field[string] `json:"CallQueueAgentMonitorMode,omitzero"`
	CallQueueAgentMonitorNotificationMode               policy.Field[string] `json:"CallQueueAgentMonitorNotificationMode,omitzero"`
	Description                                         policy.Field[string] `json:"Description,omitzero"`
	HistoricalAgentMetricsPermission                    policy.Field[string] `json:"HistoricalAgentMetricsPermission,omitzero"`
	HistoricalAutoAttendantMetricsPermission            policy.Field[string] `json:"HistoricalAutoAttendantMetricsPermission,omitzero"`
	HistoricalCallQueueMetricsPermission                policy.Field[string] `json:"HistoricalCallQueueMetricsPermission,omitzero"`
	RealTimeAgentMetricsPermission                      policy.Field[string] `json:"RealTimeAgentMetricsPermission,omitzero"`
	RealTimeAutoAttendantMetricsPermission              policy.Field[string] `json:"RealTimeAutoAttendantMetricsPermission,omitzero"`
	RealTimeCallQueueMetricsPermission                  policy.Field[string] `json:"RealTimeCallQueueMetricsPermission,omitzero"`
}

// CreateRequest permits no settings, matching the compiled New processing.
type CreateRequest struct {
	Settings
	policy.Controls
}

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "voice applications policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "voice applications policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
