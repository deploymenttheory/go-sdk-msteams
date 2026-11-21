package meetingbranding_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

func TestUnitPinnedNestedBrandingProjection(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-model-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Cases []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &capture) != nil {
		t.Fatal("invalid pinned branding observations")
	}
	seen := 0
	for _, c := range capture.Cases {
		if c.Name != "provider-branding-empty-lists" && c.Name != "provider-branding-nested-lists" {
			continue
		}
		seen++
		var p resource.Policy
		if json.Unmarshal(c.Wire, &p) != nil || p.Identity == nil || *p.Identity != "tag:Fixture" {
			t.Fatal("pinned branding policy projection differs")
		}
		encoded, err := json.Marshal(p)
		var original, roundtrip any
		if err != nil || json.Unmarshal(c.Wire, &original) != nil || json.Unmarshal(encoded, &roundtrip) != nil || !reflect.DeepEqual(original, roundtrip) {
			t.Fatal("pinned branding snapshot changed")
		}
		if c.Name == "provider-branding-empty-lists" {
			if p.MeetingBackgroundImages == nil || p.MeetingBrandingThemes == nil || p.NDIAssuranceSlateImages == nil || len(p.MeetingBackgroundImages)+len(p.MeetingBrandingThemes)+len(p.NDIAssuranceSlateImages) != 0 {
				t.Fatal("explicit empty asset arrays lost")
			}
			continue
		}
		if len(p.MeetingBackgroundImages) != 1 || len(p.MeetingBrandingThemes) != 1 || len(p.NDIAssuranceSlateImages) != 1 {
			t.Fatal("nested branding assets lost")
		}
		image, theme, slate := p.MeetingBackgroundImages[0], p.MeetingBrandingThemes[0], p.NDIAssuranceSlateImages[0]
		if image == nil || image.Order == nil || *image.Order != 0 || image.IsRequired == nil || *image.IsRequired || image.IsHidden != nil || theme == nil || theme.Enabled == nil || *theme.Enabled || theme.MeetingReactions == nil || len(theme.MeetingReactions) != 0 || slate == nil || slate.IsDefault == nil || *slate.IsDefault {
			t.Fatal("nullable/zero/default nested data changed")
		}
	}
	if seen != 2 {
		t.Fatal("pinned branding observations missing")
	}
}

func TestUnitBrandingSnapshotsPreserveFuturePresenceAndRedactFormatting(t *testing.T) {
	t.Parallel()
	const raw = `{"Identity":"fixture-private","MeetingBackgroundImages":[{"Id":"image","Order":-1,"Future":{"Enabled":false}}],"MeetingBrandingThemes":[{"Identity":"theme","Enabled":null,"MeetingReactions":[{"Id":"reaction","ReactionType":"Future","ReactionSpriteImageUrl":null,"Future":0}]}],"Future":[]}`
	var p resource.Policy
	if json.Unmarshal([]byte(raw), &p) != nil || p.EnableMeetingOptionsThemeOverride != nil || p.NDIAssuranceSlateImages != nil || p.DefaultTheme != nil {
		t.Fatal("omitted policy fields acquired defaults")
	}
	image, theme := p.MeetingBackgroundImages[0], p.MeetingBrandingThemes[0]
	if *image.Order != -1 || string(image.RawFields["Future"]) != `{"Enabled":false}` || theme.Enabled != nil || len(theme.MeetingReactions) != 1 || string(theme.MeetingReactions[0].RawFields["Future"]) != "0" {
		t.Fatal("nested future/nullable fields lost")
	}
	encoded, err := json.Marshal(p)
	var a, b any
	if err != nil || json.Unmarshal([]byte(raw), &a) != nil || json.Unmarshal(encoded, &b) != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("future snapshot round trip differs")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, p), "fixture-private") {
			t.Fatal("snapshot formatting exposed tenant values")
		}
	}
}

func TestUnitPinnedAssetConstructorSnapshots(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/branding-asset-models.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Cases []struct {
			Kind     string          `json:"kind"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &capture) != nil || len(capture.Cases) != 31 {
		t.Fatal("invalid pinned asset observations")
	}
	accepted := 0
	for _, c := range capture.Cases {
		if c.Rejected {
			// Captured setter validation is input for the pending local builders;
			// snapshot decoding does not claim to implement that validation.
			continue
		}
		accepted++
		var value any
		switch c.Kind {
		case "image":
			value = new(resource.BackgroundImage)
		case "slate":
			value = new(resource.AssuranceSlate)
		case "reaction":
			value = new(resource.Reaction)
		case "theme":
			value = new(resource.BrandingTheme)
		default:
			t.Fatal("unknown captured asset kind")
		}
		if json.Unmarshal(c.Wire, value) != nil {
			t.Fatal("pinned asset could not be decoded")
		}
		encoded, err := json.Marshal(value)
		var want, got any
		if err != nil || json.Unmarshal(c.Wire, &want) != nil || json.Unmarshal(encoded, &got) != nil || !reflect.DeepEqual(want, got) {
			t.Fatal("pinned asset projection changed")
		}
	}
	if accepted != 16 {
		t.Fatal("pinned accepted asset count differs")
	}
}
