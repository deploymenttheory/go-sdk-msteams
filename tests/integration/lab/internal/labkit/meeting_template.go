package labkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_configuration"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
)

func MeetingTemplateConfigurationProjection(value *configuration.Configuration) map[string]any {
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

func MeetingTemplatePermissionProjection(value *permission.Policy) map[string]any {
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

func MeetingTemplateReadEnabled(t *testing.T) lab.Config {
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

type MeetingTemplatePolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

// Oracle output is canonical PowerShell JSON. These strict producer types are
// intentionally separate from the source-compatible SDK received-JSON coercion.
type meetingTemplateOracleOption struct {
	Identity, Name, Value *string
	IsLocked, IsHidden    *bool
}
type meetingTemplateOracleTemplate struct {
	Identity, Name, Description, Category *string
	TeamsMeetingOptions                   []*meetingTemplateOracleOption
}
type meetingTemplateOracleConfiguration struct {
	Description           *string
	TeamsMeetingTemplates []*meetingTemplateOracleTemplate
}

var MeetingTemplatePolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-meeting-template-policy$`)

func DecodeMeetingTemplatePolicyOracle(raw []byte, kind, tenant string) (MeetingTemplatePolicyObservation, error) {
	return DecodeMeetingTemplatePolicyOracleForIdentity(raw, kind, tenant, "Global")
}

func DecodeMeetingTemplatePolicyOracleForIdentity(raw []byte, kind, tenant, identity string) (MeetingTemplatePolicyObservation, error) {
	var value MeetingTemplatePolicyObservation
	if !strings.EqualFold(identity, "Global") && (kind != "permission" || !MeetingTemplatePolicyOwnedIdentity.MatchString(identity)) {
		return value, ErrOracle
	}
	if len(raw) > MaxOracleBytes || !DirectoryID.MatchString(tenant) || (kind != "configuration" && kind != "permission" && kind != "firstparty") {
		return value, ErrOracle
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	identityMatches := strings.EqualFold(value.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = SamePolicyName(value.Identity, identity)
	}
	if !value.Success || !strings.EqualFold(value.Tenant, tenant) || !identityMatches {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	names := []string{"Description", "TeamsMeetingTemplates"}
	if kind == "permission" {
		names = []string{"Description", "DefaultMeetingTemplateId", "HiddenMeetingTemplates"}
	}
	if len(value.Settings) != len(names) {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	for _, name := range names {
		if _, ok := value.Settings[name]; !ok {
			return MeetingTemplatePolicyObservation{}, ErrOracle
		}
	}
	settings, err := json.Marshal(value.Settings)
	if err != nil {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	if kind != "permission" {
		var canonical meetingTemplateOracleConfiguration
		if json.Unmarshal(settings, &canonical) != nil {
			return MeetingTemplatePolicyObservation{}, ErrOracle
		}
		var model configuration.Configuration
		err = json.Unmarshal(settings, &model)
	} else {
		var model permission.Policy
		err = json.Unmarshal(settings, &model)
	}
	if err != nil {
		return MeetingTemplatePolicyObservation{}, ErrOracle
	}
	return value, nil
}

func MeetingTemplatePolicyOracle(t *testing.T, ctx context.Context, kind, tenant string) MeetingTemplatePolicyObservation {
	return MeetingTemplatePolicyOracleForIdentity(t, ctx, kind, tenant, "Global")
}

func MeetingTemplatePolicyOracleForIdentity(t *testing.T, ctx context.Context, kind, tenant, identity string) MeetingTemplatePolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity, "kind": kind})
	if err != nil {
		t.Fatal("invalid template policy oracle selector")
	}
	defer clear(input)
	cmd := OracleCommand(ctx, "read-meeting-template-policies.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned template policy read oracle process failed")
	}
	value, err := DecodeMeetingTemplatePolicyOracleForIdentity(output.Data, kind, tenant, identity)
	if err != nil {
		t.Fatal("pinned template policy projection unavailable")
	}
	return value
}
