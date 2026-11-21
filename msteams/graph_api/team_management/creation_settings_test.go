package teammanagement_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

// Each typed input selects one setting independently, so moving, omitting or
// inverting a flag fails against its independent request JSON vector.
var booleanSettings = map[string]func(*bool) team.Settings{
	"AllowCreateUpdateChannels": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateChannels: v}}
	},
	"AllowDeleteChannels": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowDeleteChannels: v}}
	},
	"AllowAddRemoveApps": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowAddRemoveApps: v}}
	},
	"AllowCreateUpdateRemoveTabs": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateRemoveTabs: v}}
	},
	"AllowCreateUpdateRemoveConnectors": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateRemoveConnectors: v}}
	},
	"AllowCreatePrivateChannels": func(v *bool) team.Settings {
		return team.Settings{MemberSettings: &team.MemberSettings{AllowCreatePrivateChannels: v}}
	},
	"AllowGuestCreateUpdateChannels": func(v *bool) team.Settings {
		return team.Settings{GuestSettings: &team.GuestSettings{AllowCreateUpdateChannels: v}}
	},
	"AllowGuestDeleteChannels": func(v *bool) team.Settings {
		return team.Settings{GuestSettings: &team.GuestSettings{AllowDeleteChannels: v}}
	},
	"AllowUserEditMessages": func(v *bool) team.Settings {
		return team.Settings{MessagingSettings: &team.MessagingSettings{AllowUserEditMessages: v}}
	},
	"AllowUserDeleteMessages": func(v *bool) team.Settings {
		return team.Settings{MessagingSettings: &team.MessagingSettings{AllowUserDeleteMessages: v}}
	},
	"AllowOwnerDeleteMessages": func(v *bool) team.Settings {
		return team.Settings{MessagingSettings: &team.MessagingSettings{AllowOwnerDeleteMessages: v}}
	},
	"AllowTeamMentions": func(v *bool) team.Settings {
		return team.Settings{MessagingSettings: &team.MessagingSettings{AllowTeamMentions: v}}
	},
	"AllowChannelMentions": func(v *bool) team.Settings {
		return team.Settings{MessagingSettings: &team.MessagingSettings{AllowChannelMentions: v}}
	},
	"AllowGiphy": func(v *bool) team.Settings { return team.Settings{FunSettings: &team.FunSettings{AllowGiphy: v}} },
	"AllowStickersAndMemes": func(v *bool) team.Settings {
		return team.Settings{FunSettings: &team.FunSettings{AllowStickersAndMemes: v}}
	},
	"AllowCustomMemes": func(v *bool) team.Settings { return team.Settings{FunSettings: &team.FunSettings{AllowCustomMemes: v}} },
	"ShowInTeamsSearchAndSuggestions": func(v *bool) team.Settings {
		return team.Settings{DiscoverySettings: &team.DiscoverySettings{ShowInTeamsSearchAndSuggestions: v}}
	},
}

func TestUnitSettingsInputsCoverBothPinnedApplicationParameterSets(t *testing.T) {
	t.Parallel()
	var runtime struct {
		Commands []struct {
			Name string
			Sets []struct {
				Name       string
				Parameters []struct{ Name string }
			} `json:"parameter_sets"`
		}
	}
	var ledger struct {
		Common map[string]string `json:"common_parameters"`
	}
	data, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil || json.Unmarshal(data, &runtime) != nil {
		t.Fatal("pinned parameter metadata unavailable")
	}
	data, err = os.ReadFile("../../../contracts/coverage/8.1.0.json")
	if err != nil || json.Unmarshal(data, &ledger) != nil {
		t.Fatal("coverage common-parameter mappings unavailable")
	}
	want := map[string]bool{"GiphyContentRating": true}
	for name := range booleanSettings {
		want[name] = true
	}
	checked := 0
	for _, command := range runtime.Commands {
		if command.Name != "New-Team" && command.Name != "Set-Team" {
			continue
		}
		for _, set := range command.Sets {
			if command.Name == "New-Team" && set.Name != "CreateTeam" {
				continue
			}
			otherInputs := map[string]bool{"DisplayName": true, "Description": true, "Classification": true, "MailNickName": true, "Visibility": true}
			if command.Name == "New-Team" {
				for _, name := range []string{"Owner", "Template", "RetainCreatedGroup"} {
					otherInputs[name] = true
				}
			} else {
				otherInputs["GroupId"] = true
			}
			observed := map[string]bool{}
			for _, p := range set.Parameters {
				if _, common := ledger.Common[p.Name]; !common && !otherInputs[p.Name] {
					observed[p.Name] = true
				}
			}
			if !reflect.DeepEqual(observed, want) {
				t.Fatal("a pinned application settings parameter has no typed input/field case")
			}
			checked++
		}
	}
	if checked != 2 {
		t.Fatal("both pinned settings-bearing parameter sets must be checked")
	}
}

