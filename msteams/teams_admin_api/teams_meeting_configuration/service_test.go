package teamsmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

// The live acceptance update: a preparation GET, then one sparse PATCH of
// only the two bound strings, answered by the actual empty 204.
func TestUnitUpdateUsesActualPreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	run(t, contracttest.Case{Name: "update", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CustomFooterText":"msteams-sdk acceptance footer","HelpURL":"https://example.invalid/msteams-sdk-help"}`), 204, string(readResponseFixture(t, "update_204.body"))),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{
			CustomFooterText: policy.Set("msteams-sdk acceptance footer"),
			HelpURL:          policy.Set("https://example.invalid/msteams-sdk-help"),
		}})
		if e == nil && (v == nil || v.Before == nil || v.Before.CustomFooterText == nil || *v.Before.CustomFooterText != "" || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted || v.After != nil) {
			t.Fatal("preparation or mutation acceptance receipt lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}

// Restoration of a null original sends an explicit JSON null, as the live
// restore of HelpURL did.
func TestUnitRestoreSendsExplicitNullAndEmptyString(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "restore", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(readResponseFixture(t, "update_readback.json"))),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CustomFooterText":"","HelpURL":null}`), 204, string(readResponseFixture(t, "update_204.body"))),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{
			CustomFooterText: policy.Set(""), HelpURL: policy.Null[string](),
		}})
		return contracttest.Observation{Response: r, Err: e}
	}})
}

// Every pinned Set parameter binds by its exact name. False, positive integers and empty
// values are sent rather than dropped, explicit null is retained and Force
// keeps the compiled string form.
func TestUnitEverySettingNullFalsePositiveEmptyAndForceRemainExplicit(t *testing.T) {
	t.Parallel()
	c1 := "c1"
	name := "Sensitive meetings"
	published := true
	for _, tc := range []struct {
		name, body string
		settings   resource.Settings
	}{
		{"all-bound", `{"Identity":"Global","LogoURL":"https://example.invalid/logo.png","LegalURL":"https://example.invalid/legal","HelpURL":"https://example.invalid/help","CustomFooterText":"footer","DisableAnonymousJoin":false,"DisableAppInteractionForAnonymousUsers":true,"EnableQoS":false,"EnableAttributedTranscripts":true,"EnableGraphTranscriptAccess":false,"EnableWorkIQTranscriptAccess":true,"ClientAudioPort":50000,"ClientAudioPortRange":20,"ClientVideoPort":50020,"ClientVideoPortRange":20,"ClientAppSharingPort":50040,"ClientAppSharingPortRange":4294967295,"ClientMediaPortRangeEnabled":false,"LimitPresenterRolePermissions":true,"FeedbackSurveyForAnonymousUsers":"Disabled","ReportMeeting":"Disabled","PublishedEntraAuthenticationContexts":[{"Identity":null,"Id":"c1","PublishedName":"Sensitive meetings","PublishedDescription":null,"EntraName":null,"EntraDescription":null,"Published":true}],"Force":"False"}`, resource.Settings{
			ScalarSettings: resource.ScalarSettings{
				LogoURL: policy.Set("https://example.invalid/logo.png"), LegalURL: policy.Set("https://example.invalid/legal"), HelpURL: policy.Set("https://example.invalid/help"), CustomFooterText: policy.Set("footer"),
				DisableAnonymousJoin: policy.Set(false), DisableAppInteractionForAnonymousUsers: policy.Set(true), EnableQoS: policy.Set(false),
				EnableAttributedTranscripts: policy.Set(true), EnableGraphTranscriptAccess: policy.Set(false), EnableWorkIQTranscriptAccess: policy.Set(true),
				ClientAudioPort: policy.Set[uint32](50000), ClientAudioPortRange: policy.Set[uint32](20), ClientVideoPort: policy.Set[uint32](50020), ClientVideoPortRange: policy.Set[uint32](20),
				ClientAppSharingPort: policy.Set[uint32](50040), ClientAppSharingPortRange: policy.Set[uint32](4294967295), ClientMediaPortRangeEnabled: policy.Set(false),
				LimitPresenterRolePermissions: policy.Set(true), FeedbackSurveyForAnonymousUsers: policy.Set("Disabled"), ReportMeeting: policy.Set("Disabled"),
			},
			PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{{ID: &c1, PublishedName: &name, Published: &published}}),
		}},
		{"null", `{"Identity":"Global","LogoURL":null,"EnableWorkIQTranscriptAccess":null,"ClientVideoPort":null,"PublishedEntraAuthenticationContexts":[],"Force":"False"}`, resource.Settings{
			ScalarSettings:                       resource.ScalarSettings{LogoURL: policy.Null[string](), EnableWorkIQTranscriptAccess: policy.Null[bool](), ClientVideoPort: policy.Null[uint32]()},
			PublishedEntraAuthenticationContexts: policy.Null[[]*resource.PublishedEntraAuthenticationContext](),
		}},
		{"positive-empty", `{"Identity":"Global","CustomFooterText":"","ClientAudioPort":1,"EnableQoS":false,"PublishedEntraAuthenticationContexts":[],"Force":"False"}`, resource.Settings{
			ScalarSettings:                       resource.ScalarSettings{CustomFooterText: policy.Set(""), ClientAudioPort: policy.Set[uint32](1), EnableQoS: policy.Set(false)},
			PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{}),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: tc.settings, Controls: policy.Controls{Force: new(false)}})
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

