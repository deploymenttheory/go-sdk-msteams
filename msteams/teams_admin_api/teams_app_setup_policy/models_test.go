package teamsappsetuppolicy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

type modelFixture struct {
	Policies []struct {
		Input       string
		Observation struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"policies"`
	Elements []struct {
		Field, Input string
		Observation  struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"elements"`
	Lists []struct {
		Field, Name string
		Observation struct {
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
	if json.Unmarshal(raw, &v) != nil || len(v.Policies) != 12 || len(v.Elements) != 56 || len(v.Lists) != 48 {
		t.Fatal("invalid setup model fixture")
	}
	return v
}

// UseNumber prevents a broken Int64 implementation from passing via float rounding.
func sameJSON(a, b []byte) bool {
	var x, y any
	left, right := json.NewDecoder(bytes.NewReader(a)), json.NewDecoder(bytes.NewReader(b))
	left.UseNumber()
	right.UseNumber()
	return left.Decode(&x) == nil && right.Decode(&y) == nil && reflect.DeepEqual(x, y)
}
func appsProjection(v []*resource.App) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
func pinsProjection(v []*resource.PinnedApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Order": x.Order}
		}
	}
	return out
}
func customsProjection(v []*resource.CustomizationApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "AdditionalCustomizationId": x.AdditionalCustomizationID}
		}
	}
	return out
}
func projection(v resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Identity": v.Identity, "Description": v.Description, "AllowSideLoading": v.AllowSideLoading, "AllowUserPinning": v.AllowUserPinning, "AppPresetList": appsProjection(v.AppPresetList), "PinnedAppBarApps": pinsProjection(v.PinnedAppBarApps), "PinnedMessageBarApps": pinsProjection(v.PinnedMessageBarApps), "AppPresetMeetingList": appsProjection(v.AppPresetMeetingList), "AdditionalCustomizationApps": customsProjection(v.AdditionalCustomizationApps), "PinnedCallingBarApps": appsProjection(v.PinnedCallingBarApps)})
	return raw
}
func TestAppSetupPinnedPolicyProjections(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Policies {
		var v resource.Policy
		var e error
		switch c.Input {
		case "constructor":
		case "long-description":
			v.Description = new(strings.Repeat("x", 2000))
			v.AllowSideLoading = new(false)
		default:
			e = json.Unmarshal([]byte(c.Input), &v)
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatal("pinned setup policy acceptance changed")
		}
		if e == nil && !sameJSON(projection(v), c.Observation.Value) {
			t.Fatal("pinned setup policy projection changed")
		}
	}
}
func TestAppSetupPinnedElementProjections(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Elements {
		var v any
		var e error
		decode := strings.HasPrefix(c.Input, "json:") || c.Input == "null"
		body := strings.TrimPrefix(c.Input, "json:")
		if c.Input == "null" {
			body = `{"Id":null}`
		}
		switch c.Field {
		case "PinnedAppBarApps", "PinnedMessageBarApps":
			var x *resource.PinnedApp
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewPinnedApp(c.Input, largeOrder)
			}
			if e == nil {
				v = pinsProjection([]*resource.PinnedApp{x})[0]
			}
		case "AdditionalCustomizationApps":
			var x *resource.CustomizationApp
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewCustomizationApp(c.Input, "fixture-customization")
			}
			if e == nil {
				v = customsProjection([]*resource.CustomizationApp{x})[0]
			}
		default:
			var x *resource.App
			if decode {
				e = json.Unmarshal([]byte(body), &x)
			} else {
				x = resource.NewApp(c.Input)
			}
			if e == nil {
				v = appsProjection([]*resource.App{x})[0]
			}
		}
		if (e != nil) != (c.Observation.Failure != nil) {
			t.Fatalf("pinned setup element acceptance changed: %s %s", c.Field, c.Input)
		}
		if e == nil {
			raw, _ := json.Marshal(v)
			if !sameJSON(raw, c.Observation.Value) {
				t.Fatal("pinned setup element projection changed")
			}
		}
	}
}
func TestAppSetupRawSnapshotsAndInt64StayExact(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"fixture-private","PinnedAppBarApps":[{"Id":"a","Order":9007199254740993,"Future":null}],"pinnedappbarapps":[{"Id":"b","Order":9223372036854775807}],"PinnedAppBarApps":null,"Future":false}`)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil || len(v.PinnedAppBarApps) != 2 || v.PinnedAppBarApps[0].Order != largeOrder || v.PinnedAppBarApps[1].Order != 9223372036854775807 {
		t.Fatal("readonly append or exact Int64 decoding changed")
	}
	clear(raw)
	encoded, e := json.Marshal(v)
	var again resource.Policy
	if e != nil || json.Unmarshal(encoded, &again) != nil || !sameJSON(projection(v), projection(again)) {
		t.Fatal("raw ownership/repeated properties lost")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", v, v), "fixture-private") {
		t.Fatal("setup formatting exposed data")
	}
	for _, body := range []string{`{"PinnedAppBarApps":[{"Order":null,"Order":1}]}`, `{"AppPresetList":[{"Id":null,"Id":"good"}]}`, `{"AdditionalCustomizationApps":[{"Id":"a","AdditionalCustomizationId":null,"AdditionalCustomizationId":"good"}]}`, `{"PinnedMessageBarApps":[{"Order":9223372036854775808}]}`} {
		var x resource.Policy
		if json.Unmarshal([]byte(body), &x) == nil {
			t.Fatal("earlier setter failure or overflow accepted")
		}
	}
}
