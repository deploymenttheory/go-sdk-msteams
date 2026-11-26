package conversationrole_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
)

func TestUnitConversationLocalBuildersMatchActualPinnedOutputAndCopyNullableValues(t *testing.T) {
	t.Parallel()
	parameters := parameters(t)
	raw, e := json.Marshal(parameters.Permissions)
	if e != nil {
		t.Fatal(e)
	}
	capture, e := os.ReadFile("testdata/pinned-projections.json")
	if e != nil {
		t.Fatal(e)
	}
	var source struct {
		BuilderBody string `json:"builder_body"`
	}
	if json.Unmarshal(capture, &source) != nil || source.BuilderBody == "" {
		t.Fatal("invalid source snapshot")
	}
	var actual, expected any
	if json.Unmarshal(raw, &actual) != nil || json.Unmarshal([]byte(source.BuilderBody), &expected) != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("local DTO builder differed from actual pinned local output at %v", differentPaths("", actual, expected))
	}
	flag := false
	chat := roles.NewChatPermissions(roles.ChatPermissionsValues{AddParticipants: &flag})
	flag = true
	if chat.AddParticipants == nil || *chat.AddParticipants {
		t.Fatal("nullable constructor retained caller pointer instead of CLR value")
	}
	if _, e := roles.NewPermissions(nil, parameters.Permissions.Calling); !errors.Is(e, roles.ErrInput) {
		t.Fatal("required local constructor input accepted nil")
	}
}
func TestUnitConversationPresenceMatchesCLRNullAndMinimumDateBehavior(t *testing.T) {
	t.Parallel()
	var value roles.Role
	if json.Unmarshal([]byte(`{}`), &value) != nil || value.Permissions != nil || value.CreatedAt.Time().Year() != 1 || value.CreatedAt.Kind() != roles.DateTimeUnspecified {
		t.Fatal("omitted CLR values changed")
	}
	for _, raw := range []string{`null`, `[]`, `{"Permissions":null}`, `{"Permissions":{"Chat":null}}`, `{"Permissions":{"Calling":null}}`, `{"CreatedAt":null}`, `{"UpdatedAt":"bad-date"}`, `{"Id":"not-a-guid"}`, `{"IsAssignable":"false"}`} {
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("invalid known CLR projection was accepted")
		}
	}
	raw := `{"Id":"00000000-0000-0000-0000-000000000000\n","Permissions":{"Chat":{"AddParticipants":null,"SendMessage":false,"Future":[]},"Calling":{"SendAudio":true}},"Future":null}`
	if json.Unmarshal([]byte(raw), &value) != nil || value.Permissions == nil || value.Permissions.Chat == nil || value.Permissions.Chat.AddParticipants != nil || value.Permissions.Chat.SendMessage == nil || *value.Permissions.Chat.SendMessage {
		t.Fatal("model GUID/final-LF or nullable flag behavior changed")
	}
	encoded, e := json.Marshal(value)
	var actual, expected any
	if e != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(raw), &expected) != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatal("nested raw null/false/empty/unknown presence lost")
	}
}
func TestUnitConversationDateKindsAndPrecisionPreservePinnedRequestBoundary(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"2027-01-01T00:00:00", "2027-01-01T00:00:00Z", "2027-01-01T00:00:00.1234567Z"} {
		value, e := roles.ParseDateTime(raw)
		if e != nil {
			t.Fatal(e)
		}
		encoded, e := json.Marshal(value)
		if e != nil || string(encoded) != `"`+raw+`"` {
			t.Fatal("source date representation changed")
		}
	}
	if _, e := roles.NewDateTime(time.Date(2027, 1, 1, 0, 0, 0, 1, time.UTC), roles.DateTimeUTC); !errors.Is(e, roles.ErrInput) {
		t.Fatal("finer than CLR100ns precision silently truncated")
	}
}

func differentPaths(path string, actual, expected any) []string {
	if reflect.DeepEqual(actual, expected) {
		return nil
	}
	a, aok := actual.(map[string]any)
	b, bok := expected.(map[string]any)
	if !aok || !bok {
		return []string{path}
	}
	var result []string
	for key, value := range a {
		result = append(result, differentPaths(path+"/"+key, value, b[key])...)
	}
	for key := range b {
		if _, ok := a[key]; !ok {
			result = append(result, path+"/"+key)
		}
	}
	return result
}
