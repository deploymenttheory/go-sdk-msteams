package teamsbyodanddeskspolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_byod_and_desks_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// sameExceptDeviceDataCollection reports whether two snapshots differ only in
// DeviceDataCollection and server-maintained metadata, proving the sparse update
// did not reset other returned fields.
func sameExceptDeviceDataCollection(before, after map[string]json.RawMessage) bool {
	if len(before) != len(after) {
		return false
	}
	for name, value := range before {
		switch name {
		case "DeviceDataCollection", "ConfigMetadata", "DataSource":
			if _, ok := after[name]; !ok {
				return false
			}
			continue
		}
		if !bytes.Equal(value, after[name]) {
			return false
		}
	}
	return true
}

func deviceDataCollectionIs(value *resource.Policy, want string) bool {
	return value != nil && value.DeviceDataCollection != nil && *value.DeviceDataCollection == want
}

// TestAccBYODAndDesksPolicyCRUD manages one uniquely named unassigned
// policy. It never changes Global or grants a policy to a tenant, user or group.
func TestAccBYODAndDesksPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("BYOD and desks CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsBYODAndDesksPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private BYOD and desks recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"byod-and-desks-policy": byodAndDesksPolicyCleanup{service}}); err != nil {
			t.Error("BYOD and desks policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("BYOD and desks recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("byod-and-desks-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("byod-and-desks-policy", name)
	if err != nil {
		t.Fatal("BYOD and desks create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, resource.CreateRequest{Settings: resource.Settings{DeviceDataCollection: policy.Set("Disabled")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("BYOD and desks policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "create", response, replacements)
	var value *resource.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && deviceDataCollectionIs(read, "Disabled")
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("BYOD and desks non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("BYOD and desks creation receipt not journaled")
	}
	compare := func(value *resource.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("BYOD and desks tenant binding unavailable")
		}
		actual := byodAndDesksPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesBYODAndDesksPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell BYOD and desks policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	created := value.RawFields
	updated, response, err := service.UpdateByIdentity(ctx, identity, resource.UpdateRequest{Settings: resource.Settings{DeviceDataCollection: policy.Set("Enabled")}})
	if err != nil || updated == nil || !updated.HTTPAccepted || !deviceDataCollectionIs(updated.Before, "Disabled") {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("BYOD and desks sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := deviceDataCollectionIs(read, "Enabled") && sameExceptDeviceDataCollection(created, read.RawFields)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("BYOD and desks update or unchanged-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("BYOD and desks policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned BYOD and desks policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "delete", deletedResponse, replacements)
	}
	if err := (byodAndDesksPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("BYOD and desks deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("BYOD and desks deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_byod_and_desks_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("BYOD and desks policy create/read/update/list/delete verified")
}
