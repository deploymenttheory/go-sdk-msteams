package teamsmediaconnectivitypolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_connectivity_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// sameExceptDirectConnection reports whether two snapshots differ only in
// DirectConnection and server-maintained metadata, proving the sparse update
// did not reset other returned fields.
func sameExceptDirectConnection(before, after map[string]json.RawMessage) bool {
	if len(before) != len(after) {
		return false
	}
	for name, value := range before {
		switch name {
		case "DirectConnection", "ConfigMetadata", "DataSource":
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

func directConnectionIs(value *resource.Policy, want string) bool {
	return value != nil && value.DirectConnection != nil && *value.DirectConnection == want
}

// TestAccMediaConnectivityPolicyCRUD manages one uniquely named unassigned
// policy. It never changes Global or grants a policy to a tenant, user or group.
func TestAccMediaConnectivityPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("media connectivity CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMediaConnectivityPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private media connectivity recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"media-connectivity-policy": mediaConnectivityPolicyCleanup{service}}); err != nil {
			t.Error("media connectivity policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("media connectivity recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("media-connectivity-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("media-connectivity-policy", name)
	if err != nil {
		t.Fatal("media connectivity create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, resource.CreateRequest{Settings: resource.Settings{DirectConnection: policy.Set("Disabled")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("media connectivity policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "create", response, replacements)
	var value *resource.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && directConnectionIs(read, "Disabled")
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("media connectivity non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("media connectivity creation receipt not journaled")
	}
	compare := func(value *resource.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("media connectivity tenant binding unavailable")
		}
		actual := mediaConnectivityPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesMediaConnectivityPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell media connectivity policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	created := value.RawFields
	updated, response, err := service.UpdateByIdentity(ctx, identity, resource.UpdateRequest{Settings: resource.Settings{DirectConnection: policy.Set("Enabled")}})
	if err != nil || updated == nil || !updated.HTTPAccepted || !directConnectionIs(updated.Before, "Disabled") {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("media connectivity sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := directConnectionIs(read, "Enabled") && sameExceptDirectConnection(created, read.RawFields)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("media connectivity update or unchanged-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("media connectivity policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned media connectivity policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "delete", deletedResponse, replacements)
	}
	if err := (mediaConnectivityPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("media connectivity deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("media connectivity deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_media_connectivity_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("media connectivity policy create/read/update/list/delete verified")
}
