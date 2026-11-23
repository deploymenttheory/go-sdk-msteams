package teamssharedcallingroutingpolicyacc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	sharedCalling "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shared_calling_routing_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccSharedCallingRoutingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func testSharedCallingPolicyCRUD(t *testing.T, populated bool) {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("shared calling CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies
	numbers := []string{}
	var allNumbers []string
	account := policy.Field[string]{}
	if populated {
		file := os.Getenv("MSTEAMS_LAB_SHARED_CALLING_FIXTURE_FILE")
		if file == "" {
			t.Skip("populated shared calling requires a bound resource account and callback-number fixture")
		}
		raw, e := labkit.PrivateFile(file)
		if e != nil {
			t.Fatal("private shared-calling fixture unavailable")
		}
		defer clear(raw)
		var fixture struct {
			TenantID          string   `json:"tenant_id"`
			ResourceAccountID string   `json:"resource_account_id"`
			EmergencyNumbers  []string `json:"emergency_numbers"`
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.ResourceAccountID) || len(fixture.EmergencyNumbers) < 2 {
			t.Fatal("shared-calling fixture binding or values differ")
		}
		for _, number := range fixture.EmergencyNumbers {
			if !strings.HasPrefix(number, "+") || strings.ContainsAny(number, " \r\n\t") {
				t.Fatal("callback fixture must use actual assigned E.164 telephone inventory")
			}
		}
		allNumbers = append(allNumbers, fixture.EmergencyNumbers...)
		numbers = append(numbers, allNumbers[0])
		account = policy.Set(fixture.ResourceAccountID)
	}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private shared calling recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"shared-calling-policy": sharedCallingPolicyCleanup{service}}); err != nil {
			t.Error("shared calling policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("shared calling recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("shared-calling-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("shared-calling-policy", name)
	if err != nil {
		t.Fatal("shared calling create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	for i, number := range allNumbers {
		replacements[number] = fmt.Sprintf("+120255501%02d", i+1)
	}
	_, response, err := service.CreateByIdentity(ctx, name, sharedCalling.CreateRequest{ScalarSettings: sharedCalling.ScalarSettings{Description: policy.Set("SDK lab non-default initial description"), ResourceAccount: account}, EmergencyNumbers: policy.ReplaceStrings(numbers...)})
	labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "create", response, replacements)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("shared calling policy create failed; recovery journal retained")
	}
	var value *sharedCalling.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == "SDK lab non-default initial description" && slices.Equal(read.EmergencyNumbers, numbers)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("shared calling non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("shared calling creation receipt not journaled")
	}
	compare := func(value *sharedCalling.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("shared calling tenant binding unavailable")
		}
		actual := sharedCallingPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesSharedCallingRoutingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell shared calling policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, sharedCalling.UpdateRequest{ScalarSettings: sharedCalling.ScalarSettings{Description: policy.Set("SDK lab non-default updated description")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("shared calling sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == "SDK lab non-default updated description" && slices.Equal(read.EmergencyNumbers, numbers)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("shared calling update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	if populated {
		for _, stage := range []struct {
			name string
			edit *policy.ListEdit
			want []string
		}{
			{"emergency_list_edit", policy.ChangeStrings([]string{allNumbers[1]}, []string{allNumbers[0]}), []string{allNumbers[1]}},
			{"emergency_list_clear", policy.ReplaceStrings(), []string{}},
		} {
			_, r, e := service.UpdateByIdentity(ctx, identity, sharedCalling.UpdateRequest{EmergencyNumbers: stage.edit})
			labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", stage.name, r, replacements)
			if e != nil {
				labkit.LogAccessResponse(t, r, e)
				t.Fatal("shared calling list edit failed")
			}
			if e := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
				read, r, e := service.GetByIdentity(ctx, identity)
				matched := e == nil && read != nil && slices.Equal(read.EmergencyNumbers, stage.want) && read.Description != nil && *read.Description == "SDK lab non-default updated description"
				if matched {
					value = read
					labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", stage.name+"_readback", r, replacements)
				}
				return matched, e
			}); e != nil {
				t.Fatal("shared calling list edit read-back did not converge")
			}
			compare(value)
		}
		t.Log("shared calling populated list edit/clear: passed")
	}
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("shared calling policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned shared calling policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (sharedCallingPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("shared calling deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("shared calling deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_shared_calling_routing_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("shared calling policy create/read/update/list/delete and omitted-field preservation verified")
}

// TestAccSharedCallingRoutingPolicyCRUD qualifies an unassigned policy without
// resource-account or callback-number associations. It does not qualify a populated list.
func TestAccSharedCallingRoutingPolicyCRUD(t *testing.T) { testSharedCallingPolicyCRUD(t, false) }

// TestAccSharedCallingRoutingPolicyPopulatedCRUD requires real bound telephone inventory.
func TestAccSharedCallingRoutingPolicyPopulatedCRUD(t *testing.T) {
	testSharedCallingPolicyCRUD(t, true)
}
