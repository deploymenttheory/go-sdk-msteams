package labkit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_configuration"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
)

func TestUnitMeetingTemplateOraclesUseAcceptanceFields(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	for _, kind := range []string{"configuration", "permission"} {
		service := "teams_meeting_template_configuration"
		if kind == "permission" {
			service = "teams_meeting_template_permission_policy"
		}
		directory, err := os.OpenRoot(Root + "/msteams/teams_admin_api/" + service + "/testdata")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := directory.Close(); err != nil {
				t.Error("template fixture directory close failed")
			}
		})
		raw, err := directory.ReadFile("get_global.json")
		if err != nil {
			t.Fatal(err)
		}
		var projection map[string]any
		if kind == "configuration" {
			var model configuration.Configuration
			if json.Unmarshal(raw, &model) != nil {
				t.Fatal("invalid acceptance response")
			}
			projection = MeetingTemplateConfigurationProjection(&model)
		} else {
			var model permission.Policy
			if json.Unmarshal(raw, &model) != nil {
				t.Fatal("invalid acceptance response")
			}
			projection = MeetingTemplatePermissionProjection(&model)
		}
		report := map[string]any{"success": true, "tenant_id": tenant, "identity": "Global", "settings": projection}
		valid, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeMeetingTemplatePolicyOracle(valid, kind, tenant); err != nil {
			t.Fatal("actual template response projection rejected")
		}
		invalidProjection := make(map[string]any, len(projection))
		for name, value := range projection {
			invalidProjection[name] = value
		}
		invalidProjection["Description"] = false
		invalidType, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": "Global", "settings": invalidProjection})
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range [][]byte{append(append([]byte(nil), valid...), []byte("{}")...), []byte(strings.Replace(string(valid), tenant, "00000000-0000-4000-8000-000000000004", 1)), []byte(strings.Replace(string(valid), `"identity":"Global"`, `"identity":"Tag:Global"`, 1)), invalidType} {
			if _, err := DecodeMeetingTemplatePolicyOracle(bad, kind, tenant); err == nil {
				t.Fatal("invalid template oracle projection accepted")
			}
		}
		if _, err := DecodeMeetingTemplatePolicyOracleForIdentity(valid, kind, tenant, "other-policy"); err == nil {
			t.Fatal("foreign template oracle selector accepted")
		}
	}
}
