package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsFirstPartyMeetingTemplateConfiguration"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsFirstPartyMeetingTemplateConfiguration"
const set = "Set-CsTeamsFirstPartyMeetingTemplateConfiguration"

func TestUnitReadAndFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	run(t, testutil.Case{Name: "read", Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, 200, string(raw))}, Want: testutil.Expected{JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
		if e == nil && (v == nil || v.Identity == nil || *v.Identity != "Global" || len(v.TeamsMeetingTemplates) == 0) {
			t.Fatal("typed first-party templates lost from the real response")
		}
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
	list := readResponseFixture(t, "list.json")
	run(t, testutil.Case{Name: "list", Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(list))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.List(ctx)
		if e == nil && (len(v) != 1 || v[0] == nil || v[0].Identity == nil || *v[0].Identity != "Global") {
			t.Fatal("bare collection selection differs")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"all", "*", 1}, {"identity", "Glo?al", 1}, {"case", "[gG]lobal", 1}, {"tag", "Glob*", 1}, {"missing", "missing", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(list))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(v) != tc.count {
					t.Fatal("local filter selection differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitUpdateUsesAcceptancePreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	accepted := readResponseFixture(t, "update_204.body")
	const updated = "SDK lab non-default updated description"
	run(t, testutil.Case{Name: "update", Steps: []testutil.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","Description":"`+updated+`"}`), 204, string(accepted)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(updated)}})
		if e == nil && (v == nil || v.Before == nil || v.Before.Description != nil || len(v.Before.TeamsMeetingTemplates) == 0 || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted) {
			t.Fatal("update preparation and acceptance receipts lost")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
	readback := readResponseFixture(t, "update_readback.json")
	run(t, testutil.Case{Name: "update-readback", Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, 200, string(readback))}, Want: testutil.Expected{JSON: readback}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
		var original resource.Configuration
		if json.Unmarshal(before, &original) != nil {
			t.Fatal("invalid preparation fixture")
		}
		if e == nil && (v == nil || v.Description == nil || *v.Description != updated || !jsonEqual(mustMarshal(t, v.TeamsMeetingTemplates), mustMarshal(t, original.TeamsMeetingTemplates))) {
			t.Fatal("persisted Description or omitted template list differs")
		}
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}

// expectedTemplates projects the real response templates through the pinned
// serializer's typed properties, independently of the package DTO.
func expectedTemplates(t *testing.T, raw []byte, keep func(int) bool) []map[string]any {
	t.Helper()
	var response struct {
		Templates []map[string]any `json:"TeamsMeetingTemplates"`
	}
	if json.Unmarshal(raw, &response) != nil || len(response.Templates) < 2 {
		t.Fatal("permanent populated template list unavailable")
	}
	out := []map[string]any{}
	for i, template := range response.Templates {
		if !keep(i) {
			continue
		}
		options := []map[string]any{}
		items, _ := template["TeamsMeetingOptions"].([]any)
		for _, item := range items {
			option := item.(map[string]any)
			options = append(options, map[string]any{"Identity": option["Identity"], "IsLocked": option["IsLocked"], "IsHidden": option["IsHidden"], "Value": option["Value"], "Name": option["Name"]})
		}
		out = append(out, map[string]any{"Identity": template["Identity"], "TeamsMeetingOptions": options, "Description": template["Description"], "Name": template["Name"], "Category": template["Category"]})
	}
	return out
}

func TestUnitTemplateEditsUseTheActualTemplateList(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var original resource.Configuration
	if json.Unmarshal(before, &original) != nil || len(original.TeamsMeetingTemplates) < 2 {
		t.Fatal("permanent populated template list unavailable")
	}
	first := original.TeamsMeetingTemplates[0]
	added := resource.NewTemplate("fixture-template")
	for _, tc := range []struct {
		name string
		edit *resource.TemplateEdit
		want []map[string]any
	}{
		{"clear", resource.ClearTemplates(), []map[string]any{}},
		{"empty-delta", resource.ChangeTemplates(nil, nil), []map[string]any{}},
		{"replace", resource.ReplaceTemplates(first), expectedTemplates(t, before, func(i int) bool { return i == 0 })},
		{"remove-first-name-match", resource.ChangeTemplates(nil, []*resource.Template{resource.NewTemplate(*first.Name)}), expectedTemplates(t, before, func(i int) bool { return i != 0 })},
		{"add", resource.ChangeTemplates([]*resource.Template{added}, nil), append(expectedTemplates(t, before, func(int) bool { return true }), map[string]any{"Identity": nil, "TeamsMeetingOptions": []map[string]any{}, "Description": nil, "Name": "fixture-template", "Category": nil})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Binding the template list also serializes the prepared root Description.
			body, err := json.Marshal(map[string]any{"Identity": "Global", "Description": nil, "TeamsMeetingTemplates": tc.want})
			if err != nil {
				t.Fatal(err)
			}
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, string(before)), compiledWire("PATCH", selected, set, body, 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{TeamsMeetingTemplates: tc.edit})
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
	t.Run("bound-description-wins", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{"Identity": "Global", "Description": "", "TeamsMeetingTemplates": []any{}})
		if err != nil {
			t.Fatal(err)
		}
		run(t, testutil.Case{Name: "bound-description", Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, string(before)), compiledWire("PATCH", selected, set, body, 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("")}, TeamsMeetingTemplates: resource.ClearTemplates()})
			return testutil.Observation{Response: r, Err: e}
		}})
	})
}

func TestUnitNullableSettingsAndControlsStayBound(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "nullable", Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","Description":null,"Force":"False"}`), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Null[string]()}}, Controls: policy.Controls{Force: new(false)}})
		return testutil.Observation{Response: r, Err: e}
	}})
	run(t, testutil.Case{Name: "patch-templates", Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","TeamsMeetingTemplates":[],"Force":"True"}`), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{TeamsMeetingTemplates: policy.Set[[]*resource.Template](nil)}, Controls: policy.Controls{Force: new(true)}})
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitMalformedReadDoesNotSendAnUpdate(t *testing.T) {
	t.Parallel()
	for _, invalid := range []string{`null`, `[]`, `{"Description":{}}`, `{"TeamsMeetingTemplates":{}}`, `{"TeamsMeetingTemplates":[{"Name":null}]}`, `{"TeamsMeetingTemplates":[{"Name":"a","TeamsMeetingOptions":[{"Name":null}]}]}`} {
		t.Run(fmt.Sprintf("shape-%x", invalid), func(t *testing.T) {
			run(t, testutil.Case{Name: "bad-read", Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, invalid)}, Want: testutil.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("x")}})
				if v != nil && (v.PatchInvoked || v.HTTPAccepted) {
					t.Fatal("malformed read triggered mutation")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitValidationAndCancellationMakeNoRequests(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations
		for _, id := range []string{"", " ", "scope/", "/missing", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached transport")
			}
			if _, _, e := s.PatchByIdentity(ctx, id, resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Set("x")}}}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid write identity reached transport")
			}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("unbound update accepted")
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Controls: policy.Controls{Force: new(true)}}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("Force alone accepted as a setting")
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{TeamsMeetingTemplates: &resource.TemplateEdit{}}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid template edit reached transport")
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("empty patch accepted")
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{TeamsMeetingTemplates: policy.Null[[]*resource.Template]()}}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("null template array is not a source DTO value")
		}
		for _, pattern := range []string{"", "[bad"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid filter reached transport")
			}
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, _, e := s.UpdateByIdentity(canceled, "Global", resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		if _, _, e := s.GetByIdentity(canceled, "Global"); !errors.Is(e, context.Canceled) {
			t.Fatal("read cancellation lost")
		}
		for _, empty := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := empty.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return testutil.Observation{}
	}})
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
