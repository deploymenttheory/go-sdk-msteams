package teamsguestmeetingconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
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
	raw, err := root.ReadFile("msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/read_model_cases.json")
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
		if row.Command != "Get-CsTeamsGuestMeetingConfiguration" {
			continue
		}
		count++
		t.Run(fmt.Sprintf("%s-%d", row.Kind, count), func(t *testing.T) {
			var model any = new(resource.Configuration)
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
	if count != 160 {
		t.Fatal("source rows disappeared")
	}
}

func TestUnitConfigurationCoercedReadUsesRootClient(t *testing.T) {
	t.Parallel()
	const body = `{"Identity":37,"AllowIPVideo":"False","ScreenSharingMode":9007199254740993,"AllowMeetNow":1,"Future":{"keep":true}}`
	run(t, testutil.Case{Name: "own-coerced-root-read", Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Global", "Get-CsTeamsGuestMeetingConfiguration", nil, 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.GetByIdentity(ctx, "Global")
		if err != nil || value == nil || response == nil || !bytes.Equal(response.Body(), []byte(body)) || !bytes.Equal(value.RawJSON, []byte(body)) {
			t.Fatal("root read lost its model, receipt or raw snapshot")
		}
		if value.Identity == nil || *value.Identity != "37" || value.AllowIPVideo == nil || *value.AllowIPVideo || value.AllowMeetNow == nil || !*value.AllowMeetNow || value.ScreenSharingMode == nil || *value.ScreenSharingMode != "9007199254740993" || value.RawFields["Future"] == nil {
			t.Fatal("source scalar projection lost exact values or future fields")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
}
