package teamsmeetingconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes omission, explicit null and bound false, zero
// or empty values for the 20 scalar Set parameters. The compiled setters impose
// no enum, URL, length or port-range rule, so none is invented here.
type ScalarSettings struct {
	LogoURL                                policy.Field[string] `json:"LogoURL,omitzero"`
	LegalURL                               policy.Field[string] `json:"LegalURL,omitzero"`
	HelpURL                                policy.Field[string] `json:"HelpURL,omitzero"`
	CustomFooterText                       policy.Field[string] `json:"CustomFooterText,omitzero"`
	DisableAnonymousJoin                   policy.Field[bool]   `json:"DisableAnonymousJoin,omitzero"`
	DisableAppInteractionForAnonymousUsers policy.Field[bool]   `json:"DisableAppInteractionForAnonymousUsers,omitzero"`
	EnableQoS                              policy.Field[bool]   `json:"EnableQoS,omitzero"`
	EnableAttributedTranscripts            policy.Field[bool]   `json:"EnableAttributedTranscripts,omitzero"`
	EnableGraphTranscriptAccess            policy.Field[bool]   `json:"EnableGraphTranscriptAccess,omitzero"`
	EnableWorkIQTranscriptAccess           policy.Field[bool]   `json:"EnableWorkIQTranscriptAccess,omitzero"`
	ClientAudioPort                        policy.Field[uint32] `json:"ClientAudioPort,omitzero"`
	ClientAudioPortRange                   policy.Field[uint32] `json:"ClientAudioPortRange,omitzero"`
	ClientVideoPort                        policy.Field[uint32] `json:"ClientVideoPort,omitzero"`
	ClientVideoPortRange                   policy.Field[uint32] `json:"ClientVideoPortRange,omitzero"`
	ClientAppSharingPort                   policy.Field[uint32] `json:"ClientAppSharingPort,omitzero"`
	ClientAppSharingPortRange              policy.Field[uint32] `json:"ClientAppSharingPortRange,omitzero"`
	ClientMediaPortRangeEnabled            policy.Field[bool]   `json:"ClientMediaPortRangeEnabled,omitzero"`
	LimitPresenterRolePermissions          policy.Field[bool]   `json:"LimitPresenterRolePermissions,omitzero"`
	FeedbackSurveyForAnonymousUsers        policy.Field[string] `json:"FeedbackSurveyForAnonymousUsers,omitzero"`
	ReportMeeting                          policy.Field[string] `json:"ReportMeeting,omitzero"`
}

func (ScalarSettings) String() string     { return "meeting configuration settings (redacted)" }
func (v ScalarSettings) GoString() string { return v.String() }

// Settings is one PATCH body. A bound PublishedEntraAuthenticationContexts is
// the complete replacement list the compiled cmdlet sends after applying its
// local PSListModifier. Bound null or a nil replacement normalizes to [].
type Settings struct {
	ScalarSettings
	PublishedEntraAuthenticationContexts policy.Field[[]*PublishedEntraAuthenticationContext] `json:"PublishedEntraAuthenticationContexts,omitzero"`
}

