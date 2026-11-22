package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

type templatePermissionPolicyCleanup struct{ service *permission.Service }

func (h templatePermissionPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if labkit.IsHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h templatePermissionPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	if err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if value == nil || value.Identity == nil {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (templatePermissionPolicyCleanup) Restore(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

// TestAccTemplatePermissionPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccTemplatePermissionPolicyCRUD(t *testing.T) {
	testTemplatePermissionCRUD(t, false)
}

// TestAccTemplatePermissionPopulatedCRUD keeps the complete list scenario
// executable while its required authentication is being qualified.
func TestAccTemplatePermissionPopulatedCRUD(t *testing.T) {
	testTemplatePermissionCRUD(t, true)
}

func testTemplatePermissionCRUD(t *testing.T, populated bool) {
	t.Helper()
	prefix := ""
	if populated {
		prefix = "populated_"
	}
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
	service := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies
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
		listed, _, e := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
		if e != nil || listed == nil || listed.IsIncomplete {
			t.Fatal("complete team-template inventory required")
		}
		for _, item := range listed.Items {
			if item.ID != nil && strings.HasPrefix(*item.ID, "com.microsoft.teams.template.") {
				templateIDs = append(templateIDs, *item.ID)
			}
		}
		if len(templateIDs) < 2 {
			t.Fatal("two existing public team-template IDs required")
		}
		templateIDs = templateIDs[:2]
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
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"template-permission-policy": templatePermissionPolicyCleanup{service}}); err != nil {
			t.Error("permission policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("permission recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("template-permission-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("template-permission-policy", name)
	if err != nil {
		t.Fatal("permission create intent not journaled")
	}
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, permission.CreateRequest{ScalarSettings: permission.ScalarSettings{Description: policy.Set(initial)}, HiddenTemplates: initialEdit})
	labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"create", response, replacements)
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
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && hasTeamHiddenTemplateIDs(read, expectedIDs...)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"create_readback", response, replacements)
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
		actual := templatePermissionPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesTemplatePermissionPolicyGetterProjection(value, actual.Settings) {
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
	labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && hasTeamHiddenTemplateIDs(read, expectedIDs...)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"update_readback", response, replacements)
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
			_, response, err = service.UpdateByIdentity(ctx, identity, permission.UpdateRequest{HiddenTemplates: stage.edit})
			if err != nil {
				labkit.LogAccessResponse(t, response, err)
				t.Fatal("permission hidden-list update failed")
			}
			labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+stage.label, response, replacements)
			if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
				read, response, err := service.GetByIdentity(ctx, identity)
				if err != nil {
					return false, err
				}
				matched := read != nil && read.Description != nil && *read.Description == updated && hasTeamHiddenTemplateIDs(read, stage.ids...)
				if matched {
					value = read
					labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+stage.label+"_readback", response, replacements)
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
	labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (templatePermissionPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("permission deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("permission deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_template_permission_policy", prefix+"get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("permission policy create/read/update/list/delete and omitted-field preservation verified")
}

func hasTeamHiddenTemplateIDs(value *permission.Policy, ids ...string) bool {
	if value == nil || value.HiddenTemplates == nil || len(value.HiddenTemplates) != len(ids) {
		return false
	}
	for index, id := range ids {
		item := value.HiddenTemplates[index]
		if item == nil || item.ID == nil || *item.ID != id {
			return false
		}
	}
	return true
}
