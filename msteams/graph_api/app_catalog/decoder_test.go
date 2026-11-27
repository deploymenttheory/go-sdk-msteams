package appcatalog

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func sourceFixture(t *testing.T, name string, into any) {
	t.Helper()
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(name)
	if err != nil || json.Unmarshal(raw, into) != nil {
		t.Fatal("source observation fixture unavailable")
	}
}

type stringProjection struct {
	ID                 *string `json:"id"`
	ExternalID         *string `json:"externalId"`
	DisplayName        *string `json:"displayName"`
	DistributionMethod *string `json:"distributionMethod"`
}

func projected(value *TeamsApp) stringProjection {
	if value == nil {
		return stringProjection{}
	}
	return stringProjection{value.ID, value.ExternalID, value.DisplayName, value.DistributionMethod}
}

func TestUnitAppCatalogPinnedNullableModels(t *testing.T) {
	var fixture struct {
		Cases []struct {
			JSON       string          `json:"json"`
			Serialized *string         `json:"reserialized"`
			Exception  json.RawMessage `json:"exception"`
			Null       bool            `json:"object_is_null"`
		} `json:"model_cases"`
	}
	sourceFixture(t, "pinned-models.json", &fixture)
	if len(fixture.Cases) != 13 {
		t.Fatal("source cases missing")
	}
	for _, item := range fixture.Cases {
		// JsonConvert accepts empty text as a null scalar; Create handles this
		// separately because encoding/json requires a JSON document.
		if item.JSON == "" {
			continue
		}
		var app *TeamsApp
		err := json.Unmarshal([]byte(item.JSON), &app)
		if len(item.Exception) > 0 && string(item.Exception) != "null" {
			if err == nil {
				t.Fatal("source-rejected shape decoded")
			}
			continue
		}
		if err != nil || (app == nil) != item.Null {
			t.Fatal("source nullable model changed")
		}
		if item.Serialized != nil {
			var want stringProjection
			if json.Unmarshal([]byte(*item.Serialized), &want) != nil || !reflect.DeepEqual(projected(app), want) {
				t.Fatal("source projected strings changed")
			}
		}
		if item.JSON == `{"id":"a","future":{"keep":true},"appDefinitions":[{}]}` && app.RawFields["future"] == nil {
			t.Fatal("native raw extension was lost")
		}
	}
}

func TestUnitAppCatalogPinnedStringLexemesAndAliases(t *testing.T) {
	var numeric struct {
		Cases []struct {
			JSON string `json:"input_json"`
			stringProjection
		} `json:"numeric_cases"`
	}
	sourceFixture(t, "pinned-numbers.json", &numeric)
	if len(numeric.Cases) != 10 {
		t.Fatal("source numeric cases missing")
	}
	for _, item := range numeric.Cases {
		var app *TeamsApp
		if json.Unmarshal([]byte(item.JSON), &app) != nil || app == nil || !reflect.DeepEqual(projected(app), item.stringProjection) {
			t.Fatal("opaque string lexeme changed")
		}
	}
	var aliases struct {
		Cases []struct {
			JSON string `json:"json"`
			stringProjection
		} `json:"coercion_cases"`
	}
	sourceFixture(t, "pinned-output.json", &aliases)
	if len(aliases.Cases) < 6 {
		t.Fatal("source alias cases missing")
	}
	for _, item := range aliases.Cases {
		var app *TeamsApp
		if json.Unmarshal([]byte(item.JSON), &app) != nil || app == nil || !reflect.DeepEqual(projected(app), item.stringProjection) {
			t.Fatal("source alias/coercion changed")
		}
	}
}

func TestUnitAppCatalogPinnedEnvelopeReuse(t *testing.T) {
	var fixture struct {
		Cases []struct {
			JSON string   `json:"input_json"`
			IDs  []string `json:"ids"`
			Null bool     `json:"objects_is_null"`
			Next *string  `json:"next_link"`
		} `json:"cases"`
	}
	sourceFixture(t, "pinned-duplicates.json", &fixture)
	if len(fixture.Cases) != 7 {
		t.Fatal("source duplicate cases missing")
	}
	for _, item := range fixture.Cases {
		page, err := decode([]byte(item.JSON))
		if item.Null {
			if !errors.Is(err, client.ErrDecode) {
				t.Fatal("null value became empty success")
			}
			continue
		}
		if err != nil || len(page.Items) != len(item.IDs) {
			t.Fatal("source collection reuse changed")
		}
		for index, id := range item.IDs {
			if page.Items[index] == nil || page.Items[index].ID == nil || *page.Items[index].ID != id {
				t.Fatal("source wire-order identity changed")
			}
		}
		next := ""
		if item.Next != nil {
			next = *item.Next
		}
		if page.Next != next {
			t.Fatal("source scalar alias precedence changed")
		}
	}
}

func TestUnitAppCatalogPinnedContinuationProjection(t *testing.T) {
	var fixture struct {
		Rows []struct {
			JSON       string  `json:"input_json"`
			Next       *string `json:"next_link"`
			Terminates bool    `json:"terminates"`
			Exception  *string `json:"exception"`
		} `json:"rows"`
	}
	sourceFixture(t, "pinned-continuations.json", &fixture)
	if len(fixture.Rows) < 10 {
		t.Fatal("source continuation cases missing")
	}
	for _, item := range fixture.Rows {
		page, err := decode([]byte(item.JSON))
		if item.Exception != nil {
			if !errors.Is(err, client.ErrDecode) {
				t.Fatal("invalid continuation accepted")
			}
			continue
		}
		next := ""
		if !item.Terminates && item.Next != nil {
			next = *item.Next
		}
		if err != nil || page.Next != next {
			t.Fatal("source nullable/lexical continuation changed")
		}
	}
	for _, raw := range []string{"", " ", "null", "{}", `{"value":null}`, `{"value":{}}`} {
		if _, err := decode([]byte(raw)); !errors.Is(err, client.ErrDecode) {
			t.Fatal("missing collection became empty success")
		}
	}
	page, err := decode([]byte(`{"VALUE":[null,{}, {"id":"a"},null],"@ODATA.NEXTLINK":"\t\n"}`))
	if err != nil || len(page.Items) != 4 || page.Items[0] != nil || page.Items[1] == nil || page.Items[1].ID != nil || page.Items[2] == nil || page.Items[2].ID == nil || *page.Items[2].ID != "a" || page.Items[3] != nil || page.Next != "" {
		t.Fatal("public nullable output cardinality changed")
	}
}
