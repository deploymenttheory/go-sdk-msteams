// Package teamsmeetingconfiguration manages compiled Teams meeting configuration.
package teamsmeetingconfiguration

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// contextID is the pinned PublishedEntraAuthenticationContext.Id setter pattern
// (culture-invariant, case-sensitive): c1 through c99, with .NET's optional final LF.
var contextID = regexp.MustCompile(`^c([1-9]|[1-9][0-9])$`)

// PublishedEntraAuthenticationContext is one element of the compiled
// PublishedEntraAuthenticationContexts list. Every field is nullable. A value
// decoded from a response keeps its exact JSON so an Add edit re-sends
// existing elements unchanged.
type PublishedEntraAuthenticationContext struct {
	Identity             *string         `json:"Identity"`
	ID                   *string         `json:"Id"`
	PublishedName        *string         `json:"PublishedName"`
	PublishedDescription *string         `json:"PublishedDescription"`
	EntraName            *string         `json:"EntraName"`
	EntraDescription     *string         `json:"EntraDescription"`
	Published            *bool           `json:"Published"`
	RawJSON              json.RawMessage `json:"-"`
}

func (v *PublishedEntraAuthenticationContext) UnmarshalJSON(raw []byte) error {
	type plain PublishedEntraAuthenticationContext
	var next plain
	if json.Unmarshal(raw, &next) != nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = PublishedEntraAuthenticationContext(next)
	return nil
}

func (v PublishedEntraAuthenticationContext) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain PublishedEntraAuthenticationContext
	return json.Marshal(plain(v))
}

// Validate applies the pinned Id setter rule to an assigned Id. A nil Id is
// the unassigned constructor default and is not checked locally.
func (v *PublishedEntraAuthenticationContext) Validate() error {
	if v == nil {
		return ErrInput
	}
	if v.ID != nil && !contextID.MatchString(strings.TrimSuffix(*v.ID, "\n")) {
		return ErrInput
	}
	return nil
}

func (PublishedEntraAuthenticationContext) String() string {
	return "published Entra authentication context (redacted)"
}
func (v PublishedEntraAuthenticationContext) GoString() string { return v.String() }

// Configuration exposes the 21 nullable compiled getters and an owned
// complete response snapshot. The constructor-initialized context list is
// never nil after decoding; a JSON null leaves it empty, as the module does.
type Configuration struct {
	Identity                               *string                                `json:"Identity"`
	PublishedEntraAuthenticationContexts   []*PublishedEntraAuthenticationContext `json:"PublishedEntraAuthenticationContexts"`
	LogoURL                                *string                                `json:"LogoURL"`
	LegalURL                               *string                                `json:"LegalURL"`
	HelpURL                                *string                                `json:"HelpURL"`
	CustomFooterText                       *string                                `json:"CustomFooterText"`
	DisableAnonymousJoin                   *bool                                  `json:"DisableAnonymousJoin"`
	DisableAppInteractionForAnonymousUsers *bool                                  `json:"DisableAppInteractionForAnonymousUsers"`
	EnableQoS                              *bool                                  `json:"EnableQoS"`
	EnableAttributedTranscripts            *bool                                  `json:"EnableAttributedTranscripts"`
	EnableGraphTranscriptAccess            *bool                                  `json:"EnableGraphTranscriptAccess"`
	EnableWorkIQTranscriptAccess           *bool                                  `json:"EnableWorkIQTranscriptAccess"`
	ClientAudioPort                        *uint32                                `json:"ClientAudioPort"`
	ClientAudioPortRange                   *uint32                                `json:"ClientAudioPortRange"`
	ClientVideoPort                        *uint32                                `json:"ClientVideoPort"`
	ClientVideoPortRange                   *uint32                                `json:"ClientVideoPortRange"`
	ClientAppSharingPort                   *uint32                                `json:"ClientAppSharingPort"`
	ClientAppSharingPortRange              *uint32                                `json:"ClientAppSharingPortRange"`
	ClientMediaPortRangeEnabled            *bool                                  `json:"ClientMediaPortRangeEnabled"`
	LimitPresenterRolePermissions          *bool                                  `json:"LimitPresenterRolePermissions"`
	FeedbackSurveyForAnonymousUsers        *string                                `json:"FeedbackSurveyForAnonymousUsers"`
	ReportMeeting                          *string                                `json:"ReportMeeting"`
	RawFields                              map[string]json.RawMessage             `json:"-"`
	RawJSON                                json.RawMessage                        `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	// The list keeps the module's constructor-initialized empty value when the
	// response omits it or sends null.
	if next.PublishedEntraAuthenticationContexts == nil {
		next.PublishedEntraAuthenticationContexts = []*PublishedEntraAuthenticationContext{}
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(next)
	return nil
}

func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string     { return "meeting configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }
