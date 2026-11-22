package teamsmobilitypolicyacc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	mobility "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_mobility_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMobilityPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccMobilityPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("mobility CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMobilityPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private mobility recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"mobility-policy": mobilityPolicyCleanup{service}}); err != nil {
			t.Error("mobility policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("mobility recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("mobility-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("mobility-policy", name)
	if err != nil {
		t.Fatal("mobility create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, mobility.CreateRequest{Settings: mobility.Settings{Description: policy.Set("SDK lab non-default initial description"), IPVideoMobileMode: policy.Set("AllNetworks")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("mobility policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_mobility_policy", "create", response, replacements)
	var value *mobility.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == "SDK lab non-default initial description" && read.IPVideoMobileMode != nil && *read.IPVideoMobileMode == "AllNetworks"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_mobility_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("mobility non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("mobility creation receipt not journaled")
	}
	compare := func(value *mobility.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("mobility tenant binding unavailable")
		}
		actual := mobilityPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesMobilityPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell mobility policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, mobility.UpdateRequest{Settings: mobility.Settings{Description: policy.Set("SDK lab non-default updated description")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("mobility sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_mobility_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == "SDK lab non-default updated description" && read.IPVideoMobileMode != nil && *read.IPVideoMobileMode == "AllNetworks"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_mobility_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("mobility update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("mobility policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned mobility policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_mobility_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_mobility_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (mobilityPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("mobility deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("mobility deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_mobility_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("mobility policy create/read/update/list/delete and omitted-field preservation verified")
}
