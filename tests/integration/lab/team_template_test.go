package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

const templateLocale = "en-US"

type templateCleanup struct{ service *template.Service }

func ownedTemplateReference(reference string) bool {
	parts := strings.Split(reference, "/")
	return len(parts) == 7 && parts[0] == "" && parts[1] == "api" && strings.EqualFold(parts[2], "teamtemplates") &&
		parts[3] == "v1.0" && parts[4] != "" && parts[4] != "." && parts[4] != ".." && parts[5] == "Tenant" && strings.EqualFold(parts[6], templateLocale)
}

func ownedTemplateName(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-template")
}

func isHTTP(err error, status int) bool {
	var apiError *client.APIError
	return errors.As(err, &apiError) && apiError != nil && apiError.StatusCode == status
}

func (h templateCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedTemplateName(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	items, _, err := h.service.List(ctx, nil)
	if err != nil {
		return "", err
	}
	if items.IsIncomplete {
		return "", lab.ErrUnresolved
	}
	count, reference := 0, ""
	for _, item := range items.Items {
		if item.Name == nil || *item.Name != entry.Fixture {
			continue
		}
		count++
		if item.ODataID == nil || !ownedTemplateReference(*item.ODataID) || (item.Scope != nil && !strings.EqualFold(*item.Scope, "Tenant")) {
			return "", lab.ErrUnresolved
		}
		reference = *item.ODataID
	}
	if count > 1 {
		return "", lab.ErrUnresolved
	}
	if count == 0 {
		// A planned creation has no previously verified reference. An empty
		// projection/list alone cannot prove that an accepted/lost creation
		// never existed; retain it for later reconciliation.
		return "", lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, reference)
	if err != nil {
		return "", err
	}
	if value.DisplayName == nil || *value.DisplayName != entry.Fixture {
		return "", lab.ErrUnresolved
	}
	return reference, nil
}

func (h templateCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedTemplateName(entry.Fixture) || !ownedTemplateReference(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if isHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if value.DisplayName == nil || *value.DisplayName != entry.Fixture {
		return lab.ErrUnresolved
	}
	_, err = h.service.DeleteByID(ctx, entry.ResourceID)
	if isHTTP(err, 404) {
		return nil
	}
	return err
}

func (templateCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func comparePowerShellTemplate(t *testing.T, ctx context.Context, reference, name, short, description string, channels []string) {
	t.Helper()
	if os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") != "1" {
		return
	}
	module := os.Getenv("MSTEAMS_LAB_MODULE_PATH")
	if module == "" || !ownedTemplateName(name) || !ownedTemplateReference(reference) {
		t.Fatal("explicit pinned module and owned template selectors are required")
	}
	command := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-team-template.ps1")
	command.Env = []string{}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"),
		"MSTEAMS_LAB_TEMPLATE_FIXTURE="+name, "MSTEAMS_LAB_TEMPLATE_ID="+reference)
	output, err := command.Output()
	defer clear(output)
	if err != nil {
		t.Fatal("pinned PowerShell template read failed; no output retained")
	}
	var observed struct {
		Success      bool     `json:"success"`
		DisplayName  string   `json:"display_name"`
		Short        string   `json:"short_description"`
		Description  string   `json:"description"`
		ChannelNames []string `json:"channel_names"`
	}
	if json.Unmarshal(output, &observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell could not read the owned template")
	}
	if observed.DisplayName != name || observed.Short != short || observed.Description != description || len(observed.ChannelNames) != len(channels) {
		t.Fatal("PowerShell and Go template fields differ")
	}
	for index, channel := range channels {
		if observed.ChannelNames[index] != channel {
			t.Fatal("PowerShell and Go template channel ordering differs")
		}
	}
	t.Log("PowerShell 8.1.0 / Go template read parity: passed")
}

func templateReadBack(t *testing.T, ctx context.Context, service *template.Service, reference, name, short, description string) *template.Template {
	t.Helper()
	value, _, err := service.GetByID(ctx, reference)
	if err != nil {
		t.Fatal("owned template read-back failed")
	}
	if value.DisplayName == nil || *value.DisplayName != name || value.ShortDescription == nil || *value.ShortDescription != short || value.Description == nil || *value.Description != description {
		t.Fatal("template read-back lost non-default create/update fields")
	}
	names := []string{}
	for _, channel := range value.Channels {
		if channel != nil && channel.DisplayName != nil {
			names = append(names, *channel.DisplayName)
		}
	}
	if len(names) != 1 || !strings.EqualFold(names[0], "General") {
		t.Fatal("template General channel normalization was not observed")
	}
	comparePowerShellTemplate(t, ctx, reference, name, short, description, names)
	return value
}

// TestAccTeamTemplateCRUD creates one uniquely named tenant template. It creates
// no team, changes no permission policy, and never edits a Microsoft template.
func TestAccTeamTemplateCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("template CRUD requires explicit read/mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"expanded", "json-body"} {
		t.Run(variant, func(t *testing.T) {
			runTemplateCRUD(t, cfg, variant == "json-body")
		})
	}
}

