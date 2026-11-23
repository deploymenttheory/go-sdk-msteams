package teamsmessagingpolicyacc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMessagingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccMessagingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("messaging CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMessagingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private messaging recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"messaging-policy": messagingPolicyCleanup{service}}); err != nil {
			t.Error("messaging policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("messaging recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("messaging-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("messaging-policy", name)
	if err != nil {
		t.Fatal("messaging create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, messaging.CreateRequest{Settings: messaging.Settings{Description: policy.Set(initial), AllowGiphy: policy.Set(false)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("messaging policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "create", response, replacements)
	var value *messaging.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowGiphy != nil && !*read.AllowGiphy
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_messaging_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("messaging non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("messaging creation receipt not journaled")
	}
	compare := func(value *messaging.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("messaging tenant binding unavailable")
		}
		actual := messagingPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesMessagingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell messaging policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, messaging.UpdateRequest{Settings: messaging.Settings{Description: policy.Set(updated)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("messaging sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowGiphy != nil && !*read.AllowGiphy
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_messaging_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("messaging update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("messaging policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned messaging policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_messaging_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (messagingPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("messaging deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("messaging deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("messaging policy create/read/update/list/delete and omitted-field preservation verified")
}
