package teamsmeetingtemplatepermissionpolicyacc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMeetingTemplatePermissionReadOnly(t *testing.T) {
	cfg := labkit.MeetingTemplateReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting template permission Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("template permission selected identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting template permission list failed")
	}
	var selected *permission.Policy
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous template permission Global")
			}
			selected = item
		}
	}
	if selected == nil || !labkit.SameProjection(labkit.MeetingTemplatePermissionProjection(value), labkit.MeetingTemplatePermissionProjection(selected)) {
		t.Fatal("template permission read/list getter projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("template tenant binding unavailable")
		}
		actual := labkit.MeetingTemplatePolicyOracle(t, ctx, "permission", tenant)
		if !labkit.SameProjection(labkit.MeetingTemplatePermissionProjection(value), actual.Settings) {
			t.Fatal("Go and PowerShell template permission getters differ")
		}
	}
	t.Log("meeting template permission selected reads verified")
}

// Read actual existing first-party IDs for the unassigned permission-policy test.
// PowerShell is a test fixture lookup here; SDK methods never execute it.
func existingMeetingTemplateIDs(t *testing.T, ctx context.Context, tenant string) []string {
	t.Helper()
	actual := labkit.MeetingTemplatePolicyOracle(t, ctx, "firstparty", tenant)
	raw, err := json.Marshal(actual.Settings)
	var configuration struct {
		Templates []struct{ Name *string } `json:"TeamsMeetingTemplates"`
	}
	if err != nil || json.Unmarshal(raw, &configuration) != nil {
		t.Fatal("first-party meeting template fixture lookup could not be decoded")
	}
	var ids []string
	seen := map[string]bool{}
	for _, item := range configuration.Templates {
		if item.Name != nil && strings.HasPrefix(*item.Name, "firstparty_") && !seen[*item.Name] {
			ids = append(ids, *item.Name)
			seen[*item.Name] = true
			if len(ids) == 2 {
				return ids
			}
		}
	}
	t.Fatal("two existing first-party meeting templates are required for populated-list acceptance")
	return nil
}
