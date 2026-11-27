package teamstemplatepermissionpolicy_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

func TestUnitPopulatedCRUDAndListEditsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	load := func(name string) (*resource.Policy, []byte) {
		raw := readResponseFixture(t, "populated_"+name+".json")
		var v resource.Policy
		if json.Unmarshal(raw, &v) != nil {
			t.Fatal("invalid actual populated response")
		}
		return &v, raw
	}
	before, beforeRaw := load("create_readback")
	updated, updatedRaw := load("update_readback")
	edited, editedRaw := load("list_edit_readback")
	cleared, clearedRaw := load("list_clear_readback")
	if before.Identity == nil || *before.Identity != "Tag:fixture-policy" || len(before.HiddenTemplates) != 1 || len(edited.HiddenTemplates) != 1 || len(cleared.HiddenTemplates) != 0 || before.HiddenTemplates[0] == nil || edited.HiddenTemplates[0] == nil || before.HiddenTemplates[0].ID == nil || edited.HiddenTemplates[0].ID == nil || *before.HiddenTemplates[0].ID == *edited.HiddenTemplates[0].ID {
		t.Fatal("actual populated fixtures must retain two distinct template IDs and a cleared list")
	}
	id := *before.Identity
	selected := collection + "/configuration/Tag%3Afixture-policy"
	body := func(fields map[string]any) json.RawMessage {
		raw, e := json.Marshal(fields)
		if e != nil {
			t.Fatal(e)
		}
		return raw
	}
	templates := func(values []*resource.HiddenTemplate) []map[string]*string {
		out := make([]map[string]*string, len(values))
		for i, v := range values {
			if v != nil {
				out[i] = map[string]*string{"Identity": v.Identity, "Id": v.ID}
			}
		}
		return out
	}
	missing := readResponseFixture(t, "populated_get_deleted_404.json")
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(missing, &envelope) != nil {
		t.Fatal("invalid actual deletion error fixture")
	}
	steps := []testutil.Exchange{
		compiledWire("POST", collection, "New-CsTeamsTemplatePermissionPolicy", body(map[string]any{"Identity": "fixture-policy", "Description": before.Description, "HiddenTemplates": templates(before.HiddenTemplates)}), 201, string(readResponseFixture(t, "populated_create_201.body"))),
		compiledWire("GET", selected, "Get-CsTeamsTemplatePermissionPolicy", nil, 200, string(beforeRaw)),
		compiledWire("GET", selected, "Set-CsTeamsTemplatePermissionPolicy", nil, 200, string(beforeRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsTemplatePermissionPolicy", body(map[string]any{"Identity": id, "Description": updated.Description}), 204, string(readResponseFixture(t, "populated_update_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsTemplatePermissionPolicy", nil, 200, string(updatedRaw)),
		compiledWire("GET", selected, "Set-CsTeamsTemplatePermissionPolicy", nil, 200, string(updatedRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsTemplatePermissionPolicy", body(map[string]any{"Identity": id, "HiddenTemplates": templates(edited.HiddenTemplates)}), 204, string(readResponseFixture(t, "populated_list_edit_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsTemplatePermissionPolicy", nil, 200, string(editedRaw)),
		compiledWire("GET", selected, "Set-CsTeamsTemplatePermissionPolicy", nil, 200, string(editedRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsTemplatePermissionPolicy", body(map[string]any{"Identity": id, "HiddenTemplates": templates(cleared.HiddenTemplates)}), 204, string(readResponseFixture(t, "populated_list_clear_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsTemplatePermissionPolicy", nil, 200, string(clearedRaw)),
		compiledWire("GET", collection, "Get-CsTeamsTemplatePermissionPolicy", nil, 200, string(readResponseFixture(t, "populated_owned_list.json"))),
		compiledWire("DELETE", selected, "Remove-CsTeamsTemplatePermissionPolicy", body(map[string]any{"Identity": id}), 204, string(readResponseFixture(t, "populated_delete_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsTemplatePermissionPolicy", nil, 404, string(missing)),
	}
	run(t, testutil.Case{Name: "populated-crud", Steps: steps, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies
		check := func(want *resource.Policy) {
			got, _, e := s.GetByIdentity(ctx, id)
			if e != nil || got == nil || !reflect.DeepEqual(got.RawFields, want.RawFields) || len(got.HiddenTemplates) != len(want.HiddenTemplates) || !reflect.DeepEqual(got.Description, want.Description) {
				t.Fatal("populated read lost actual settings or raw fields")
			}
			for i, v := range got.HiddenTemplates {
				if v == nil || !reflect.DeepEqual(v.ID, want.HiddenTemplates[i].ID) {
					t.Fatal("populated read lost typed template IDs")
				}
			}
		}
		if _, _, e := s.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(*before.Description)}, HiddenTemplates: resource.ReplaceHiddenTemplates(resource.NewHiddenTemplate(*before.HiddenTemplates[0].ID))}); e != nil {
			t.Fatal("populated create request differs")
		}
		check(before)
		changes := []resource.UpdateRequest{
			{ScalarSettings: resource.ScalarSettings{Description: policy.Set(*updated.Description)}},
			{HiddenTemplates: resource.ChangeHiddenTemplates([]*resource.HiddenTemplate{resource.NewHiddenTemplate(*edited.HiddenTemplates[0].ID)}, []*resource.HiddenTemplate{resource.NewHiddenTemplate(*before.HiddenTemplates[0].ID)})},
			{HiddenTemplates: resource.ClearHiddenTemplates()},
		}
		for i, want := range []*resource.Policy{updated, edited, cleared} {
			receipt, r, e := s.UpdateByIdentity(ctx, id, changes[i])
			if e != nil || r == nil || r.StatusCode() != 204 || receipt == nil || !receipt.PatchInvoked || !receipt.HTTPAccepted || receipt.ReadResponse == nil {
				t.Fatal("populated update lost request or preparation receipt")
			}
			check(want)
		}
		values, _, e := s.List(ctx)
		if e != nil {
			t.Fatal("populated list failed")
		}
		count := 0
		for _, v := range values {
			if v != nil && v.Identity != nil && *v.Identity == id {
				count++
			}
		}
		if count != 1 {
			t.Fatal("populated owned policy missing or ambiguous")
		}
		if _, _, e := s.DeleteByIdentity(ctx, id, policy.Controls{}); e != nil {
			t.Fatal("populated delete differs")
		}
		_, r, e := s.GetByIdentity(ctx, id)
		return testutil.Observation{Response: r, Err: e}
	}})
}