func (Settings) String() string     { return "meeting configuration settings (redacted)" }
func (v Settings) GoString() string { return v.String() }

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH
// composition. A nonnil PublishedEntraAuthenticationContexts edit is applied
// to the list read during preparation.
type UpdateRequest struct {
	ScalarSettings
	PublishedEntraAuthenticationContexts *AuthenticationContextEdit
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "meeting configuration update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

// AuthenticationContextEdit is the compiled cmdlet's PSListModifier for
// PublishedEntraAuthenticationContexts, applied locally to the preparation
// read. Remove succeeds on an empty list; the source comparator rejects it
// when any current element is examined. Helper-stage errors retain preparation.
type AuthenticationContextEdit struct {
	replace []*PublishedEntraAuthenticationContext
	add     []*PublishedEntraAuthenticationContext
	remove  []*PublishedEntraAuthenticationContext
	delta   bool
	valid   bool
}

// ReplaceAuthenticationContexts replaces the whole list. No values clears it.
func ReplaceAuthenticationContexts(values ...*PublishedEntraAuthenticationContext) *AuthenticationContextEdit {
	return &AuthenticationContextEdit{replace: cloneAuthenticationContexts(values), valid: true}
}

// AddAuthenticationContexts appends values to the current list without
// deduplication. As in the pinned PSListModifierHelper, an Add with no values
// is an empty modifier and clears the list.
func AddAuthenticationContexts(values ...*PublishedEntraAuthenticationContext) *AuthenticationContextEdit {
	return ChangeAuthenticationContexts(values, nil)
}

// ChangeAuthenticationContexts removes first, then appends Add entries. The
// source comparator rejects removal from a nonempty list; an empty-list removal
// is a no-op before additions. Results and input snapshots are owned copies.
func ChangeAuthenticationContexts(add, remove []*PublishedEntraAuthenticationContext) *AuthenticationContextEdit {
	return &AuthenticationContextEdit{add: cloneAuthenticationContexts(add), remove: cloneAuthenticationContexts(remove), delta: true, valid: true}
}

// Validate rejects a zero edit and assigned Ids the pinned setter refuses.
// Null container members are allowed; null Remove fails at the Apply stage.
func (v *AuthenticationContextEdit) Validate() error {
	if v == nil || !v.valid {
		return ErrInput
	}
	for _, list := range [][]*PublishedEntraAuthenticationContext{v.replace, v.add, v.remove} {
		for _, item := range list {
			if item == nil {
				continue
			}
			if err := item.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Apply returns the complete list the module would send. It is not atomic with
// the preceding resource read and provides no concurrency/ETag guarantee.
func (v *AuthenticationContextEdit) Apply(current []*PublishedEntraAuthenticationContext) ([]*PublishedEntraAuthenticationContext, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if !v.delta {
		return cloneAuthenticationContexts(v.replace), nil
	}
	if len(v.add) == 0 && len(v.remove) == 0 {
		return []*PublishedEntraAuthenticationContext{}, nil
	}
	for _, item := range v.remove {
		if item == nil || len(current) > 0 {
			return nil, ErrInput
		}
	}
	values := cloneAuthenticationContexts(current)
	return append(values, cloneAuthenticationContexts(v.add)...), nil
}

func (AuthenticationContextEdit) String() string {
	return "authentication context list edit (redacted)"
}
func (v AuthenticationContextEdit) GoString() string { return v.String() }

func cloneAuthenticationContexts(values []*PublishedEntraAuthenticationContext) []*PublishedEntraAuthenticationContext {
	out := make([]*PublishedEntraAuthenticationContext, len(values))
	cloneString := func(v *string) *string {
		if v == nil {
			return nil
		}
		return new(*v)
	}
	for i, v := range values {
		if v == nil {
			continue
		}
		x := *v
		x.Identity = cloneString(v.Identity)
		x.ID = cloneString(v.ID)
		x.PublishedName = cloneString(v.PublishedName)
		x.PublishedDescription = cloneString(v.PublishedDescription)
		x.EntraName = cloneString(v.EntraName)
		x.EntraDescription = cloneString(v.EntraDescription)
		if v.Published != nil {
			x.Published = new(*v.Published)
		}
		x.RawJSON = append(json.RawMessage(nil), v.RawJSON...)
		out[i] = &x
	}
	return out
}

// A received model marshals its raw snapshot for inspection. Writes instead
// project exactly the seven own source properties, including null defaults.
type authenticationContextWrite struct {
	Identity             *string `json:"Identity"`
	ID                   *string `json:"Id"`
	PublishedName        *string `json:"PublishedName"`
	PublishedDescription *string `json:"PublishedDescription"`
	EntraName            *string `json:"EntraName"`
	EntraDescription     *string `json:"EntraDescription"`
	Published            *bool   `json:"Published"`
}

func meetingFields(input Settings) (map[string]json.RawMessage, error) {
	fields, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, err
	}
	if input.PublishedEntraAuthenticationContexts.IsSet() {
		values, _ := input.PublishedEntraAuthenticationContexts.Value()
		wire := make([]*authenticationContextWrite, len(values))
		for i, v := range values {
			if v == nil {
				continue
			}
			if err := v.Validate(); err != nil {
				return nil, err
			}
			wire[i] = &authenticationContextWrite{v.Identity, v.ID, v.PublishedName, v.PublishedDescription, v.EntraName, v.EntraDescription, v.Published}
		}
		raw, err := json.Marshal(wire)
		if err != nil {
			return nil, ErrInput
		}
		fields["PublishedEntraAuthenticationContexts"] = raw
	}
	return fields, nil
}
