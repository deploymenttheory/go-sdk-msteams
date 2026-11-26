// Package teamsmeetingconfiguration manages compiled Teams meeting configuration.
package teamsmeetingconfiguration

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// contextID is the pinned PublishedEntraAuthenticationContext.Id setter pattern
// (culture-invariant, case-sensitive): c1 through c99, with .NET's optional final LF.
var contextID = regexp.MustCompile(`^c([1-9]|[1-9][0-9])$`)

// PublishedEntraAuthenticationContext is one element of the compiled
// PublishedEntraAuthenticationContexts list. Every field is nullable. A value
// decoded from a response keeps its exact JSON for inspection. Typed write
// projection is separate from this raw received view.
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
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var wire struct {
		Identity             policyrp.ReadString `json:"Identity"`
		ID                   readContextID       `json:"Id"`
		PublishedName        policyrp.ReadString `json:"PublishedName"`
		PublishedDescription policyrp.ReadString `json:"PublishedDescription"`
		EntraName            policyrp.ReadString `json:"EntraName"`
		EntraDescription     policyrp.ReadString `json:"EntraDescription"`
		Published            policyrp.ReadBool   `json:"Published"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	*v = PublishedEntraAuthenticationContext{Identity: wire.Identity.Value, ID: wire.ID.Value, PublishedName: wire.PublishedName.Value, PublishedDescription: wire.PublishedDescription.Value, EntraName: wire.EntraName.Value, EntraDescription: wire.EntraDescription.Value, Published: wire.Published.Value, RawJSON: append(json.RawMessage(nil), raw...)}
	return nil
}

type readContextID struct{ Value *string }

func (v *readContextID) UnmarshalJSON(raw []byte) error {
	var text policyrp.ReadString
	if json.Unmarshal(raw, &text) != nil || text.Value == nil || !contextID.MatchString(strings.TrimSuffix(*text.Value, "\n")) {
		return client.ErrDecode
	}
	v.Value = text.Value
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
	var wire struct {
		Identity                               policyrp.ReadString `json:"Identity"`
		LogoURL                                policyrp.ReadString `json:"LogoURL"`
		LegalURL                               policyrp.ReadString `json:"LegalURL"`
		HelpURL                                policyrp.ReadString `json:"HelpURL"`
		CustomFooterText                       policyrp.ReadString `json:"CustomFooterText"`
		DisableAnonymousJoin                   policyrp.ReadBool   `json:"DisableAnonymousJoin"`
		DisableAppInteractionForAnonymousUsers policyrp.ReadBool   `json:"DisableAppInteractionForAnonymousUsers"`
		EnableQoS                              policyrp.ReadBool   `json:"EnableQoS"`
		EnableAttributedTranscripts            policyrp.ReadBool   `json:"EnableAttributedTranscripts"`
		EnableGraphTranscriptAccess            policyrp.ReadBool   `json:"EnableGraphTranscriptAccess"`
		EnableWorkIQTranscriptAccess           policyrp.ReadBool   `json:"EnableWorkIQTranscriptAccess"`
		ClientAudioPort                        readMeetingPort     `json:"ClientAudioPort"`
		ClientAudioPortRange                   readMeetingPort     `json:"ClientAudioPortRange"`
		ClientVideoPort                        readMeetingPort     `json:"ClientVideoPort"`
		ClientVideoPortRange                   readMeetingPort     `json:"ClientVideoPortRange"`
		ClientAppSharingPort                   readMeetingPort     `json:"ClientAppSharingPort"`
		ClientAppSharingPortRange              readMeetingPort     `json:"ClientAppSharingPortRange"`
		ClientMediaPortRangeEnabled            policyrp.ReadBool   `json:"ClientMediaPortRangeEnabled"`
		LimitPresenterRolePermissions          policyrp.ReadBool   `json:"LimitPresenterRolePermissions"`
		FeedbackSurveyForAnonymousUsers        policyrp.ReadString `json:"FeedbackSurveyForAnonymousUsers"`
		ReportMeeting                          policyrp.ReadString `json:"ReportMeeting"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	value := Configuration{
		Identity:                               wire.Identity.Value,
		LogoURL:                                wire.LogoURL.Value,
		LegalURL:                               wire.LegalURL.Value,
		HelpURL:                                wire.HelpURL.Value,
		CustomFooterText:                       wire.CustomFooterText.Value,
		DisableAnonymousJoin:                   wire.DisableAnonymousJoin.Value,
		DisableAppInteractionForAnonymousUsers: wire.DisableAppInteractionForAnonymousUsers.Value,
		EnableQoS:                              wire.EnableQoS.Value,
		EnableAttributedTranscripts:            wire.EnableAttributedTranscripts.Value,
		EnableGraphTranscriptAccess:            wire.EnableGraphTranscriptAccess.Value,
		EnableWorkIQTranscriptAccess:           wire.EnableWorkIQTranscriptAccess.Value,
		ClientAudioPort:                        wire.ClientAudioPort.Value,
		ClientAudioPortRange:                   wire.ClientAudioPortRange.Value,
		ClientVideoPort:                        wire.ClientVideoPort.Value,
		ClientVideoPortRange:                   wire.ClientVideoPortRange.Value,
		ClientAppSharingPort:                   wire.ClientAppSharingPort.Value,
		ClientAppSharingPortRange:              wire.ClientAppSharingPortRange.Value,
		ClientMediaPortRangeEnabled:            wire.ClientMediaPortRangeEnabled.Value,
		LimitPresenterRolePermissions:          wire.LimitPresenterRolePermissions.Value,
		FeedbackSurveyForAnonymousUsers:        wire.FeedbackSurveyForAnonymousUsers.Value,
		ReportMeeting:                          wire.ReportMeeting.Value,
		PublishedEntraAuthenticationContexts:   []*PublishedEntraAuthenticationContext{},
	}
	fields, err := readObject(raw, func(key string, body json.RawMessage) error {
		if !strings.EqualFold(key, "PublishedEntraAuthenticationContexts") {
			return nil
		}
		token := bytes.TrimSpace(body)
		if len(token) == 0 || token[0] == '{' {
			return client.ErrDecode
		}
		if token[0] != '[' {
			return nil
		}
		var items []*PublishedEntraAuthenticationContext
		if json.Unmarshal(body, &items) != nil {
			return client.ErrDecode
		}
		value.PublishedEntraAuthenticationContexts = append(value.PublishedEntraAuthenticationContexts, items...)
		return nil
	})
	if err != nil {
		return client.ErrDecode
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
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

// readObject visits each property in wire order so list aliases accumulate and
// an invalid earlier setter cannot be hidden by a later property occurrence.
func readObject(raw []byte, visit func(string, json.RawMessage) error) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if _, err := decoder.Token(); err != nil {
		return nil, client.ErrDecode
	}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, client.ErrDecode
		}
		name, ok := token.(string)
		if !ok {
			return nil, client.ErrDecode
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || visit(name, value) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}

// readMeetingPort applies the own DTO's positiveInteger restriction after the
// nullable UInt32 conversion. The full nonzero UInt32 range is valid; this is
// not a networking port limit and introduces no cross-field/range constraint.
type readMeetingPort struct{ Value *uint32 }

func (v *readMeetingPort) UnmarshalJSON(raw []byte) error {
	var value policyrp.ReadUint32
	if json.Unmarshal(raw, &value) != nil || value.Value != nil && *value.Value == 0 {
		return client.ErrDecode
	}
	v.Value = value.Value
	return nil
}
