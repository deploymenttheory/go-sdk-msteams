package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

// Own pinned DTO projections are local model observations, not tenant captures.
func TestUnitMeetingTemplateOwnReadVectors(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile("msteams/teams_admin_api/teams_first_party_meeting_template_configuration/testdata/template_read_cases.json")
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
		t.Fatal("own source packet unavailable")
	}
	count := 0
	for _, row := range packet.Rows {
		if row.Command != "Get-CsTeamsFirstPartyMeetingTemplateConfiguration" {
			continue
		}
		count++
		t.Run(fmt.Sprintf("%s-%d", row.Kind, count), func(t *testing.T) {
			var value any
			switch row.Kind {
			case "configuration":
				value = &resource.Configuration{}
			case "template":
				value = &resource.Template{}
			case "option":
				value = &resource.Option{}
			default:
				t.Fatal("unknown own model")
			}
			err := json.Unmarshal([]byte(row.Input), value)
			if row.Error != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("own rejected read was accepted")
				}
				return
			}
			if err != nil || row.Projection == nil {
				t.Fatalf("own accepted read was rejected: %v", err)
			}
			var want any
			if json.Unmarshal([]byte(*row.Projection), &want) != nil {
				t.Fatal("invalid own projection")
			}
			if !reflect.DeepEqual(templateProjection(reflect.ValueOf(value)), want) {
				t.Fatalf("typed projection differs: input %s", row.Input)
			}
			got := reflect.ValueOf(value).Elem()
			if !bytes.Equal(got.FieldByName("RawJSON").Bytes(), []byte(row.Input)) {
				t.Fatal("read coercion changed raw snapshot")
			}
			var fields map[string]json.RawMessage
			if json.Unmarshal([]byte(row.Input), &fields) != nil || !reflect.DeepEqual(got.FieldByName("RawFields").Interface(), fields) {
				t.Fatal("read coercion changed raw fields")
			}
		})
	}
	if count != 140 {
		t.Fatalf("own rows disappeared: %d", count)
	}
}

// Project typed values without using MarshalJSON's deliberate raw-body view.
func templateProjection(v reflect.Value) any {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return templateProjection(v.Elem())
	}
	switch v.Kind() {
	case reflect.Struct:
		m := map[string]any{}
		for i := range v.NumField() {
			tag := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
			if tag == "-" || tag == "" {
				continue
			}
			m[tag] = templateProjection(v.Field(i))
		}
		return m
	case reflect.Slice:
		if v.IsNil() {
			return nil
		}
		a := make([]any, v.Len())
		for i := range v.Len() {
			a[i] = templateProjection(v.Index(i))
		}
		return a
	default:
		return v.Interface()
	}
}
