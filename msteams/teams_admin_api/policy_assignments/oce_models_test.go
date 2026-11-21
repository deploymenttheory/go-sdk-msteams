package policyassignments

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestOCEProfilesMatchPinnedEnvironmentAndRegionTables(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-assignment-oce.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Profiles []struct {
			Environment string
			Origin      string
			Regions     []Region
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Profiles) != 8 {
		t.Fatal("incomplete pinned profiles")
	}
	profiles := oceProfiles()
	if len(profiles) != len(fixture.Profiles) {
		t.Fatal("profile count changed")
	}
	for _, observed := range fixture.Profiles {
		found := false
		for _, profile := range profiles {
			if string(profile.environment) == observed.Environment {
				found = true
				if profile.origin != observed.Origin || !reflect.DeepEqual(profile.regions, observed.Regions) {
					t.Fatal("profile origin or region order differs from pinned source")
				}
			}
		}
		if !found {
			t.Fatal("conditional environment dropped")
		}
	}
}

func TestPinnedOCEModelObservationsPreserveReceivedFields(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-assignment-oce.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Models []struct {
			Name, Model       string
			Input, Projection json.RawMessage
			Rejected          bool
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 12 {
		t.Fatal("incomplete OCE models")
	}
	for _, row := range fixture.Models {
		t.Run(row.Name, func(t *testing.T) {
			if row.Rejected {
				t.Fatal("unexpected pinned model rejection")
			}
			if string(row.Input) == "null" {
				var value *GroupAssignment
				if json.Unmarshal(row.Input, &value) != nil || value != nil {
					t.Fatal("nullable GPA reference changed")
				}
				return
			}
			var value any
			switch row.Model {
			case "GroupAssignment":
				value = &GroupAssignment{}
			case "GroupMembers":
				value = &OCEGroupMember{}
			case "UserMembership":
				value = &OCEUserMembership{}
			case "BatchDefinition":
				value = &OCEBatchDefinition{}
			case "BatchStatus", "ExtendedBatchStatus":
				value = &OCEBatchStatus{}
			default:
				t.Fatal("unknown OCE model")
			}
			if json.Unmarshal(row.Input, value) != nil {
				t.Fatal("native OCE model rejected observed fields")
			}
			var typed any
			switch v := value.(type) {
			case *GroupAssignment:
				type plain GroupAssignment
				typed = plain(*v)
			case *OCEGroupMember:
				type plain OCEGroupMember
				typed = plain(*v)
			case *OCEUserMembership:
				type plain OCEUserMembership
				typed = plain(*v)
			case *OCEBatchDefinition:
				type plain OCEBatchDefinition
				typed = plain(*v)
			case *OCEBatchStatus:
				type plain OCEBatchStatus
				typed = plain(*v)
			}
			encoded, e := json.Marshal(typed)
			if e != nil {
				t.Fatal(e)
			}
			var actual, projection, input any
			if json.Unmarshal(encoded, &actual) != nil || json.Unmarshal(row.Projection, &projection) != nil || json.Unmarshal(row.Input, &input) != nil {
				t.Fatal("invalid model projection")
			}
			// CLR defaults are observable in the oracle but must not manufacture
			// fields absent from the wire in a native read snapshot.
			if !containsProjection(actual, receivedProjection(input, projection)) {
				t.Fatal("native typed fields differ from pinned received values")
			}
			encoded, e = json.Marshal(value)
			if e != nil {
				t.Fatal(e)
			}
			if json.Unmarshal(encoded, &actual) != nil || !reflect.DeepEqual(actual, input) {
				t.Fatal("OCE snapshot lost wire presence or future fields")
			}
		})
	}
}

func receivedProjection(input, projection any) any {
	switch expected := projection.(type) {
	case map[string]any:
		result := map[string]any{}
		observed, ok := input.(map[string]any)
		if !ok {
			return result
		}
		for key, value := range expected {
			for received, original := range observed {
				if strings.EqualFold(key, received) {
					result[key] = receivedProjection(original, value)
					break
				}
			}
		}
		return result
	case []any:
		observed, ok := input.([]any)
		if !ok || len(observed) != len(expected) {
			return projection
		}
		result := make([]any, len(expected))
		for i, value := range expected {
			result[i] = receivedProjection(observed[i], value)
		}
		return result
	default:
		return projection
	}
}

func TestOCEReadModelsDoNotMaterializeDefaultsOrLeakData(t *testing.T) {
	t.Parallel()
	var status OCEBatchStatus
	if json.Unmarshal([]byte(`{"CompletedCount":0,"ErrorCount":null,"future":"fixture-private","UserState":[{"Id":"fixture-private","AlternateId":null}]}`), &status) != nil {
		t.Fatal("OCE snapshot rejected")
	}
	if status.CompletedCount == nil || *status.CompletedCount != 0 || status.ErrorCount != nil || status.NotStartedCount != nil || status.CreatedBy != nil || len(status.UserState) != 1 || status.UserState[0].AlternateID != nil || string(status.RawFields["ErrorCount"]) != "null" {
		t.Fatal("OCE wire presence changed")
	}
	for _, value := range []any{status, status.UserState[0], OCEGroupMember{RawFields: status.RawFields}, OCEUserMembership{RawFields: status.RawFields}, OCEBatchDefinition{RawFields: status.RawFields}, OCEBatchCommand{RawFields: status.RawFields}, RegionAttempt{}, RegionalLookup[OCEBatchStatus]{Value: &status}, RegionalList[OCEBatchStatus]{Values: []*OCEBatchStatus{&status}}, OCEActionResult{Value: new("fixture-private")}, OCEContext{TenantID: "fixture-private"}} {
		if strings.Contains(fmt.Sprintf("%v %#v", value, value), "fixture-private") {
			t.Fatal("OCE diagnostic formatter leaked data")
		}
	}
}
