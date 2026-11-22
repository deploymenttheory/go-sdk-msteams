package teamsmeetingtemplatepermissionpolicyacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMeetingTemplatePermissionPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccMeetingTemplatePermissionPolicyCRUD(t *testing.T) {
	testMeetingTemplatePermissionCRUD(t, false)
}

// TestAccMeetingTemplatePermissionPopulatedCRUD keeps the complete list scenario
// executable while its required authentication is being qualified.
func TestAccMeetingTemplatePermissionPopulatedCRUD(t *testing.T) {
	testMeetingTemplatePermissionCRUD(t, true)
}

func testMeetingTemplatePermissionCRUD(t *testing.T, populated bool) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("permission CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("permission tenant binding unavailable")
	}
	var templateIDs []string
	initialEdit := permission.ClearHiddenTemplates()
	var expectedIDs []string
	if populated {
		templateIDs = existingMeetingTemplateIDs(t, ctx, tenant)
		expectedIDs = templateIDs[:1]
		initialEdit = permission.ReplaceHiddenTemplates(permission.NewHiddenTemplate(templateIDs[0]))
	}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private permission recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"meeting-template-policy": meetingTemplatePolicyCleanup{service}}); err != nil {
			t.Error("permission policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("permission recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("meeting-template-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("meeting-template-policy", name)
	if err != nil {
		t.Fatal("permission create intent not journaled")
	}
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, permission.CreateRequest{ScalarSettings: permission.ScalarSettings{Description: policy.Set(initial)}, HiddenMeetingTemplates: initialEdit})
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "create", response, replacements)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("permission policy create failed; recovery journal retained")
	}
	var value *permission.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && hasHiddenTemplateIDs(read, expectedIDs...)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("permission non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("permission creation receipt not journaled")
	}
	compare := func(value *permission.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("permission tenant binding unavailable")
		}
		actual := labkit.MeetingTemplatePolicyOracleForIdentity(t, ctx, "permission", tenant, identity)
		if !labkit.SameProjection(labkit.MeetingTemplatePermissionProjection(value), actual.Settings) {
			t.Fatal("SDK and PowerShell permission policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, permission.UpdateRequest{ScalarSettings: permission.ScalarSettings{Description: policy.Set(updated)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("permission sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("permission update request metadata missing")
	}
	sent, ok := response.Request.Body.([]byte)
	var fields map[string]json.RawMessage
	if !ok || json.Unmarshal(sent, &fields) != nil || len(fields) != 2 || fields["Identity"] == nil || fields["Description"] == nil {
		t.Fatal("description-only request included an unbound setting")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && hasHiddenTemplateIDs(read, expectedIDs...)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("permission update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	if populated {
		for _, stage := range []struct {
			edit          *permission.HiddenTemplateEdit
			ids           []string
			label, marker string
		}{
			{permission.ChangeHiddenTemplates([]*permission.HiddenTemplate{permission.NewHiddenTemplate(templateIDs[1])}, []*permission.HiddenTemplate{permission.NewHiddenTemplate(templateIDs[0])}), []string{templateIDs[1]}, "list_edit", "policy hidden-list edit/read-back: passed"},
			{permission.ClearHiddenTemplates(), []string{}, "list_clear", "policy hidden-list clear/read-back: passed"},
		} {
			_, response, err = service.UpdateByIdentity(ctx, identity, permission.UpdateRequest{HiddenMeetingTemplates: stage.edit})
			if err != nil {
				labkit.LogAccessResponse(t, response, err)
				t.Fatal("permission hidden-list update failed")
			}
			labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", stage.label, response, replacements)
			if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
				read, response, err := service.GetByIdentity(ctx, identity)
				if err != nil {
					return false, err
				}
				matched := read != nil && read.Description != nil && *read.Description == updated && hasHiddenTemplateIDs(read, stage.ids...)
				if matched {
					value = read
					labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", stage.label+"_readback", response, replacements)
				}
				return matched, nil
			}); err != nil {
				t.Fatal("permission hidden-list edit or clear did not persist")
			}
			compare(value)
			t.Log(stage.marker)
		}
	}

	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("permission policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned permission policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (meetingTemplatePolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("permission deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("permission deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("permission policy create/read/update/list/delete and omitted-field preservation verified")
}

func hasHiddenTemplateIDs(value *permission.Policy, ids ...string) bool {
	if value == nil || value.HiddenMeetingTemplates == nil || len(value.HiddenMeetingTemplates) != len(ids) {
		return false
	}
	for index, id := range ids {
		item := value.HiddenMeetingTemplates[index]
		if item == nil || item.ID == nil || *item.ID != id {
			return false
		}
	}
	return true
}
