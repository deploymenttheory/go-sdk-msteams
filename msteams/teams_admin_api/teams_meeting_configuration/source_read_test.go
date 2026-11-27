package teamsmeetingconfiguration_test

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

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

// typedReadProjection bypasses the intentional raw-snapshot MarshalJSON view.
func typedReadProjection(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		return typedReadProjection(v.Elem())
	}
	switch v.Kind() {
	case reflect.Struct:
		out := map[string]any{}
		for i := range v.NumField() {
			f := v.Type().Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = f.Name
			}
			out[tag] = typedReadProjection(v.Field(i))
		}
		return out
	case reflect.Slice:
		if v.IsNil() {
			return nil
		}
		out := make([]any, v.Len())
		for i := range v.Len() {
			out[i] = typedReadProjection(v.Index(i))
		}
		return out
	default:
		return v.Interface()
	}
}

func TestUnitConfigurationOwnCompiledReadVectors(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile("msteams/teams_admin_api/teams_meeting_configuration/testdata/read_model_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var packet struct {
		Rows []struct {
			Command    string  `json:"command"`
			Kind       string  `json:"kind"`
			Input      string  `json:"input_json"`
			Projection *string `json:"projected_json"`
			Error      *string `json:"error"`
		} `json:"observations"`
	}
	if json.Unmarshal(raw, &packet) != nil {
		t.Fatal("actual pinned DTO observations unavailable")
	}
	count := 0
	for _, row := range packet.Rows {
		if row.Command != "Get-CsTeamsMeetingConfiguration" {
			continue
		}
		count++
		t.Run(fmt.Sprintf("%s-%d", row.Kind, count), func(t *testing.T) {
			var model any
			if row.Kind == "context" {
				model = new(resource.PublishedEntraAuthenticationContext)
			} else {
				model = new(resource.Configuration)
			}
			err := json.Unmarshal([]byte(row.Input), model)
			if row.Error != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source rejected this input but native model accepted it")
				}
				return
			}
			if err != nil || row.Projection == nil {
				t.Fatal("source-accepted input rejected")
			}
			got, err := json.Marshal(typedReadProjection(reflect.ValueOf(model)))
			var left, right any
			if err != nil || json.Unmarshal(got, &left) != nil || json.Unmarshal([]byte(*row.Projection), &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("typed complete DTO projection differs from actual source")
			}
			if !bytes.Equal(reflect.ValueOf(model).Elem().FieldByName("RawJSON").Bytes(), bytes.TrimSpace([]byte(row.Input))) {
				t.Fatal("read conversion changed the owned raw document")
			}
		})
	}
	if count != 565 {
		t.Fatal("source rows disappeared")
	}
}

func TestUnitConfigurationCoercedReadUsesRootClient(t *testing.T) {
	t.Parallel()
	const body = `{"Identity":false,"ClientAudioPort":1.5,"EnableQoS":"true","LogoURL":true,"PublishedEntraAuthenticationContexts":[{"Id":"c1","PublishedName":37,"Published":"True","Future":1},null],"publishedentraauthenticationcontexts":[{}],"FutureRoot":1}`
	run(t, testutil.Case{Name: "own-coerced-root-read", Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Global", "Get-CsTeamsMeetingConfiguration", nil, 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.GetByIdentity(ctx, "Global")
		if err != nil || value == nil || response == nil || !bytes.Equal(response.Body(), []byte(body)) || !bytes.Equal(value.RawJSON, []byte(body)) {
			t.Fatal("root read lost its model, receipt or raw snapshot")
		}
		if value.Identity == nil || *value.Identity != "false" || value.ClientAudioPort == nil || *value.ClientAudioPort != 2 || value.EnableQoS == nil || !*value.EnableQoS || value.LogoURL == nil || *value.LogoURL != "true" || len(value.PublishedEntraAuthenticationContexts) != 3 || value.PublishedEntraAuthenticationContexts[1] != nil || value.PublishedEntraAuthenticationContexts[2] == nil || value.PublishedEntraAuthenticationContexts[2].ID != nil || value.RawFields["FutureRoot"] == nil {
			t.Fatal("source scalar/list/default projection differs")
		}
		item := value.PublishedEntraAuthenticationContexts[0]
		if item == nil || item.ID == nil || *item.ID != "c1" || item.PublishedName == nil || *item.PublishedName != "37" || item.Published == nil || !*item.Published || !bytes.Contains(item.RawJSON, []byte(`"Future":1`)) {
			t.Fatal("context read lost typed and raw distinction")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
}

func TestUnitMeetingContextSetterFailureKeepsPreparationReceiptWithoutPatch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, body string }{
		{"explicit-null", `{"PublishedEntraAuthenticationContexts":[{"Id":null}]}`},
		{"invalid-first-alias", `{"PublishedEntraAuthenticationContexts":[{"Id":"invalid","id":"c1"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "context-setter-" + tc.name, Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, tc.body)}, Want: testutil.Expected{Status: 200, Body: []byte(tc.body), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{DisableAnonymousJoin: policy.Set(false)}})
				if !errors.Is(err, client.ErrDecode) || response == nil || !bytes.Equal(response.Body(), []byte(tc.body)) || result == nil || result.Before != nil || result.ReadResponse != response || result.PatchInvoked || result.HTTPAccepted || result.After != nil {
					t.Fatal("source setter failure lost receipt or sent a mutation")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
		})
	}
}
