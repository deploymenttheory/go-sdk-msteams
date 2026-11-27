package aiagents

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func generatedFixture(t *testing.T, name string, value any) {
	t.Helper()
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(name)
	if err != nil || json.Unmarshal(raw, value) != nil {
		t.Fatal("pinned data observation unavailable")
	}
}
func decodeGeneratedModel(model string, raw []byte) (any, error) {
	var value any
	switch {
	case strings.HasSuffix(model, ".AiAgentConfigurationDtoModel"):
		value = &Agent{}
	case strings.HasSuffix(model, ".GetAiAgentConfigurationsResponse"):
		value = &AgentList{}
	case strings.HasSuffix(model, ".AiAgentQueryResult"):
		value = &QueryResult{}
	case strings.HasSuffix(model, ".AiAgentDescription"):
		value = &PlatformAgentDescription{}
	case strings.HasSuffix(model, ".AiAgent"):
		value = &PlatformAgent{}
	default:
		return nil, client.ErrDecode
	}
	err := json.Unmarshal(raw, value)
	return value, err
}

// These snapshots use the exact getter names and selected fields captured by
// the inert pinned probes, not wire aliases or RawFields. Configuration-list
// nil collections are normalized to public empty output by the wrapper.
func generatedSnapshot(value any) any {
	switch model := value.(type) {
	case *Agent:
		if model == nil {
			return nil
		}
		return map[string]any{"AgentId": model.AgentID, "Name": model.Name, "AgentType": model.AgentType, "AgentTargetTagTemplateId": model.AgentTargetTagTemplateID, "ConfigurationId": model.ConfigurationID}
	case *AgentList:
		var agents []any
		for _, item := range model.Agents {
			agents = append(agents, generatedSnapshot(item))
		}
		return map[string]any{"AgentCount": model.AgentCount, "AiAgent": agents}
	case *QueryResult:
		var agents []any
		if model.Agents != nil {
			agents = make([]any, len(model.Agents))
			for index, item := range model.Agents {
				agents[index] = generatedSnapshot(item)
			}
		}
		return map[string]any{"AgentCount": model.AgentCount, "ContinuationToken": model.ContinuationToken, "Agents": agents}
	case *PlatformAgent:
		if model == nil {
			return nil
		}
		return map[string]any{"Channel": model.Channels, "IsTeamsIvrEnabled": model.IsTeamsIvrEnabled, "Cid": model.CID, "BotHandle": model.BotHandle, "Description": generatedSnapshot(model.Description)}
	case *PlatformAgentDescription:
		if model == nil {
			return nil
		}
		return map[string]any{"AgentId": model.AgentID, "Description": model.Description, "IsTrusted": model.IsTrusted, "StarRating": model.StarRating, "AgentType": model.AgentType, "DisplayName": model.DisplayName}
	default:
		return nil
	}
}

func TestUnitGeneratedExactDuplicateAndCaseOrdering(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Model, Case string
			JSON        string `json:"input_json"`
			Exception   json.RawMessage
			Properties  map[string]json.RawMessage
		} `json:"cases"`
	}
	generatedFixture(t, "pinned-ordering.json", &fixture)
	if len(fixture.Cases) != 81 {
		t.Fatal("pinned ordering cases missing")
	}
	rejected := 0
	for _, item := range fixture.Cases {
		t.Run(item.Case, func(t *testing.T) {
			value, err := decodeGeneratedModel(item.Model, []byte(item.JSON))
			if len(item.Exception) > 0 && string(item.Exception) != "null" {
				rejected++
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("exact duplicate accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("case-distinct source object rejected")
			}
			got, marshalErr := json.Marshal(generatedSnapshot(value))
			source, sourceErr := json.Marshal(item.Properties)
			var sourceValue any
			if sourceErr != nil || json.Unmarshal(source, &sourceValue) != nil {
				t.Fatal("source getter snapshot unavailable")
			}
			// Canonicalize nested source objects too: RawMessage preserves their
			// original property order, which has no semantic significance here.
			want, sourceErr := json.Marshal(sourceValue)
			if marshalErr != nil || sourceErr != nil || string(got) != string(want) {
				t.Fatalf("source getter projection differs for %s", item.Model)
			}
		})
	}
	if rejected != 56 {
		t.Fatal("not every exact duplicate vector executed")
	}
	// The root node parser rejects duplicates even in ignored nested fields.
	if _, err := generatedNode([]byte(`{"future":{"x":1,"x":2}}`)); !errors.Is(err, client.ErrDecode) {
		t.Fatal("ignored nested duplicate accepted")
	}
}

