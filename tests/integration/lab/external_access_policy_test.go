package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

type externalPolicyCleanup struct{ service *external.Service }

func (h externalPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	var found string
	for _, v := range values {
		if v != nil && v.Identity != nil && samePolicyName(*v.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *v.Identity
		}
	}
	return found, nil
}
func (h externalPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	var api *client.APIError
	if errors.As(err, &api) && api.StatusCode == 404 {
		return nil
	}
	return err
}
func (externalPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

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
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
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
	initial := "SDK lab initial external policy description"
	if _, _, err := service.Create(ctx, external.CreateRequest{Identity: name, Settings: external.Settings{Fields: external.Fields{Description: policy.Set(initial), EnableFederationAccess: policy.Set(false)}}}); err != nil {
		t.Fatal("external policy creation failed; private recovery journal retained")
	}
	value, _, err := service.GetByIdentity(ctx, name)
	if err != nil || value == nil || value.Identity == nil || !samePolicyName(*value.Identity, name) || value.Description == nil || *value.Description != initial || value.EnableFederationAccess == nil || *value.EnableFederationAccess {
		t.Fatal("external policy non-default create read-back differs")
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
		compareFederation(t, "external-policy", value.RawFields, federationOracleForIdentity(t, ctx, "external-policy", tenant, identity))
	}
	compare(value)
	updated := "SDK lab updated external policy description"
	if _, _, err := service.UpdateByIdentity(ctx, identity, external.UpdateRequest{Fields: external.Fields{Description: policy.Set(updated)}}); err != nil {
		t.Fatal("external policy update failed; cleanup journal retained")
	}
	value, _, err = service.GetByIdentity(ctx, identity)
	if err != nil || value == nil || value.Description == nil || *value.Description != updated || value.EnableFederationAccess == nil || *value.EnableFederationAccess {
		t.Fatal("external policy update read-back or omission differs")
	}
	compare(value)
	values, _, err := service.List(ctx)
	if err != nil {
		t.Fatal("external policy list failed")
	}
	found := 0
	for _, v := range values {
		if v != nil && v.Identity != nil && samePolicyName(*v.Identity, name) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned external policy missing/ambiguous in list")
	}
	if _, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		t.Fatal("owned external policy deletion failed")
	}
	_, _, err = service.GetByIdentity(ctx, identity)
	var api *client.APIError
	if !errors.As(err, &api) || api.StatusCode != 404 {
		t.Fatal("deleted external policy absence unverified")
	}
	t.Log("owned unassigned external policy CRUD/read-back/absence verified; effective policy and assignments remain unqualified")
}
