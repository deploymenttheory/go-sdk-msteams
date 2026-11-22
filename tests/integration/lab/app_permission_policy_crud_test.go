package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	appPermission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

type appPermissionPolicyCleanup struct{ service *appPermission.Service }

func (h appPermissionPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h appPermissionPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (appPermissionPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccAppPermissionPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccAppPermissionPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("app permission CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private app permission recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"app-permission-policy": appPermissionPolicyCleanup{service}}); err != nil {
			t.Error("app permission policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("app permission recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("app-permission-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("app-permission-policy", name)
	if err != nil {
		t.Fatal("app permission create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, appPermission.CreateRequest{ScalarSettings: appPermission.ScalarSettings{Description: policy.Set(initial), DefaultCatalogAppsType: policy.Set("AllowedAppList")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("app permission policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "create", response, replacements)
	var value *appPermission.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.DefaultCatalogAppsType != nil && *read.DefaultCatalogAppsType == "AllowedAppList"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_app_permission_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("app permission non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("app permission creation receipt not journaled")
	}
	compare := func(value *appPermission.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("app permission tenant binding unavailable")
		}
		actual := appPermissionOracleForIdentity(t, ctx, tenant, identity)
		if !matchesAppPermissionCRUDProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell app permission policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, appPermission.UpdateRequest{ScalarSettings: appPermission.ScalarSettings{Description: policy.Set(updated)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("app permission sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.DefaultCatalogAppsType != nil && *read.DefaultCatalogAppsType == "AllowedAppList"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_app_permission_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("app permission update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("app permission policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned app permission policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_app_permission_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (appPermissionPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("app permission deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("app permission deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("app permission policy create/read/update/list/delete and omitted-field preservation verified")
}

func matchesAppPermissionCRUDProjection(value *appPermission.Policy, settings map[string]json.RawMessage) bool {
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	a, b := json.NewDecoder(bytes.NewReader(appPermissionProjection(value))), json.NewDecoder(bytes.NewReader(raw))
	a.UseNumber()
	b.UseNumber()
	return a.Decode(&left) == nil && b.Decode(&right) == nil && reflect.DeepEqual(left, right)
}
