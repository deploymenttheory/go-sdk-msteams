package lab_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

type audioPolicyCleanup struct{ service *audio.Service }

func (h audioPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if labkit.SamePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if labkit.IsHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h audioPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if labkit.SamePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (audioPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccAudioConferencingPolicyCRUD changes only an owned unassigned policy.
// It makes no calls, assigns no users and references no telephone numbers.
func TestAccAudioConferencingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("audio policy CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private audio policy recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"audio-policy": audioPolicyCleanup{service}}); err != nil {
			t.Error("audio policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("audio policy recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("audio-policy")
	if err != nil {
		t.Fatal("audio policy fixture name unavailable")
	}
	entry, err := journal.PlanCreate("audio-policy", name)
	if err != nil {
		t.Fatal("audio policy creation intent unavailable")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	response, err := service.CreateByIdentity(ctx, name, audio.Settings{AllowTollFreeDialin: new(false), MeetingInvitePhoneNumbers: audio.ReplacePhoneNumbers()})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("audio policy create failed")
	}
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "create", response, replacements)
	var value *audio.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && labkit.SamePolicyName(read.Identity, name) && read.AllowTollFreeDialin != nil && !*read.AllowTollFreeDialin && read.MeetingInvitePhoneNumbers != nil && len(read.MeetingInvitePhoneNumbers) == 0
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("audio policy non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("audio policy receipt not journaled")
	}
	compare := func(value *audio.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("audio policy tenant binding unavailable")
		}
		actual := audioConferencingPolicyOracleForIdentity(t, ctx, tenant, identity)
		for _, key := range audioConferencingPolicyFields {
			var left, right any
			if decodeAudioPolicyValue(value.RawFields[key], &left) != nil || decodeAudioPolicyValue(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("SDK and PowerShell audio policy projections differ")
			}
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	response, err = service.UpdateByIdentity(ctx, identity, audio.Settings{AllowTollFreeDialin: new(true)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("audio policy sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.AllowTollFreeDialin != nil && *read.AllowTollFreeDialin && read.MeetingInvitePhoneNumbers != nil && len(read.MeetingInvitePhoneNumbers) == 0
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("audio policy changed or omitted setting did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, response, err := service.List(ctx)
	if err != nil {
		t.Fatal("audio policy list failed")
	}
	count := 0
	for _, item := range values {
		if labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned audio policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "owned_list", response, replacements)
	t.Log("policy list: passed")
	response, err = service.DeleteByIdentity(ctx, identity)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned audio policy delete failed")
	}
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "delete", response, replacements)
	if err := (audioPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("audio policy deletion or independent absence not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted audio policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
}
