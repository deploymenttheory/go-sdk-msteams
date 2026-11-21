package teamsappsetuppolicy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

type modelFixture struct {
	Policies []struct {
		Input       string
		Observation struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"policies"`
	Elements []struct {
		Field, Input string
		Observation  struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"elements"`
	Lists []struct {
		Field, Name string
		Observation struct {
			Failure *string
			Values  json.RawMessage
		}
	} `json:"lists"`
}

func models(t *testing.T) modelFixture {
	t.Helper()

	v := modelFixture{
		Policies: []struct {
			Input       string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "long-description",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":false,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{\"AppPresetList\":null,\"PinnedAppBarApps\":null,\"PinnedMessageBarApps\":null,\"AppPresetMeetingList\":null,\"AdditionalCustomizationApps\":null,\"PinnedCallingBarApps\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{\"PinnedAppBarApps\":[{\"Id\":\"a\",\"Order\":1}],\"PinnedAppBarApps\":[{\"Id\":\"b\",\"Order\":2}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"b\",\"Identity\":null,\"Order\":2}],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{\"AllowSideLoading\":false,\"AllowUserPinning\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":false,\"AllowUserPinning\":true,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{\"PinnedMessageBarApps\":[null]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[null]}"),
				},
			},

			{
				Input: "{\"PinnedAppBarApps\":[{\"Id\":\"a\",\"Order\":9007199254740993}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationApps\":[],\"AllowSideLoading\":null,\"AllowUserPinning\":null,\"AppPresetList\":[],\"AppPresetMeetingList\":[],\"Description\":null,\"Identity\":null,\"PinnedAppBarApps\":[{\"Id\":\"a\",\"Identity\":null,\"Order\":9007199254740993}],\"PinnedCallingBarApps\":[],\"PinnedMessageBarApps\":[]}"),
				},
			},

			{
				Input: "{\"PinnedMessageBarApps\":[{\"Id\":\"a\",\"Order\":null}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Input: "{\"AdditionalCustomizationApps\":[{\"Id\":\"a\",\"AdditionalCustomizationId\":null}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Input: "{\"PinnedCallingBarApps\":[{\"Id\":null}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Input: "{\"AppPresetList\":{\"Id\":\"bad-shape\"}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
		},
		Elements: []struct {
			Field, Input string
			Observation  struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Field: "AppPresetList",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetList",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null,\"Order\":9007199254740993}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null,\"Order\":9007199254740993}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":0}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":-1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":-1}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":9223372036854775807}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":9223372036854775807}"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.InvalidCastException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.InvalidOperationException"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":\"2\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":2}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null,\"Order\":9007199254740993}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null,\"Order\":9007199254740993}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":0}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":0}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":-1}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":-1}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":9223372036854775807}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":9223372036854775807}"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.InvalidCastException"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":9223372036854775808}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.InvalidOperationException"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Input: "json:{\"Id\":\"a\",\"Order\":\"2\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"a\",\"Identity\":null,\"Order\":2}"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationId\":\"fixture-customization\",\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationId\":\"fixture-customization\",\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationId\":null,\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationId\":null,\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{\"Id\":\"a\",\"AdditionalCustomizationId\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Input: "json:{\"Id\":\"a\",\"AdditionalCustomizationId\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AdditionalCustomizationId\":\"\",\"Id\":\"a\",\"Identity\":null}"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "json:{\"Id\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Input: "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},
		},
		Lists: []struct {
			Field, Name string
			Observation struct {
				Failure *string
				Values  json.RawMessage
			}
		}{

			{
				Field: "AppPresetList",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetList",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedAppBarApps",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"A\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "PinnedMessageBarApps",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":2},{\"Id\":\"a\",\"Identity\":null,\"Order\":1},{\"Id\":\"A\",\"Identity\":null,\"Order\":1},{\"Id\":\"a\",\"Identity\":null,\"Order\":1}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AppPresetMeetingList",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"two\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"two\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"two\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"two\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "AdditionalCustomizationApps",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"two\",\"Id\":\"a\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"A\",\"Identity\":null},{\"AdditionalCustomizationId\":\"one\",\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "remove-three",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Field: "PinnedCallingBarApps",
				Name:  "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},
		},
	}

	return v
}

