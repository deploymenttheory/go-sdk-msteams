package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	cortana "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

type cortanaPolicyCleanup struct{ service *cortana.Service }

func (h cortanaPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h cortanaPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (cortanaPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccCortanaPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccCortanaPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("Cortana CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsCortanaPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private Cortana recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"cortana-policy": cortanaPolicyCleanup{service}}); err != nil {
			t.Error("Cortana policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("Cortana recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("cortana-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("cortana-policy", name)
	if err != nil {
		t.Fatal("Cortana create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, cortana.Settings{Description: new(initial), AllowCortanaAmbientListening: new(false)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("Cortana policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_cortana_policy", "create", response, replacements)
	var value *cortana.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && samePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowCortanaAmbientListening != nil && !*read.AllowCortanaAmbientListening
		if matched {
			value = read
			writeResponseFixture(t, "teams_cortana_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("Cortana non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("Cortana creation receipt not journaled")
	}
	compare := func(value *cortana.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("Cortana tenant binding unavailable")
		}
		actual := cortanaPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesCortanaProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell Cortana policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, cortana.UpdateRequest{Description: new(updated)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("Cortana sparse update failed")
	}
	writeResponseFixture(t, "teams_cortana_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowCortanaAmbientListening != nil && !*read.AllowCortanaAmbientListening
		if matched {
			value = read
			writeResponseFixture(t, "teams_cortana_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("Cortana update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("Cortana policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && samePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned Cortana policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_cortana_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_cortana_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (cortanaPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("Cortana deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("Cortana deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_cortana_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("Cortana policy create/read/update/list/delete and omitted-field preservation verified")
}
