package meetingbranding_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

func TestUnitBrandingLocalBuildersMatchTypedSetters(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name, Kind, URI string
			Rejected        bool
			Wire            json.RawMessage
		}
	}{
		Cases: []struct {
			Name, Kind, URI string
			Rejected        bool
			Wire            json.RawMessage
		}{

			{
				Name:     "image-https",
				Kind:     "image",
				URI:      "https://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/image.png\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Name:     "image-uppercase",
				Kind:     "image",
				URI:      "HTTPS://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"HTTPS://example.invalid/image.png\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Name:     "image-http",
				Kind:     "image",
				URI:      "http://example.invalid/image.png",
				Rejected: true,
			},

			{
				Name:     "image-relative",
				Kind:     "image",
				URI:      "image.png",
				Rejected: true,
			},

			{
				Name:     "image-empty",
				Kind:     "image",
				URI:      "",
				Rejected: true,
			},

			{
				Name:     "image-empty-https",
				Kind:     "image",
				URI:      "https://",
				Rejected: true,
			},

			{
				Name:     "image-space",
				Kind:     "image",
				URI:      "https:// ",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https:// \",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Name:     "image-over-1024",
				Kind:     "image",
				URI:      "https://example.invalid/" + strings.Repeat("a", 2048),
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Name:     "image-final-newline",
				Kind:     "image",
				URI:      "https://example.invalid/image.png\n",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/image.png\\n\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Name:     "image-interior-newline",
				Kind:     "image",
				URI:      "https://example.invalid/image\n.png",
				Rejected: true,
			},

			{
				Name:     "slate-https",
				Kind:     "slate",
				URI:      "https://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/image.png\"}"),
			},

			{
				Name:     "slate-uppercase",
				Kind:     "slate",
				URI:      "HTTPS://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"HTTPS://example.invalid/image.png\"}"),
			},

			{
				Name:     "slate-http",
				Kind:     "slate",
				URI:      "http://example.invalid/image.png",
				Rejected: true,
			},

			{
				Name:     "slate-relative",
				Kind:     "slate",
				URI:      "image.png",
				Rejected: true,
			},

			{
				Name:     "slate-empty",
				Kind:     "slate",
				URI:      "",
				Rejected: true,
			},

			{
				Name:     "slate-empty-https",
				Kind:     "slate",
				URI:      "https://",
				Rejected: true,
			},

			{
				Name:     "slate-space",
				Kind:     "slate",
				URI:      "https:// ",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https:// \"}"),
			},

			{
				Name:     "slate-over-1024",
				Kind:     "slate",
				URI:      "https://example.invalid/" + strings.Repeat("a", 2048),
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\"}"),
			},

			{
				Name:     "slate-final-newline",
				Kind:     "slate",
				URI:      "https://example.invalid/image.png\n",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/image.png\\n\"}"),
			},

			{
				Name:     "slate-interior-newline",
				Kind:     "slate",
				URI:      "https://example.invalid/image\n.png",
				Rejected: true,
			},

			{
				Name:     "reaction-https",
				Kind:     "reaction",
				URI:      "https://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/image.png\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Name:     "reaction-uppercase",
				Kind:     "reaction",
				URI:      "HTTPS://example.invalid/image.png",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"HTTPS://example.invalid/image.png\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Name:     "reaction-http",
				Kind:     "reaction",
				URI:      "http://example.invalid/image.png",
				Rejected: true,
			},

			{
				Name:     "reaction-relative",
				Kind:     "reaction",
				URI:      "image.png",
				Rejected: true,
			},

			{
				Name:     "reaction-empty",
				Kind:     "reaction",
				URI:      "",
				Rejected: true,
			},

			{
				Name:     "reaction-empty-https",
				Kind:     "reaction",
				URI:      "https://",
				Rejected: true,
			},

			{
				Name:     "reaction-space",
				Kind:     "reaction",
				URI:      "https:// ",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https:// \",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Name:     "reaction-over-1024",
				Kind:     "reaction",
				URI:      "https://example.invalid/" + strings.Repeat("a", 2048),
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Name:     "reaction-final-newline",
				Kind:     "reaction",
				URI:      "https://example.invalid/image.png\n",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/image.png\\n\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Name:     "reaction-interior-newline",
				Kind:     "reaction",
				URI:      "https://example.invalid/image\n.png",
				Rejected: true,
			},

			{
				Name:     "theme-open-fields",
				Kind:     "theme",
				Rejected: false,
				Wire:     json.RawMessage("{\"BackgroundImageDarkPreAuthUri\":null,\"BackgroundImageDarkUri\":null,\"BackgroundImageLightPreAuthUri\":null,\"BackgroundImageLightUri\":null,\"BrandAccentColor\":\"future-color\",\"ChatBackgroundImageAlignment\":\"FutureAlignment\",\"ChatBackgroundImageDarkUri\":null,\"ChatBackgroundImageLightUri\":null,\"DisableOriginalReactions\":false,\"DisplayName\":null,\"EnableChatBackground\":false,\"Enabled\":false,\"Identity\":\"fixture-theme\",\"LogoImageDarkPreAuthUri\":null,\"LogoImageDarkUri\":null,\"LogoImageLightPreAuthUri\":null,\"LogoImageLightUri\":\"relative-or-future\",\"MeetingInviteLogoImageDarkPreAuthUri\":null,\"MeetingInviteLogoImageLightPreAuthUri\":null,\"MeetingReactions\":[]}"),
			},
		},
	}

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
			run(t, c)
		}
	}
}