func sameJSON(a, b []byte) bool {
	var x, y any
	left, right := json.NewDecoder(bytes.NewReader(a)), json.NewDecoder(bytes.NewReader(b))
	left.UseNumber()
	right.UseNumber()
	return left.Decode(&x) == nil && right.Decode(&y) == nil && reflect.DeepEqual(x, y)
}
func appsProjection(v []*resource.App) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
func pinsProjection(v []*resource.PinnedApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Order": x.Order}
		}
	}
	return out
}
func customsProjection(v []*resource.CustomizationApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "AdditionalCustomizationId": x.AdditionalCustomizationID}
		}
	}
	return out
}
func projection(v resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Identity": v.Identity, "Description": v.Description, "AllowSideLoading": v.AllowSideLoading, "AllowUserPinning": v.AllowUserPinning, "AppPresetList": appsProjection(v.AppPresetList), "PinnedAppBarApps": pinsProjection(v.PinnedAppBarApps), "PinnedMessageBarApps": pinsProjection(v.PinnedMessageBarApps), "AppPresetMeetingList": appsProjection(v.AppPresetMeetingList), "AdditionalCustomizationApps": customsProjection(v.AdditionalCustomizationApps), "PinnedCallingBarApps": appsProjection(v.PinnedCallingBarApps)})
	return raw
}
func TestUnitAppSetupTypedPolicyProjections(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Policies {
		var v resource.Policy
		var e error
		switch c.Input {
		case "constructor":
		case "long-description":
			v.Description = new(strings.Repeat("x", 2000))
			v.AllowSideLoading = new(false)
		default:
			e = json.Unmarshal([]byte(c.Input), &v)
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatal("pinned setup policy acceptance changed")
		}
		if e == nil && !sameJSON(projection(v), c.Observation.Value) {
			t.Fatal("pinned setup policy projection changed")
		}
	}
}
func TestUnitAppSetupTypedElementProjections(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Elements {
		var v any
		var e error
		decode := strings.HasPrefix(c.Input, "json:") || c.Input == "null"
		body := strings.TrimPrefix(c.Input, "json:")
		if c.Input == "null" {
			body = `{"Id":null}`
		}
		switch c.Field {
		case "PinnedAppBarApps", "PinnedMessageBarApps":
			var x *resource.PinnedApp
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewPinnedApp(c.Input, largeOrder)
			}
			if e == nil {
				v = pinsProjection([]*resource.PinnedApp{x})[0]
			}
		case "AdditionalCustomizationApps":
			var x *resource.CustomizationApp
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewCustomizationApp(c.Input, "fixture-customization")
			}
			if e == nil {
				v = customsProjection([]*resource.CustomizationApp{x})[0]
			}
		default:
			var x *resource.App
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewApp(c.Input)
			}
			if e == nil {
				v = appsProjection([]*resource.App{x})[0]
			}
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatalf("pinned setup element acceptance changed: %s %s", c.Field, c.Input)
		}
		if e == nil {
			raw, _ := json.Marshal(v)
			if !sameJSON(raw, c.Observation.Value) {
				t.Fatal("pinned setup element projection changed")
			}
		}
	}
}
func TestUnitAppSetupRawSnapshotsAndInt64StayExact(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"fixture-private","PinnedAppBarApps":[{"Id":"a","Order":9007199254740993,"Future":null}],"pinnedappbarapps":[{"Id":"b","Order":9223372036854775807}],"PinnedAppBarApps":null,"Future":false}`)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil || len(v.PinnedAppBarApps) != 2 || v.PinnedAppBarApps[0].Order != largeOrder || v.PinnedAppBarApps[1].Order != 9223372036854775807 {
		t.Fatal("readonly append or exact Int64 decoding changed")
	}
	clear(raw)
	encoded, e := json.Marshal(v)
	var again resource.Policy
	if e != nil || json.Unmarshal(encoded, &again) != nil || !sameJSON(projection(v), projection(again)) {
		t.Fatal("raw ownership/repeated properties lost")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", v, v), "fixture-private") {
		t.Fatal("setup formatting exposed data")
	}
	for _, body := range []string{`{"PinnedAppBarApps":[{"Order":null,"Order":1}]}`, `{"AppPresetList":[{"Id":null,"Id":"good"}]}`, `{"AdditionalCustomizationApps":[{"Id":"a","AdditionalCustomizationId":null,"AdditionalCustomizationId":"good"}]}`, `{"PinnedMessageBarApps":[{"Order":9223372036854775808}]}`} {
		var x resource.Policy
		if json.Unmarshal([]byte(body), &x) == nil {
			t.Fatal("earlier setter failure or overflow accepted")
		}
	}
}
