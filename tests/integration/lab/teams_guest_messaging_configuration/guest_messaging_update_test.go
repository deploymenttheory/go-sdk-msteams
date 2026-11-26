package teamsguestmessagingconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	guestmessaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccGuestMessagingConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid guest messaging configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("guest messaging configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid guest messaging configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsGuestMessagingConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private guest messaging recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"guest-messaging-configuration": guestMessagingConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("guest messaging configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("guest messaging recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || !guestMessagingComplete(before) || alternateRating(*before.GiphyRatingType) == "" {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("supported original guest messaging setting unavailable; update refused")
	}
	compare := func(v *guestmessaging.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(guestMessagingProjection(v), guestMessagingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell guest messaging configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	original := guestMessagingSnapshot{AllowMemes: before.AllowMemes, GiphyRatingType: before.GiphyRatingType}
	snapshot, err := json.Marshal(original)
	if err != nil {
		t.Fatal("guest messaging original settings could not be journaled")
	}
	entry, err := journal.PlanRestore("guest-messaging-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("guest messaging restoration intent not journaled")
	}
	wantMemes, wantRating := !*before.AllowMemes, alternateRating(*before.GiphyRatingType)
	want := guestMessagingSnapshot{AllowMemes: &wantMemes, GiphyRatingType: &wantRating}
	response, err = s.UpdateGlobal(ctx, guestmessaging.UpdateRequest{AllowMemes: want.AllowMemes, GiphyRatingType: want.GiphyRatingType})
	labkit.WriteResponseFixture(t, "teams_guest_messaging_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("guest messaging configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("guest messaging update receipt not journaled")
	}
	var updated *guestmessaging.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && want.matches(v)
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "teams_guest_messaging_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("guest messaging configuration update did not persist")
	}
	if !labkit.SameProjection(unchangedSettings(before), unchangedSettings(updated)) {
		t.Fatal("guest messaging update changed omitted settings")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("guest messaging original settings restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || !original.matches(restored) || !labkit.SameProjection(guestMessagingProjection(before), guestMessagingProjection(restored)) {
		t.Fatal("guest messaging original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_guest_messaging_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitGuestMessagingRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":"Moderate"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":null,"giphy_rating_type":"Moderate"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":""}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":"false","giphy_rating_type":"Moderate"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":"Moderate","extra":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_memes":true,"giphy_rating_type":"Moderate"} {}`)},
	} {
		if !errors.Is((guestMessagingConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}

// alternateRating picks a different rating from the three names the 8.1.0
// help documents (Moderate, Strict, NoRestriction). An undocumented original
// returns "" so the update is refused rather than guessing a value.
func alternateRating(original string) string {
	switch {
	case strings.EqualFold(original, "Moderate"):
		return "Strict"
	case strings.EqualFold(original, "Strict"), strings.EqualFold(original, "NoRestriction"):
		return "Moderate"
	}
	return ""
}

// unchangedSettings projects the eight getter settings the update leaves unbound.
func unchangedSettings(v *guestmessaging.Configuration) map[string]any {
	p := guestMessagingProjection(v)
	delete(p, "AllowMemes")
	delete(p, "GiphyRatingType")
	return p
}
