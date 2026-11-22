package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	updatePolicy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
)

type updateManagementPolicyCleanup struct{ service *updatePolicy.Service }

func (h updateManagementPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if samePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if isHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h updateManagementPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	if err != nil && !isHTTP(err, 404) {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !isHTTP(err, 404) {
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
			if samePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (updateManagementPolicyCleanup) Restore(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func updateManagementProjection(v *updatePolicy.Policy) map[string]any {
	return map[string]any{"Description": v.Description, "AllowManagedUpdates": v.AllowManagedUpdates, "AllowPreview": v.AllowPreview, "UpdateDayOfWeek": v.UpdateDayOfWeek, "UpdateTime": v.UpdateTime, "UpdateTimeOfDay": v.UpdateTimeOfDay, "AllowPublicPreview": v.AllowPublicPreview, "UseNewTeamsClient": v.UseNewTeamsClient, "BlockLegacyAuthorization": v.BlockLegacyAuthorization, "DisabledInProductMessages": v.DisabledInProductMessages}
}
func TestAccUpdateManagementPolicyCRUD(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid update-management lab configuration")
	}
	if e := cfg.Require(lab.Mutation); errors.Is(e, lab.ErrDisabled) {
		t.Skip("update-management CRUD requires mutation opt-in and recovery directory")
	} else if e != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies
	journal, e := lab.NewJournal(cfg.JournalDirectory, binding)
	if e != nil {
		t.Fatal("private update-management journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, e := journal.Recover(ctx, map[string]lab.Handler{"update-management-policy": updateManagementPolicyCleanup{service}}); e != nil {
			t.Error("update-management cleanup unresolved; private intent retained")
		}
		if e := journal.Close(); e != nil {
			t.Error("update-management journal close failed")
		}
	})
	name, e := journal.FixtureName("update-management-policy")
	if e != nil {
		t.Fatal(e)
	}
	entry, e := journal.PlanCreate("update-management-policy", name)
	if e != nil {
		t.Fatal("update-management create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	// These are Microsoft's documented campaign categories, not tenant phone/account fixtures.
	first, second := "91382d07-8b89-444c-bbcb-cfe43133af33", "edf2633e-9827-44de-b34c-8b8b9717e84c"
	_, r, e := service.CreateByIdentity(ctx, name, updatePolicy.CreateRequest{ScalarSettings: updatePolicy.ScalarSettings{Description: policy.Set("SDK lab non-default initial description"), AllowManagedUpdates: policy.Set(true), UpdateDayOfWeek: policy.Set(int64(3)), UpdateTime: policy.Set("19:30"), AllowPublicPreview: policy.Set("Enabled")}, DisabledInProductMessages: policy.ReplaceStrings(first)})
	writeResponseFixture(t, "teams_update_management_policy", "create", r, replacements)
	if e != nil {
		logAccessResponse(t, r, e)
		t.Fatal("update-management create failed; intent retained")
	}
	var current *updatePolicy.Policy
	if e := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := service.GetByIdentity(ctx, name)
		if isHTTP(e, 404) {
			return false, nil
		}
		if e != nil {
			return false, e
		}
		matched := v != nil && v.Identity != nil && samePolicyName(*v.Identity, name) && v.Description != nil && *v.Description == "SDK lab non-default initial description" && v.AllowManagedUpdates != nil && *v.AllowManagedUpdates && v.UpdateDayOfWeek != nil && *v.UpdateDayOfWeek == 3 && v.UpdateTime != nil && *v.UpdateTime == "19:30" && v.AllowPublicPreview != nil && *v.AllowPublicPreview == "Enabled" && len(v.DisabledInProductMessages) == 1 && v.DisabledInProductMessages[0] == first
		if matched {
			current = v
			writeResponseFixture(t, "teams_update_management_policy", "create_readback", r, replacements)
		}
		return matched, nil
	}); e != nil {
		t.Fatal("update-management non-default create read-back did not converge")
	}
	identity := *current.Identity
	if e := journal.Applied(entry, identity); e != nil {
		t.Fatal("update-management receipt not journaled")
	}
	compare := func(v *updatePolicy.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("update-management tenant binding unavailable")
		}
		actual := updateManagementPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !sameTemplateProjection(updateManagementProjection(v), actual.Settings) {
			t.Fatal("SDK and PowerShell update-management getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(current)
	original := updateManagementProjection(current)
	_, r, e = service.UpdateByIdentity(ctx, identity, updatePolicy.UpdateRequest{ScalarSettings: updatePolicy.ScalarSettings{Description: policy.Set("SDK lab non-default updated description")}})
	writeResponseFixture(t, "teams_update_management_policy", "update", r, replacements)
	if e != nil {
		logAccessResponse(t, r, e)
		t.Fatal("update-management sparse update failed")
	}
	original["Description"] = new("SDK lab non-default updated description")
	if e := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := service.GetByIdentity(ctx, identity)
		matched := e == nil && v != nil && sameTemplateProjection(original, updateManagementProjection(v))
		if matched {
			current = v
			writeResponseFixture(t, "teams_update_management_policy", "update_readback", r, replacements)
		}
		return matched, e
	}); e != nil {
		t.Fatal("update-management sparse omission did not persist")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(current)
	if current.UpdateTimeOfDay == nil {
		t.Fatal("update-management DateTime getter missing")
	}
	clock := current.UpdateTimeOfDay.Time().Add(45 * time.Minute)
	date, e := updatePolicy.NewDateTime(clock, current.UpdateTimeOfDay.Kind())
	if e != nil {
		t.Fatal("update-management DateTime construction failed")
	}
	_, r, e = service.UpdateByIdentity(ctx, identity, updatePolicy.UpdateRequest{ScalarSettings: updatePolicy.ScalarSettings{UpdateTimeOfDay: policy.Set(date)}})
	writeResponseFixture(t, "teams_update_management_policy", "datetime_update", r, replacements)
	if e != nil {
		logAccessResponse(t, r, e)
		t.Fatal("update-management DateTime update failed")
	}
	if e := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := service.GetByIdentity(ctx, identity)
		matched := e == nil && v != nil && v.UpdateTimeOfDay != nil && v.UpdateTimeOfDay.Time().Hour() == clock.Hour() && v.UpdateTimeOfDay.Time().Minute() == clock.Minute()
		if matched {
			current = v
			writeResponseFixture(t, "teams_update_management_policy", "datetime_readback", r, replacements)
		}
		return matched, e
	}); e != nil {
		t.Fatal("update-management DateTime update did not persist")
	}
	t.Log("update-management DateTime update/read-back: passed")
	compare(current)
	for _, stage := range []struct {
		label, marker string
		edit          *policy.ListEdit
		want          []string
	}{
		{"list_edit", "update-management list edit/read-back: passed", policy.ChangeStrings([]string{second}, []string{first}), []string{second}},
		{"list_clear", "update-management list clear/read-back: passed", policy.ReplaceStrings(), []string{}},
	} {
		_, r, e = service.UpdateByIdentity(ctx, identity, updatePolicy.UpdateRequest{DisabledInProductMessages: stage.edit})
		writeResponseFixture(t, "teams_update_management_policy", stage.label, r, replacements)
		if e != nil {
			logAccessResponse(t, r, e)
			t.Fatal("update-management list update failed")
		}
		want := updateManagementProjection(current)
		want["DisabledInProductMessages"] = stage.want
		if e := observeUntil(ctx, func(ctx context.Context) (bool, error) {
			v, r, e := service.GetByIdentity(ctx, identity)
			matched := e == nil && v != nil && sameTemplateProjection(want, updateManagementProjection(v))
			if matched {
				current = v
				writeResponseFixture(t, "teams_update_management_policy", stage.label+"_readback", r, replacements)
			}
			return matched, e
		}); e != nil {
			t.Fatal("update-management list edit or clear did not persist")
		}
		t.Log(stage.marker)
		compare(current)
	}
	values, r, e := service.List(ctx)
	if e != nil {
		t.Fatal("update-management list failed")
	}
	count := 0
	for _, v := range values {
		if v != nil && v.Identity != nil && samePolicyName(*v.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned update-management policy missing or ambiguous")
	}
	writeResponseFixture(t, "teams_update_management_policy", "owned_list", r, replacements)
	t.Log("policy list: passed")
	_, r, e = service.DeleteByIdentity(ctx, identity, policy.Controls{})
	writeResponseFixture(t, "teams_update_management_policy", "delete", r, replacements)
	if e != nil {
		t.Fatal("update-management delete failed")
	}
	if e := (updateManagementPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); e != nil {
		t.Fatal("update-management independent absence not verified")
	}
	_, r, e = service.GetByIdentity(ctx, identity)
	if !isHTTP(e, 404) {
		t.Fatal("update-management deleted read did not return404")
	}
	writeResponseFixture(t, "teams_update_management_policy", "get_deleted", r, replacements)
	t.Log("policy delete/absence: passed")
}
