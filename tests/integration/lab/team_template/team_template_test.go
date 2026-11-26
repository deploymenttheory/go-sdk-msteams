package teamtemplateacc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func comparePowerShellTemplate(t *testing.T, ctx context.Context, reference, name, short, description string, channels []string, definition template.Definition) {
	t.Helper()
	if os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") != "1" {
		return
	}
	module := os.Getenv("MSTEAMS_LAB_MODULE_PATH")
	if module == "" || !ownedTemplateName(name) || !ownedTemplateReference(reference) {
		t.Fatal("explicit pinned module and owned template selectors are required")
	}
	command := labkit.OracleCommand(ctx, "read-team-template.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"),
		"MSTEAMS_LAB_TEMPLATE_FIXTURE="+name, "MSTEAMS_LAB_TEMPLATE_ID="+reference)
	var output labkit.OracleBuffer
	command.Stdout = &output
	command.Stderr = io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell template read failed; no output retained")
	}
	var observed struct {
		Success      bool            `json:"success"`
		DisplayName  string          `json:"display_name"`
		Short        string          `json:"short_description"`
		Description  string          `json:"description"`
		ChannelNames []string        `json:"channel_names"`
		Settings     map[string]bool `json:"settings"`
	}
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success {
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
	if !reflect.DeepEqual(observed.Settings, templateSettings(definition)) {
		t.Fatal("PowerShell and Go template boolean settings differ")
	}
	t.Log("PowerShell 8.1.0 / Go template read parity: passed")
}

func templateReadBack(t *testing.T, ctx context.Context, service *template.Service, reference, name, short, description, fixtureLabel string, replacements map[string]string) *template.Template {
	t.Helper()
	value, response, err := service.GetByID(ctx, reference)
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
	labkit.WriteResponseFixture(t, "team_template", fixtureLabel, response, replacements)
	comparePowerShellTemplate(t, ctx, reference, name, short, description, names, value.Definition)
	return value
}

// templateSettings follows the pinned ITeam*Settings properties. Missing settings
// objects normalize to the generated boolean defaults; the true phase proves
// every selected field can carry a non-default value before false is checked.
func templateSettings(d template.Definition) map[string]bool {
	if d.MemberSettings == nil {
		d.MemberSettings = &template.MemberSettings{}
	}
	if d.GuestSettings == nil {
		d.GuestSettings = &template.GuestSettings{}
	}
	if d.FunSettings == nil {
		d.FunSettings = &template.FunSettings{}
	}
	if d.MessagingSettings == nil {
		d.MessagingSettings = &template.MessagingSettings{}
	}
	if d.DiscoverySettings == nil {
		d.DiscoverySettings = &template.DiscoverySettings{}
	}
	return map[string]bool{
		"member_add_remove_apps":        d.MemberSettings.AllowAddRemoveApps,
		"member_private_channels":       d.MemberSettings.AllowCreatePrivateChannels,
		"member_create_update_channels": d.MemberSettings.AllowCreateUpdateChannels,
		"member_connectors":             d.MemberSettings.AllowCreateUpdateRemoveConnectors,
		"member_tabs":                   d.MemberSettings.AllowCreateUpdateRemoveTabs,
		"member_delete_channels":        d.MemberSettings.AllowDeleteChannels,
		"member_upload_custom_app":      d.MemberSettings.UploadCustomApp,
		"guest_create_update_channels":  d.GuestSettings.AllowCreateUpdateChannels,
		"guest_delete_channels":         d.GuestSettings.AllowDeleteChannels,
		"fun_custom_memes":              d.FunSettings.AllowCustomMemes,
		"fun_giphy":                     d.FunSettings.AllowGiphy,
		"fun_stickers_memes":            d.FunSettings.AllowStickersAndMemes,
		"messaging_channel_mentions":    d.MessagingSettings.AllowChannelMentions,
		"messaging_owner_delete":        d.MessagingSettings.AllowOwnerDeleteMessages,
		"messaging_team_mentions":       d.MessagingSettings.AllowTeamMentions,
		"messaging_user_delete":         d.MessagingSettings.AllowUserDeleteMessages,
		"messaging_user_edit":           d.MessagingSettings.AllowUserEditMessages,
		"discovery_search":              d.DiscoverySettings.ShowInTeamsSearchAndSuggestions,
	}
}

