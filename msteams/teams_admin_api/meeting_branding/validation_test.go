package meetingbranding_test

import (
	"encoding/json"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
)

func TestAssetReadValidationFollowsCaseInsensitiveSetterOrder(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		value any
		raw   string
	}{
		{new(resource.BackgroundImage), `{"imageuri":"http://example.invalid/a"}`},
		{new(resource.BackgroundImage), `{"order":null}`},
		{new(resource.BackgroundImage), `{"ImageUri":"http://example.invalid/a","ImageUri":"https://example.invalid/a"}`},
		{new(resource.BackgroundImage), `{"ImageUri":"http://example.invalid/a","imageuri":"https://example.invalid/a"}`},
		{new(resource.Reaction), `{"reactionimageurl":null}`},
		{new(resource.Reaction), `{"ReactionSpriteImageUrl":"http://example.invalid/a"}`},
		{new(resource.BrandingTheme), `{"identity":null}`},
		{new(resource.AssuranceSlate), `{"ndiimageuri":null}`},
	} {
		if json.Unmarshal([]byte(tc.raw), tc.value) == nil {
			t.Fatal("invalid asset setter input bypassed case/order validation")
		}
	}
	var image resource.BackgroundImage
	if json.Unmarshal([]byte(`{"imageuri":"https://example.invalid/a","ImageUri":"HTTPS://example.invalid/b","ORDER":0}`), &image) != nil || image.ImageURI == nil || *image.ImageURI != "HTTPS://example.invalid/b" || image.Order == nil || *image.Order != 0 {
		t.Fatal("valid case variants or duplicate setter values rejected")
	}
}
