package licensereport_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	report "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/change_notification_license_report"
)

func ptr[T any](v T) *T         { return &v }
func num(v string) *json.Number { n := json.Number(v); return &n }
func TestUnitPinnedRowsKeepSourceWrapMissingNullAndDisplayFallback(t *testing.T) {
	t.Parallel()
	activity := []report.Activity{{UserPrincipalName: ptr("a"), PrivateChatMessageCount: num("2147483647"), TeamChatMessageCount: num("1")}, {UserPrincipalName: ptr("b")}, {UserPrincipalName: ptr("c")}}
	users := []report.DirectoryUser{{ID: userID, UserPrincipalName: ptr("a"), DisplayName: ptr("")}}
	values := map[string][]*report.PinnedLicense{"a": nil, "b": {}}
	rows, e := report.ProjectPinnedRows(activity, users, values)
	if e != nil || len(rows) != 2 || rows[0].NumberOfMessagesSent != -2147483648 || rows[0].DisplayName != "" || rows[0].HasRequiredLicensePlan != "" || rows[1].DisplayName != "Not Found" || rows[1].HasRequiredLicensePlan != "False" {
		t.Fatal("exact source output projection changed")
	}
}
func TestUnitPinnedNestedActualDefaultsAndOrdinalShortCircuit(t *testing.T) {
	t.Parallel()
	activity := []report.Activity{{UserPrincipalName: ptr("a")}}
	for _, scenario := range []struct {
		name  string
		value []*report.PinnedLicense
		want  string
		bad   bool
	}{
		{"null-service-plans", []*report.PinnedLicense{{ServicePlans: nil}}, "", true},
		{"empty-service-plans", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{}}}, "False", false},
		{"null-license", []*report.PinnedLicense{nil}, "", true},
		{"null-plan", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{nil}}}, "", true},
		{"missing-id", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{{ID: nil}}}}, "False", false},
		{"empty-id", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{{ID: ptr("")}}}}, "False", false},
		{"uppercase-id", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{{ID: ptr(strings.ToUpper(report.SourcePlanID))}}}}, "False", false},
		{"match-before-null", []*report.PinnedLicense{{ServicePlans: []*report.PinnedPlan{{ID: ptr(report.SourcePlanID)}, nil}}, nil}, "True", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			rows, e := report.ProjectPinnedRows(activity, nil, map[string][]*report.PinnedLicense{"a": scenario.value})
			if scenario.bad {
				if !errors.Is(e, report.ErrSourceShape) {
					t.Fatal("source malformed nested value became a verdict")
				}
				return
			}
			if e != nil || len(rows) != 1 || rows[0].HasRequiredLicensePlan != scenario.want {
				t.Fatal("source nested/default/ordinal rule changed")
			}
		})
	}
}
func TestUnitPinnedCountRepresentabilityDoesNotConstrainNativeDTO(t *testing.T) {
	t.Parallel()
	activity := []report.Activity{{UserPrincipalName: ptr("a"), PrivateChatMessageCount: num("2147483648"), TeamChatMessageCount: num("0")}}
	if _, e := report.ProjectPinnedRows(activity, nil, map[string][]*report.PinnedLicense{"a": {}}); !errors.Is(e, report.ErrSourceRange) {
		t.Fatal("source Int32 overflow was silently coerced")
	}
	activity[0].PrivateChatMessageCount = nil
	activity[0].RawFields = map[string]json.RawMessage{"privateChatMessageCount": json.RawMessage("null")}
	if _, e := report.ProjectPinnedRows(activity, nil, map[string][]*report.PinnedLicense{"a": {}}); !errors.Is(e, report.ErrSourceShape) {
		t.Fatal("explicit null Int32 became missing default zero")
	}
}