func setTemplateSettings(d *template.Definition, enabled bool) {
	d.MemberSettings = &template.MemberSettings{AllowAddRemoveApps: enabled, AllowCreatePrivateChannels: enabled,
		AllowCreateUpdateChannels: enabled, AllowCreateUpdateRemoveConnectors: enabled, AllowCreateUpdateRemoveTabs: enabled,
		AllowDeleteChannels: enabled, UploadCustomApp: enabled}
	d.GuestSettings = &template.GuestSettings{AllowCreateUpdateChannels: enabled, AllowDeleteChannels: enabled}
	d.FunSettings = &template.FunSettings{AllowCustomMemes: enabled, AllowGiphy: enabled, AllowStickersAndMemes: enabled}
	d.MessagingSettings = &template.MessagingSettings{AllowChannelMentions: enabled, AllowOwnerDeleteMessages: enabled,
		AllowTeamMentions: enabled, AllowUserDeleteMessages: enabled, AllowUserEditMessages: enabled}
	d.DiscoverySettings = &template.DiscoverySettings{ShowInTeamsSearchAndSuggestions: enabled}
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
	sdk, binding, _ := labkit.SDKClient(t)
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
	prefix := "expanded_"
	if body {
		prefix = "json_body_"
	}
	replacements := map[string]string{name: "fixture-template"}
	_, response, err := service.Create(ctx, identity.Locale, definition)
	labkit.WriteResponseFixture(t, "team_template", prefix+"create", response, replacements)
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
	initial := templateReadBack(t, ctx, service, reference, name, short, description, prefix+"create_readback", replacements)
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
	_, response, err = service.UpdateByID(ctx, identity.ODataID, update)
	labkit.WriteResponseFixture(t, "team_template", prefix+"update", response, replacements)
	if err != nil {
		t.Fatal("live template update failed")
	}
	templateReadBack(t, ctx, service, reference, name, updatedShort, updatedDescription, prefix+"update_readback", replacements)
	t.Log("template update/read-back: passed")

	// Explicit empty must remain distinct from an omitted pointer in both
	// expanded input and the pinned JSON Body conversion.
	empty := ""
	clearDescription, err := template.NewDefinition(name, updatedShort)
	if err != nil {
		t.Fatal("template empty-description input construction failed")
	}
	clearDescription.Description, clearDescription.Channels = &empty, initial.Channels
	if body {
		clearDescription = labBodyDefinition(t, clearDescription)
	}
	_, response, err = service.UpdateByID(ctx, identity.ODataID, clearDescription)
	labkit.WriteResponseFixture(t, "team_template", prefix+"empty_description_update", response, replacements)
	if err != nil {
		t.Fatal("live template empty-description update failed")
	}
	templateReadBack(t, ctx, service, reference, name, updatedShort, empty, prefix+"empty_description_readback", replacements)
	t.Log("template empty description/read-back: passed")
	for _, enabled := range []bool{true, false} {
		label := "settings_false"
		if enabled {
			label = "settings_true"
		}
		settings, err := template.NewDefinition(name, updatedShort)
		if err != nil {
			t.Fatal("template settings input construction failed")
		}
		settings.Description, settings.Channels = &empty, initial.Channels
		setTemplateSettings(&settings, enabled)
		if body {
			settings = labBodyDefinition(t, settings)
		}
		_, response, err = service.UpdateByID(ctx, identity.ODataID, settings)
		labkit.WriteResponseFixture(t, "team_template", prefix+label+"_update", response, replacements)
		if err != nil {
			t.Fatal("live template boolean settings update failed")
		}
		observed := templateReadBack(t, ctx, service, reference, name, updatedShort, empty, prefix+label+"_readback", replacements)
		for _, value := range templateSettings(observed.Definition) {
			if value != enabled {
				t.Fatal("template boolean settings read-back lost explicit input")
			}
		}
		if enabled {
			t.Log("template non-default boolean settings/read-back: passed")
		} else {
			t.Log("template explicit false settings/read-back: passed")
		}
	}
	listed, response, err := service.List(ctx, nil)
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
	labkit.WriteResponseFixture(t, "team_template", prefix+"owned_list", response, replacements)
	t.Log("template list: passed")
	response, err = service.DeleteByID(ctx, reference)
	labkit.WriteResponseFixture(t, "team_template", prefix+"delete", response, replacements)
	if err != nil {
		t.Fatal("live template delete failed")
	}
	if err := cleanup.absent(ctx, lab.Entry{Fixture: name, ResourceID: reference}); err != nil {
		t.Fatal("template selected/list absence not verified")
	}
	listed, response, err = service.List(ctx, nil)
	if err != nil || listed == nil || listed.IsIncomplete {
		t.Fatal("complete post-delete template list unavailable")
	}
	for _, value := range listed.Items {
		if value.Name != nil && *value.Name == name {
			t.Fatal("deleted template remains in independent list")
		}
	}
	labkit.WriteResponseFixture(t, "team_template", prefix+"list_deleted", response, replacements)
	_, response, err = service.GetByID(ctx, reference)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted template absence was not verified")
	}
	var failure *template.Error
	if !errors.As(err, &failure) || failure.Envelope == nil || len(failure.Envelope.Errors) != 1 || failure.Envelope.Errors[0].Code == nil {
		t.Fatal("actual template errorCode was not exposed by the typed error")
	}
	codeRetained := false
	for _, code := range failure.APIError.Codes {
		if code == *failure.Envelope.Errors[0].Code {
			codeRetained = true
		}
	}
	if !codeRetained {
		t.Fatal("actual template string errorCode was lost by shared normalization")
	}
	labkit.WriteResponseFixture(t, "team_template", prefix+"get_deleted", response, replacements)
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
	sdk, binding, _ := labkit.SDKClient(t)
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
