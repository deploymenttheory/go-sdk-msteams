package teamsapppermissionpolicy_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

type modelFixture struct {
	Policies []struct {
		Input       string
		Observation struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"policies"`
	Apps []struct {
		Catalog, Input string
		Observation    struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"apps"`
	Lists []struct {
		Catalog, Name string
		Observation   struct {
			Failure *string
			Values  json.RawMessage
		}
	} `json:"lists"`
}

func models(t *testing.T) modelFixture {
	t.Helper()
	raw, e := os.ReadFile("../../../" + sourceModels)
	if e != nil {
		t.Fatal(e)
	}
	var v modelFixture
	if json.Unmarshal(raw, &v) != nil || len(v.Policies) != 10 || len(v.Apps) != 21 || len(v.Lists) != 21 {
		t.Fatal("invalid app permission model fixture")
	}
	return v
}
func sameJSON(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
func appsProjection(v []*resource.CatalogApp) []map[string]*string {
	out := make([]map[string]*string, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]*string{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
func policyProjection(v resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Identity": v.Identity, "Description": v.Description, "DefaultCatalogAppsType": v.DefaultCatalogAppsType, "GlobalCatalogAppsType": v.GlobalCatalogAppsType, "PrivateCatalogAppsType": v.PrivateCatalogAppsType, "DefaultCatalogApps": appsProjection(v.DefaultCatalogApps), "GlobalCatalogApps": appsProjection(v.GlobalCatalogApps), "PrivateCatalogApps": appsProjection(v.PrivateCatalogApps)})
	return raw
}

func TestAppPermissionPinnedReadonlyPolicyModels(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Policies {
		var value resource.Policy
		var e error
		switch c.Input {
		case "constructor":
		case "future-types-long-description":
			value.Description = new(strings.Repeat("x", 2000))
			value.DefaultCatalogAppsType = new("FutureMode")
			value.GlobalCatalogAppsType = new("")
		default:
			e = json.Unmarshal([]byte(c.Input), &value)
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatal("pinned policy setter acceptance changed")
		}
		if e == nil && !sameJSON(policyProjection(value), c.Observation.Value) {
			t.Fatal("pinned readonly-list projection changed")
		}
	}
}
func TestAppPermissionPinnedCatalogModels(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Apps {
		var value *resource.CatalogApp
		var e error
		if strings.HasPrefix(c.Input, "json:") {
			e = json.Unmarshal([]byte(strings.TrimPrefix(c.Input, "json:")), &value)
		} else if c.Input == "null" {
			e = json.Unmarshal([]byte(`{"Id":null}`), &value)
		} else {
			value = resource.NewCatalogApp(c.Input)
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatal("pinned catalog setter acceptance changed")
		}
		if e == nil {
			raw, _ := json.Marshal(appsProjection([]*resource.CatalogApp{value})[0])
			if !sameJSON(raw, c.Observation.Value) {
				t.Fatal("pinned catalog projection changed")
			}
		}
	}
}
func TestAppPermissionSnapshotsRetainRepeatedPropertiesAndOwnership(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"Identity":"fixture-private","DefaultCatalogApps":null,"Future":false}`, `{"DefaultCatalogApps":[{"Id":"a","Future":null}],"defaultcatalogapps":[{"Id":"b"}],"DefaultCatalogApps":null}`, `{"GlobalCatalogApps":[{"Id":"x"}],"GlobalCatalogApps":[{"Id":"y"}],"PrivateCatalogApps":[null]}`} {
		buffer := []byte(raw)
		var value resource.Policy
		if json.Unmarshal(buffer, &value) != nil {
			t.Fatal("valid snapshot rejected")
		}
		clear(buffer)
		encoded, e := json.Marshal(value)
		var again resource.Policy
		if e != nil || json.Unmarshal(encoded, &again) != nil || !bytesProjectionEqual(value, again) {
			t.Fatal("duplicate properties or raw buffer ownership lost")
		}
		if strings.Contains(fmt.Sprintf("%v %#v", value, value), "fixture-private") {
			t.Fatal("policy formatting exposed data")
		}
	}
	for _, raw := range []string{`[]`, `{"DefaultCatalogApps":[{"Id":null,"Id":"good"}]}`, `{"GlobalCatalogApps":{"Id":"bad-shape"}}`, `{"PrivateCatalogApps":[{"Id":false}]}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) == nil {
			t.Fatal("invalid catalog setter/shape accepted")
		}
	}
}
func bytesProjectionEqual(a, b resource.Policy) bool {
	return sameJSON(policyProjection(a), policyProjection(b))
}

func TestAppPermissionPinnedCatalogEdits(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Lists {
		a, duplicate, upper, b := resource.NewCatalogApp("a"), resource.NewCatalogApp("a"), resource.NewCatalogApp("A"), resource.NewCatalogApp("b")
		current := []*resource.CatalogApp{a, duplicate, upper, b}
		var edit *resource.CatalogEdit
		switch c.Name {
		case "null":
			edit = resource.ClearCatalogApps()
		case "empty-operators":
			edit = resource.ChangeCatalogApps(nil, nil)
		case "replace-duplicates":
			edit = resource.ReplaceCatalogApps(b, b, a)
		case "remove-one":
			edit = resource.ChangeCatalogApps(nil, []*resource.CatalogApp{duplicate})
		case "remove-two":
			edit = resource.ChangeCatalogApps(nil, []*resource.CatalogApp{duplicate, duplicate})
		case "add-existing":
			edit = resource.ChangeCatalogApps([]*resource.CatalogApp{duplicate, duplicate}, nil)
		case "remove-and-add":
			edit = resource.ChangeCatalogApps([]*resource.CatalogApp{duplicate}, []*resource.CatalogApp{duplicate})
		default:
			t.Fatal("unmapped catalog edit")
		}
		v, e := edit.Apply(current)
		raw, _ := json.Marshal(appsProjection(v))
		if e != nil || c.Observation.Failure != nil || !sameJSON(raw, c.Observation.Values) {
			t.Fatal("pinned catalog list behavior changed")
		}
		if len(current) != 4 || *current[0].ID != "a" {
			t.Fatal("catalog edit mutated caller source")
		}
	}
}
func TestCatalogBuildersOwnInputsAndIgnoreNonKeyFields(t *testing.T) {
	t.Parallel()
	a := resource.NewCatalogApp("a")
	a.Identity = new("first")
	b := resource.NewCatalogApp("a")
	b.Identity = new("different")
	a.RawFields = map[string]json.RawMessage{"Future": json.RawMessage(`{"a":[]}`)}
	edit := resource.ReplaceCatalogApps(a)
	*a.ID = "changed"
	a.RawFields["Future"][0] = 'x'
	out, e := edit.Apply(nil)
	if e != nil || *out[0].ID != "a" || !json.Valid(out[0].RawFields["Future"]) {
		t.Fatal("catalog edit aliased caller state")
	}
	v, e := resource.ChangeCatalogApps(nil, []*resource.CatalogApp{b}).Apply(out)
	if e != nil || len(v) != 0 {
		t.Fatal("catalog comparison used Identity or future data")
	}
	if _, e := resource.ChangeCatalogApps(nil, []*resource.CatalogApp{b}).Apply([]*resource.CatalogApp{nil}); e == nil {
		t.Fatal("null comparator silently skipped")
	}
}
