package teamsfilespolicy_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

func TestUnitFilesPolicyModelInputs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name, Input string
		Invalid     bool
		Expected    json.RawMessage
	}{
		{Name: "constructor", Input: "constructor", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "empty-object", Input: "{}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "all-settings", Input: "{\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}")},
		{Name: "NativeFileEntryPoints-null", Input: "{\"NativeFileEntryPoints\":null}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "NativeFileEntryPoints-empty", Input: "{\"NativeFileEntryPoints\":\"\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":\"\",\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "NativeFileEntryPoints-future", Input: "{\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\",\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "SPChannelFilesTab-null", Input: "{\"SPChannelFilesTab\":null}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "SPChannelFilesTab-empty", Input: "{\"SPChannelFilesTab\":\"\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":\"\",\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "SPChannelFilesTab-future", Input: "{\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "DefaultFileUploadAppId-null", Input: "{\"DefaultFileUploadAppId\":null}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "DefaultFileUploadAppId-empty", Input: "{\"DefaultFileUploadAppId\":\"\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":\"\",\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "DefaultFileUploadAppId-future", Input: "{\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "FileSharingInChatswithExternalUsers-null", Input: "{\"FileSharingInChatswithExternalUsers\":null}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "FileSharingInChatswithExternalUsers-empty", Input: "{\"FileSharingInChatswithExternalUsers\":\"\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":\"\"}")},
		{Name: "FileSharingInChatswithExternalUsers-future", Input: "{\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\"}")},
		{Name: "long-native-entry-points", Input: "{\"NativeFileEntryPoints\":\"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":\"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\",\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "case-duplicate", Input: "{\"NativeFileEntryPoints\":\"first\",\"nativefileentrypoints\":\"second\"}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":\"second\",\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "duplicate-null", Input: "{\"SPChannelFilesTab\":\"first\",\"spchannelfilestab\":null}", Invalid: false, Expected: json.RawMessage("{\"Identity\":null,\"NativeFileEntryPoints\":null,\"SPChannelFilesTab\":null,\"DefaultFileUploadAppId\":null,\"FileSharingInChatswithExternalUsers\":null}")},
		{Name: "invalid-string-object", Input: "{\"DefaultFileUploadAppId\":{}}", Invalid: true, Expected: json.RawMessage("null")},
		{Name: "invalid-string-array", Input: "{\"FileSharingInChatswithExternalUsers\":[]}", Invalid: true, Expected: json.RawMessage("null")},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var value resource.Policy
			var e error
			if c.Input != "constructor" {
				e = json.Unmarshal([]byte(c.Input), &value)
			}
			if c.Invalid {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatal("invalid scalar input accepted")
				}
				return
			}
			if e != nil {
				t.Fatal("valid typed model rejected")
			}
			type projection resource.Policy
			encoded, e := json.Marshal(projection(value))
			var got, want map[string]any
			if e != nil || json.Unmarshal(encoded, &got) != nil || json.Unmarshal(c.Expected, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("typed model projection differs")
			}
		})
	}
}

func TestUnitFilesPolicySnapshotsOwnFutureData(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","NativeFileEntryPoints":"","SPChannelFilesTab":null,"DefaultFileUploadAppId":"not-a-guid","Future":9007199254740993}`)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.NativeFileEntryPoints == nil || *value.NativeFileEntryPoints != "" || value.SPChannelFilesTab != nil || value.DefaultFileUploadAppID == nil || *value.DefaultFileUploadAppID != "not-a-guid" {
		t.Fatal("nullable string projection lost")
	}

	clear(raw)
	if string(value.RawFields["Future"]) != "9007199254740993" {
		t.Fatal("future snapshot lost")
	}
	value.RawFields["Future"][0] = 'x'
	encoded, e := json.Marshal(value)
	if e != nil || !json.Valid(encoded) {
		t.Fatal("original snapshot borrowed projected map storage")
	}
}
