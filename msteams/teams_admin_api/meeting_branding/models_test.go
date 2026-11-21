package meetingbranding_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

func TestUnitTypedNestedBrandingProjection(t *testing.T) {
	t.Parallel()

	inputVectors := struct {
		Cases []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		} `json:"cases"`
	}{
		Cases: []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		}{

			{
				Name: "appids-omitted",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-scalar",
				Wire: json.RawMessage("{\"AppIds\":[\"fixture-app\"],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-replacement",
				Wire: json.RawMessage("{\"AppIds\":[\"fixture-a\",\"fixture-b\"],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-empty-replacement",
				Wire: json.RawMessage("{\"AppIds\":[],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-null-replacement",
				Wire: json.RawMessage("{\"AppIds\":[],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-operators",
				Wire: json.RawMessage("{\"AppIds\":{\"Add\":[\"fixture-a\"],\"Remove\":[\"fixture-b\"]},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-empty-operators",
				Wire: json.RawMessage("{\"AppIds\":{},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-legacy-object-replacement",
				Wire: json.RawMessage("{\"AppIds\":[{\"AppId\":\"fixture-app\",\"Enabled\":false}],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-legacy-object-operators",
				Wire: json.RawMessage("{\"AppIds\":{\"Add\":[{\"AppId\":\"fixture-app\",\"Enabled\":false}]},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-replace-operator-rejected",
			},

			{
				Name: "appids-numeric-item-rejected",
			},

			{
				Name: "appids-ps-list-modifier-rejected",
			},

			{
				Name: "migration-request-field-casing",
				Wire: json.RawMessage("{\"authority\":\"Tenant\",\"groupid\":\"00000000-0000-4000-8000-000000000001\",\"policyname\":\"Fixture\",\"policytype\":\"FixturePolicy\"}"),
			},

			{
				Name: "migration-state-NotStarted",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"NotStarted\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-PopulationInProgress",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"PopulationInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-GpaCreationInProgress",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"GpaCreationInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-WaitingForGpaProcessed",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"WaitingForGpaProcessed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-CleanUpInProgress",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"CleanUpInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-Completed",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"Completed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-Failed",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"Failed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "migration-state-ConnectionLost",
				Wire: json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"ConnectionLost\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "provider-bound-omitted",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-bound-empty-and-false",
				Wire: json.RawMessage("{\"Description\":\"\",\"EnableFederationAccess\":false,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-OrganizationDefault",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":0,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-AllowAllExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":1,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-AllowSpecificExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":2,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-BlockSpecificExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":3,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-BlockAllExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":4,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-voicemail-timespan",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MaximumRecordingLength\":\"00:01:30\"}"),
			},

			{
				Name: "provider-voicemail-explicit-null",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MaximumRecordingLength\":null}"),
			},

			{
				Name: "provider-branding-empty-lists",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MeetingBackgroundImages\":[],\"MeetingBrandingThemes\":[],\"NdiAssuranceSlateImages\":[]}"),
			},

			{
				Name: "provider-branding-nested-lists",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MeetingBackgroundImages\":[{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/background.png\",\"IsHidden\":null,\"IsRequired\":false,\"Name\":\"Fixture\",\"Order\":0}],\"MeetingBrandingThemes\":[{\"BackgroundImageDarkPreAuthUri\":null,\"BackgroundImageDarkUri\":null,\"BackgroundImageLightPreAuthUri\":null,\"BackgroundImageLightUri\":null,\"BrandAccentColor\":null,\"ChatBackgroundImageAlignment\":null,\"ChatBackgroundImageDarkUri\":null,\"ChatBackgroundImageLightUri\":null,\"DisableOriginalReactions\":null,\"DisplayName\":null,\"EnableChatBackground\":null,\"Enabled\":false,\"Identity\":\"fixture-theme\",\"LogoImageDarkPreAuthUri\":null,\"LogoImageDarkUri\":null,\"LogoImageLightPreAuthUri\":null,\"LogoImageLightUri\":null,\"MeetingInviteLogoImageDarkPreAuthUri\":null,\"MeetingInviteLogoImageLightPreAuthUri\":null,\"MeetingReactions\":[]}],\"NdiAssuranceSlateImages\":[{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":false,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/slate.png\"}]}"),
			},

			{
				Name: "provider-user-clear",
				Wire: json.RawMessage("{\"additionalParameters\":null,\"policyName\":null}"),
			},

			{
				Name: "provider-tenant-clear",
				Wire: json.RawMessage("{\"PolicyName\":null,\"additionalParameters\":null,\"forceSwitchPresent\":false}"),
			},

			{
				Name: "provider-group-clear-zero-rank",
				Wire: json.RawMessage("{\"policyName\":null,\"priority\":0}"),
			},

			{
				Name: "provider-identity-tagged",
				Wire: json.RawMessage("\"tag%3AFixture%20/%20Policy\""),
			},

			{
				Name: "provider-identity-nested",
				Wire: json.RawMessage("\"parent/Fixture%20%2F%20Child\""),
			},
		},
	}

	seen := 0
	for _, c := range inputVectors.Cases {
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

func TestUnitTypedAssetConstructorSnapshots(t *testing.T) {
	t.Parallel()

	inputVectors := struct {
		Cases []struct {
			Kind     string          `json:"kind"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		} `json:"cases"`
	}{
		Cases: []struct {
			Kind     string          `json:"kind"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		}{

			{
				Kind:     "image",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/image.png\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Kind:     "image",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"HTTPS://example.invalid/image.png\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Kind:     "image",
				Rejected: true,
			},

			{
				Kind:     "image",
				Rejected: true,
			},

			{
				Kind:     "image",
				Rejected: true,
			},

			{
				Kind:     "image",
				Rejected: true,
			},

			{
				Kind:     "image",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https:// \",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Kind:     "image",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Kind:     "image",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/image.png\\n\",\"IsHidden\":null,\"IsRequired\":null,\"Name\":\"Fixture\",\"Order\":-1}"),
			},

			{
				Kind:     "image",
				Rejected: true,
			},

			{
				Kind:     "slate",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/image.png\"}"),
			},

			{
				Kind:     "slate",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"HTTPS://example.invalid/image.png\"}"),
			},

			{
				Kind:     "slate",
				Rejected: true,
			},

			{
				Kind:     "slate",
				Rejected: true,
			},

			{
				Kind:     "slate",
				Rejected: true,
			},

			{
				Kind:     "slate",
				Rejected: true,
			},

			{
				Kind:     "slate",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https:// \"}"),
			},

			{
				Kind:     "slate",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\"}"),
			},

			{
				Kind:     "slate",
				Rejected: false,
				Wire:     json.RawMessage("{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":null,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/image.png\\n\"}"),
			},

			{
				Kind:     "slate",
				Rejected: true,
			},

			{
				Kind:     "reaction",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/image.png\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Kind:     "reaction",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"HTTPS://example.invalid/image.png\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Kind:     "reaction",
				Rejected: true,
			},

			{
				Kind:     "reaction",
				Rejected: true,
			},

			{
				Kind:     "reaction",
				Rejected: true,
			},

			{
				Kind:     "reaction",
				Rejected: true,
			},

			{
				Kind:     "reaction",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https:// \",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Kind:     "reaction",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Kind:     "reaction",
				Rejected: false,
				Wire:     json.RawMessage("{\"BrandedDisplayName\":null,\"Id\":\"fixture-reaction\",\"Identity\":null,\"ReactionImageUrl\":\"https://example.invalid/image.png\\n\",\"ReactionSpriteImageUrl\":null,\"ReactionType\":\"FutureReaction\"}"),
			},

			{
				Kind:     "reaction",
				Rejected: true,
			},

			{
				Kind:     "theme",
				Rejected: false,
				Wire:     json.RawMessage("{\"BackgroundImageDarkPreAuthUri\":null,\"BackgroundImageDarkUri\":null,\"BackgroundImageLightPreAuthUri\":null,\"BackgroundImageLightUri\":null,\"BrandAccentColor\":\"future-color\",\"ChatBackgroundImageAlignment\":\"FutureAlignment\",\"ChatBackgroundImageDarkUri\":null,\"ChatBackgroundImageLightUri\":null,\"DisableOriginalReactions\":false,\"DisplayName\":null,\"EnableChatBackground\":false,\"Enabled\":false,\"Identity\":\"fixture-theme\",\"LogoImageDarkPreAuthUri\":null,\"LogoImageDarkUri\":null,\"LogoImageLightPreAuthUri\":null,\"LogoImageLightUri\":\"relative-or-future\",\"MeetingInviteLogoImageDarkPreAuthUri\":null,\"MeetingInviteLogoImageLightPreAuthUri\":null,\"MeetingReactions\":[]}"),
			},
		},
	}

	accepted := 0
	for _, c := range inputVectors.Cases {
		if c.Rejected {
			// Requestd setter validation is input for the pending local builders;
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
			t.Fatal("unknown requestInputs asset kind")
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