func TestUnitBrandingObjectEditsMatchTypedCoreExtension(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Lists []struct {
			Kind, Name  string
			Empty       bool
			Observation struct {
				Rejected bool
				IDs      []*string `json:"ids"`
			}
		}
	}{
		Lists: []struct {
			Kind, Name  string
			Empty       bool
			Observation struct {
				Rejected bool
				IDs      []*string `json:"ids"`
			}
		}{

			{
				Kind:  "policy",
				Name:  "null",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "empty-replace",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "replace",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("b"),
						new("b"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "empty-delta",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "add",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
						new("a"),
						new("b"),
						new("a"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "remove",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: true,
					IDs: []*string{
						new("a"),
						new("a"),
						new("b"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "remove-missing",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: true,
					IDs: []*string{
						new("a"),
						new("a"),
						new("b"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "null-item",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						nil,
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "null",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "empty-replace",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "replace",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("b"),
						new("b"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "empty-delta",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "add",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
					},
				},
			},

			{
				Kind:  "policy",
				Name:  "remove",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "remove-missing",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "policy",
				Name:  "null-item",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						nil,
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "null",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "empty-replace",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "replace",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("b"),
						new("b"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "empty-delta",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "add",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
						new("a"),
						new("a:different"),
						new("a"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "remove",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
						new("a:different"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "remove-different",
				Empty: false,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
						new("a"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "null",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "empty-replace",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "replace",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("b"),
						new("b"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "empty-delta",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "add",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs: []*string{
						new("a"),
					},
				},
			},

			{
				Kind:  "reaction",
				Name:  "remove",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},

			{
				Kind:  "reaction",
				Name:  "remove-different",
				Empty: true,
				Observation: struct {
					Rejected bool
					IDs      []*string `json:"ids"`
				}{
					Rejected: false,
					IDs:      []*string{},
				},
			},
		},
	}

	for _, f := range fixture.Lists {
		command := "Set-CsTeamsMeetingBrandingPolicy"
		if f.Kind == "reaction" {
			command = "New-CsTeamsMeetingBrandingTheme"
		}
		name := f.Kind + "-" + f.Name + "-populated"
		if f.Empty {
			name = f.Kind + "-" + f.Name + "-empty"
		}
		run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Invoke: func(context.Context, *msteams.Client) contracttest.Observation {
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
		}})
	}
}

func TestUnitBrandingAssetsOwnPointersAndProjectKnownFields(t *testing.T) {
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
