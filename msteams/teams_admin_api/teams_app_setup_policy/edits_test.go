package teamsappsetuppolicy_test

import (
	"encoding/json"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

func appCase(name string, current []*resource.App) ([]*resource.App, error) {
	a, b := current[1], current[4]
	var edit *resource.AppEdit
	switch name {
	case "null":
		edit = resource.ClearApps()
	case "empty-operators":
		edit = resource.ChangeApps(nil, nil)
	case "remove-one":
		edit = resource.ChangeApps(nil, []*resource.App{a})
	case "remove-two":
		edit = resource.ChangeApps(nil, []*resource.App{a, a})
	case "remove-three":
		edit = resource.ChangeApps(nil, []*resource.App{a, a, a})
	case "replace-duplicates":
		edit = resource.ReplaceApps(b, b, current[0])
	case "add-existing":
		edit = resource.ChangeApps([]*resource.App{a, a}, nil)
	case "remove-and-add":
		edit = resource.ChangeApps([]*resource.App{a}, []*resource.App{a})
	}
	return edit.Apply(current)
}
func pinCase(name string, current []*resource.PinnedApp) ([]*resource.PinnedApp, error) {
	a, b := current[1], current[4]
	var edit *resource.PinEdit
	switch name {
	case "null":
		edit = resource.ClearPins()
	case "empty-operators":
		edit = resource.ChangePins(nil, nil)
	case "remove-one":
		edit = resource.ChangePins(nil, []*resource.PinnedApp{a})
	case "remove-two":
		edit = resource.ChangePins(nil, []*resource.PinnedApp{a, a})
	case "remove-three":
		edit = resource.ChangePins(nil, []*resource.PinnedApp{a, a, a})
	case "replace-duplicates":
		edit = resource.ReplacePins(b, b, current[0])
	case "add-existing":
		edit = resource.ChangePins([]*resource.PinnedApp{a, a}, nil)
	case "remove-and-add":
		edit = resource.ChangePins([]*resource.PinnedApp{a}, []*resource.PinnedApp{a})
	}
	return edit.Apply(current)
}
func customCase(name string, current []*resource.CustomizationApp) ([]*resource.CustomizationApp, error) {
	a, b := current[1], current[4]
	var edit *resource.CustomizationEdit
	switch name {
	case "null":
		edit = resource.ClearCustomizations()
	case "empty-operators":
		edit = resource.ChangeCustomizations(nil, nil)
	case "remove-one":
		edit = resource.ChangeCustomizations(nil, []*resource.CustomizationApp{a})
	case "remove-two":
		edit = resource.ChangeCustomizations(nil, []*resource.CustomizationApp{a, a})
	case "remove-three":
		edit = resource.ChangeCustomizations(nil, []*resource.CustomizationApp{a, a, a})
	case "replace-duplicates":
		edit = resource.ReplaceCustomizations(b, b, current[0])
	case "add-existing":
		edit = resource.ChangeCustomizations([]*resource.CustomizationApp{a, a}, nil)
	case "remove-and-add":
		edit = resource.ChangeCustomizations([]*resource.CustomizationApp{a}, []*resource.CustomizationApp{a})
	}
	return edit.Apply(current)
}
func TestUnitAppSetupPinnedListComparisons(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Lists {
		var value any
		var e error
		switch c.Field {
		case "PinnedAppBarApps", "PinnedMessageBarApps":
			current := []*resource.PinnedApp{resource.NewPinnedApp("a", 1), resource.NewPinnedApp("a", 1), resource.NewPinnedApp("a", 2), resource.NewPinnedApp("a", 1), resource.NewPinnedApp("A", 1)}
			v, err := pinCase(c.Name, current)
			e = err
			value = pinsProjection(v)
			if len(current) != 5 || current[0].Order != 1 {
				t.Fatal("pin edit changed caller source")
			}
		case "AdditionalCustomizationApps":
			current := []*resource.CustomizationApp{resource.NewCustomizationApp("a", "one"), resource.NewCustomizationApp("a", "one"), resource.NewCustomizationApp("a", "one"), resource.NewCustomizationApp("a", "two"), resource.NewCustomizationApp("A", "one")}
			v, err := customCase(c.Name, current)
			e = err
			value = customsProjection(v)
			if len(current) != 5 || *current[0].AdditionalCustomizationID != "one" {
				t.Fatal("customization edit changed caller source")
			}
		default:
			current := []*resource.App{resource.NewApp("a"), resource.NewApp("a"), resource.NewApp("a"), resource.NewApp("a"), resource.NewApp("A")}
			v, err := appCase(c.Name, current)
			e = err
			value = appsProjection(v)
			if len(current) != 5 || *current[0].ID != "a" {
				t.Fatal("app edit changed caller source")
			}
		}
		raw, _ := json.Marshal(value)
		if e != nil || c.Observation.Failure != nil || !sameJSON(raw, c.Observation.Values) {
			t.Fatalf("pinned setup edit changed: %s %s", c.Field, c.Name)
		}
	}
}
func TestUnitAppSetupEditsOwnDataAndRejectZeroValues(t *testing.T) {
	t.Parallel()
	a := resource.NewApp("a")
	a.Identity = new("first")
	a.RawFields = map[string]json.RawMessage{"Future": json.RawMessage(`{"a":[]}`)}
	edit := resource.ReplaceApps(a)
	*a.ID = "changed"
	a.RawFields["Future"][0] = 'x'
	v, e := edit.Apply(nil)
	if e != nil || *v[0].ID != "a" || !json.Valid(v[0].RawFields["Future"]) {
		t.Fatal("app edit aliased caller state")
	}
	remove := resource.NewApp("a")
	remove.Identity = new("different")
	if v, e := resource.ChangeApps(nil, []*resource.App{remove}).Apply(v); e != nil || len(v) != 0 {
		t.Fatal("app equality used Identity")
	}
	for _, edit := range []*resource.AppEdit{nil, {}} {
		if _, e := edit.Apply(nil); e == nil {
			t.Fatal("nil/zero app edit accepted")
		}
	}
	for _, edit := range []*resource.PinEdit{nil, {}} {
		if _, e := edit.Apply(nil); e == nil {
			t.Fatal("nil/zero pin edit accepted")
		}
	}
	for _, edit := range []*resource.CustomizationEdit{nil, {}} {
		if _, e := edit.Apply(nil); e == nil {
			t.Fatal("nil/zero customization edit accepted")
		}
	}
	if _, e := resource.ChangePins(nil, []*resource.PinnedApp{resource.NewPinnedApp("a", 1)}).Apply([]*resource.PinnedApp{nil}); e == nil {
		t.Fatal("null pin comparator silently skipped")
	}
}
