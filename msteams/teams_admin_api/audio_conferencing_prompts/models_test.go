package audioconferencingprompts

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

const fixtureID = "11111111-1111-1111-1111-111111111111"
const secondID = "22222222-2222-2222-2222-222222222222"

type observation struct {
	Failure *string         `json:"failure"`
	Value   json.RawMessage `json:"value"`
	Values  json.RawMessage `json:"values"`
}
type modelCase struct {
	Kind, Input, Name string
	Observation       observation
}

func readModels(t *testing.T) struct {
	Models       []modelCase
	Constructors []struct {
		Kind, Name, ID, Locale string
		TypeOrLocale           string `json:"type_or_locale"`
		NullID                 bool   `json:"null_id"`
		NullTypeOrLocale       bool   `json:"null_type_or_locale"`
		NullLocale             bool   `json:"null_locale"`
		Observation            observation
	}
	Lists []modelCase
} {
	t.Helper()
	v := struct {
		Models       []modelCase
		Constructors []struct {
			Kind, Name, ID, Locale string
			TypeOrLocale           string `json:"type_or_locale"`
			NullID                 bool   `json:"null_id"`
			NullTypeOrLocale       bool   `json:"null_type_or_locale"`
			NullLocale             bool   `json:"null_locale"`
			Observation            observation
		}
		Lists []modelCase
	}{
		Models: []modelCase{
			{
				Kind:  "configuration",
				Input: "constructor",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[],\"Prompts\":[]}"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{}",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[],\"Prompts\":[]}"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Identity\":null,\"Prompts\":null,\"Packages\":null}",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[],\"Prompts\":[]}"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Prompts\":[null],\"Packages\":[null]}",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[null],\"Prompts\":[null]}"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Prompts\":[{}],\"prompts\":[{}],\"Packages\":[{}],\"Packages\":[{}]}",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}],\"Prompts\":[{\"AudioPrompt\":null,\"Id\":null,\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null},{\"AudioPrompt\":null,\"Id\":null,\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}]}"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Prompts\":{\"Id\":\"bad-envelope\"}}",
				Observation: observation{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Prompts\":[{\"Id\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\"}]}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "configuration",
				Input: "{\"Future\":false}",
				Observation: observation{
					Value: json.RawMessage("{\"Identity\":null,\"Packages\":[],\"Prompts\":[]}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "constructor",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":null,\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":null,\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Identity\":null,\"Name\":null,\"AudioPrompt\":null,\"TextPrompt\":null}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":null,\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":null}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Type\":null}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Locale\":null}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"\"}",
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"future-id\"}",
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\"}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"ABCDEF12-1234-1234-1234-123456789ABC\"}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":\"ABCDEF12-1234-1234-1234-123456789ABC\",\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\\n\"}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\\n\",\"Identity\":null,\"Locale\":null,\"Name\":null,\"TextPrompt\":null,\"Type\":null}"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\\r\\n\"}",
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\"}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "prompt",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Type\":\"\",\"Locale\":\"future-locale\",\"Name\":\"\",\"AudioPrompt\":\"\",\"TextPrompt\":\"\"}",
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":\"\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"Locale\":\"future-locale\",\"Name\":\"\",\"TextPrompt\":\"\",\"Type\":\"\"}"),
				},
			},
			{
				Kind:  "package",
				Input: "constructor",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":null}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":\"\"}",
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":\"future-id\"}",
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\"}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":\"11111111-1111-1111-1111-111111111111\\n\"}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\\n\",\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"DefaultLocale\":null}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"DefaultLocale\":\"\"}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":\"\",\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"DefaultLocale\":\"future-locale\",\"Name\":null}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":\"future-locale\",\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"InboundStartRecordingPrompt\":null,\"InboundEndRecordingPrompt\":null,\"OutboundStartRecordingPrompt\":null,\"OutboundEndRecordingPrompt\":null,\"MeetingStartRecordingPrompt\":null,\"MeetingEndRecordingPrompt\":null}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"InboundStartRecordingPrompt\":[\"a\",\"a\",null,\"\"],\"inboundStartRecordingPrompt\":[\"b\"]}",
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":null,\"Id\":null,\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[\"a\",\"a\",null,\"\",\"b\"],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"InboundStartRecordingPrompt\":{}}",
				Observation: observation{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
					Value:   json.RawMessage("null"),
				},
			},
			{
				Kind:  "package",
				Input: "{\"Id\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\"}",
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},
		},
		Constructors: []struct {
			Kind, Name, ID, Locale string
			TypeOrLocale           string `json:"type_or_locale"`
			NullID                 bool   `json:"null_id"`
			NullTypeOrLocale       bool   `json:"null_type_or_locale"`
			NullLocale             bool   `json:"null_locale"`
			Observation            observation
		}{

			{
				Kind:             "prompt",
				Name:             "valid",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"Locale\":\"xx-future\",\"Name\":null,\"TextPrompt\":null,\"Type\":\"FutureType\"}"),
				},
			},

			{
				Kind:             "prompt",
				Name:             "empty-fields",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "",
				TypeOrLocale:     "",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Value: json.RawMessage("{\"AudioPrompt\":null,\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"Locale\":\"\",\"Name\":null,\"TextPrompt\":null,\"Type\":\"\"}"),
				},
			},

			{
				Kind:             "prompt",
				Name:             "null-id",
				ID:               "",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           true,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "prompt",
				Name:             "invalid-id",
				ID:               "future-id",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "prompt",
				Name:             "null-type-or-default-locale",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "xx-future",
				TypeOrLocale:     "",
				NullID:           false,
				NullTypeOrLocale: true,
				NullLocale:       false,
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "prompt",
				Name:             "null-locale",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       true,
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "package",
				Name:             "valid",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":\"FutureType\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},

			{
				Kind:             "package",
				Name:             "empty-fields",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "",
				TypeOrLocale:     "",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":\"\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},

			{
				Kind:             "package",
				Name:             "null-id",
				ID:               "",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           true,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "package",
				Name:             "invalid-id",
				ID:               "future-id",
				Locale:           "xx-future",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       false,
				Observation: observation{
					Failure: new("XsdClassGenerator.PropertyArgumentException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "package",
				Name:             "null-type-or-default-locale",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "xx-future",
				TypeOrLocale:     "",
				NullID:           false,
				NullTypeOrLocale: true,
				NullLocale:       false,
				Observation: observation{
					Failure: new("System.ArgumentNullException"),
					Value:   json.RawMessage("null"),
				},
			},

			{
				Kind:             "package",
				Name:             "null-locale",
				ID:               "11111111-1111-1111-1111-111111111111",
				Locale:           "",
				TypeOrLocale:     "FutureType",
				NullID:           false,
				NullTypeOrLocale: false,
				NullLocale:       true,
				Observation: observation{
					Value: json.RawMessage("{\"DefaultLocale\":\"FutureType\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":null,\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":null,\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}"),
				},
			},
		},
		Lists: []modelCase{
			{
				Kind: "prompt",
				Name: "null",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "prompt",
				Name: "empty-operators",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "prompt",
				Name: "replace-duplicates",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "replace-null",
				Observation: observation{
					Values: json.RawMessage("[null]"),
				},
			},
			{
				Kind: "prompt",
				Name: "remove-one",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "remove-two",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "add-existing",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "add-null",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},null]"),
				},
			},
			{
				Kind: "prompt",
				Name: "remove-null",
				Observation: observation{
					Failure: new("System.NullReferenceException"),
					Values:  json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "remove-and-add",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "remove-from-empty",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "prompt",
				Name: "null-existing",
				Observation: observation{
					Failure: new("System.NullReferenceException"),
					Values:  json.RawMessage("[null,{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "identity-ignored",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-Id",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-Name",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-AudioPrompt",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-TextPrompt",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-Type",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "prompt",
				Name: "different-Locale",
				Observation: observation{
					Values: json.RawMessage("[{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"a\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"},{\"AudioPrompt\":\"audio\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"Locale\":\"xx-future\",\"Name\":\"b\",\"TextPrompt\":\"text\",\"Type\":\"FutureType\"}]"),
				},
			},
			{
				Kind: "package",
				Name: "null",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "package",
				Name: "empty-operators",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "package",
				Name: "replace-duplicates",
				Observation: observation{
					Values: json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "replace-null",
				Observation: observation{
					Values: json.RawMessage("[null]"),
				},
			},
			{
				Kind: "package",
				Name: "remove-one",
				Observation: observation{
					Failure: new("System.ArgumentException"),
					Values:  json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "remove-two",
				Observation: observation{
					Failure: new("System.ArgumentException"),
					Values:  json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "add-existing",
				Observation: observation{
					Values: json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "add-null",
				Observation: observation{
					Values: json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},null]"),
				},
			},
			{
				Kind: "package",
				Name: "remove-null",
				Observation: observation{
					Failure: new("System.ArgumentException"),
					Values:  json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "remove-and-add",
				Observation: observation{
					Failure: new("System.ArgumentException"),
					Values:  json.RawMessage("[{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
			{
				Kind: "package",
				Name: "remove-from-empty",
				Observation: observation{
					Values: json.RawMessage("[]"),
				},
			},
			{
				Kind: "package",
				Name: "null-existing",
				Observation: observation{
					Failure: new("System.ArgumentException"),
					Values:  json.RawMessage("[null,{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"a\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]},{\"DefaultLocale\":\"xx-future\",\"Id\":\"11111111-1111-1111-1111-111111111111\",\"Identity\":\"ignored\",\"InboundEndRecordingPrompt\":[],\"InboundStartRecordingPrompt\":[],\"MeetingEndRecordingPrompt\":[],\"MeetingStartRecordingPrompt\":[],\"Name\":\"b\",\"OutboundEndRecordingPrompt\":[],\"OutboundStartRecordingPrompt\":[]}]"),
				},
			},
		},
	}

	return v
}
func equalJSON(t *testing.T, got any, want json.RawMessage) {
	t.Helper()
	raw, e := json.Marshal(got)
	if e != nil {
		t.Fatal(e)
	}
	var a, b any
	if json.Unmarshal(raw, &a) != nil || json.Unmarshal(want, &b) != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("known DTO differs from pinned observation\ngot %s\nwant %s", raw, want)
	}
}
func configurationDTO(v *Configuration) map[string]any {
	return map[string]any{"Identity": v.Identity, "Prompts": project(v.Prompts, promptDTO), "Packages": project(v.Packages, packageDTO)}
}
func TestUnitTypedPromptModels(t *testing.T) {
	for i, c := range readModels(t).Models {
		t.Run(c.Kind+"/"+c.Input, func(t *testing.T) {
			var got any
			var e error
			switch c.Kind {
			case "configuration":
				var v Configuration
				if c.Input != "constructor" {
					e = json.Unmarshal([]byte(c.Input), &v)
				}
				got = configurationDTO(&v)
			case "prompt":
				var v Prompt
				if c.Input != "constructor" {
					e = json.Unmarshal([]byte(c.Input), &v)
				}
				got = promptDTO(&v)
			case "package":
				var v PromptPackage
				if c.Input != "constructor" {
					e = json.Unmarshal([]byte(c.Input), &v)
				}
				got = packageDTO(&v)
			default:
				t.Fatal("unknown model kind")
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatalf("case %d accepted rejected model: %v", i, e)
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			equalJSON(t, got, c.Observation.Value)
		})
	}
}
func TestUnitTypedPromptConstructors(t *testing.T) {
	for _, c := range readModels(t).Constructors {
		t.Run(c.Kind+"/"+c.Name, func(t *testing.T) {
			if c.NullID || c.NullTypeOrLocale || c.Kind == "prompt" && c.NullLocale {

				if c.Observation.Failure == nil {
					t.Fatal("pinned null constructor unexpectedly succeeds")
				}
				return
			}
			var got any
			var e error
			if c.Kind == "prompt" {
				var v *Prompt
				v, e = NewPrompt(c.ID, c.TypeOrLocale, c.Locale)
				got = promptDTO(v)
			} else {
				var v *PromptPackage
				v, e = NewPromptPackage(c.ID, c.TypeOrLocale)
				got = packageDTO(v)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, ErrInput) {
					t.Fatal("invalid source constructor accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			equalJSON(t, got, c.Observation.Value)
		})
	}
}
func samplePrompt(name string) *Prompt {
	return &Prompt{Identity: new("ignored"), ID: new(fixtureID), Name: new(name), AudioPrompt: new("audio"), TextPrompt: new("text"), Type: new("FutureType"), Locale: new("xx-future")}
}
func samplePackage(name string) *PromptPackage {
	return &PromptPackage{Identity: new("ignored"), ID: new(fixtureID), Name: new(name), DefaultLocale: new("xx-future")}
}
func promptCase(name string, removed, other *Prompt) *PromptEdit {
	switch name {
	case "null", "empty-operators":
		return ClearPrompts()
	case "replace-duplicates":
		return ReplacePrompts(other, other, removed)
	case "replace-null":
		return ReplacePrompts(nil)
	case "remove-two":
		return ChangePrompts(nil, []*Prompt{removed, removed})
	case "add-existing":
		return ChangePrompts([]*Prompt{removed, removed}, nil)
	case "add-null":
		return ChangePrompts([]*Prompt{nil}, nil)
	case "remove-null":
		return ChangePrompts(nil, []*Prompt{nil})
	case "remove-and-add":
		return ChangePrompts([]*Prompt{other}, []*Prompt{removed})
	default:
		return ChangePrompts(nil, []*Prompt{removed})
	}
}
func packageCase(name string, removed, other *PromptPackage) *PackageEdit {
	switch name {
	case "null", "empty-operators":
		return ClearPackages()
	case "replace-duplicates":
		return ReplacePackages(other, other, removed)
	case "replace-null":
		return ReplacePackages(nil)
	case "remove-two":
		return ChangePackages(nil, []*PromptPackage{removed, removed})
	case "add-existing":
		return ChangePackages([]*PromptPackage{removed, removed}, nil)
	case "add-null":
		return ChangePackages([]*PromptPackage{nil}, nil)
	case "remove-null":
		return ChangePackages(nil, []*PromptPackage{nil})
	case "remove-and-add":
		return ChangePackages([]*PromptPackage{other}, []*PromptPackage{removed})
	default:
		return ChangePackages(nil, []*PromptPackage{removed})
	}
}
func TestUnitTypedPromptAndPackageEdits(t *testing.T) {
	for _, c := range readModels(t).Lists {
		t.Run(c.Kind+"/"+c.Name, func(t *testing.T) {
			var got any
			var e error
			if c.Kind == "prompt" {
				current := []*Prompt{samplePrompt("a"), samplePrompt("a"), samplePrompt("b")}
				removed := samplePrompt("a")
				switch c.Name {
				case "identity-ignored":
					removed.Identity = new("other-metadata")
				case "different-Id":
					removed.ID = new(secondID)
				case "different-Name":
					removed.Name = new("different")
				case "different-AudioPrompt":
					removed.AudioPrompt = new("different")
				case "different-TextPrompt":
					removed.TextPrompt = new("different")
				case "different-Type":
					removed.Type = new("different")
				case "different-Locale":
					removed.Locale = new("different")
				case "remove-from-empty":
					current = nil
				case "null-existing":
					current = append([]*Prompt{nil}, current...)
				}
				before, _ := json.Marshal(project(current, promptDTO))
				var out []*Prompt
				out, e = promptCase(c.Name, removed, samplePrompt("b")).Apply(current)
				got = project(out, promptDTO)
				after, _ := json.Marshal(project(current, promptDTO))
				if string(before) != string(after) {
					t.Fatal("edit mutated caller inputs")
				}
			} else {
				current := []*PromptPackage{samplePackage("a"), samplePackage("a"), samplePackage("b")}
				if c.Name == "remove-from-empty" {
					current = nil
				}
				if c.Name == "null-existing" {
					current = append([]*PromptPackage{nil}, current...)
				}
				before, _ := json.Marshal(project(current, packageDTO))
				var out []*PromptPackage
				out, e = packageCase(c.Name, samplePackage("a"), samplePackage("b")).Apply(current)
				got = project(out, packageDTO)
				after, _ := json.Marshal(project(current, packageDTO))
				if string(before) != string(after) {
					t.Fatal("edit mutated caller inputs")
				}
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, ErrInput) {
					t.Fatal("source-rejected comparison succeeded")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			equalJSON(t, got, c.Observation.Values)
		})
	}
}
func TestUnitEditsOwnNestedDataAndRawSnapshots(t *testing.T) {
	raw := []byte(`{"Id":"11111111-1111-1111-1111-111111111111","InboundStartRecordingPrompt":["first",null],"Future":{"keep":false}}`)
	var v PromptPackage
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("fixture")
	}
	e := ReplacePackages(&v)
	*v.InboundStartRecordingPrompt[0] = "changed"
	*v.ID = secondID
	v.RawJSON[0] = ' '
	v.RawFields["Future"][0] = ' '
	out, err := e.Apply(nil)
	if err != nil {
		t.Fatal(err)
	}
	if *out[0].InboundStartRecordingPrompt[0] != "first" || *out[0].ID != fixtureID || out[0].RawJSON[0] != '{' || out[0].RawFields["Future"][0] != '{' {
		t.Fatal("edit borrowed nested input")
	}
	*out[0].InboundStartRecordingPrompt[0] = "again"
	out[0].RawJSON[0] = ' '
	again, err := e.Apply(nil)
	if err != nil || *again[0].InboundStartRecordingPrompt[0] != "first" || again[0].RawJSON[0] != '{' {
		t.Fatal("Apply result borrowed edit storage")
	}
	var p *PromptEdit
	var pkg *PackageEdit
	if !errors.Is(p.Validate(), ErrInput) || !errors.Is(pkg.Validate(), ErrInput) {
		t.Fatal("nil receiver")
	}
	if _, err := p.Apply(nil); !errors.Is(err, ErrInput) {
		t.Fatal("nil prompt Apply")
	}
	if _, err := pkg.Apply(nil); !errors.Is(err, ErrInput) {
		t.Fatal("nil package Apply")
	}
	if new(PromptEdit).Validate() == nil || new(PackageEdit).Validate() == nil {
		t.Fatal("zero edit accepted")
	}
}
func TestUnitRawDuplicatesAndTypedArraysStayIndependent(t *testing.T) {
	raw := []byte(`{"Prompts":[{}],"prompts":[{}],"Packages":[{"InboundStartRecordingPrompt":["a"],"InboundStartRecordingPrompt":["b"]}],"Future":false}`)
	original := append([]byte(nil), raw...)
	var v Configuration
	if json.Unmarshal(raw, &v) != nil || len(v.Prompts) != 2 || len(v.Packages) != 1 || len(v.Packages[0].InboundStartRecordingPrompt) != 2 {
		t.Fatal("readonly arrays did not append")
	}
	raw[0] = ' '
	got, e := json.Marshal(v)
	if e != nil || string(got) != string(original) {
		t.Fatal("raw presence/duplicates changed or borrowed")
	}
}
