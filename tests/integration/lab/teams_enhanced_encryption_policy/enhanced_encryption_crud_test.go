package teamsenhancedencryptionpolicyacc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	encryption "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccEnhancedEncryptionPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccEnhancedEncryptionPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("enhanced encryption CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private enhanced encryption recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"enhanced-encryption-policy": enhancedEncryptionPolicyCleanup{service}}); err != nil {
			t.Error("enhanced encryption policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("enhanced encryption recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("enhanced-encryption-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("enhanced-encryption-policy", name)
	if err != nil {
		t.Fatal("enhanced encryption create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, encryption.Settings{Description: new(initial), CallingEndToEndEncryptionEnabledType: new("Disabled")})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("enhanced encryption policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "create", response, replacements)
	var value *encryption.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && labkit.SamePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.CallingEndToEndEncryptionEnabledType != nil && *read.CallingEndToEndEncryptionEnabledType == "Disabled"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("enhanced encryption non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("enhanced encryption creation receipt not journaled")
	}
	compare := func(value *encryption.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("enhanced encryption tenant binding unavailable")
		}
		actual := enhancedEncryptionOracleForIdentity(t, ctx, tenant, identity)
		if !matchesEnhancedEncryptionProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell enhanced encryption policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, encryption.UpdateRequest{Description: new(updated)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("enhanced encryption sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.CallingEndToEndEncryptionEnabledType != nil && *read.CallingEndToEndEncryptionEnabledType == "Disabled"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("enhanced encryption update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("enhanced encryption policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned enhanced encryption policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (enhancedEncryptionPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("enhanced encryption deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("enhanced encryption deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_enhanced_encryption_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("enhanced encryption policy create/read/update/list/delete and omitted-field preservation verified")
}
