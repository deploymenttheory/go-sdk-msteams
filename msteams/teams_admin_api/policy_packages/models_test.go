package policypackages

import (
	"encoding/json"
	"fmt"
	"os"
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
func TestUnitPinnedPackageModelsAndRankDefaults(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-package-models.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Model       string
			Input, Projection json.RawMessage
			Rejected          bool
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 18 {
		t.Fatal("incomplete model capture")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			if row.Rejected {
				t.Fatal("unexpected pinned model rejection")
			}
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
				t.Fatal("typed model differs from pinned projection")
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

func TestUnitPackageWrapperPolicyPairsRemainLocalAndOrdered(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-package-binding.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Command string
			Input         []string
			Failure       *string `json:"failure_type"`
			Calls         []struct {
				Parameters struct{ PolicyList []PolicyPair }
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 9 {
		t.Fatal("incomplete wrapper capture")
	}
	checked := 0
	for _, row := range fixture.Cases {
		if row.Command == "group" {
			continue
		}
		checked++
		pairs, e := ParsePolicyList(row.Input)
		if row.Failure != nil {
			if e == nil {
				t.Fatal("delimiter-free pair accepted")
			}
			continue
		}
		if e != nil || len(row.Calls) != 1 || !reflect.DeepEqual(pairs, row.Calls[0].Parameters.PolicyList) {
			t.Fatal("local pair conversion differs from unmodified wrapper")
		}
	}
	if checked != 5 {
		t.Fatal("wrapper cases dropped")
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
