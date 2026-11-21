package teamscallholdpolicy_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func TestUnitCallHoldTypedModels(t *testing.T) {

	fixture := struct {
		Models []struct {
			Input       string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}{
		Models: []struct {
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
					Value: json.RawMessage("{\"AudioFileId\":null,\"Description\":null,\"Identity\":null,\"StreamingSourceAuthType\":null,\"StreamingSourceUrl\":null}"),
				},
			},

			{
				Input: "long-description",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":\"future-file\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":null,\"StreamingSourceAuthType\":\"FutureAuth\",\"StreamingSourceUrl\":\"future://fixture\"}"),
				},
			},

			{
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":null,\"Description\":null,\"Identity\":null,\"StreamingSourceAuthType\":null,\"StreamingSourceUrl\":null}"),
				},
			},

			{
				Input: "{\"Identity\":null,\"Description\":null,\"AudioFileId\":null,\"StreamingSourceUrl\":null,\"StreamingSourceAuthType\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":null,\"Description\":null,\"Identity\":null,\"StreamingSourceAuthType\":null,\"StreamingSourceUrl\":null}"),
				},
			},

			{
				Input: "{\"Identity\":\"Global\",\"Description\":\"\",\"AudioFileId\":\"future-file\",\"StreamingSourceUrl\":\"future://fixture\",\"StreamingSourceAuthType\":\"FutureAuth\",\"Future\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":\"future-file\",\"Description\":\"\",\"Identity\":\"Global\",\"StreamingSourceAuthType\":\"FutureAuth\",\"StreamingSourceUrl\":\"future://fixture\"}"),
				},
			},

			{
				Input: "{\"Description\":\"first\",\"description\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":null,\"Description\":\"second\",\"Identity\":null,\"StreamingSourceAuthType\":null,\"StreamingSourceUrl\":null}"),
				},
			},

			{
				Input: "{\"Description\":\"first\",\"Description\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AudioFileId\":null,\"Description\":null,\"Identity\":null,\"StreamingSourceAuthType\":null,\"StreamingSourceUrl\":null}"),
				},
			},

			{
				Input: "{\"AudioFileId\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},

			{
				Input: "{\"StreamingSourceUrl\":[]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},

			{
				Input: "{\"StreamingSourceAuthType\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	for _, c := range fixture.Models {
		t.Run(c.Input, func(t *testing.T) {
			var v resource.Policy
			var e error
			switch c.Input {
			case "constructor":
			case "long-description":
				v = resource.Policy{Description: new(strings.Repeat("x", 2000)), AudioFileID: new("future-file"), StreamingSourceURL: new("future://fixture"), StreamingSourceAuthType: new("FutureAuth")}
			default:
				e = json.Unmarshal([]byte(c.Input), &v)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatal("source-rejected shape accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			dto := map[string]any{"Identity": v.Identity, "Description": v.Description, "AudioFileId": v.AudioFileID, "StreamingSourceUrl": v.StreamingSourceURL, "StreamingSourceAuthType": v.StreamingSourceAuthType}
			raw, e := json.Marshal(dto)
			if e != nil {
				t.Fatal(e)
			}
			var got, want any
			if json.Unmarshal(raw, &got) != nil || json.Unmarshal(c.Observation.Value, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("model projection differs: got %s want %s", raw, c.Observation.Value)
			}
		})
	}
}

func TestUnitCallHoldRawSnapshotsRetainDuplicatesAndOwnership(t *testing.T) {
	raw := []byte(`{"Description":"first","description":"second","AudioFileId":null,"Future":false}`)
	original := append([]byte(nil), raw...)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil || v.Description == nil || *v.Description != "second" || v.AudioFileID != nil {
		t.Fatal("nullable/duplicate projection changed")
	}
	raw[0] = ' '
	*v.Description = "changed"
	encoded, e := json.Marshal(v)
	if e != nil || string(encoded) != string(original) {
		t.Fatal("original snapshot was borrowed or replaced by projection edits")
	}
}
