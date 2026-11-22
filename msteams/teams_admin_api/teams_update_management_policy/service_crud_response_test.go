package teamsupdatemanagementpolicy_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
)

func TestUnitOwnedCRUDDateTimeAndListEditsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	load := func(label string) (*resource.Policy, []byte) {
		raw := readResponseFixture(t, label+".json")
		var v resource.Policy
		if json.Unmarshal(raw, &v) != nil {
			t.Fatal("invalid actual update-management response")
		}
		return &v, raw
	}
	before, beforeRaw := load("create_readback")
	updated, updatedRaw := load("update_readback")
	dated, datedRaw := load("datetime_readback")
	edited, editedRaw := load("list_edit_readback")
	cleared, clearedRaw := load("list_clear_readback")
	if before.Identity == nil || *before.Identity != "Tag:fixture-policy" || before.Description == nil || updated.Description == nil || before.UpdateTimeOfDay == nil || len(before.DisabledInProductMessages) != 1 || len(edited.DisabledInProductMessages) != 1 || len(cleared.DisabledInProductMessages) != 0 {
		t.Fatal("actual owned fixtures missing required settings")
	}
	id := *before.Identity
	selected := collection + "/configuration/Tag%3Afixture-policy"
	date, e := resource.NewDateTime(before.UpdateTimeOfDay.Time().Add(45*time.Minute), before.UpdateTimeOfDay.Kind())
	if e != nil {
		t.Fatal(e)
	}
	body := func(fields map[string]any) json.RawMessage {
		raw, e := json.Marshal(fields)
		if e != nil {
			t.Fatal(e)
		}
		return raw
	}
	missing := readResponseFixture(t, "get_deleted_404.json")
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(missing, &envelope) != nil {
		t.Fatal("invalid actual error fixture")
	}
	steps := []contracttest.Exchange{
		compiledWire("POST", collection, "New-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": "fixture-policy", "Description": before.Description, "AllowManagedUpdates": true, "UpdateDayOfWeek": int64(3), "UpdateTime": "19:30", "AllowPublicPreview": "Enabled", "DisabledInProductMessages": before.DisabledInProductMessages}), 201, string(readResponseFixture(t, "create_201.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(beforeRaw)),
		compiledWire("GET", selected, "Set-CsTeamsUpdateManagementPolicy", nil, 200, string(beforeRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": id, "Description": updated.Description}), 204, string(readResponseFixture(t, "update_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(updatedRaw)),
		compiledWire("GET", selected, "Set-CsTeamsUpdateManagementPolicy", nil, 200, string(updatedRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": id, "UpdateTimeOfDay": date}), 204, string(readResponseFixture(t, "datetime_update_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(datedRaw)),
		compiledWire("GET", selected, "Set-CsTeamsUpdateManagementPolicy", nil, 200, string(datedRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": id, "DisabledInProductMessages": edited.DisabledInProductMessages}), 204, string(readResponseFixture(t, "list_edit_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(editedRaw)),
		compiledWire("GET", selected, "Set-CsTeamsUpdateManagementPolicy", nil, 200, string(editedRaw)),
		compiledWire("PATCH", selected, "Set-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": id, "DisabledInProductMessages": []string{}}), 204, string(readResponseFixture(t, "list_clear_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(clearedRaw)),
		compiledWire("GET", collection, "Get-CsTeamsUpdateManagementPolicy", nil, 200, string(readResponseFixture(t, "owned_list.json"))),
		compiledWire("DELETE", selected, "Remove-CsTeamsUpdateManagementPolicy", body(map[string]any{"Identity": id}), 204, string(readResponseFixture(t, "delete_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpdateManagementPolicy", nil, 404, string(missing)),
	}
	run(t, contracttest.Case{Name: "owned-crud-datetime-list", Steps: steps, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies
		check := func(expected *resource.Policy) {
			v, _, e := s.GetByIdentity(ctx, id)
			if e != nil || v == nil || !reflect.DeepEqual(v.RawFields, expected.RawFields) || !reflect.DeepEqual(v.UpdateTimeOfDay, expected.UpdateTimeOfDay) || !reflect.DeepEqual(v.DisabledInProductMessages, expected.DisabledInProductMessages) {
				t.Fatal("actual settings or complete response fields changed")
			}
		}
		_, _, e := s.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(*before.Description), AllowManagedUpdates: policy.Set(true), UpdateDayOfWeek: policy.Set(int64(3)), UpdateTime: policy.Set("19:30"), AllowPublicPreview: policy.Set("Enabled")}, DisabledInProductMessages: policy.ReplaceStrings(before.DisabledInProductMessages...)})
		if e != nil {
			t.Fatal("owned create differs")
		}
		check(before)
		requests := []resource.UpdateRequest{
			{ScalarSettings: resource.ScalarSettings{Description: policy.Set(*updated.Description)}},
			{ScalarSettings: resource.ScalarSettings{UpdateTimeOfDay: policy.Set(date)}},
			{DisabledInProductMessages: policy.ChangeStrings(edited.DisabledInProductMessages, before.DisabledInProductMessages)},
			{DisabledInProductMessages: policy.ReplaceStrings()},
		}
		for i, expected := range []*resource.Policy{updated, dated, edited, cleared} {
			receipt, r, e := s.UpdateByIdentity(ctx, id, requests[i])
			if e != nil || receipt == nil || !receipt.HTTPAccepted || !receipt.PatchInvoked || receipt.ReadResponse == nil || r == nil || r.StatusCode() != 204 {
				t.Fatal("owned update lost preparation or acceptance")
			}
			check(expected)
		}
		values, _, e := s.List(ctx)
		if e != nil {
			t.Fatal("owned list failed")
		}
		count := 0
		for _, v := range values {
			if v != nil && v.Identity != nil && *v.Identity == id {
				count++
			}
		}
		if count != 1 {
			t.Fatal("owned policy missing or duplicated")
		}
		if _, _, e := s.DeleteByIdentity(ctx, id, policy.Controls{}); e != nil {
			t.Fatal("owned delete differs")
		}
		_, r, e := s.GetByIdentity(ctx, id)
		return contracttest.Observation{Response: r, Err: e}
	}})
}
