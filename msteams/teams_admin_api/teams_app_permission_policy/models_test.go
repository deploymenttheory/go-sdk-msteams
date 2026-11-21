package teamsapppermissionpolicy_test

import (
	"encoding/json"
	"fmt"
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

	v := modelFixture{
		Policies: []struct {
			Input       string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "future-types-long-description",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[],\"DefaultCatalogAppsType\":\"FutureMode\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":\"\",\"Identity\":null,\"PrivateCatalogApps\":[],\"Private" +
						"CatalogAppsType\":null}"),
				},
			},

			{
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":null,\"GlobalCatalogApps\":null,\"PrivateCatalogApps\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":[null]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[null],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":[{\"Id\":\"a\"}],\"DefaultCatalogApps\":[{\"Id\":\"b\"}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":[{\"Id\":null}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":[{\"Id\":null,\"Id\":\"good\"}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":[{\"Id\":\"\"}],\"GlobalCatalogApps\":[{\"Id\":\"future-app\"}],\"PrivateCatalogApps\":[{\"Id\":\" \"}]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"DefaultCatalogApps\":[{\"Id\":\"\",\"Identity\":null}],\"DefaultCatalogAppsType\":null,\"Description\":null,\"GlobalCatalogApps\":[{\"Id\":\"future-app\",\"Identity\":null}],\"GlobalCatalogAppsType\":null,\"Identity\":null,\"PrivateCatalogApps\":[{\"Id\":\" \",\"Identity\":null}],\"PrivateCatalogAppsType\":null}"),
				},
			},

			{
				Input: "{\"DefaultCatalogApps\":{\"Id\":\"wrong-envelope\"}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},
		},
		Apps: []struct {
			Catalog, Input string
			Observation    struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Catalog: "default",
				Input:   "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "default",
				Input:   "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "default",
				Input:   " ",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\" \",\"Identity\":null}"),
				},
			},

			{
				Catalog: "default",
				Input:   "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "default",
				Input:   "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Catalog: "default",
				Input:   "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "default",
				Input:   "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "global",
				Input:   "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "global",
				Input:   "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "global",
				Input:   " ",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\" \",\"Identity\":null}"),
				},
			},

			{
				Catalog: "global",
				Input:   "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "global",
				Input:   "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Catalog: "global",
				Input:   "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "global",
				Input:   "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "private",
				Input:   "null",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "private",
				Input:   "",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "private",
				Input:   " ",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\" \",\"Identity\":null}"),
				},
			},

			{
				Catalog: "private",
				Input:   "future-app",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":\"future-app\",\"Identity\":null}"),
				},
			},

			{
				Catalog: "private",
				Input:   "json:{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"Id\":null,\"Identity\":null}"),
				},
			},

			{
				Catalog: "private",
				Input:   "json:{\"Id\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},

			{
				Catalog: "private",
				Input:   "json:{\"Id\":null,\"Id\":\"good\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("System.ArgumentNullException"),
				},
			},
		},
		Lists: []struct {
			Catalog, Name string
			Observation   struct {
				Failure *string
				Values  json.RawMessage
			}
		}{

			{
				Catalog: "default",
				Name:    "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "default",
				Name:    "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "default",
				Name:    "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "default",
				Name:    "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "default",
				Name:    "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "default",
				Name:    "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "default",
				Name:    "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "global",
				Name:    "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "global",
				Name:    "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "global",
				Name:    "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "global",
				Name:    "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "global",
				Name:    "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "global",
				Name:    "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "global",
				Name:    "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "private",
				Name:    "null",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "private",
				Name:    "empty-operators",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[]"),
				},
			},

			{
				Catalog: "private",
				Name:    "replace-duplicates",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "private",
				Name:    "remove-one",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "private",
				Name:    "remove-two",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "private",
				Name:    "add-existing",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},

			{
				Catalog: "private",
				Name:    "remove-and-add",
				Observation: struct {
					Failure *string
					Values  json.RawMessage
				}{
					Values: json.RawMessage("[{\"Id\":\"a\",\"Identity\":null},{\"Id\":\"A\",\"Identity\":null},{\"Id\":\"b\",\"Identity\":null},{\"Id\":\"a\",\"Identity\":null}]"),
				},
			},
		},
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

func TestUnitAppPermissionTypedReadonlyPolicyModels(t *testing.T) {
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
func TestUnitAppPermissionTypedCatalogModels(t *testing.T) {
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
func TestUnitAppPermissionSnapshotsRetainRepeatedPropertiesAndOwnership(t *testing.T) {
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

func TestUnitAppPermissionTypedCatalogEdits(t *testing.T) {
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
func TestUnitCatalogBuildersOwnInputsAndIgnoreNonKeyFields(t *testing.T) {
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
