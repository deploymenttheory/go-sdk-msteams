package teamscallingpolicy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

type modelCase struct {
	Name, Input string
	Observation struct {
		Failure *string
		Value   json.RawMessage
	}
}

func modelCases(t *testing.T) []modelCase {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Models []modelCase }
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 144 {
		t.Fatal("invalid model fixture")
	}
	return fixture.Models
}
func exactJSON(t *testing.T, raw []byte) any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestUnitCallingPolicyKnownProjectionsMatchPinnedModels(t *testing.T) {
	t.Parallel()
	for _, m := range modelCases(t) {
		t.Run(m.Name, func(t *testing.T) {
			if strings.Contains(m.Name, "-json-1.5") || strings.Contains(m.Name, `-json-"9007199254740993"`) {
				return
			} // Explicit native JSON boundary below.
			var value resource.Policy
			var err error
			if m.Input != "constructor" {
				err = json.Unmarshal([]byte(m.Input), &value)
			}
			if m.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source-invalid model accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source-valid typed model rejected")
			}
			type known resource.Policy
			raw, err := json.Marshal(known(value))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exactJSON(t, raw), exactJSON(t, m.Observation.Value)) {
				t.Fatal("known model projection differs")
			}
		})
	}
}
func TestUnitCallingPolicyPreservesOwnedRawSnapshotAndExactInt64(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","CallingSpendUserLimit":9223372036854775807,"Future":{"n":9007199254740993},"Description":"first","description":null}`)
	saved := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.CallingSpendUserLimit == nil || *value.CallingSpendUserLimit != 9223372036854775807 || value.Description != nil {
		t.Fatal("nullable/exact projection differs")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	got, err := json.Marshal(value)
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatal("snapshot lost bytes or ownership")
	}
}
func TestUnitCallingPolicyRequiresNativeJSONTypes(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"CallingSpendUserLimit":"9007199254740993"}`, `{"CallingSpendUserLimit":1.5}`, `{"CallRecordingExpirationDays":1.5}`, `{"AllowPrivateCalling":"false"}`, `{"Description":{}}`, `null`, `[]`} {
		var value resource.Policy
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("non-native JSON type accepted")
		}
	}
}

// The source model supplies accepted final values; the provider fixture
// independently establishes sparse field binding and identity serialization.
// This tests the composition, not a claimed source binder/network execution.
func TestUnitCallingPolicyRequestsRetainAcceptedModelSettings(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for index, m := range modelCases(t) {
		if m.Observation.Failure != nil || m.Input == "constructor" || m.Input == "{}" {
			continue
		}
		var input map[string]json.RawMessage
		var known map[string]json.RawMessage
		if json.Unmarshal([]byte(m.Input), &input) != nil || json.Unmarshal(m.Observation.Value, &known) != nil {
			t.Fatal("invalid model input/projection")
		}
		bound := map[string]json.RawMessage{}
		for key := range input {
			for canonical, value := range known {
				if strings.EqualFold(key, canonical) {
					bound[canonical] = value
					break
				}
			}
		}
		raw, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		var settings resource.Settings
		if json.Unmarshal(raw, &settings) != nil {
			t.Fatal("accepted final settings not representable")
		}
		bound["Identity"] = json.RawMessage(`"Global"`)
		body, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("accepted-model-%03d", index), Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", item, setCommand, body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{}})
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
