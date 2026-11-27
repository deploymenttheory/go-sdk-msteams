package conversationrolessetting_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	settings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_roles_setting"
)

func TestUnitConversationAllThirtyObservedListTransitions(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("testdata/pinned-list-transitions.json")
	if e != nil {
		t.Fatal(e)
	}
	var source struct {
		Cases []struct {
			Current, Scenario string
			Accepted          bool
			IDs               []string
		}
	}
	if json.Unmarshal(raw, &source) != nil || len(source.Cases) != 30 {
		t.Fatal("exact source transition matrix missing")
	}
	role := func(id string) *roles.Role { return &roles.Role{ID: ptr(id)} }
	thirdID := "00000000-0000-4000-8000-000000000003"
	for _, c := range source.Cases {
		t.Run(c.Current+"-"+c.Scenario, func(t *testing.T) {
			var current []*roles.Role
			if c.Current != "null" {
				current = make([]*roles.Role, 0)
			}
			if c.Current == "nonempty" {
				current = append(current, role(firstID))
			}
			modifier := &settings.ListModifier{}
			switch c.Scenario {
			case "null":
				modifier = nil
			case "add":
				modifier.Add = []*roles.Role{role(nextID)}
			case "replace":
				modifier.Replace = []*roles.Role{role(nextID)}
			case "remove-existing":
				modifier.Remove = []*roles.Role{role(firstID)}
			case "remove-missing":
				modifier.Remove = []*roles.Role{role(nextID)}
			case "add-and-remove":
				modifier.Add = []*roles.Role{role(nextID)}
				modifier.Remove = []*roles.Role{role(firstID)}
			case "replace-and-remove":
				modifier.Replace = []*roles.Role{role(nextID)}
				modifier.Remove = []*roles.Role{role(firstID)}
			case "replace-and-add":
				modifier.Replace = []*roles.Role{role(nextID)}
				modifier.Add = []*roles.Role{role(thirdID)}
			case "all":
				modifier.Replace = []*roles.Role{role(nextID)}
				modifier.Add = []*roles.Role{role(thirdID)}
				modifier.Remove = []*roles.Role{role(firstID)}
			}
			encoded, _, e := settings.ProjectListValue(current, modifier)
			if (e == nil) != c.Accepted {
				t.Fatal("source transition acceptance changed")
			}
			if e != nil {
				return
			}
			var actual []struct {
				ID string `json:"Id"`
			}
			if json.Unmarshal(encoded, &actual) != nil {
				t.Fatal("projected list invalid")
			}
			ids := make([]string, 0, len(actual))
			for _, v := range actual {
				ids = append(ids, v.ID)
			}
			expected := c.IDs
			if expected == nil {
				expected = []string{}
			}
			if !reflect.DeepEqual(ids, expected) {
				t.Fatal("source transition result changed")
			}
			if c.Current == "nonempty" && (len(current) != 1 || current[0].ID == nil || *current[0].ID != firstID) {
				t.Fatal("local transition mutated current snapshot")
			}
		})
	}
}
