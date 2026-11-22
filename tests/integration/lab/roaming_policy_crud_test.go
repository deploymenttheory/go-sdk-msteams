package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	roaming "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_network_roaming_policy"
)

type roamingPolicyCleanup struct{ service *roaming.Service }

func (h roamingPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h roamingPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (roamingPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccNetworkRoamingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccNetworkRoamingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("NetworkRoaming CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private NetworkRoaming recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"roaming-policy": roamingPolicyCleanup{service}}); err != nil {
			t.Error("NetworkRoaming policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("NetworkRoaming recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("roaming-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("roaming-policy", name)
	if err != nil {
		t.Fatal("NetworkRoaming create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, roaming.Settings{Description: new(initial), AllowIPVideo: new(false), MediaBitRateKb: new(int64(50000))})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("NetworkRoaming policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "create", response, replacements)
	var value *roaming.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && labkit.SamePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowIPVideo != nil && !*read.AllowIPVideo && read.MediaBitRateKb != nil && *read.MediaBitRateKb == 50000
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("NetworkRoaming non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("NetworkRoaming creation receipt not journaled")
	}
	compare := func(value *roaming.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("NetworkRoaming tenant binding unavailable")
		}
		actual := roamingPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesRoamingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell NetworkRoaming policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, roaming.UpdateRequest{Description: new(updated)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("NetworkRoaming sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowIPVideo != nil && !*read.AllowIPVideo && read.MediaBitRateKb != nil && *read.MediaBitRateKb == 50000
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("NetworkRoaming update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("NetworkRoaming policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned NetworkRoaming policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (roamingPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("NetworkRoaming deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("NetworkRoaming deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_network_roaming_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("NetworkRoaming policy create/read/update/list/delete and omitted-field preservation verified")
}
