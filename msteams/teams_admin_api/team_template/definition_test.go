package teamtemplate_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

type definitionCase struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
	Wire  json.RawMessage `json:"wire"`
}

func definitionCases() []definitionCase {
	return []definitionCase{
		{Name: "template-expanded-sparse", Input: json.RawMessage("{\"DisplayName\":\"Fixture\",\"ShortDescription\":\"Fixture summary\"}"), Wire: json.RawMessage("{\"displayName\":\"Fixture\",\"shortDescription\":\"Fixture summary\"}")},
		{Name: "template-expanded-empty-arrays-false", Input: json.RawMessage("{\"App\":[],\"Category\":[],\"Channel\":[],\"DisplayName\":\"Fixture\",\"IsMembershipLimitedToOwner\":false,\"ShortDescription\":\"Fixture summary\"}"), Wire: json.RawMessage("{\"apps\":[],\"categories\":[],\"channels\":[],\"displayName\":\"Fixture\",\"isMembershipLimitedToOwners\":false,\"shortDescription\":\"Fixture summary\"}")},
		{Name: "template-json-body-default-settings", Input: json.RawMessage("\"{\\\"displayName\\\":\\\"Fixture\\\",\\\"shortDescription\\\":\\\"Fixture summary\\\"}\""), Wire: json.RawMessage("{\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"displayName\":\"Fixture\",\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false},\"shortDescription\":\"Fixture summary\"}")},
		{Name: "template-json-flattened-precedence", Input: json.RawMessage("\"{\\\"teamDefinition\\\":{\\\"displayName\\\":\\\"Definition\\\",\\\"channels\\\":[]},\\\"extendedProperties\\\":{\\\"displayName\\\":\\\"Extended\\\",\\\"shortDescription\\\":\\\"Summary\\\"}}\""), Wire: json.RawMessage("{\"channels\":[],\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"displayName\":\"Definition\",\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false},\"shortDescription\":\"Summary\"}")},
		{Name: "template-json-tab-configuration-omitted", Input: json.RawMessage("\"{\\\"displayName\\\":\\\"Fixture\\\",\\\"shortDescription\\\":\\\"Summary\\\",\\\"channels\\\":[{\\\"displayName\\\":\\\"General\\\",\\\"tabs\\\":[{\\\"name\\\":\\\"Fixture tab\\\"}]}]}\""), Wire: json.RawMessage("{\"channels\":[{\"displayName\":\"General\",\"tabs\":[{\"name\":\"Fixture tab\"}]}],\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"displayName\":\"Fixture\",\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false},\"shortDescription\":\"Summary\"}")},
		{Name: "template-json-root-precedence", Input: json.RawMessage("\"{\\\"displayName\\\":\\\"Root\\\",\\\"teamDefinition\\\":{\\\"displayName\\\":\\\"Definition\\\",\\\"shortDescription\\\":\\\"Definition summary\\\"},\\\"extendedProperties\\\":{\\\"displayName\\\":\\\"Extended\\\",\\\"shortDescription\\\":\\\"Extended summary\\\",\\\"description\\\":\\\"Extended description\\\"}}\""), Wire: json.RawMessage("{\"description\":\"Extended description\",\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"displayName\":\"Root\",\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false},\"shortDescription\":\"Definition summary\"}")},
		{Name: "template-json-root-null-masks-nested", Input: json.RawMessage("\"{\\\"displayName\\\":null,\\\"teamDefinition\\\":{\\\"displayName\\\":\\\"Definition\\\"},\\\"extendedProperties\\\":{\\\"displayName\\\":\\\"Extended\\\"}}\""), Wire: json.RawMessage("{\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false}}")},
		{Name: "template-json-null-array-items", Input: json.RawMessage("\"{\\\"apps\\\":[null,{\\\"id\\\":\\\"fixture-app\\\"},null],\\\"categories\\\":[\\\"first\\\",null,\\\"last\\\"],\\\"channels\\\":[null,{\\\"displayName\\\":\\\"General\\\",\\\"tabs\\\":[null,{\\\"name\\\":\\\"Fixture tab\\\"}]}]}\""), Wire: json.RawMessage("{\"apps\":[{\"id\":\"fixture-app\"}],\"categories\":[\"first\",\"last\"],\"channels\":[{\"displayName\":\"General\",\"tabs\":[{\"name\":\"Fixture tab\"}]}],\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false}}")},
		{Name: "template-json-empty-tab-configuration", Input: json.RawMessage("\"{\\\"channels\\\":[{\\\"tabs\\\":[{\\\"configuration\\\":{}}]}]}\""), Wire: json.RawMessage("{\"channels\":[{\"tabs\":[]}],\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false}}")},
		{Name: "template-json-other-property-types", Input: json.RawMessage("\"{\\\"displayName\\\":7,\\\"description\\\":false,\\\"isMembershipLimitedToOwners\\\":\\\"false\\\",\\\"apps\\\":[7,{\\\"id\\\":7}],\\\"categories\\\":[null,5,\\\"last\\\"],\\\"channels\\\":[4,{\\\"displayName\\\":7,\\\"isFavoriteByDefault\\\":\\\"false\\\",\\\"tabs\\\":[false,{\\\"name\\\":7,\\\"configuration\\\":{\\\"entityId\\\":7}}]}],\\\"funSettings\\\":{\\\"allowGiphy\\\":\\\"true\\\",\\\"giphyContentRating\\\":7}}\""), Wire: json.RawMessage("{\"apps\":[],\"categories\":[\"last\"],\"channels\":[{\"tabs\":[]}],\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false}}")},
		{Name: "template-json-null-settings", Input: json.RawMessage("\"{\\\"discoverySettings\\\":null,\\\"funSettings\\\":null,\\\"guestSettings\\\":null,\\\"memberSettings\\\":null,\\\"messagingSettings\\\":null}\""), Wire: json.RawMessage("{\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":false},\"funSettings\":{\"allowCustomMemes\":false,\"allowGiphy\":false,\"allowStickersAndMemes\":false},\"guestSettings\":{\"allowCreateUpdateChannels\":false,\"allowDeleteChannels\":false},\"memberSettings\":{\"allowAddRemoveApps\":false,\"allowCreatePrivateChannels\":false,\"allowCreateUpdateChannels\":false,\"allowCreateUpdateRemoveConnectors\":false,\"allowCreateUpdateRemoveTabs\":false,\"allowDeleteChannels\":false,\"uploadCustomApp\":false},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":false,\"allowTeamMentions\":false,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":false}}")},
		{Name: "template-json-full-definition", Input: json.RawMessage("\"{\\\"templateId\\\":\\\"fixture-id\\\",\\\"classification\\\":\\\"fixture-classification\\\",\\\"displayName\\\":\\\"Fixture\\\",\\\"description\\\":\\\"\\\",\\\"ownerUserObjectId\\\":\\\"fixture-owner\\\",\\\"specialization\\\":\\\"fixture-specialization\\\",\\\"visibility\\\":\\\"Private\\\",\\\"channels\\\":[{\\\"id\\\":\\\"fixture-channel\\\",\\\"displayName\\\":\\\"General\\\",\\\"description\\\":\\\"\\\",\\\"isFavoriteByDefault\\\":false,\\\"tabs\\\":[{\\\"id\\\":\\\"fixture-tab\\\",\\\"teamsAppId\\\":\\\"fixture-app\\\",\\\"name\\\":\\\"Tab\\\",\\\"sortOrderIndex\\\":\\\"0\\\",\\\"messageId\\\":\\\"fixture-message\\\",\\\"webUrl\\\":\\\"https://example.invalid/tab\\\",\\\"key\\\":\\\"fixture-key\\\",\\\"configuration\\\":{\\\"entityId\\\":\\\"fixture-entity\\\",\\\"contentUrl\\\":\\\"https://example.invalid/content\\\",\\\"removeUrl\\\":\\\"\\\",\\\"websiteUrl\\\":\\\"https://example.invalid/site\\\"}}]}],\\\"memberSettings\\\":{\\\"allowCreateUpdateChannels\\\":true,\\\"allowAddRemoveApps\\\":true,\\\"uploadCustomApp\\\":true,\\\"allowCreateUpdateRemoveTabs\\\":true,\\\"allowCreateUpdateRemoveConnectors\\\":true,\\\"allowCreatePrivateChannels\\\":true},\\\"guestSettings\\\":{\\\"allowCreateUpdateChannels\\\":true},\\\"messagingSettings\\\":{\\\"allowUserEditMessages\\\":true,\\\"allowOwnerDeleteMessages\\\":true,\\\"allowTeamMentions\\\":true},\\\"funSettings\\\":{\\\"allowGiphy\\\":true,\\\"allowCustomMemes\\\":true,\\\"giphyContentRating\\\":\\\"strict\\\"},\\\"discoverySettings\\\":{\\\"showInTeamsSearchAndSuggestions\\\":true},\\\"isMembershipLimitedToOwners\\\":false,\\\"apps\\\":[{\\\"id\\\":\\\"fixture-app\\\"}],\\\"icon\\\":\\\"fixture-icon\\\",\\\"shortDescription\\\":\\\"Summary\\\",\\\"categories\\\":[\\\"fixture-category\\\"],\\\"publishedBy\\\":\\\"fixture-publisher\\\",\\\"uri\\\":\\\"fixture-uri\\\",\\\"futureField\\\":false}\""), Wire: json.RawMessage("{\"apps\":[{\"id\":\"fixture-app\"}],\"categories\":[\"fixture-category\"],\"channels\":[{\"description\":\"\",\"displayName\":\"General\",\"id\":\"fixture-channel\",\"isFavoriteByDefault\":false,\"tabs\":[{\"configuration\":{\"contentUrl\":\"https://example.invalid/content\",\"entityId\":\"fixture-entity\",\"removeUrl\":\"\",\"websiteUrl\":\"https://example.invalid/site\"},\"id\":\"fixture-tab\",\"key\":\"fixture-key\",\"messageId\":\"fixture-message\",\"name\":\"Tab\",\"sortOrderIndex\":\"0\",\"teamsAppId\":\"fixture-app\",\"webUrl\":\"https://example.invalid/tab\"}]}],\"classification\":\"fixture-classification\",\"description\":\"\",\"discoverySettings\":{\"showInTeamsSearchAndSuggestions\":true},\"displayName\":\"Fixture\",\"funSettings\":{\"allowCustomMemes\":true,\"allowGiphy\":true,\"allowStickersAndMemes\":false,\"giphyContentRating\":\"strict\"},\"guestSettings\":{\"allowCreateUpdateChannels\":true,\"allowDeleteChannels\":false},\"icon\":\"fixture-icon\",\"isMembershipLimitedToOwners\":false,\"memberSettings\":{\"allowAddRemoveApps\":true,\"allowCreatePrivateChannels\":true,\"allowCreateUpdateChannels\":true,\"allowCreateUpdateRemoveConnectors\":true,\"allowCreateUpdateRemoveTabs\":true,\"allowDeleteChannels\":false,\"uploadCustomApp\":true},\"messagingSettings\":{\"allowChannelMentions\":false,\"allowOwnerDeleteMessages\":true,\"allowTeamMentions\":true,\"allowUserDeleteMessages\":false,\"allowUserEditMessages\":true},\"ownerUserObjectId\":\"fixture-owner\",\"publishedBy\":\"fixture-publisher\",\"shortDescription\":\"Summary\",\"specialization\":\"fixture-specialization\",\"templateId\":\"fixture-id\",\"uri\":\"fixture-uri\",\"visibility\":\"Private\"}")},
	}
}

