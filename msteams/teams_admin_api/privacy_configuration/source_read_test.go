package privacyconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

// These are actual own pinned DTO projections, not successful tenant replies.
func TestUnitConfigurationPinnedScalarReadVectors(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile("msteams/teams_admin_api/privacy_configuration/testdata/read_scalar_cases.json")
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
		if row.Command != "Get-CsPrivacyConfiguration" {
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
	if count != 85 {
		t.Fatal("pinned own DTO cases disappeared")
	}
}

func TestUnitConfigurationScalarReadUsesRootClientAndKeepsReceipt(t *testing.T) {
	t.Parallel()
	const body = `{"Identity":37,"EnablePrivacyMode":1,"AutoInitiateContacts":" False ","future":{"keep":true}}`
	run(t, testutil.Case{
		Name: "source-scalar-root-read",

		Steps: []testutil.Exchange{wire("GET", collection+"/configuration/Global", "Get-CsPrivacyConfiguration", nil, 200, body)},
		Want:  testutil.Expected{Status: 200, Body: []byte(body)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
			if err != nil || value == nil || response == nil || string(response.Body()) != body || value.RawFields["future"] == nil {
				t.Fatal("source scalar read lost its value, raw extension or receipt")
			}
			if value.Identity == nil || *value.Identity != "37" {
				t.Fatal("typed Identity conversion changed")
			}
			if value.EnablePrivacyMode == nil || *value.EnablePrivacyMode != true {
				t.Fatal("typed EnablePrivacyMode conversion changed")
			}
			if value.AutoInitiateContacts == nil || *value.AutoInitiateContacts != false {
				t.Fatal("typed AutoInitiateContacts conversion changed")
			}
			return testutil.Observation{Response: response, Err: err}
		},
	})
}
