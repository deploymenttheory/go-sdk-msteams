package teamsaipolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ai_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// Owned CRUD values. Create sets a description and disables face enrollment
// (the lab Global has no description); update enables face enrollment and
// explicitly clears the description with JSON null, omitting everything else.
const (
	createdDescription = "msteams-sdk owned AI policy"
	createdEnrollFace  = "Disabled"
	updatedEnrollFace  = "Enabled"
)

// sameExceptUpdatedFields reports whether two snapshots differ only in the
// updated Description/EnrollFace and server-maintained metadata, proving the
// sparse update did not reset other returned fields.
func sameExceptUpdatedFields(before, after map[string]json.RawMessage) bool {
	if len(before) != len(after) {
		return false
	}
	for name, value := range before {
		switch name {
		case "Description", "EnrollFace", "ConfigMetadata", "DataSource":
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

func createdValues(value *resource.Policy) bool {
	return value != nil && value.Description != nil && *value.Description == createdDescription &&
		value.EnrollFace != nil && *value.EnrollFace == createdEnrollFace
}

func updatedValues(value *resource.Policy) bool {
	return value != nil && value.Description == nil && value.EnrollFace != nil && *value.EnrollFace == updatedEnrollFace
}

// TestAccAIPolicyCRUD manages one uniquely named unassigned
// policy. It never changes Global or grants a policy to a tenant, user or group.
func TestAccAIPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("AI CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsAIPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private AI recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"ai-policy": aiPolicyCleanup{service}}); err != nil {
			t.Error("AI policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("AI recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("ai-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("ai-policy", name)
	if err != nil {
		t.Fatal("AI create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, resource.CreateRequest{Settings: resource.Settings{Description: policy.Set(createdDescription), EnrollFace: policy.Set(createdEnrollFace)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("AI policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_ai_policy", "create", response, replacements)
	var value *resource.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && createdValues(read)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_ai_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("AI non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("AI creation receipt not journaled")
	}
	compare := func(value *resource.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("AI tenant binding unavailable")
		}
		actual := aiPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesAIPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell AI policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	created := value.RawFields
	updated, response, err := service.UpdateByIdentity(ctx, identity, resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string](), EnrollFace: policy.Set(updatedEnrollFace)}})
	if err != nil || updated == nil || !updated.HTTPAccepted || !createdValues(updated.Before) {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("AI sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_ai_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := updatedValues(read) && sameExceptUpdatedFields(created, read.RawFields)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_ai_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("AI update or unchanged-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("AI policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned AI policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_ai_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_ai_policy", "delete", deletedResponse, replacements)
	}
	if err := (aiPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("AI deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("AI deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_ai_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("AI policy create/read/update/list/delete verified")
}
