package teamstemplatepermissionpolicy_test

import (
	"encoding/json"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

func TestUnitHiddenTemplateProjectionUsesPopulatedAcceptanceResponse(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "populated_create_readback.json")
	var wire struct {
		HiddenTemplates []struct {
			ID       *string `json:"Id"`
			Identity *string
		}
	}
	if json.Unmarshal(raw, &wire) != nil || len(wire.HiddenTemplates) == 0 {
		t.Fatal("populated acceptance fixture required")
	}
	var model resource.Policy
	if json.Unmarshal(raw, &model) != nil {
		t.Fatal("actual policy did not decode")
	}
	if len(model.HiddenTemplates) != len(wire.HiddenTemplates) {
		t.Fatal("typed hidden templates lost real response entries")
	}
	for i, item := range model.HiddenTemplates {
		if item == nil || !reflect.DeepEqual(item.ID, wire.HiddenTemplates[i].ID) || !reflect.DeepEqual(item.Identity, wire.HiddenTemplates[i].Identity) {
			t.Fatal("typed hidden template fields differ from actual response")
		}
	}
}
