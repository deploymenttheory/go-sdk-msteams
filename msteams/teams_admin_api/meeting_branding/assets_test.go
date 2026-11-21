package meetingbranding_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

func TestBrandingLocalBuildersMatchPinnedSetters(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + assetSource)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Kind, URI string
			Rejected        bool
			Wire            json.RawMessage
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 31 {
		t.Fatal("invalid pinned asset capture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		command := "New-CsTeamsMeetingBrandingTheme"
		if f.Kind == "image" {
			command = "New-CsTeamsMeetingBackgroundImage"
		}
		if f.Kind == "slate" {
			command = "New-CsTeamsNdiAssuranceSlate"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Source: assetSource, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Want: contracttest.Expected{JSON: f.Wire}, Invoke: func(context.Context, *msteams.Client) contracttest.Observation {
				switch f.Kind {
				case "image":
					v, e := resource.NewBackgroundImage("fixture-image", -1, "Fixture", f.URI, resource.BackgroundImageOptions{})
					return contracttest.Observation{Value: v, Err: e}
				case "slate":
					v, e := resource.NewAssuranceSlate("fixture-slate", "Fixture", f.URI, resource.AssuranceSlateOptions{})
					return contracttest.Observation{Value: v, Err: e}
				case "reaction":
					v, e := resource.NewReaction("fixture-reaction", f.URI, "FutureReaction", resource.ReactionOptions{})
					return contracttest.Observation{Value: v, Err: e}
				default:
					logo, color, alignment, no := "relative-or-future", "future-color", "FutureAlignment", false
					v, e := resource.NewBrandingTheme("fixture-theme", resource.ThemeOptions{ThemeFields: resource.ThemeFields{LogoImageLightURI: &logo, BrandAccentColor: &color, ChatBackgroundImageAlignment: &alignment, EnableChatBackground: &no, Enabled: &no, DisableOriginalReactions: &no}})
					return contracttest.Observation{Value: v, Err: e}
				}
			}}
			if f.Rejected {
				c.Want.JSON, c.Want.Error = nil, resource.ErrInput
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestBrandingObjectEditsMatchPinnedCoreExtension(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Lists []struct {
			Kind, Name  string
			Empty       bool
			Observation struct {
				Rejected bool
				IDs      []*string `json:"ids"`
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Lists) != 30 {
		t.Fatal("invalid object modifier capture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Lists {
		command := "Set-CsTeamsMeetingBrandingPolicy"
		if f.Kind == "reaction" {
			command = "New-CsTeamsMeetingBrandingTheme"
		}
		name := f.Kind + "-" + f.Name + "-populated"
		if f.Empty {
			name = f.Kind + "-" + f.Name + "-empty"
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Invoke: func(context.Context, *msteams.Client) contracttest.Observation {
			var ids []*string
			var err error
			if f.Kind == "policy" {
				a, _ := resource.NewBackgroundImage("a", 0, "Fixture", "https://example.invalid/a", resource.BackgroundImageOptions{})
				b, _ := resource.NewBackgroundImage("b", 1, "Fixture", "https://example.invalid/b", resource.BackgroundImageOptions{})
				current := []*resource.BackgroundImage{a, a, b}
				if f.Empty {
					current = nil
				}
				var edit *resource.AssetEdit[resource.BackgroundImage]
				switch f.Name {
				case "null", "empty-replace":
					edit, err = resource.ReplaceAssets[resource.BackgroundImage]()
				case "empty-delta":
					edit, err = resource.ChangeAssets[resource.BackgroundImage](nil, nil)
				case "replace":
					edit, err = resource.ReplaceAssets(b, b)
				case "add":
					edit, err = resource.ChangeAssets([]*resource.BackgroundImage{a}, nil)
				case "null-item":
					edit, err = resource.ReplaceAssets[resource.BackgroundImage](nil)
				case "remove-missing":
					missing, _ := resource.NewBackgroundImage("missing", 0, "Fixture", "https://example.invalid/a", resource.BackgroundImageOptions{})
					edit, err = resource.ChangeAssets(nil, []*resource.BackgroundImage{missing})
				default:
					edit, err = resource.ChangeAssets(nil, []*resource.BackgroundImage{a})
				}
				if err != nil {
					return contracttest.Observation{Err: err}
				}
				list, e := edit.Apply(current)
				err = e
				if e == nil {
					raw, e := json.Marshal(list)
					var values []*resource.BackgroundImage
					if e != nil || json.Unmarshal(raw, &values) != nil {
						return contracttest.Observation{Err: errors.New("invalid materialized assets")}
					}
					ids = []*string{}
					for _, value := range values {
						var id *string
						if value != nil {
							id = value.ID
						}
						ids = append(ids, id)
					}
				}
			} else {
				a, _ := resource.NewReaction("a", "https://example.invalid/a", "Like", resource.ReactionOptions{})
				display := "different"
				different, _ := resource.NewReaction("a", "https://example.invalid/a", "Like", resource.ReactionOptions{BrandedDisplayName: &display})
				b, _ := resource.NewReaction("b", "https://example.invalid/b", "Future", resource.ReactionOptions{})
				current := []*resource.Reaction{a, a, different}
				if f.Empty {
					current = nil
				}
				var edit *resource.AssetEdit[resource.Reaction]
				switch f.Name {
				case "null", "empty-replace":
					edit, err = resource.ReplaceAssets[resource.Reaction]()
				case "empty-delta":
					edit, err = resource.ChangeAssets[resource.Reaction](nil, nil)
				case "replace":
					edit, err = resource.ReplaceAssets(b, b)
				case "add":
					edit, err = resource.ChangeAssets([]*resource.Reaction{a}, nil)
				case "remove-different":
					edit, err = resource.ChangeAssets(nil, []*resource.Reaction{different})
				default:
					edit, err = resource.ChangeAssets(nil, []*resource.Reaction{a})
				}
				if err != nil {
					return contracttest.Observation{Err: err}
				}
				list, e := edit.Apply(current)
				err = e
				if e == nil {
					raw, e := json.Marshal(list)
					var values []*resource.Reaction
					if e != nil || json.Unmarshal(raw, &values) != nil {
						return contracttest.Observation{Err: errors.New("invalid materialized reactions")}
					}
					ids = []*string{}
					for _, v := range values {
						id := *v.ID
						if v.BrandedDisplayName != nil {
							id += ":" + *v.BrandedDisplayName
						}
						ids = append(ids, &id)
					}
				}
			}
			if (err != nil) != f.Observation.Rejected || err == nil && !reflect.DeepEqual(ids, f.Observation.IDs) {
				return contracttest.Observation{Err: errors.New("pinned object-list behavior changed")}
			}
			return contracttest.Observation{}
		}}))
	}
	evidence(t, results)
}

func TestBrandingAssetsOwnPointersAndProjectKnownFields(t *testing.T) {
	t.Parallel()
	no, future := false, "FutureAlignment"
	theme, err := resource.NewBrandingTheme("fixture-theme", resource.ThemeOptions{ThemeFields: resource.ThemeFields{Enabled: &no, ChatBackgroundImageAlignment: &future}})
	if err != nil {
		t.Fatal(err)
	}
	no, future = true, "changed"
	if theme.Enabled == nil || *theme.Enabled || *theme.ChatBackgroundImageAlignment != "FutureAlignment" {
		t.Fatal("local builder retained mutable option pointers")
	}
	theme.RawFields = map[string]json.RawMessage{"Future": json.RawMessage("true")}
	list, err := resource.Assets(theme)
	if err != nil {
		t.Fatal(err)
	}
	*theme.Enabled = true
	raw, err := json.Marshal(list)
	var projected []map[string]json.RawMessage
	if err != nil || json.Unmarshal(raw, &projected) != nil || string(projected[0]["Enabled"]) != "false" || projected[0]["Future"] != nil || string(projected[0]["MeetingReactions"]) != "[]" {
		t.Fatal("request ownership or known nested projection changed")
	}
}
