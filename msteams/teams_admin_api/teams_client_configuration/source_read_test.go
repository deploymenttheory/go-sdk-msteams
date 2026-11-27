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

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

// These are actual own pinned DTO projections, not successful tenant replies.
func TestUnitConfigurationPinnedScalarReadVectors(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile("msteams/teams_admin_api/teams_client_configuration/testdata/read_scalar_cases.json")
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
	count := 0
	for _, row := range packet.Rows {
		if row.Command != "Get-CsTeamsClientConfiguration" {
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
	if count != 299 {
		t.Fatal("pinned own DTO cases disappeared")
	}
}

func TestUnitConfigurationScalarReadUsesRootClientAndKeepsReceipt(t *testing.T) {
	t.Parallel()
	const body = `{"Identity":false,"RestrictedSenderList":9007199254740993,"AllowBox":"TRUE","ContentPin":true,"future":1}`
	run(t, testutil.Case{
		Name: "source-scalar-root-read",

		Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Global", "Get-CsTeamsClientConfiguration", nil, 200, body)},
		Want:  testutil.Expected{Status: 200, Body: []byte(body)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.GetByIdentity(ctx, "Global")
			if err != nil || value == nil || response == nil || string(response.Body()) != body || value.RawFields["future"] == nil {
				t.Fatal("source scalar read lost its value, raw extension or receipt")
			}
			if value.Identity == nil || *value.Identity != "false" {
				t.Fatal("typed Identity conversion changed")
			}
			if value.RestrictedSenderList == nil || *value.RestrictedSenderList != "9007199254740993" {
				t.Fatal("typed RestrictedSenderList conversion changed")
			}
			if value.AllowBox == nil || *value.AllowBox != true {
				t.Fatal("typed AllowBox conversion changed")
			}
			if value.ContentPin == nil || *value.ContentPin != "true" {
				t.Fatal("typed ContentPin conversion changed")
			}
			return testutil.Observation{Response: response, Err: err}
		},
	})
}
