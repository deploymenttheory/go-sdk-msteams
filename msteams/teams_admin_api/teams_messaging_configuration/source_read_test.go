package teamsmessagingconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
)

// These are actual own pinned DTO projections, not successful tenant replies.
func TestUnitConfigurationPinnedScalarReadVectors(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile("docs/implementation/policies/evidence/pol91-qualified-read-scalars.json")
	if err != nil {
		t.Fatal(err)
	}
	var packet struct {
		Rows []struct {
			Command    string  `json:"command"`
			Field      string  `json:"field"`
			Input      string  `json:"input_json"`
			Projection *string `json:"projected_json"`
			Error      *string `json:"error"`
		} `json:"observations"`
	}
	if json.Unmarshal(raw, &packet) != nil {
		t.Fatal("pinned source vectors unavailable")
	}
	// Supplemental observations retain their own source provenance and cover
	// numeric boundaries, URI setters and ordered nullable list members.
	supplement, err := root.ReadFile("docs/implementation/policies/evidence/pol91-qualified-read-supplement.json")
	if err != nil {
		t.Fatal(err)
	}
	var extra struct {
		Rows []struct {
			Command    string  `json:"command"`
			Field      string  `json:"field"`
			Input      string  `json:"input_json"`
			Projection *string `json:"projected_json"`
			Error      *string `json:"error"`
		} `json:"observations"`
	}
	if json.Unmarshal(supplement, &extra) != nil {
		t.Fatal("source supplement unavailable")
	}
	packet.Rows = append(packet.Rows, extra.Rows...)
	edges, err := root.ReadFile("docs/implementation/policies/evidence/pol91-enum-string-edges.json")
	if err != nil {
		t.Fatal(err)
	}
	extra.Rows = nil
	if json.Unmarshal(edges, &extra) != nil {
		t.Fatal("enum edge source unavailable")
	}
	packet.Rows = append(packet.Rows, extra.Rows...)
	count := 0
	for _, row := range packet.Rows {
		if row.Command != "Get-CsTeamsMessagingConfiguration" {
			continue
		}
		count++
		t.Run(fmt.Sprintf("%s-%d", row.Field, count), func(t *testing.T) {
			var value resource.Configuration
			err := json.Unmarshal([]byte(row.Input), &value)
			if row.Error != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source-rejected scalar was accepted")
				}
				return
			}
			if err != nil || row.Projection == nil {
				t.Fatal("source-accepted scalar was rejected")
			}
			typed := reflect.ValueOf(value)
			found := false
			for index := range typed.NumField() {
				field := typed.Type().Field(index)
				if strings.Split(field.Tag.Get("json"), ",")[0] != row.Field {
					continue
				}
				got, err := json.Marshal(typed.Field(index).Interface())
				if err != nil || !bytes.Equal(got, []byte(*row.Projection)) {
					t.Fatal("typed scalar differs from actual source projection")
				}
				found = true
			}
			if !found {
				t.Fatal("source property has no typed field")
			}
			var original map[string]json.RawMessage
			if json.Unmarshal([]byte(row.Input), &original) != nil || !bytes.Equal(value.RawFields[row.Field], original[row.Field]) {
				t.Fatal("coercion changed the retained raw field")
			}
		})
	}
	if count != 205 {
		t.Fatal("pinned own DTO cases disappeared")
	}
}

func TestUnitReceivedStorylineNumericNulDoesNotRelaxNamedValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want resource.Storyline
	}{{`{"Storyline":"1\u0000"}`, 1}, {`{"Storyline":"37\u0000"}`, 37}} {
		var value resource.Configuration
		if e := json.Unmarshal([]byte(tc.body), &value); e != nil || value.Storyline == nil {
			t.Fatal("source numeric-string trailing NUL rejected")
		}
		if *value.Storyline != tc.want {
			t.Fatal("numeric-string value changed")
		}
	}
	for _, body := range []string{`{"Storyline":"Enabled\u0000"}`, `{"Storyline":"\u0000Enabled"}`, `{"Storyline":"2147483648\u0000"}`} {
		var value resource.Configuration
		if !errors.Is(json.Unmarshal([]byte(body), &value), client.ErrDecode) {
			t.Fatal("named NUL or checked numeric overflow accepted")
		}
	}
}
