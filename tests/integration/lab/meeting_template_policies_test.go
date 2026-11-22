package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_configuration"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
)

func meetingTemplateConfigurationProjection(value *configuration.Configuration) map[string]any {
	templates := []map[string]any{}
	for _, item := range value.TeamsMeetingTemplates {
		if item == nil {
			templates = append(templates, nil)
			continue
		}
		options := []map[string]any{}
		for _, option := range item.TeamsMeetingOptions {
			if option == nil {
				options = append(options, nil)
				continue
			}
			options = append(options, map[string]any{"Identity": option.Identity, "Name": option.Name, "Value": option.Value, "IsLocked": option.IsLocked, "IsHidden": option.IsHidden})
		}
		templates = append(templates, map[string]any{"Identity": item.Identity, "Name": item.Name, "Description": item.Description, "Category": item.Category, "TeamsMeetingOptions": options})
	}
	return map[string]any{"Description": value.Description, "TeamsMeetingTemplates": templates}
}
func meetingTemplatePermissionProjection(value *permission.Policy) map[string]any {
	templates := []map[string]any{}
	for _, item := range value.HiddenMeetingTemplates {
		if item == nil {
			templates = append(templates, nil)
			continue
		}
		templates = append(templates, map[string]any{"Identity": item.Identity, "Id": item.ID})
	}
	return map[string]any{"Description": value.Description, "DefaultMeetingTemplateId": value.DefaultMeetingTemplateId, "HiddenMeetingTemplates": templates}
}
func sameTemplateProjection(left, right any) bool {
	a, e := json.Marshal(left)
	if e != nil {
		return false
	}
	b, e := json.Marshal(right)
	if e != nil {
		return false
	}
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
func meetingTemplateReadEnabled(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid template policy lab configuration")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting template policy reads require explicit opt-in")
	} else if err != nil {
		t.Fatal("invalid template policy lab enablement")
	}
	return cfg
}

type meetingTemplatePolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var meetingTemplatePolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-meeting-template-policy$`)

func decodeMeetingTemplatePolicyOracle(raw []byte, kind, tenant string) (meetingTemplatePolicyObservation, error) {
	return decodeMeetingTemplatePolicyOracleForIdentity(raw, kind, tenant, "Global")
}
func decodeMeetingTemplatePolicyOracleForIdentity(raw []byte, kind, tenant, identity string) (meetingTemplatePolicyObservation, error) {
	var value meetingTemplatePolicyObservation
	if !strings.EqualFold(identity, "Global") && (kind != "permission" || !meetingTemplatePolicyOwnedIdentity.MatchString(identity)) {
		return value, errTeamOracle
	}
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || (kind != "configuration" && kind != "permission" && kind != "firstparty") {
		return value, errTeamOracle
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(value.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(value.Identity, identity)
	}
	if !value.Success || !strings.EqualFold(value.Tenant, tenant) || !identityMatches {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	names := []string{"Description", "TeamsMeetingTemplates"}
	if kind == "permission" {
		names = []string{"Description", "DefaultMeetingTemplateId", "HiddenMeetingTemplates"}
	}
	if len(value.Settings) != len(names) {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	for _, name := range names {
		if _, ok := value.Settings[name]; !ok {
			return meetingTemplatePolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(value.Settings)
	if err != nil {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	if kind != "permission" {
		var model configuration.Configuration
		err = json.Unmarshal(settings, &model)
	} else {
		var model permission.Policy
		err = json.Unmarshal(settings, &model)
	}
	if err != nil {
		return meetingTemplatePolicyObservation{}, errTeamOracle
	}
	return value, nil
}
func meetingTemplatePolicyOracle(t *testing.T, ctx context.Context, kind, tenant string) meetingTemplatePolicyObservation {
	return meetingTemplatePolicyOracleForIdentity(t, ctx, kind, tenant, "Global")
}
func meetingTemplatePolicyOracleForIdentity(t *testing.T, ctx context.Context, kind, tenant, identity string) meetingTemplatePolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity, "kind": kind})
	if err != nil {
		t.Fatal("invalid template policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-meeting-template-policies.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned template policy read oracle process failed")
	}
	value, err := decodeMeetingTemplatePolicyOracleForIdentity(output.data, kind, tenant, identity)
	if err != nil {
		t.Fatal("pinned template policy projection unavailable")
	}
	return value
}
func TestAccMeetingTemplateConfigurationReadOnly(t *testing.T) {
	cfg := meetingTemplateReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_template_configuration", "get_global", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("meeting template configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("template configuration selected identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_template_configuration", "list", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
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
	if selected == nil || !sameTemplateProjection(meetingTemplateConfigurationProjection(value), meetingTemplateConfigurationProjection(selected)) {
		t.Fatal("template configuration read/list getter projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("template tenant binding unavailable")
		}
		actual := meetingTemplatePolicyOracle(t, ctx, "configuration", tenant)
		if !sameTemplateProjection(meetingTemplateConfigurationProjection(value), actual.Settings) {
			t.Fatal("Go and PowerShell template configuration getters differ")
		}
	}
	t.Log("meeting template configuration selected reads verified")
}
func TestAccMeetingTemplatePermissionReadOnly(t *testing.T) {
	cfg := meetingTemplateReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies
	value, response, err := service.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_template_permission_policy", "get_global", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("meeting template permission Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("template permission selected identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_meeting_template_permission_policy", "list", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
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
	if selected == nil || !sameTemplateProjection(meetingTemplatePermissionProjection(value), meetingTemplatePermissionProjection(selected)) {
		t.Fatal("template permission read/list getter projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("template tenant binding unavailable")
		}
		actual := meetingTemplatePolicyOracle(t, ctx, "permission", tenant)
		if !sameTemplateProjection(meetingTemplatePermissionProjection(value), actual.Settings) {
			t.Fatal("Go and PowerShell template permission getters differ")
		}
	}
	t.Log("meeting template permission selected reads verified")
}

func TestUnitMeetingTemplateOraclesUseAcceptanceFields(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	for _, kind := range []string{"configuration", "permission"} {
		service := "teams_meeting_template_configuration"
		if kind == "permission" {
			service = "teams_meeting_template_permission_policy"
		}
		directory, err := os.OpenRoot("../../../msteams/teams_admin_api/" + service + "/testdata")
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
			projection = meetingTemplateConfigurationProjection(&model)
		} else {
			var model permission.Policy
			if json.Unmarshal(raw, &model) != nil {
				t.Fatal("invalid acceptance response")
			}
			projection = meetingTemplatePermissionProjection(&model)
		}
		report := map[string]any{"success": true, "tenant_id": tenant, "identity": "Global", "settings": projection}
		valid, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeMeetingTemplatePolicyOracle(valid, kind, tenant); err != nil {
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
			if _, err := decodeMeetingTemplatePolicyOracle(bad, kind, tenant); err == nil {
				t.Fatal("invalid template oracle projection accepted")
			}
		}
		if _, err := decodeMeetingTemplatePolicyOracleForIdentity(valid, kind, tenant, "other-policy"); err == nil {
			t.Fatal("foreign template oracle selector accepted")
		}
	}
}

// Read actual existing first-party IDs for the unassigned permission-policy test.
// PowerShell is a test fixture lookup here; SDK methods never execute it.
func existingMeetingTemplateIDs(t *testing.T, ctx context.Context, tenant string) []string {
	t.Helper()
	actual := meetingTemplatePolicyOracle(t, ctx, "firstparty", tenant)
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
