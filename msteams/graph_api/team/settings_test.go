package team_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

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
	if !reflect.DeepEqual(encodedSettings(t, value), map[string]any{"memberSettings": map[string]any{}, "guestSettings": map[string]any{}, "messagingSettings": map[string]any{}, "funSettings": map[string]any{}, "discoverySettings": map[string]any{}}) {
		t.Fatal("explicit empty settings containers were omitted")
	}
}

func TestUnitSettingsPreserveFalseAndUseNativeStringRating(t *testing.T) {
	t.Parallel()
	yes, no, rating := true, false, team.GiphyModerate
	value := team.Settings{MemberSettings: &team.MemberSettings{AllowCreatePrivateChannels: &yes},
		GuestSettings: &team.GuestSettings{AllowDeleteChannels: &no}, MessagingSettings: &team.MessagingSettings{AllowUserEditMessages: &no},
		FunSettings:       &team.FunSettings{AllowGiphy: &no, GiphyContentRating: &rating},
		DiscoverySettings: &team.DiscoverySettings{ShowInTeamsSearchAndSuggestions: &no}}
	want := map[string]any{
		"memberSettings":    map[string]any{"allowCreatePrivateChannels": true},
		"guestSettings":     map[string]any{"allowDeleteChannels": false},
		"messagingSettings": map[string]any{"allowUserEditMessages": false},
		"funSettings":       map[string]any{"allowGiphy": false, "giphyContentRating": "moderate"},
		"discoverySettings": map[string]any{"showInTeamsSearchAndSuggestions": false},
	}
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