func equalJSON(t *testing.T, expected, actual []byte) {
	t.Helper()
	var want, got any
	if json.Unmarshal(expected, &want) != nil || json.Unmarshal(actual, &got) != nil || !reflect.DeepEqual(want, got) {
		t.Fatal("JSON differs from expected local value; no values printed")
	}
}

func TestUnitNewDefinition(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, display, short string
		invalid              bool
	}{
		{name: "missing display name", short: "Summary", invalid: true},
		{name: "missing short description", display: "Fixture", invalid: true},
		{name: "sparse expanded", display: "Fixture", short: "Summary"},
		{name: "mandatory string whitespace", display: " ", short: " "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value, err := template.NewDefinition(tc.display, tc.short)
			if tc.invalid {
				if !errors.Is(err, template.ErrInput) {
					t.Fatal("missing expanded argument accepted")
				}
				return
			}
			if err != nil || value.DisplayName == nil || *value.DisplayName != tc.display || value.ShortDescription == nil || *value.ShortDescription != tc.short || value.MemberSettings != nil {
				t.Fatal("expanded construction changed input or initialized omitted settings")
			}
		})
	}
}

func TestUnitDefinitionSerializationVectors(t *testing.T) {
	t.Parallel()
	selected := 0
	for _, fixture := range definitionCases() {
		if !strings.HasPrefix(fixture.Name, "template-json-") && !strings.HasPrefix(fixture.Name, "template-expanded-") {
			continue
		}
		selected++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			var value *template.Definition
			if strings.HasPrefix(fixture.Name, "template-json-") {
				var input string
				if json.Unmarshal(fixture.Input, &input) != nil {
					t.Fatal("model fixture input is not JSON text")
				}
				var err error
				value, err = template.ParseDefinition([]byte(input))
				if err != nil {
					t.Fatal("Go Body parser rejected model input")
				}
			} else {
				definition, err := template.NewDefinition("Fixture", "Fixture summary")
				if err != nil {
					t.Fatal(err)
				}
				value = &definition
				if fixture.Name == "template-expanded-empty-arrays-false" {
					no := false
					value.Apps, value.Categories, value.Channels = []*template.App{}, []string{}, []*template.Channel{}
					value.IsMembershipLimitedToOwners = &no
				}
			}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, fixture.Wire, raw)
		})
	}
	if selected < 12 {
		t.Fatal("expected template vectors disappeared")
	}
}

func TestUnitParseDefinitionRejectsNonObjectsAndResetsReceiver(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"", "null", "[]", "true", `"text"`, "{"} {
		t.Run("invalid "+raw, func(t *testing.T) {
			t.Parallel()
			if _, err := template.ParseDefinition([]byte(raw)); !errors.Is(err, template.ErrInput) {
				t.Fatal("invalid Body root accepted")
			}
		})
	}
	var value template.Definition
	if json.Unmarshal([]byte(`{"displayName":"First","apps":[{"id":"fixture"}]}`), &value) != nil || json.Unmarshal([]byte(`{}`), &value) != nil {
		t.Fatal("valid Body input rejected")
	}
	if value.DisplayName != nil || value.Apps != nil || value.MemberSettings == nil {
		t.Fatal("Body decoding retained prior values or lost default settings")
	}
}
