package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	voicemail "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

type voicemailPolicyCleanup struct{ service *voicemail.Service }

func (h voicemailPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h voicemailPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (voicemailPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccOnlineVoicemailPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccOnlineVoicemailPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("voicemail CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private voicemail recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"voicemail-policy": voicemailPolicyCleanup{service}}); err != nil {
			t.Error("voicemail policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("voicemail recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("voicemail-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("voicemail-policy", name)
	if err != nil {
		t.Fatal("voicemail create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.Create(ctx, voicemail.CreateRequest{Identity: name, Settings: voicemail.Settings{Description: policy.Set(initial), EnableTranscription: policy.Set(false)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("voicemail policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "online_voicemail_policy", "create", response, replacements)
	var value *voicemail.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.EnableTranscription != nil && !*read.EnableTranscription
		if matched {
			value = read
			writeResponseFixture(t, "online_voicemail_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("voicemail non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("voicemail creation receipt not journaled")
	}
	compare := func(value *voicemail.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("voicemail tenant binding unavailable")
		}
		actual := voicemailOracleForIdentity(t, ctx, tenant, identity)
		if !matchesVoicemailCRUDProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell voicemail policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, voicemail.UpdateRequest{Settings: voicemail.Settings{Description: policy.Set(updated)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("voicemail sparse update failed")
	}
	writeResponseFixture(t, "online_voicemail_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.EnableTranscription != nil && !*read.EnableTranscription
		if matched {
			value = read
			writeResponseFixture(t, "online_voicemail_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("voicemail update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("voicemail policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned voicemail policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "online_voicemail_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "online_voicemail_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (voicemailPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("voicemail deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("voicemail deleted policy did not return typed404")
	}
	writeResponseFixture(t, "online_voicemail_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("voicemail policy create/read/update/list/delete and omitted-field preservation verified")
}

func matchesVoicemailCRUDProjection(value *voicemail.Policy, settings map[string]json.RawMessage) bool {
	for _, key := range voicemailScalars {
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
