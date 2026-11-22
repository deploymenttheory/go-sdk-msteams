package teamsmessagingpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var messagingPolicyScalars = []string{"Description", "AllowUrlPreviews", "AllowOwnerDeleteMessage", "AllowUserEditMessage", "AllowUserDeleteMessage", "UsersCanDeleteBotMessages", "AllowUserDeleteChat", "AllowUserChat", "AllowRemoveUser", "AllowGiphy", "GiphyRatingType", "AllowGiphyDisplay", "AllowPasteInternetImage", "AllowMemes", "AllowImmersiveReader", "AllowStickers", "AllowUserTranslation", "ReadReceiptsEnabledType", "AllowPriorityMessages", "AllowSmartReply", "AllowSmartCompose", "ChannelsInChatListEnabledType", "AudioMessageEnabledType", "ChatPermissionRole", "AllowFullChatPermissionUserToDeleteAnyMessage", "AllowFluidCollaborate", "AllowVideoMessages", "AllowCommunicationComplianceEndUserReporting", "AllowChatWithGroup", "AllowSecurityEndUserReporting", "InOrganizationChatControl", "AllowGroupChatJoinLinks", "CreateCustomEmojis", "UseB2BInvitesToAddExternalUsers", "AllowProactiveSummaries", "DeleteCustomEmojis", "AutoShareFilesInExternalChats", "DesignerForBackgroundsAndImages", "AllowCustomGroupChatAvatars"}

type messagingPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var messagingPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-messaging-policy$`)

func decodeMessagingPolicyOracle(raw []byte, tenant string) (messagingPolicyObservation, error) {
	return decodeMessagingPolicyOracleForIdentity(raw, tenant, "Global")
}

func decodeMessagingPolicyOracleForIdentity(raw []byte, tenant, identity string) (messagingPolicyObservation, error) {
	var v messagingPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !messagingPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return messagingPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(messagingPolicyScalars) {
		return messagingPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return messagingPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range messagingPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return messagingPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return messagingPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func messagingPolicyOracle(t *testing.T, ctx context.Context, tenant string) messagingPolicyObservation {
	t.Helper()
	return messagingPolicyOracleForIdentity(t, ctx, tenant, "Global")
}

func messagingPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) messagingPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid messaging policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-messaging-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned messaging policy read failed; no output retained")
	}
	value, err := decodeMessagingPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned messaging policy projection unavailable")
	}
	return value
}

// TestAccMessagingPolicyReadOnly performs configuration reads only.
// Creates, updates, deletes, grants and effective policy behavior need separate acceptance.
func TestAccMessagingPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("messaging policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsMessagingPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsMessagingPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("messaging policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsMessagingPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_messaging_policy", "list", listResponse, nil)
	selected, selectionErr := globalMessagingPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(messagingPolicyProjection(value), messagingPolicyProjection(selected)) {
		t.Fatal("messaging policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := messagingPolicyOracle(t, ctx, tenant)
		if !matchesMessagingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned messaging policy getter projections differ")
		}

	}
	t.Log("selected messaging policy reads verified; persistence and effective policy behavior remain separately qualified")
}

func TestUnitMessagingPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := map[string]any{}
	for _, name := range messagingPolicyScalars {
		settings[name] = nil
	}
	settings["Description"] = ""
	settings["AllowGiphy"] = false
	settings["AllowMemes"] = true
	raw, err := json.Marshal(messagingPolicyObservation{Success: true, Tenant: tenant, Identity: "Global", Settings: func() map[string]json.RawMessage {
		encoded, _ := json.Marshal(settings)
		var value map[string]json.RawMessage
		_ = json.Unmarshal(encoded, &value)
		return value
	}()})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	for _, input := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"Description":""`, `"Description":false`, 1), strings.Replace(valid, `"AllowMemes":true`, `"AllowMemes":"true"`, 1), strings.Replace(valid, `,"AllowMemes":true`, "", 1)} {
		if _, e := decodeMessagingPolicyOracle([]byte(input), tenant); e == nil {
			t.Fatal("invalid messaging policy oracle accepted")
		}
	}
	if _, e := decodeMessagingPolicyOracle(raw, tenant); e != nil {
		t.Fatal("complete nullable messaging policy projection rejected")
	}
}

func messagingPolicyProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "AllowUrlPreviews": v.AllowUrlPreviews, "AllowOwnerDeleteMessage": v.AllowOwnerDeleteMessage, "AllowUserEditMessage": v.AllowUserEditMessage, "AllowUserDeleteMessage": v.AllowUserDeleteMessage, "UsersCanDeleteBotMessages": v.UsersCanDeleteBotMessages, "AllowUserDeleteChat": v.AllowUserDeleteChat, "AllowUserChat": v.AllowUserChat, "AllowRemoveUser": v.AllowRemoveUser, "AllowGiphy": v.AllowGiphy, "GiphyRatingType": v.GiphyRatingType, "AllowGiphyDisplay": v.AllowGiphyDisplay, "AllowPasteInternetImage": v.AllowPasteInternetImage, "AllowMemes": v.AllowMemes, "AllowImmersiveReader": v.AllowImmersiveReader, "AllowStickers": v.AllowStickers, "AllowUserTranslation": v.AllowUserTranslation, "ReadReceiptsEnabledType": v.ReadReceiptsEnabledType, "AllowPriorityMessages": v.AllowPriorityMessages, "AllowSmartReply": v.AllowSmartReply, "AllowSmartCompose": v.AllowSmartCompose, "ChannelsInChatListEnabledType": v.ChannelsInChatListEnabledType, "AudioMessageEnabledType": v.AudioMessageEnabledType, "ChatPermissionRole": v.ChatPermissionRole, "AllowFullChatPermissionUserToDeleteAnyMessage": v.AllowFullChatPermissionUserToDeleteAnyMessage, "AllowFluidCollaborate": v.AllowFluidCollaborate, "AllowVideoMessages": v.AllowVideoMessages, "AllowCommunicationComplianceEndUserReporting": v.AllowCommunicationComplianceEndUserReporting, "AllowChatWithGroup": v.AllowChatWithGroup, "AllowSecurityEndUserReporting": v.AllowSecurityEndUserReporting, "InOrganizationChatControl": v.InOrganizationChatControl, "AllowGroupChatJoinLinks": v.AllowGroupChatJoinLinks, "CreateCustomEmojis": v.CreateCustomEmojis, "UseB2BInvitesToAddExternalUsers": v.UseB2BInvitesToAddExternalUsers, "AllowProactiveSummaries": v.AllowProactiveSummaries, "DeleteCustomEmojis": v.DeleteCustomEmojis, "AutoShareFilesInExternalChats": v.AutoShareFilesInExternalChats, "DesignerForBackgroundsAndImages": v.DesignerForBackgroundsAndImages, "AllowCustomGroupChatAvatars": v.AllowCustomGroupChatAvatars})
	return raw
}

func globalMessagingPolicy(values []*resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, "Global") {
			if found != nil {
				return nil, labkit.ErrOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
	}
	return found, nil
}

func TestUnitMessagingPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalMessagingPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalMessagingPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesMessagingPolicyGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(messagingPolicyProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitMessagingPolicyOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"description":"fixture"}`, `{"Description":"earlier","description":"fixture"}`, `{"Description":null,"description":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range messagingPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["Description"] = json.RawMessage(`"fixture"`)
		if !matchesMessagingPolicyGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}

func TestUnitMessagingPolicyOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-messaging-policy"
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_messaging_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range messagingPolicyScalars {
		projection[field] = settings[field]
	}
	for _, tc := range []struct {
		name, selected, returned string
		valid                    bool
	}{
		{"global", "Global", "global", true},
		{"owned-bare", name, "Tag:" + name, true},
		{"owned-tagged", "Tag:" + name, name, true},
		{"foreign-selection", "other-policy", "other-policy", false},
		{"wrong-owned-policy", name, strings.Replace(name, "aaaa", "bbbb", 1), false},
		{"tagged-global", "Global", "Tag:Global", false},
		{"wrong-kind", strings.Replace(name, "messaging-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeMessagingPolicyOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