// The compiled cmdlet applies its PSListModifier to the preparation read and
// sends the complete list. Existing elements use the typed known-field projection;
// Add appends without deduplication; Replace wins; an empty modifier clears.
func TestUnitContextListEditsApplyToPreparationRead(t *testing.T) {
	t.Parallel()
	const existing = `[{"Identity":"c1","Id":"c1","PublishedName":"Existing","PublishedDescription":"kept","EntraName":"Entra","EntraDescription":null,"Published":true,"Unknown":1}]`
	const projected = `[{"Identity":"c1","Id":"c1","PublishedName":"Existing","PublishedDescription":"kept","EntraName":"Entra","EntraDescription":null,"Published":true}]`
	c1, c2 := "c1", "c2"
	added := &resource.PublishedEntraAuthenticationContext{ID: &c2}
	duplicate := &resource.PublishedEntraAuthenticationContext{ID: &c1}
	const addedJSON = `{"Identity":null,"Id":"c2","PublishedName":null,"PublishedDescription":null,"EntraName":null,"EntraDescription":null,"Published":null}`
	const duplicateJSON = `{"Identity":null,"Id":"c1","PublishedName":null,"PublishedDescription":null,"EntraName":null,"EntraDescription":null,"Published":null}`
	for _, tc := range []struct {
		name, current, sent string
		edit                *resource.AuthenticationContextEdit
	}{
		{"add-to-existing", existing, `[` + projected[1:len(projected)-1] + `,` + addedJSON + `]`, resource.AddAuthenticationContexts(added)},
		{"add-duplicate-id", existing, `[` + projected[1:len(projected)-1] + `,` + duplicateJSON + `]`, resource.AddAuthenticationContexts(duplicate)},
		{"add-to-actual-empty", "[]", `[` + addedJSON + `]`, resource.AddAuthenticationContexts(added)},
		{"add-to-null", "null", `[` + addedJSON + `]`, resource.AddAuthenticationContexts(added)},
		{"replace", existing, `[` + addedJSON + `]`, resource.ReplaceAuthenticationContexts(added)},
		{"replace-empty-clears", existing, `[]`, resource.ReplaceAuthenticationContexts()},
		{"empty-add-clears", existing, `[]`, resource.AddAuthenticationContexts()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
				compiledWire("GET", selected, set, nil, 200, withContexts(t, tc.current)),
				compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","PublishedEntraAuthenticationContexts":`+tc.sent+`}`), 204, string(readResponseFixture(t, "update_204.body"))),
			}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: tc.edit})
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitFilterUsesActualCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"global", "Glo?al", 1}, {"wildcard", "*", 1}, {"missing", "missing", 0}, {"literal-whitespace", " ", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: get, ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, get, nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(values) != tc.count {
					t.Fatal("local filter result differs")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
