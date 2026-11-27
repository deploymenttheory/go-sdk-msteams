package teamsmessagingconfiguration_test

import (
	"encoding/json"

	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
)

func TestUnitStorylineRetainsSourceEnumValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		want       resource.Storyline
	}{{"named", `"Enabled"`, resource.StorylineEnabled}, {"case", `"disabled"`, resource.StorylineDisabled}, {"numeric", "37", 37}, {"numeric-string", `"-7"`, -7}} {
		t.Run(tc.name, func(t *testing.T) {
			var value resource.Storyline
			if json.Unmarshal([]byte(tc.body), &value) != nil || value != tc.want {
				t.Fatal("source enum representation lost")
			}
		})
	}
	for _, invalid := range []string{`"UnknownName"`, "2147483648", "1.5", "{}", "true"} {
		var value resource.Storyline
		if json.Unmarshal([]byte(invalid), &value) == nil {
			t.Fatal("invalid source enum accepted")
		}
	}
}
