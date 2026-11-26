package teamsguestmessagingconfigurationacc

import (
	"context"
	"strings"
	"testing"

	guestmessaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccGuestMessagingConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsGuestMessagingConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_guest_messaging_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("guest messaging configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !guestMessagingComplete(value) {
		t.Fatal("guest messaging configuration identity or setting differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_guest_messaging_configuration", "list", response, nil)
	var selected *guestmessaging.Configuration
	for i := range values {
		item := &values[i]
		if item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous guest messaging configuration Global")
			}
			selected = item
		}
	}
	if err != nil || selected == nil || !labkit.SameProjection(guestMessagingProjection(value), guestMessagingProjection(selected)) {
		t.Fatal("guest messaging configuration read/list getters differ")
	}
	if !labkit.SameConfigurationFields(value.RawFields, selected.RawFields) {
		t.Log(labkit.ResponseComparisonFailure(value.RawFields, selected.RawFields))
		t.Fatal("guest messaging configuration read/list snapshots differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		if !labkit.SameProjection(guestMessagingProjection(value), guestMessagingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell guest messaging configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	t.Log("guest messaging configuration read/list and pinned getter comparison verified")
}
