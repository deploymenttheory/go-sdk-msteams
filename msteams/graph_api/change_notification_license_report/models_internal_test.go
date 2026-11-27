package licensereport

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func number(v string) *json.Number { n := json.Number(v); return &n }
func text(v string) *string        { return &v }
func TestUnitNativeCountRequiresJSONGrammarAndBoundsOnlyComputation(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"2/1", "0x1p4097", "\"2\"", "null", "", "1e4097", "1e-4097", "0.1", strings.Repeat("9", 4097)} {
		if _, e := integer(number(raw)); !errors.Is(e, ErrCount) {
			t.Fatal("non-JSON/noninteger/over-budget count accepted")
		}
	}
	for raw, want := range map[string]string{"-0": "0", "1e3": "1000", "1.00": "1", "9223372036854775808": "9223372036854775808"} {
		value, e := integer(number(raw))
		if e != nil || value.String() != want {
			t.Fatal("valid precise JSON integer changed")
		}
	}
}
func TestUnitNativePlanUnknownIsNeverNegative(t *testing.T) {
	t.Parallel()
	for _, items := range [][]LicenseDetail{{{ServicePlans: nil}}, {{ServicePlans: []*ServicePlan{nil}}}, {{ServicePlans: []*ServicePlan{{ID: nil}}}}, {{ServicePlans: []*ServicePlan{{ID: text("")}}}}} {
		value, complete, e := matchingPlan(items)
		if e == nil || complete || value != nil {
			t.Fatal("absent or unparseable plan data became negative")
		}
	}
	for _, items := range [][]LicenseDetail{{}, {{ServicePlans: []*ServicePlan{}}}, {{ServicePlans: []*ServicePlan{{ID: text("00000000-0000-4000-8000-000000000001")}}}}} {
		value, complete, e := matchingPlan(items)
		if e != nil || !complete || value == nil || *value {
			t.Fatal("complete empty/nonmatching data did not support false")
		}
	}
	value, complete, e := matchingPlan([]LicenseDetail{{ServicePlans: []*ServicePlan{{ID: text(SourcePlanID), ProvisioningStatus: text("Disabled")}}}})
	if e != nil || !complete || value == nil || !*value {
		t.Fatal("source presence predicate was replaced by effective entitlement")
	}
}
func TestUnitNativeModelRetainsNumberWidthAndNestedPresence(t *testing.T) {
	t.Parallel()
	raw := `{"userId":null,"privateChatMessageCount":9223372036854775808,"teamChatMessageCount":1,"future":[]}`
	var activity Activity
	if json.Unmarshal([]byte(raw), &activity) != nil || activity.PrivateChatMessageCount == nil || activity.PrivateChatMessageCount.String() != "9223372036854775808" || string(activity.RawFields["userId"]) != "null" {
		t.Fatal("native width or presence lost")
	}
	encoded, e := json.Marshal(activity)
	if e != nil || string(encoded) != raw {
		t.Fatal("owned report snapshot changed")
	}
	for _, raw := range []string{`{"id":"a"}`, `{"id":"a","servicePlans":null}`, `{"id":"a","servicePlans":{}}`} {
		var value LicenseDetail
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("missing/null/malformed plan collection accepted as empty")
		}
	}
	var value LicenseDetail
	if json.Unmarshal([]byte(`{"id":"a","servicePlans":[null,{"servicePlanId":null,"future":[]}],"future":null}`), &value) != nil || len(value.ServicePlans) != 2 || value.ServicePlans[0] != nil || value.ServicePlans[1].ID != nil {
		t.Fatal("nested null plan/id presence lost")
	}
}
