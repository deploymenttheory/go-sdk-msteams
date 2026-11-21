package policypackages

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func subset(actual, expected any) bool {
	switch want := expected.(type) {
	case map[string]any:
		got, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for k, v := range want {
			x, ok := got[k]
			if !ok || !subset(x, v) {
				return false
			}
		}
		return true
	case []any:
		got, ok := actual.([]any)
		if !ok || len(got) != len(want) {
			return false
		}
		for i, v := range want {
			if !subset(got[i], v) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(actual, expected)
	}
}
func TestUnitPackageModelProjectionVectors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name, Model       string
		Input, Projection json.RawMessage
	}{
		{Name: "package-empty", Model: "Package", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "package-values", Model: "Package", Input: json.RawMessage("{\"Name\":\"Fixture\",\"Description\":\"\",\"PackageType\":\"Future\",\"RecommendationType\":null,\"Policies\":{\"TeamsMeetingPolicy\":\"FixtureMeeting\",\"Future\":{\"flag\":false},\"nullable\":null},\"future\":0}"), Projection: json.RawMessage("{\"Description\":\"\",\"Name\":\"Fixture\",\"PackageType\":\"Future\",\"Policies\":{\"TeamsMeetingPolicy\":\"FixtureMeeting\",\"Future\":{\"flag\":false}}}")},
		{Name: "summary-values", Model: "Summary", Input: json.RawMessage("{\"Name\":\"Fixture\",\"Policies\":{},\"Description\":null,\"PackageType\":\"Future\"}"), Projection: json.RawMessage("{\"Name\":\"Fixture\",\"PackageType\":\"Future\"}")},
		{Name: "recommendation-values", Model: "Recommendation", Input: json.RawMessage("{\"Name\":\"Fixture\",\"RecommendationType\":\"Future\",\"Policies\":{\"TeamsMeetingPolicy\":\"FixtureMeeting\"}}"), Projection: json.RawMessage("{\"Name\":\"Fixture\",\"Policies\":{\"TeamsMeetingPolicy\":\"FixtureMeeting\"},\"RecommendationType\":\"Future\"}")},
		{Name: "custom-empty", Model: "CustomResponse", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "custom-values", Model: "CustomResponse", Input: json.RawMessage("{\"ResponseMessage\":\"\",\"future\":false}"), Projection: json.RawMessage("{\"ResponseMessage\":\"\"}")},
		{Name: "group-empty", Model: "GroupResponse", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "group-partial", Model: "GroupResponse", Input: json.RawMessage("{\"ResponseMessage\":\"FixturePartial\",\"FailedAssignPolicies\":[\"FixturePolicy\"],\"FailedCreatePolicies\":[],\"future\":false}"), Projection: json.RawMessage("{\"FailedAssignPolicies\":[\"FixturePolicy\"],\"FailedCreatePolicies\":[],\"ResponseMessage\":\"FixturePartial\"}")},
		{Name: "user-empty", Model: "UserResponse", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "user-partial", Model: "UserResponse", Input: json.RawMessage("{\"ResponseMessage\":\"FixturePartial\",\"FailedUserIds\":[\"fixture/a\"],\"FailedUserErrorMessages\":{\"LegacyUser\":[\"fixture-private\"],\"Future\":[\"fixture-future\"]},\"FailedCreatePolicies\":[],\"FailedAssignPolicies\":null,\"future\":false}"), Projection: json.RawMessage("{\"FailedUserErrorMessages\":{\"LegacyUser\":[\"fixture-private\"]},\"FailedCreatePolicies\":[],\"FailedUserIds\":[\"fixture/a\"],\"ResponseMessage\":\"FixturePartial\"}")},
		{Name: "batch-empty", Model: "BatchResponse", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "batch-values", Model: "BatchResponse", Input: json.RawMessage("{\"OperationId\":\"fixture-operation\",\"ResponseMessage\":\"\",\"FailedCreatePolicies\":[],\"future\":false}"), Projection: json.RawMessage("{\"FailedCreatePolicies\":[],\"OperationId\":\"fixture-operation\",\"ResponseMessage\":\"\"}")},
		{Name: "pair-empty", Model: "PolicyPair", Input: json.RawMessage("{}"), Projection: json.RawMessage("{}")},
		{Name: "pair-values", Model: "PolicyPair", Input: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"PolicyName\":\"\",\"future\":false}"), Projection: json.RawMessage("{\"PolicyName\":\"\",\"PolicyType\":\"FixtureType\"}")},
		{Name: "rank-default", Model: "Ranking", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"Rank\":0}")},
		{Name: "rank-zero", Model: "Ranking", Input: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":0}"), Projection: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":0}")},
		{Name: "rank-negative", Model: "Ranking", Input: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":-1}"), Projection: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":-1}")},
		{Name: "rank-null", Model: "Ranking", Input: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":null}"), Projection: json.RawMessage("{\"PolicyType\":\"FixtureType\",\"Rank\":0}")},
	}
	for _, row := range cases {
		t.Run(row.Name, func(t *testing.T) {
			var value any
			switch row.Model {
			case "Package", "Summary", "Recommendation":
				value = &Package{}
			case "CustomResponse":
				value = &MessageResult{}
			case "GroupResponse", "UserResponse", "BatchResponse":
				value = &AssignmentResult{}
			case "PolicyPair":
				value = &PolicyPair{}
			case "Ranking":
				value = &PolicyRanking{}
			default:
				t.Fatal("unknown model")
			}
			if json.Unmarshal(row.Input, value) != nil {
				t.Fatal("native model rejected observed input")
			}
			var view any
			switch v := value.(type) {
			case *Package:
				type plain Package
				view = plain(*v)
			case *MessageResult:
				type plain MessageResult
				view = plain(*v)
			case *AssignmentResult:
				type plain AssignmentResult
				view = plain(*v)
			default:
				view = value
			}
			encoded, e := json.Marshal(view)
			if e != nil {
				t.Fatal(e)
			}
			var got, want any
			if json.Unmarshal(encoded, &got) != nil || json.Unmarshal(row.Projection, &want) != nil || !subset(got, want) {
				t.Fatal("typed model differs from expected projection")
			}
			switch value.(type) {
			case *Package, *MessageResult, *AssignmentResult:
				encoded, e = json.Marshal(value)
				if e != nil {
					t.Fatal(e)
				}
				if json.Unmarshal(encoded, &got) != nil || json.Unmarshal(row.Input, &want) != nil || !reflect.DeepEqual(got, want) {
					t.Fatal("snapshot lost received null/absence/future data")
				}
			}
		})
	}
}

func TestUnitPackagePolicyPairParsingAndOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   []string
		want    []PolicyPair
		invalid bool
	}{
		{name: "create-delimiters", input: []string{" A , B ", "A.B", "A:B", "A;B", "A B", "A\tB"}, want: []PolicyPair{{PolicyType: new(""), PolicyName: new("A , B")}, {PolicyType: new("A"), PolicyName: new("B")}, {PolicyType: new("A"), PolicyName: new("B")}, {PolicyType: new("A"), PolicyName: new("B")}, {PolicyType: new("A"), PolicyName: new("B")}, {PolicyType: new("A"), PolicyName: new("B")}}, invalid: false},
		{name: "update-duplicates", input: []string{"Type,Name", "Type,Name"}, want: []PolicyPair{{PolicyType: new("Type"), PolicyName: new("Name")}, {PolicyType: new("Type"), PolicyName: new("Name")}}, invalid: false},
		{name: "create-first-split-only", input: []string{"Type,,Name", "Type:Name:Child"}, want: []PolicyPair{{PolicyType: new("Type"), PolicyName: new(",Name")}, {PolicyType: new("Type"), PolicyName: new("Name:Child")}}, invalid: false},
		{name: "create-empty-pieces", input: []string{",Name", "Type,", ","}, want: []PolicyPair{{PolicyType: new(""), PolicyName: new("Name")}, {PolicyType: new("Type"), PolicyName: new("")}, {PolicyType: new(""), PolicyName: new("")}}, invalid: false},
		{name: "create-missing-delimiter", input: []string{"TypeName"}, want: []PolicyPair{}, invalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pairs, err := ParsePolicyList(tc.input)
			if tc.invalid {
				if err == nil {
					t.Fatal("delimiter-free pair accepted")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(pairs, tc.want) {
				t.Fatal("policy pair parsing or order differs")
			}
		})
	}
	if _, e := ParsePolicyList(nil); e == nil {
		t.Fatal("unbound required policy list accepted")
	}
}

func TestUnitPackageSnapshotsAndRequestsRedactDiagnostics(t *testing.T) {
	t.Parallel()
	var value AssignmentResult
	if json.Unmarshal([]byte(`{"FailedUserErrorMessages":{"LegacyUser":["fixture-private"],"Future":false},"OperationId":null,"FailedUserIds":[],"ResponseMessage":"fixture-private"}`), &value) != nil {
		t.Fatal("snapshot rejected")
	}
	if value.OperationID != nil || value.FailedUserIDs == nil || value.FailedUserErrorMessages == nil || string(value.FailedUserErrorMessages.RawFields["Future"]) != "false" {
		t.Fatal("null/empty/future distinction lost")
	}
	for _, v := range []any{value, value.FailedUserErrorMessages, Package{Name: new("fixture-private")}, MessageResult{ResponseMessage: new("fixture-private")}, CustomPackageRequest{Name: new("fixture-private")}, GroupAssignmentRequest{GroupID: "fixture-private"}, UserAssignmentRequest{UserIDs: []string{"fixture-private"}}, Identity{PackageName: "fixture-private"}} {
		if strings.Contains(fmt.Sprintf("%v %#v", v, v), "fixture-private") {
			t.Fatal("diagnostic formatting leaked package data")
		}
	}
}
