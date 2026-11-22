package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	workload "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

type workLoadPolicyCleanup struct{ service *workload.Service }

func (h workLoadPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value.Identity == "" {
			return "", lab.ErrUnresolved
		}
		if samePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
		// An empty list match alone cannot establish absence after a lost POST.
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
func (h workLoadPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
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
			if value.Identity == "" {
				return false, lab.ErrUnresolved
			}
			if samePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (workLoadPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccWorkLoadPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccWorkLoadPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("WorkLoad CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private WorkLoad recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"work-load-policy": workLoadPolicyCleanup{service}}); err != nil {
			t.Error("WorkLoad policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("WorkLoad recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("work-load-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("work-load-policy", name)
	if err != nil {
		t.Fatal("WorkLoad create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, workload.Settings{Description: new(initial), AllowCalling: new(false), AllowCallingPinned: new(false), AllowMeeting: new(false), AllowMeetingPinned: new(false), AllowMessaging: new(false), AllowMessagingPinned: new(false)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("WorkLoad policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_work_load_policy", "create", response, replacements)
	var value *workload.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && samePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && allWorkloadDisabled(read)
		if matched {
			value = read
			writeResponseFixture(t, "teams_work_load_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("WorkLoad non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("WorkLoad creation receipt not journaled")
	}
	compare := func(value *workload.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("WorkLoad tenant binding unavailable")
		}
		actual := workLoadPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesWorkLoadPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell WorkLoad policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	workLoadExpected := workLoadProjection(value)
	workLoadExpected["Description"] = &updated
	response, err = service.UpdateByIdentity(ctx, identity, workload.UpdateRequest{Description: new(updated)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("WorkLoad sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("workload update receipt missing")
	}
	var sent map[string]json.RawMessage
	encoded, ok := response.Request.Body.([]byte)
	if !ok || json.Unmarshal(encoded, &sent) != nil || len(sent) != 2 || sent["Identity"] == nil || sent["Description"] == nil {
		t.Fatal("workload sparse update included an unbound field")
	}
	writeResponseFixture(t, "teams_work_load_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && sameTemplateProjection(workLoadExpected, workLoadProjection(read))
		if matched {
			value = read
			writeResponseFixture(t, "teams_work_load_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("WorkLoad update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("WorkLoad policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && samePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned WorkLoad policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_work_load_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_work_load_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (workLoadPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("WorkLoad deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("WorkLoad deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_work_load_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("WorkLoad policy create/read/update/list/delete and omitted-field preservation verified")
}

func allWorkloadDisabled(v *workload.Policy) bool {
	for _, flag := range []*bool{v.AllowCalling, v.AllowCallingPinned, v.AllowMeeting, v.AllowMeetingPinned, v.AllowMessaging, v.AllowMessagingPinned} {
		if flag == nil || *flag {
			return false
		}
	}
	return true
}
