package teamsmeetingtemplateconfigurationacc

import (
	"context"
	"strings"
	"testing"

	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMeetingTemplateConfigurationReadOnly(t *testing.T) {
	cfg := labkit.MeetingTemplateReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_template_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting template configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("template configuration selected identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_template_configuration", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting template configuration list failed")
	}
	var selected *configuration.Configuration
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous template configuration Global")
			}
			selected = item
		}
	}
	if selected == nil || !labkit.SameProjection(labkit.MeetingTemplateConfigurationProjection(value), labkit.MeetingTemplateConfigurationProjection(selected)) {
		t.Fatal("template configuration read/list getter projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("template tenant binding unavailable")
		}
		actual := labkit.MeetingTemplatePolicyOracle(t, ctx, "configuration", tenant)
		if !labkit.SameProjection(labkit.MeetingTemplateConfigurationProjection(value), actual.Settings) {
			t.Fatal("Go and PowerShell template configuration getters differ")
		}
	}
	t.Log("meeting template configuration selected reads verified")
}
