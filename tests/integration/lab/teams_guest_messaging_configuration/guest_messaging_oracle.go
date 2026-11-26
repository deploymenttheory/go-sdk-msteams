package teamsguestmessagingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	guestmessaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func guestMessagingProjection(v *guestmessaging.Configuration) map[string]any {
	return map[string]any{
		"AllowUserEditMessage":      v.AllowUserEditMessage,
		"AllowUserDeleteMessage":    v.AllowUserDeleteMessage,
		"UsersCanDeleteBotMessages": v.UsersCanDeleteBotMessages,
		"AllowUserDeleteChat":       v.AllowUserDeleteChat,
		"AllowUserChat":             v.AllowUserChat,
		"AllowGiphy":                v.AllowGiphy,
		"GiphyRatingType":           v.GiphyRatingType,
		"AllowMemes":                v.AllowMemes,
		"AllowImmersiveReader":      v.AllowImmersiveReader,
		"AllowStickers":             v.AllowStickers,
	}
}

// guestMessagingComplete reports whether every public setting was read.
func guestMessagingComplete(v *guestmessaging.Configuration) bool {
	if v == nil {
		return false
	}
	for _, value := range guestMessagingProjection(v) {
		if reflect.ValueOf(value).IsNil() {
			return false
		}
	}
	return true
}

// guestMessagingOracle runs the pinned 8.1.0 Global getter in a dedicated
// read-only PowerShell process and returns its ten public getter settings.
func guestMessagingOracle(t *testing.T, ctx context.Context, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": "guest-messaging"})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-guest-messaging-configuration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned guest messaging configuration read process failed")
	}
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		t.Fatal("pinned guest messaging configuration projection unavailable")
	}
	if len(result.Settings) != len(guestMessagingProjection(&guestmessaging.Configuration{})) {
		t.Fatal("incomplete guest messaging configuration getter projection")
	}
	for name := range guestMessagingProjection(&guestmessaging.Configuration{}) {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("incomplete guest messaging configuration getter projection")
		}
	}
	return result.Settings
}