func TestUnitGeneratedNumericLexemesAndRanges(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Model, Property string
			JSON            string `json:"input_json"`
			Exception       json.RawMessage
			Null            bool    `json:"is_null"`
			Expected        *string `json:"value_invariant_string"`
		} `json:"cases"`
	}
	generatedFixture(t, "pinned-numbers.json", &fixture)
	if len(fixture.Cases) != 32 {
		t.Fatal("pinned number cases missing")
	}
	for index, item := range fixture.Cases {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			value, err := decodeGeneratedModel(item.Model, []byte(item.JSON))
			if len(item.Exception) > 0 && string(item.Exception) != "null" {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source integer format/range error accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source numeric value rejected")
			}
			var actual *string
			switch model := value.(type) {
			case *AgentList:
				if model.AgentCount != nil {
					text := strconv.FormatInt(*model.AgentCount, 10)
					actual = &text
				}
			case *QueryResult:
				if model.AgentCount != nil {
					text := strconv.FormatInt(int64(*model.AgentCount), 10)
					actual = &text
				}
			case *PlatformAgent:
				if model.CID != nil {
					text := strconv.FormatInt(*model.CID, 10)
					actual = &text
				}
			case *PlatformAgentDescription:
				if (model.StarRating == nil) != item.Null {
					t.Fatal("source Double nullability changed")
				}
				if model.StarRating != nil {
					expected, e := strconv.ParseFloat(*item.Expected, 64)
					if e != nil || *model.StarRating != expected {
						t.Fatal("source Double value changed")
					}
				}
				return
			}
			if (actual == nil) != item.Null || actual != nil && *actual != *item.Expected {
				t.Fatal("source integer projection changed")
			}
		})
	}
	for raw, want := range map[string]float64{"1e309": math.Inf(1), "-1e309": math.Inf(-1), "1e-400": 0} {
		var value PlatformAgentDescription
		if json.Unmarshal([]byte(`{"starRating":`+raw+`}`), &value) != nil || value.StarRating == nil || *value.StarRating != want {
			t.Fatal("source Double edge changed")
		}
	}
}

func TestUnitGeneratedNullableArraysAndLegacyKnowledgeAreDistinct(t *testing.T) {
	var fixture struct {
		Arrays []struct {
			Type, Property string
			JSON           string `json:"input_json"`
			Items          []struct {
				Value *string
				Null  bool `json:"is_null"`
			}
		} `json:"generated_arrays"`
	}
	generatedFixture(t, "pinned-arrays.json", &fixture)
	if len(fixture.Arrays) != 4 {
		t.Fatal("pinned array cases missing")
	}
	for _, item := range fixture.Arrays {
		value, err := decodeGeneratedModel(item.Type, []byte(item.JSON))
		if err != nil {
			t.Fatal("source array rejected")
		}
		var actual []*string
		switch model := value.(type) {
		case *PlatformAgent:
			actual = model.Channels
		case *PlatformAgentDescription:
			switch item.Property {
			case "Capability":
				actual = model.Capabilities
			case "Category":
				actual = model.Categories
			case "Keyword":
				actual = model.Keywords
			default:
				t.Fatal("source array property unknown")
			}
		}
		if len(actual) != len(item.Items) {
			t.Fatal("source array position dropped")
		}
		for index, want := range item.Items {
			if (actual[index] == nil) != want.Null || actual[index] != nil && *actual[index] != *want.Value {
				t.Fatal("generated array item stringified or dropped")
			}
		}
	}
	values, err := knowledgeStrings([]byte(`[1.20,1E+03,9007199254740993,true,null]`))
	if err != nil || len(values) != 5 || values[0] == nil || *values[0] != "1.20" || values[1] == nil || *values[1] != "1E+03" || values[2] == nil || *values[2] != "9007199254740993" || values[3] == nil || *values[3] != "true" || values[4] != nil {
		t.Fatal("legacy string reader lost lexeme/null")
	}
	for _, raw := range []string{"", " ", "null", "{}", `[{}]`} {
		if _, err := knowledgeStrings([]byte(raw)); !errors.Is(err, client.ErrDecode) {
			t.Fatal("source knowledge Count/parser failure became empty success")
		}
	}
}

func TestUnitGeneratedProjectionDefaultsAndRemoveMinimum(t *testing.T) {
	for _, raw := range []string{`{}`, `{"AIAgent":null}`, `{"AIAgent":{}}`, `{"AIAgent":{"Name":1,"name":"ignored"}}`} {
		value, err := decodeAgent([]byte(raw))
		if err != nil || value == nil || value.Name != nil || value.AgentID != nil || value.ConfigurationID != nil {
			t.Fatal("source empty domain projection changed")
		}
	}
	for _, raw := range []string{"", " \t", "[]"} {
		value, err := decodeAgent([]byte(raw))
		if err != nil || value != nil {
			t.Fatal("source null response became failure/model")
		}
	}
	for _, raw := range []string{`{}`, `{"AIAgents":null}`} {
		var value AgentList
		if json.Unmarshal([]byte(raw), &value) != nil || value.Agents == nil || len(value.Agents) != 0 {
			t.Fatal("source empty config projection changed")
		}
	}
	var configs AgentList
	if json.Unmarshal([]byte(`{"AIAgents":[null,{}]}`), &configs) == nil {
		t.Fatal("null configuration domain row silently emitted")
	}
	var result QueryResult
	if json.Unmarshal([]byte(`{"agents":[null,{}]}`), &result) != nil || len(result.Agents) != 2 || result.Agents[0] != nil || result.Agents[1] == nil || result.Agents[1].Description == nil {
		t.Fatal("catalog null/lazy distinction lost")
	}
	for _, raw := range []string{"", "\t", "[]", "{}", `{"Diagnostics":[]}`} {
		if _, err := generatedNode([]byte(raw)); err != nil {
			t.Fatal("source Remove accepted null/minimal result rejected")
		}
	}
	for _, raw := range []string{"null", "1"} {
		if _, err := generatedNode([]byte(raw)); !errors.Is(err, client.ErrDecode) {
			t.Fatal("source Remove scalar error accepted")
		}
	}
}
