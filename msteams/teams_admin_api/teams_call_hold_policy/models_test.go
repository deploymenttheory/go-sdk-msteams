package teamscallholdpolicy_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func TestCallHoldPinnedModels(t *testing.T) {
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Models []struct {
			Input       string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 10 {
		t.Fatal("incomplete call hold models")
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

func TestCallHoldRawSnapshotsRetainDuplicatesAndOwnership(t *testing.T) {
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