func labBodyDefinition(t *testing.T, value template.Definition) template.Definition {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal("lab Body input encoding failed")
	}
	defer clear(raw)
	parsed, err := template.ParseDefinition(raw)
	if err != nil {
		t.Fatal("lab Body input parsing failed")
	}
	return *parsed
}

func runTemplateCRUD(t *testing.T, cfg lab.Config, body bool) {
	t.Helper()
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamTemplates
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private template recovery journal unavailable")
	}
	cleanup := templateCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		_, cleanupErr := journal.Recover(ctx, map[string]lab.Handler{"team-template": cleanup})
		if cleanupErr != nil {
			t.Error("template cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private template journal close failed")
		}
	})
	name, err := journal.FixtureName("template")
	if err != nil {
		t.Fatal("template fixture name unavailable")
	}
	entry, err := journal.PlanCreate("team-template", name)
	if err != nil {
		t.Fatal("template creation was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	short, description := "SDK lab non-default initial summary", "SDK lab non-default initial description"
	definition, err := template.NewDefinition(name, short)
	if err != nil {
		t.Fatal("template expanded input construction failed")
	}
	definition.Description = &description
	identity := template.Identity{Locale: templateLocale}
	if body {
		definition = labBodyDefinition(t, definition)
	}
	_, _, err = service.Create(ctx, identity.Locale, definition)
	if err != nil {
		t.Fatal("live template creation failed; recovery journal retained")
	}
	// Resolve through a complete list and verify the detail name even when a
	// receipt has no @odata.id. The SDK never guesses that URL from a bare ID.
	reference, err := cleanup.Resolve(ctx, lab.Entry{Fixture: name})
	if err != nil || reference == "" {
		t.Fatal("created template reference unresolved; private journal retained")
	}
	if err := journal.Applied(entry, reference); err != nil {
		t.Fatal("created template receipt was not journaled")
	}
	identity.ODataID = reference
	initial := templateReadBack(t, ctx, service, reference, name, short, description)
	t.Log("template create/read-back and General channel: passed")
	updatedShort, updatedDescription := "SDK lab non-default updated summary", "SDK lab non-default updated description"
	update, err := template.NewDefinition(name, updatedShort)
	if err != nil {
		t.Fatal("template update input construction failed")
	}
	update.Description, update.Channels = &updatedDescription, initial.Channels
	if body {
		update = labBodyDefinition(t, update)
	}
	if _, _, err := service.UpdateByID(ctx, identity.ODataID, update); err != nil {
		t.Fatal("live template update failed")
	}
	templateReadBack(t, ctx, service, reference, name, updatedShort, updatedDescription)
	t.Log("template update/read-back: passed")
	listed, _, err := service.List(ctx, nil)
	if err != nil || listed.IsIncomplete {
		t.Fatal("complete live template listing unavailable")
	}
	found := 0
	for _, item := range listed.Items {
		if item.ODataID != nil && *item.ODataID == reference && item.Name != nil && *item.Name == name {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned template missing or ambiguous in live list")
	}
	t.Log("template list: passed")
	if _, err := service.DeleteByID(ctx, reference); err != nil {
		t.Fatal("live template delete failed")
	}
	_, _, err = service.GetByID(ctx, reference)
	if !isHTTP(err, 404) {
		t.Fatal("deleted template absence was not verified")
	}
	t.Log("template delete/absence: passed")
}

// TestAccTeamTemplateRecover is an explicit operator-selected recovery path for
// an interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccTeamTemplateRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for template recovery")
	}
	sdk, binding, _ := labSDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("private recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, map[string]lab.Handler{"team-template": templateCleanup{service: sdk.TeamsAdminAPI.TeamTemplates}}); err != nil {
		t.Fatal("template recovery incomplete; private journal retained")
	}
	t.Log("owned template recovery: passed")
}
