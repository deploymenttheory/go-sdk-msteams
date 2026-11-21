package teammanagement_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const settingsSource = "contracts/endpoints/8.1.0/fixtures/core-team-settings-fields.json"

// Each typed input selects one setting independently, so moving, omitting or
// inverting a flag fails against the actual pinned ToTeam serializer output.
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

func TestSettingsInputsCoverBothPinnedApplicationParameterSets(t *testing.T) {
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

func TestCreationAndUpdatePreserveEveryPinnedSettingsField(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../" + settingsSource)
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Version       string `json:"baseline_version"`
		Qualification string
		Cases         []struct {
			Parameter string
			Value     json.RawMessage
			Wire      map[string]map[string]any
		}
	}
	if json.Unmarshal(data, &capture) != nil || capture.Version != "8.1.0" || capture.Qualification != "offline-model-serialization" || len(capture.Cases) != 36 {
		t.Fatal("complete pinned settings-field capture unavailable")
	}
	seen := map[string]bool{}
	var creates, updates []contracttest.Result
	for index, row := range capture.Cases {
		var input team.Settings
		var label string
		if setter, ok := booleanSettings[row.Parameter]; ok {
			var value bool
			if json.Unmarshal(row.Value, &value) != nil {
				t.Fatal("boolean field capture invalid")
			}
			input, label = setter(&value), strconv.FormatBool(value)
		} else if row.Parameter == "GiphyContentRating" {
			var value team.GiphyContentRating
			if json.Unmarshal(row.Value, &value) != nil {
				t.Fatal("rating capture invalid")
			}
			wantNumeric := float64(1)
			if value == team.GiphyStrict {
				wantNumeric = 0
			} else if value != team.GiphyModerate {
				t.Fatal("unknown pinned rating")
			}
			if row.Wire["funSettings"]["giphyContentRating"] != wantNumeric {
				t.Fatal("pinned numeric rating observation changed")
			}
			// The reviewed native model follows Graph's string contract. Keep the
			// immutable capture numeric; change only this comparison copy.
			row.Wire["funSettings"]["giphyContentRating"] = string(value)
			input, label = team.Settings{FunSettings: &team.FunSettings{GiphyContentRating: &value}}, string(value)
		} else {
			t.Fatal("unmapped captured field")
		}
		name := strings.ToLower(row.Parameter) + "-" + label
		if seen[name] {
			t.Fatal("duplicate field/value observation")
		}
		seen[name] = true
		encoded, err := json.Marshal(row.Wire)
		if err != nil {
			t.Fatal(err)
		}
		for _, create := range []bool{true, false} {
			t.Run(map[bool]string{true: "create", false: "update"}[create]+"/"+name, func(t *testing.T) {
				command, fixtureName, writeIndex := "Set-Team", "team-update-readback", 1
				if create {
					command, fixtureName, writeIndex = "New-Team", "team-create-owned-group", 2
				}
				steps := fixture(t, fixtureName)
				steps[writeIndex].Request.JSON = encoded
				var readback map[string]any
				if json.Unmarshal(steps[len(steps)-1].Reply.Body, &readback) != nil {
					t.Fatal("paired team fixture unavailable")
				}
				for container, properties := range row.Wire {
					readback[container] = properties
				}
				steps[len(steps)-1].Reply.Body, err = json.Marshal(readback)
				if err != nil {
					t.Fatal(err)
				}
				c := contracttest.Case{Name: map[bool]string{true: "create-", false: "update-"}[create] + name,
					Source: settingsSource + "#/cases/" + strconv.Itoa(index), Steps: steps,
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						var result management.Result
						var callErr error
						if create {
							request := creationRequest()
							request.Settings = input
							created, err := sdk.GraphAPI.TeamManagement.Create(ctx, request)
							result, callErr = created.Result, err
						} else {
							empty := ""
							result, callErr = sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, management.UpdateRequest{Properties: group.UpdateRequest{Description: &empty}, Settings: input})
						}
						var value any
						if result.Value != nil {
							value = result.Value.Settings
						}
						return contracttest.Observation{Value: value, Response: result.Response, Err: callErr}
					}, Want: contracttest.Expected{Status: 200, JSON: encoded}}
				if create {
					c.Mappings = []contracttest.Mapping{{Command: "New-Team", ParameterSet: "CreateTeam"}}
					creates = append(creates, runCase(t, command, c))
				} else {
					updates = append(updates, runCase(t, command, c))
				}
			})
		}
	}
	evidence(t, "team-create-settings-service.json", creates)
	evidence(t, "team-update-settings-service.json", updates)
}
