package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	callPark "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

type callParkPolicyCleanup struct{ service *callPark.Service }

func (h callParkPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value.Identity == "" {
			return "", lab.ErrUnresolved
		}
		if labkit.SamePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
		// An empty list match alone cannot establish absence after a lost POST.
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
func (h callParkPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
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
			if value.Identity == "" {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (callParkPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccCallParkPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccCallParkPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("call park CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsCallParkPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private call park recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"call-park-policy": callParkPolicyCleanup{service}}); err != nil {
			t.Error("call park policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("call park recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("call-park-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("call-park-policy", name)
	if err != nil {
		t.Fatal("call park create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, callPark.Settings{Description: new(initial), AllowCallPark: new(true), ParkTimeoutSeconds: new(int64(180))})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("call park policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_call_park_policy", "create", response, replacements)
	var value *callPark.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && labkit.SamePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowCallPark != nil && *read.AllowCallPark && read.ParkTimeoutSeconds != nil && *read.ParkTimeoutSeconds == 180
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_call_park_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("call park non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("call park creation receipt not journaled")
	}
	compare := func(value *callPark.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("call park tenant binding unavailable")
		}
		actual := callParkPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesCallParkGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell call park policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, callPark.UpdateRequest{Description: new(updated)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("call park sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_call_park_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowCallPark != nil && *read.AllowCallPark && read.ParkTimeoutSeconds != nil && *read.ParkTimeoutSeconds == 180
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_call_park_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("call park update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("call park policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned call park policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_call_park_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_call_park_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (callParkPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("call park deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("call park deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_call_park_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("call park policy create/read/update/list/delete and omitted-field preservation verified")
}
