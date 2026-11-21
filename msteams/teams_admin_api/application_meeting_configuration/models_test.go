package applicationmeetingconfiguration_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
)

func TestChangesOwnInputsAndRetainEmptyOperators(t *testing.T) {
	t.Parallel()
	ids := []string{"fixture"}
	replacement := configuration.ReplaceAppIDs(ids)
	add := []string{"a"}
	remove := []string{"r"}
	delta := configuration.ChangeAppIDs(add, remove)
	legacy := json.RawMessage(`{"AppId":"fixture","DataSource":"private","Enabled":false}`)
	object, err := configuration.LegacyAppIDs([]json.RawMessage{legacy})
	if err != nil {
		t.Fatal(err)
	}
	objectDelta, err := configuration.ChangeLegacyAppIDs([]json.RawMessage{legacy}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ids[0] = "changed"
	add[0] = "changed"
	remove[0] = "changed"
	clear(legacy)
	for _, tc := range []struct {
		value *configuration.AppIDChange
		want  string
	}{
		{replacement, `["fixture"]`}, {delta, `{"Add":["a"],"Remove":["r"]}`}, {object, `[{"AppId":"fixture","Enabled":false}]`}, {objectDelta, `{"Add":[{"AppId":"fixture","Enabled":false}]}`}, {configuration.ChangeAppIDs(nil, nil), `{}`}, {configuration.ReplaceAppIDs(nil), `[]`},
	} {
		encoded, err := json.Marshal(tc.value)
		if err != nil {
			t.Fatal(err)
		}
		var actual, want any
		if json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(tc.want), &want) != nil || !reflect.DeepEqual(actual, want) {
			t.Fatal("change serialization or ownership differs")
		}
	}
	for _, raw := range []string{"null", "[]", "42", `"text"`, `{"bad":`} {
		if _, err := configuration.LegacyAppIDs([]json.RawMessage{json.RawMessage(raw)}); !errors.Is(err, configuration.ErrInput) {
			t.Fatal("invalid legacy object accepted")
		}
		if _, err := configuration.ChangeLegacyAppIDs(nil, []json.RawMessage{json.RawMessage(raw)}); !errors.Is(err, configuration.ErrInput) {
			t.Fatal("invalid legacy delta accepted")
		}
	}
	if _, err := json.Marshal(configuration.AppIDChange{}); !errors.Is(err, configuration.ErrInput) {
		t.Fatal("zero change serialized")
	}
}

func TestSnapshotRetainsNullOmissionEmptyAndFutureFields(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"Identity":"Global"}`, `{"Identity":"Global","AllowRemoveParticipantAppIds":null}`, `{"Identity":"Global","AllowRemoveParticipantAppIds":[]}`, `{"Identity":"Global","AllowRemoveParticipantAppIds":["a",""]}`, `{"Identity":"Global","AllowRemoveParticipantAppIds":[{"Future":false}],"Future":{"value":null}}`} {
		var value configuration.Configuration
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("valid response rejected")
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var actual, want map[string]json.RawMessage
		if json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(raw), &want) != nil || !reflect.DeepEqual(actual, want) {
			t.Fatal("snapshot changed received values")
		}
	}
}
