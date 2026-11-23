package teamsmeetingbroadcastconfigurationacc

import (
	"context"
	"strings"
	"testing"

	broadcast "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func selectGlobal(t *testing.T, values []broadcast.Configuration) *broadcast.Configuration {
	t.Helper()
	var selected *broadcast.Configuration
	for i := range values {
		item := &values[i]
		if item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous meeting broadcast configuration Global")
			}
			selected = item
		}
	}
	return selected
}

func TestAccMeetingBroadcastConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global", broadcast.ReadOptions{})
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || value.AllowSdnProviderForBroadcastMeeting == nil || value.SupportURL == nil {
		t.Fatal("meeting broadcast configuration identity or setting differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global", broadcast.ReadOptions{})
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "list", response, nil)
	selected := selectGlobal(t, values)
	if err != nil || selected == nil || !labkit.SameProjection(meetingBroadcastProjection(value), meetingBroadcastProjection(selected)) {
		t.Fatal("meeting broadcast configuration read/list getters differ")
	}
	if !labkit.SameConfigurationFields(value.RawFields, selected.RawFields) {
		t.Log(labkit.ResponseComparisonFailure(value.RawFields, selected.RawFields))
		t.Fatal("meeting broadcast configuration read/list snapshots differ")
	}
	exposed, response, err := service.GetByIdentity(ctx, "Global", broadcast.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "get_global_sdn", response, nil)
	if err != nil || exposed == nil || exposed.Identity == nil || !strings.EqualFold(*exposed.Identity, "Global") {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast configuration exposed SDN read failed")
	}
	for name, ok := range map[string]bool{"SdnName": exposed.SdnName != nil, "SdnApiToken": exposed.SdnAPIToken != nil, "SdnAttendeeFallbackCount": exposed.SdnAttendeeFallbackCount != nil} {
		if !ok {
			t.Fatal("exposed SDN field missing: " + name)
		}
	}
	if !labkit.SameProjection(meetingBroadcastProjection(value), meetingBroadcastProjection(exposed)) {
		t.Fatal("meeting broadcast configuration default/exposed getters differ")
	}
	exposedList, _, err := service.List(ctx, broadcast.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
	exposedSelected := selectGlobal(t, exposedList)
	if err != nil || exposedSelected == nil || !labkit.SameProjection(meetingBroadcastSDNProjection(exposed), meetingBroadcastSDNProjection(exposedSelected)) {
		t.Fatal("meeting broadcast configuration exposed read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		pinned := meetingBroadcastOracle(t, ctx, tenant)
		if !labkit.SameProjection(meetingBroadcastProjection(value), pinned.Settings) {
			t.Fatal("Go and PowerShell meeting broadcast configuration getters differ")
		}
		if !labkit.SameProjection(meetingBroadcastSDNProjection(exposed), pinned.SDN) {
			t.Fatal("Go and PowerShell exposed SDN meeting broadcast getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	t.Log("meeting broadcast configuration read/list, exposed SDN read and pinned getter comparison verified")
}
