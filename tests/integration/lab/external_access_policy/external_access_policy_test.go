package externalaccesspolicyacc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccExternalAccessPolicyCRUD creates one uniquely named unassigned policy.
// It never grants it, changes Global, or edits tenant federation configuration.
func TestAccExternalAccessPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("external policy CRUD requires explicit mutation opt-in and a private recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ExternalAccessPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, stop := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer stop()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"external-access-policy": externalPolicyCleanup{service}}); err != nil {
			t.Error("external policy cleanup incomplete; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("private journal close failed")
		}
	})
	name, err := journal.FixtureName("external-access")
	if err != nil {
		t.Fatal("owned fixture unavailable")
	}
	entry, err := journal.PlanCreate("external-access-policy", name)
	if err != nil {
		t.Fatal("create intent not journaled")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	initial := "SDK lab non-default initial description"
	if _, createResponse, err := service.Create(ctx, external.CreateRequest{Identity: name, Settings: external.Settings{Fields: external.Fields{Description: policy.Set(initial), EnableFederationAccess: policy.Set(false)}}}); err != nil {
		labkit.LogAccessResponse(t, createResponse, err)
		t.Fatal("external policy creation failed; private recovery journal retained")
	} else {
		labkit.WriteResponseFixture(t, "external_access_policy", "create", createResponse, map[string]string{name: "fixture-policy"})
	}
	var value *external.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.EnableFederationAccess != nil && !*read.EnableFederationAccess
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "external_access_policy", "create_readback", response, map[string]string{name: "fixture-policy"})
		}
		return matched, nil
	}); err != nil {
		t.Fatal("external policy non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("creation receipt not journaled")
	}
	compare := func(value *external.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("tenant binding unavailable")
		}
		labkit.CompareFederation(t, "external-policy", value.RawFields, labkit.FederationOracleForIdentity(t, ctx, "external-policy", tenant, identity))
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	if _, response, err := service.UpdateByIdentity(ctx, identity, external.UpdateRequest{Fields: external.Fields{Description: policy.Set(updated)}}); err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("external policy update failed; cleanup journal retained")
	} else {
		labkit.WriteResponseFixture(t, "external_access_policy", "update", response, map[string]string{name: "fixture-policy"})
	}
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.EnableFederationAccess != nil && !*read.EnableFederationAccess
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "external_access_policy", "update_readback", response, map[string]string{name: "fixture-policy"})
		}
		return matched, nil
	}); err != nil {
		t.Fatal("external policy update read-back or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("external policy list failed")
	}
	found := 0
	for _, v := range values {
		if v != nil && v.Identity != nil && labkit.SamePolicyName(*v.Identity, name) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned external policy missing/ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "external_access_policy", "owned_list", listResponse, map[string]string{name: "fixture-policy"})
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "external_access_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (externalPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("external policy deletion and independent absence were not verified")
	}
	_, response, err := service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted external policy absence unverified")
	}
	labkit.WriteResponseFixture(t, "external_access_policy", "get_deleted", response, map[string]string{name: "fixture-policy"})
	t.Log("policy delete/absence: passed")

	t.Log("owned unassigned external policy CRUD/read-back/absence verified; effective policy and assignments remain unqualified")
}
