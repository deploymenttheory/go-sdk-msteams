package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	branding "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

type brandingPolicyCleanup struct{ service *branding.Service }

func (h brandingPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if samePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if isHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h brandingPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	if err != nil && !isHTTP(err, 404) {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !isHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if value == nil || value.Identity == nil {
				return false, lab.ErrUnresolved
			}
			if samePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (brandingPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccMeetingBrandingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccMeetingBrandingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("branding CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.MeetingBrandingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private branding recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"branding-policy": brandingPolicyCleanup{service}}); err != nil {
			t.Error("branding policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("branding recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("branding-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("branding-policy", name)
	if err != nil {
		t.Fatal("branding create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.Create(ctx, branding.CreateRequest{Identity: name, Settings: branding.Settings{Fields: branding.Fields{EnableMeetingOptionsThemeOverride: policy.Set(false), RequireBackgroundEffect: policy.Set(false)}}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("branding policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "meeting_branding", "create", response, replacements)
	var value *branding.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.EnableMeetingOptionsThemeOverride != nil && !*read.EnableMeetingOptionsThemeOverride && read.RequireBackgroundEffect != nil && !*read.RequireBackgroundEffect
		if matched {
			value = read
			writeResponseFixture(t, "meeting_branding", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("branding non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("branding creation receipt not journaled")
	}
	compare := func(value *branding.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("branding tenant binding unavailable")
		}
		actual := brandingOracleForIdentity(t, ctx, tenant, identity)
		if !matchesBrandingCRUDProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell branding policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, branding.UpdateRequest{Fields: branding.Fields{EnableMeetingOptionsThemeOverride: policy.Set(true)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("branding sparse update failed")
	}
	writeResponseFixture(t, "meeting_branding", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.EnableMeetingOptionsThemeOverride != nil && *read.EnableMeetingOptionsThemeOverride && read.RequireBackgroundEffect != nil && !*read.RequireBackgroundEffect
		if matched {
			value = read
			writeResponseFixture(t, "meeting_branding", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("branding update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("branding policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned branding policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "meeting_branding", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "meeting_branding", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (brandingPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("branding deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("branding deleted policy did not return typed404")
	}
	writeResponseFixture(t, "meeting_branding", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("branding policy create/read/update/list/delete and omitted-field preservation verified")
}

func matchesBrandingCRUDProjection(value *branding.Policy, settings map[string]json.RawMessage) bool {
	for _, key := range brandingScalars {
		want := value.RawFields[key]
		if len(want) == 0 {
			want = json.RawMessage("null")
		}
		var left, right any
		if json.Unmarshal(want, &left) != nil || json.Unmarshal(settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
			return false
		}
	}
	return true
}
