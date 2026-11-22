package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	appSetup "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

type appSetupPolicyCleanup struct{ service *appSetup.Service }

func (h appSetupPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h appSetupPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (appSetupPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccAppSetupPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccAppSetupPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("app setup CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsAppSetupPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private app setup recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"app-setup-policy": appSetupPolicyCleanup{service}}); err != nil {
			t.Error("app setup policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("app setup recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("app-setup-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("app-setup-policy", name)
	if err != nil {
		t.Fatal("app setup create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, appSetup.CreateRequest{ScalarSettings: appSetup.ScalarSettings{Description: policy.Set(initial), AllowUserPinning: policy.Set(false)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("app setup policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_app_setup_policy", "create", response, replacements)
	var value *appSetup.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowUserPinning != nil && !*read.AllowUserPinning
		if matched {
			value = read
			writeResponseFixture(t, "teams_app_setup_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("app setup non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("app setup creation receipt not journaled")
	}
	compare := func(value *appSetup.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("app setup tenant binding unavailable")
		}
		actual := appSetupOracleForIdentity(t, ctx, tenant, identity)
		if !matchesAppSetupCRUDProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell app setup policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, appSetup.UpdateRequest{ScalarSettings: appSetup.ScalarSettings{Description: policy.Set(updated)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("app setup sparse update failed")
	}
	writeResponseFixture(t, "teams_app_setup_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowUserPinning != nil && !*read.AllowUserPinning
		if matched {
			value = read
			writeResponseFixture(t, "teams_app_setup_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("app setup update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("app setup policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned app setup policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_app_setup_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_app_setup_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (appSetupPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("app setup deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("app setup deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_app_setup_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("app setup policy create/read/update/list/delete and omitted-field preservation verified")
}

func matchesAppSetupCRUDProjection(value *appSetup.Policy, settings map[string]json.RawMessage) bool {
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	a, b := json.NewDecoder(bytes.NewReader(appSetupProjection(value))), json.NewDecoder(bytes.NewReader(raw))
	a.UseNumber()
	b.UseNumber()
	return a.Decode(&left) == nil && b.Decode(&right) == nil && reflect.DeepEqual(left, right)
}