func TestUnitSettingsSerializeEveryInputField(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		parameter string
		value     json.RawMessage
		want      json.RawMessage
	}{
		{parameter: "AllowCreateUpdateChannels", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateChannels":false}}`)},
		{parameter: "AllowCreateUpdateChannels", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateChannels":true}}`)},
		{parameter: "AllowDeleteChannels", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowDeleteChannels":false}}`)},
		{parameter: "AllowDeleteChannels", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowDeleteChannels":true}}`)},
		{parameter: "AllowAddRemoveApps", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowAddRemoveApps":false}}`)},
		{parameter: "AllowAddRemoveApps", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowAddRemoveApps":true}}`)},
		{parameter: "AllowCreateUpdateRemoveTabs", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateRemoveTabs":false}}`)},
		{parameter: "AllowCreateUpdateRemoveTabs", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateRemoveTabs":true}}`)},
		{parameter: "AllowCreateUpdateRemoveConnectors", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateRemoveConnectors":false}}`)},
		{parameter: "AllowCreateUpdateRemoveConnectors", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowCreateUpdateRemoveConnectors":true}}`)},
		{parameter: "AllowCreatePrivateChannels", value: json.RawMessage(`false`), want: json.RawMessage(`{"memberSettings":{"allowCreatePrivateChannels":false}}`)},
		{parameter: "AllowCreatePrivateChannels", value: json.RawMessage(`true`), want: json.RawMessage(`{"memberSettings":{"allowCreatePrivateChannels":true}}`)},
		{parameter: "AllowGuestCreateUpdateChannels", value: json.RawMessage(`false`), want: json.RawMessage(`{"guestSettings":{"allowCreateUpdateChannels":false}}`)},
		{parameter: "AllowGuestCreateUpdateChannels", value: json.RawMessage(`true`), want: json.RawMessage(`{"guestSettings":{"allowCreateUpdateChannels":true}}`)},
		{parameter: "AllowGuestDeleteChannels", value: json.RawMessage(`false`), want: json.RawMessage(`{"guestSettings":{"allowDeleteChannels":false}}`)},
		{parameter: "AllowGuestDeleteChannels", value: json.RawMessage(`true`), want: json.RawMessage(`{"guestSettings":{"allowDeleteChannels":true}}`)},
		{parameter: "AllowUserEditMessages", value: json.RawMessage(`false`), want: json.RawMessage(`{"messagingSettings":{"allowUserEditMessages":false}}`)},
		{parameter: "AllowUserEditMessages", value: json.RawMessage(`true`), want: json.RawMessage(`{"messagingSettings":{"allowUserEditMessages":true}}`)},
		{parameter: "AllowUserDeleteMessages", value: json.RawMessage(`false`), want: json.RawMessage(`{"messagingSettings":{"allowUserDeleteMessages":false}}`)},
		{parameter: "AllowUserDeleteMessages", value: json.RawMessage(`true`), want: json.RawMessage(`{"messagingSettings":{"allowUserDeleteMessages":true}}`)},
		{parameter: "AllowOwnerDeleteMessages", value: json.RawMessage(`false`), want: json.RawMessage(`{"messagingSettings":{"allowOwnerDeleteMessages":false}}`)},
		{parameter: "AllowOwnerDeleteMessages", value: json.RawMessage(`true`), want: json.RawMessage(`{"messagingSettings":{"allowOwnerDeleteMessages":true}}`)},
		{parameter: "AllowTeamMentions", value: json.RawMessage(`false`), want: json.RawMessage(`{"messagingSettings":{"allowTeamMentions":false}}`)},
		{parameter: "AllowTeamMentions", value: json.RawMessage(`true`), want: json.RawMessage(`{"messagingSettings":{"allowTeamMentions":true}}`)},
		{parameter: "AllowChannelMentions", value: json.RawMessage(`false`), want: json.RawMessage(`{"messagingSettings":{"allowChannelMentions":false}}`)},
		{parameter: "AllowChannelMentions", value: json.RawMessage(`true`), want: json.RawMessage(`{"messagingSettings":{"allowChannelMentions":true}}`)},
		{parameter: "AllowGiphy", value: json.RawMessage(`false`), want: json.RawMessage(`{"funSettings":{"allowGiphy":false}}`)},
		{parameter: "AllowGiphy", value: json.RawMessage(`true`), want: json.RawMessage(`{"funSettings":{"allowGiphy":true}}`)},
		{parameter: "AllowStickersAndMemes", value: json.RawMessage(`false`), want: json.RawMessage(`{"funSettings":{"allowStickersAndMemes":false}}`)},
		{parameter: "AllowStickersAndMemes", value: json.RawMessage(`true`), want: json.RawMessage(`{"funSettings":{"allowStickersAndMemes":true}}`)},
		{parameter: "AllowCustomMemes", value: json.RawMessage(`false`), want: json.RawMessage(`{"funSettings":{"allowCustomMemes":false}}`)},
		{parameter: "AllowCustomMemes", value: json.RawMessage(`true`), want: json.RawMessage(`{"funSettings":{"allowCustomMemes":true}}`)},
		{parameter: "ShowInTeamsSearchAndSuggestions", value: json.RawMessage(`false`), want: json.RawMessage(`{"discoverySettings":{"showInTeamsSearchAndSuggestions":false}}`)},
		{parameter: "ShowInTeamsSearchAndSuggestions", value: json.RawMessage(`true`), want: json.RawMessage(`{"discoverySettings":{"showInTeamsSearchAndSuggestions":true}}`)},
		{parameter: "GiphyContentRating", value: json.RawMessage(`"moderate"`), want: json.RawMessage(`{"funSettings":{"giphyContentRating":"moderate"}}`)},
		{parameter: "GiphyContentRating", value: json.RawMessage(`"strict"`), want: json.RawMessage(`{"funSettings":{"giphyContentRating":"strict"}}`)},
	} {
		t.Run(row.parameter+"/"+string(row.value), func(t *testing.T) {
			var input team.Settings
			if setter, ok := booleanSettings[row.parameter]; ok {
				var value bool
				if json.Unmarshal(row.value, &value) != nil {
					t.Fatal("invalid boolean test vector")
				}
				input = setter(&value)
			} else {
				var value team.GiphyContentRating
				if json.Unmarshal(row.value, &value) != nil {
					t.Fatal("invalid rating test vector")
				}
				input = team.Settings{FunSettings: &team.FunSettings{GiphyContentRating: &value}}
			}
			actual, err := json.Marshal(input)
			var got, want map[string]any
			if err != nil || json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.want, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("typed settings input changed field location, value or omission")
			}
		})
	}
}
