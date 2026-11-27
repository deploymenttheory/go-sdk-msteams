package teamscustombannertextacc

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccBannerTextCRUDByID(t *testing.T)       { runOwnedRecordCRUD(t, true) }
func TestAccBannerTextCRUDByIdentity(t *testing.T) { runOwnedRecordCRUD(t, false) }

// runOwnedRecordCRUD creates an unassigned record. Both route variants use the
// same random, journal-derived UUID and never update a pre-existing record.
func runOwnedRecordCRUD(t *testing.T, byID bool) {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("record CRUD requires mutation opt-in and private recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private record recovery journal unavailable")
	}
	cleanup := ownedRecordCleanup{service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"banner-text": cleanup}); err != nil {
			t.Error("record cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("record recovery journal close failed")
		}
	})
	fixture, err := journal.FixtureName("banner-text")
	if err != nil {
		t.Fatal("record fixture name unavailable")
	}
	id, err := ownedRecordID(fixture)
	if err != nil {
		t.Fatal("owned record UUID unavailable")
	}
	identity := "Global/" + id
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	// Check for a collision before journaling any permission to delete.
	if _, _, err := service.GetByID(ctx, id); !labkit.IsHTTP(err, 404) {
		t.Fatal("owned record preflight did not establish absence")
	}
	entry, err := journal.PlanCreate("banner-text", fixture)
	if err != nil {
		t.Fatal("record create intent not journaled")
	}
	label := "crud_identity"
	if byID {
		label = "crud_id"
	}
	replacements := map[string]string{id: "00000000-0000-4000-8000-000000000031", fixture: "fixture-record"}
	input := resource.CreateRequest{Text: "SDK lab non-default initial description", Description: policy.Set(fixture)}
	create := service.CreateByIdentity
	selector := identity
	if byID {
		create = service.CreateByID
		selector = id
	}
	_, createResponse, responseErr := create(ctx, selector, input)
	labkit.WriteResponseFixture(t, "teams_custom_banner_text", label+"_create", createResponse, replacements)
	// The expected UUID is journaled before read-back; recovery also resolves a
	// lost create response by UUID and the exact private ownership marker.
	if responseErr != nil {
		labkit.LogAccessResponse(t, createResponse, responseErr)
		t.Fatal("owned record create failed; recovery required")
	}
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("record creation receipt not journaled")
	}
	read := func(stage string, updated bool) *resource.Banner {
		var found *resource.Banner
		if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
			v, response, err := service.GetByIdentity(ctx, identity)
			if labkit.IsHTTP(err, 404) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			if v == nil || v.Identity == nil || !strings.EqualFold(*v.Identity, identity) || v.Description == nil || *v.Description != fixture || !matchesOwnedRecord(v, updated) {
				return false, nil
			}
			found = v
			labkit.WriteResponseFixture(t, "teams_custom_banner_text", label+"_"+stage, response, replacements)
			return true, nil
		}); err != nil {
			t.Fatal("record read-back or omitted ownership marker did not converge")
		}
		return found
	}
	compare := func(v *resource.Banner) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		listed, _, err := service.List(ctx)
		if err != nil {
			t.Fatal("banner comparison list failed")
		}
		actual := bannerOracle(t, ctx, tenant, []string{identity}, len(listed))
		var want, got any
		if json.Unmarshal(bannerProjection(v), &want) != nil || json.Unmarshal(actual[identity], &got) != nil || !reflect.DeepEqual(want, got) {
			t.Fatal("SDK and PowerShell record projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	value := read("create_readback", false)
	t.Log("policy create/read-back: passed")
	compare(value)
	update := resource.UpdateRequest{Settings: resource.Settings{Text: policy.Set("SDK lab non-default updated description")}}
	updateCall := service.UpdateByIdentity
	if byID {
		updateCall = service.UpdateByID
	}
	_, updateResponse, err := updateCall(ctx, selector, update)
	if err != nil {
		labkit.LogAccessResponse(t, updateResponse, err)
		t.Fatal("owned record update failed")
	}
	value = read("update_readback", true)
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, response, err := service.List(ctx)
	if err != nil {
		t.Fatal("record list failed")
	}
	count := 0
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, identity) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned record missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_custom_banner_text", label+"_list", response, replacements)
	t.Log("policy list: passed")
	if byID {
		_, response, err = service.DeleteByID(ctx, id, policy.Controls{})
	} else {
		_, response, err = service.DeleteByIdentity(ctx, identity, policy.Controls{})
	}
	if err != nil {
		t.Fatal("owned record delete failed")
	}
	labkit.WriteResponseFixture(t, "teams_custom_banner_text", label+"_delete", response, replacements)
	if err := cleanup.Delete(ctx, lab.Entry{Fixture: fixture, ResourceID: identity}); err != nil {
		t.Fatal("record deletion and independent absence not verified")
	}
	_, response, err = service.GetByID(ctx, id)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted record did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_custom_banner_text", label+"_get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
}

func matchesOwnedRecord(v *resource.Banner, updated bool) bool {
	want := "SDK lab non-default initial description"
	if updated {
		want = "SDK lab non-default updated description"
	}
	return v.Text != nil && *v.Text == want
}
