package teamsmeetingconfigurationacc

import (
	"context"
	"strings"
	"testing"

	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMeetingConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("meeting configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_configuration", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting configuration Global list request failed")
	}
	var selected *meeting.Configuration
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous meeting configuration Global")
			}
			selected = item
		}
	}
	if selected == nil || !labkit.SameProjection(meetingProjection(value), meetingProjection(selected)) {
		t.Fatal("meeting configuration read/list getters differ")
	}
	if !labkit.SameConfigurationFields(value.RawFields, selected.RawFields) {
		t.Log(labkit.ResponseComparisonFailure(value.RawFields, selected.RawFields))
		t.Fatal("meeting configuration read/list snapshots differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		if !labkit.SameProjection(meetingProjection(value), meetingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell meeting configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	t.Log("meeting configuration read/list and pinned getter comparison verified")
}
