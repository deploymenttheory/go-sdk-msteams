package teamsclientconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
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
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 60 {
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
func TestClientConfigurationKnownProjectionsMatchPinnedModels(t *testing.T) {
	t.Parallel()
	for _, m := range modelCases(t) {
		t.Run(m.Name, func(t *testing.T) {

			var value resource.Configuration
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
			type known resource.Configuration
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
func TestClientConfigurationPreservesOwnedRawSnapshot(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","AllowEmailIntoChannel":false,"ContentPin":"0000","Future":{"n":9007199254740993},"RestrictedSenderList":"first","restrictedsenderlist":null}`)
	saved := append([]byte(nil), raw...)
	var value resource.Configuration
	if json.Unmarshal(raw, &value) != nil || value.AllowEmailIntoChannel == nil || *value.AllowEmailIntoChannel || value.ContentPin == nil || *value.ContentPin != "0000" || value.RestrictedSenderList != nil {
		t.Fatal("nullable/opaque projection differs")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	got, err := json.Marshal(value)
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatal("snapshot lost bytes or ownership")
	}
}

func TestClientConfigurationRequiresNativeJSONTypes(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"AllowEmailIntoChannel":"false"}`, `{"RestrictedSenderList":false}`, `{"ContentPin":{}}`, `null`, `[]`} {
		var value resource.Configuration
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("non-native JSON type accepted")
		}
	}
}

// The source model supplies accepted final values; the provider fixture
// independently establishes sparse field binding and identity serialization.
// This tests the composition, not a claimed source binder/network execution.
func TestClientConfigurationRequestsRetainAcceptedModelSettings(t *testing.T) {
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
			_, r, e := sdk.TeamsAdminAPI.TeamsClientConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{}})
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
