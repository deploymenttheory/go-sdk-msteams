package teamsfeedbackpolicy_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestFeedbackPolicyModelsMatchPinnedSetters(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 26 {
		t.Fatal("incomplete feedback model fixture")
	}
	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var value resource.Policy
			var e error
			if c.Input != "constructor" {
				e = json.Unmarshal([]byte(c.Input), &value)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatal("source setter failure accepted")
				}
				return
			}
			if e != nil {
				t.Fatal("source-compatible model rejected")
			}
			type projection resource.Policy
			encoded, e := json.Marshal(projection(value))
			var got, want map[string]any
			if e != nil || json.Unmarshal(encoded, &got) != nil || json.Unmarshal(c.Observation.Value, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("source typed projection differs")
			}
		})
	}
}

func TestFeedbackPolicySnapshotsOwnFutureData(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","AllowLogCollection":false,"AllowEmailCollection":null,"UserInitiatedMode":"","Future":9007199254740993}`)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.AllowLogCollection == nil || *value.AllowLogCollection || value.AllowEmailCollection != nil || value.UserInitiatedMode == nil || *value.UserInitiatedMode != "" {
		t.Fatal("nullable false/empty projection lost")
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
