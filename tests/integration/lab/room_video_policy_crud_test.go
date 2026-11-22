package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	roomVideo "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_room_video_tele_conferencing_policy"
)

type roomVideoPolicyCleanup struct{ service *roomVideo.Service }

func (h roomVideoPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value.Identity == "" {
			return "", lab.ErrUnresolved
		}
		if samePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
		// An empty list match alone cannot establish absence after a lost POST.
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
func (h roomVideoPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
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
			if value.Identity == "" {
				return false, lab.ErrUnresolved
			}
			if samePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (roomVideoPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccRoomVideoTeleConferencingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccRoomVideoTeleConferencingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("room video CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private room video recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"room-video-policy": roomVideoPolicyCleanup{service}}); err != nil {
			t.Error("room video policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("room video recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("room-video-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("room-video-policy", name)
	if err != nil {
		t.Fatal("room video create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	response, err := service.CreateByIdentity(ctx, name, roomVideo.Settings{Description: new(initial), Enabled: new(false)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("room video policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "create", response, replacements)
	var value *roomVideo.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && samePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.Enabled != nil && !*read.Enabled
		if matched {
			value = read
			writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("room video non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("room video creation receipt not journaled")
	}
	compare := func(value *roomVideo.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("room video tenant binding unavailable")
		}
		actual := roomVideoPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesRoomVideoTeleConferencingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell room video policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, roomVideo.UpdateRequest{Description: new(updated)})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("room video sparse update failed")
	}
	writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.Enabled != nil && !*read.Enabled
		if matched {
			value = read
			writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("room video update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("room video policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && samePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned room video policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (roomVideoPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("room video deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("room video deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_room_video_tele_conferencing_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("room video policy create/read/update/list/delete and omitted-field preservation verified")
}
