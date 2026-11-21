package team_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

func pinnedSettings(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-team-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		Version string `json:"baseline_version"`
		Cases   []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &artifact) != nil || artifact.Version != "8.1.0" {
		t.Fatal("pinned settings artifact unavailable")
	}
	for _, item := range artifact.Cases {
		if item.Name == name {
			var fields map[string]any
			if json.Unmarshal(item.Wire, &fields) != nil || fields == nil {
				t.Fatal("selected settings fixture is not an object")
			}
			return fields
		}
	}
	t.Fatal("pinned settings case disappeared")
	return nil
}

func encodedSettings(t *testing.T, value team.Settings) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if json.Unmarshal(raw, &fields) != nil {
		t.Fatal("native settings are not JSON")
	}
	return fields
}

func TestUnitSettingsOmissionAndEmptyContainers(t *testing.T) {
	t.Parallel()
	if len(encodedSettings(t, team.Settings{})) != 0 {
		t.Fatal("native sparse input initialized omitted settings")
	}
	value := team.Settings{MemberSettings: &team.MemberSettings{}, GuestSettings: &team.GuestSettings{},
		MessagingSettings: &team.MessagingSettings{}, FunSettings: &team.FunSettings{}, DiscoverySettings: &team.DiscoverySettings{}}
	if !reflect.DeepEqual(encodedSettings(t, value), pinnedSettings(t, "team-settings-omitted")) {
		t.Fatal("explicit five-container composition differs from pinned observation")
	}
}

func TestUnitSettingsPreserveFalseAndUseNativeStringRating(t *testing.T) {
	t.Parallel()
	yes, no, rating := true, false, team.GiphyModerate
	value := team.Settings{MemberSettings: &team.MemberSettings{AllowCreatePrivateChannels: &yes},
		GuestSettings: &team.GuestSettings{AllowDeleteChannels: &no}, MessagingSettings: &team.MessagingSettings{AllowUserEditMessages: &no},
		FunSettings:       &team.FunSettings{AllowGiphy: &no, GiphyContentRating: &rating},
		DiscoverySettings: &team.DiscoverySettings{ShowInTeamsSearchAndSuggestions: &no}}
	want := pinnedSettings(t, "team-settings-explicit-values")
	fun, ok := want["funSettings"].(map[string]any)
	if !ok || fun["giphyContentRating"] != float64(1) {
		t.Fatal("pinned numeric discrepancy changed; do not silently normalize the artifact")
	}
	// The accepted native composition deliberately follows Graph's string enum.
	// This modifies only the local comparison value, never the pinned artifact.
	fun["giphyContentRating"] = "moderate"
	if !reflect.DeepEqual(encodedSettings(t, value), want) {
		t.Fatal("native settings lost false values or sent a numeric Giphy rating")
	}
}

func TestUnitGiphyContentRatingRetainsFutureValues(t *testing.T) {
	t.Parallel()
	var value team.FunSettings
	if json.Unmarshal([]byte(`{"giphyContentRating":"future-rating"}`), &value) != nil || value.GiphyContentRating == nil || *value.GiphyContentRating != "future-rating" {
		t.Fatal("future string enum value was discarded")
	}
}
