package policyassignments

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestUnitOCEEnvironmentAndRegionTables(t *testing.T) {
	t.Parallel()
	profilesWant := []struct {
		Environment, Origin string
		Regions             []Region
	}{
		{Environment: "INT", Origin: "https://dev.api.interfaces.records.teams.microsoft.com", Regions: []Region{"INT"}},
		{Environment: "MultiTenant", Origin: "https://api.interfaces.records.teams.microsoft.com", Regions: []Region{"NOAM", "EMEA", "APAC"}},
		{Environment: "Ag08", Origin: "https://api.interfaces.records.teams.eaglex.ic.gov", Regions: []Region{"usnat"}},
		{Environment: "Ag09", Origin: "https://api.interfaces.records.teams.microsoft.scloud", Regions: []Region{"ussec"}},
		{Environment: "Gallatin", Origin: "https://api.interfaces.records.teams.microsoftonline.cn", Regions: []Region{"Gallatin"}},
		{Environment: "Itar", Origin: "https://api.interfaces.records.gov.teams.microsoft.us", Regions: []Region{"gcch"}},
		{Environment: "GCCH", Origin: "https://api.interfaces.records.gov.teams.microsoft.us", Regions: []Region{"gcch"}},
		{Environment: "DOD", Origin: "https://api.interfaces.records.dod.teams.microsoft.us", Regions: []Region{"dod"}},
	}
	profiles := oceProfiles()
	if len(profiles) != len(profilesWant) {
		t.Fatal("profile count changed")
	}
	for _, observed := range profilesWant {
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

func TestUnitOCEModelVectorsPreserveReceivedFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name, Model       string
		Input, Projection json.RawMessage
	}{
		{Name: "group-empty", Model: "GroupAssignment", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"groupId\":\"00000000-0000-0000-0000-000000000000\",\"priority\":0,\"policyType\":null,\"policyName\":null,\"createdBy\":\"\",\"createdTime\":null}")},
		{Name: "group-values", Model: "GroupAssignment", Input: json.RawMessage("{\"groupId\":\"00000000-0000-4000-8000-000000000005\",\"priority\":0,\"policyName\":\"\",\"createdBy\":null,\"createdTime\":null,\"future\":false}"), Projection: json.RawMessage("{\"groupId\":\"00000000-0000-4000-8000-000000000005\",\"priority\":0,\"policyType\":null,\"policyName\":\"\",\"createdBy\":null,\"createdTime\":null}")},
		{Name: "member-empty", Model: "GroupMembers", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"@odata.type\":null,\"id\":\"00000000-0000-0000-0000-000000000000\"}")},
		{Name: "member-values", Model: "GroupMembers", Input: json.RawMessage("{\"id\":\"00000000-0000-4000-8000-000000000007\",\"@odata.type\":\"#fixture\",\"future\":false}"), Projection: json.RawMessage("{\"@odata.type\":\"#fixture\",\"id\":\"00000000-0000-4000-8000-000000000007\"}")},
		{Name: "membership-empty", Model: "UserMembership", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"id\":\"00000000-0000-0000-0000-000000000000\",\"groupTypes\":null,\"mailEnabled\":false,\"securityEnabled\":false,\"visibility\":null,\"createdDateTime\":null,\"deletedDateTime\":null}")},
		{Name: "membership-values", Model: "UserMembership", Input: json.RawMessage("{\"id\":\"00000000-0000-4000-8000-000000000005\",\"groupTypes\":[],\"mailEnabled\":false,\"securityEnabled\":true,\"visibility\":\"Future\",\"createdDateTime\":null}"), Projection: json.RawMessage("{\"id\":\"00000000-0000-4000-8000-000000000005\",\"groupTypes\":[],\"mailEnabled\":false,\"securityEnabled\":true,\"visibility\":\"Future\",\"createdDateTime\":null,\"deletedDateTime\":null}")},
		{Name: "definition-empty", Model: "BatchDefinition", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"Region\":null,\"OperationId\":\"00000000-0000-0000-0000-000000000000\",\"OperationName\":null,\"OperationType\":null,\"Commands\":null,\"Ids\":null,\"AcceptTime\":0}")},
		{Name: "definition-values", Model: "BatchDefinition", Input: json.RawMessage("{\"OperationId\":\"00000000-0000-4000-8000-000000000004\",\"OperationName\":\"\",\"OperationType\":\"Future\",\"AcceptTime\":0,\"Ids\":[],\"Commands\":[{\"Name\":\"Fixture\",\"RetryCount\":0,\"Properties\":{\"flag\":false},\"Description\":\"\"}]}"), Projection: json.RawMessage("{\"Region\":null,\"OperationId\":\"00000000-0000-4000-8000-000000000004\",\"OperationName\":\"\",\"OperationType\":\"Future\",\"Commands\":[{\"Name\":\"Fixture\",\"RetryCount\":0,\"Properties\":{\"flag\":false},\"Description\":\"\"}],\"Ids\":[],\"AcceptTime\":0}")},
		{Name: "status-empty", Model: "BatchStatus", Input: json.RawMessage("{}"), Projection: json.RawMessage("{\"Region\":null,\"OperationName\":null,\"OverallStatus\":null,\"NotStartedCount\":0,\"InProgressCount\":0,\"CompletedCount\":0,\"ErrorCount\":0,\"CreatedBy\":\"00000000-0000-0000-0000-000000000000\",\"CreatedTime\":null,\"CompletedTime\":null,\"OperationId\":\"00000000-0000-0000-0000-000000000000\"}")},
		{Name: "status-values", Model: "BatchStatus", Input: json.RawMessage("{\"OperationId\":\"00000000-0000-4000-8000-000000000004\",\"OverallStatus\":\"Future\",\"CompletedCount\":0,\"CreatedTime\":\"fixture-time\",\"CompletedTime\":null}"), Projection: json.RawMessage("{\"Region\":null,\"OperationName\":null,\"OverallStatus\":\"Future\",\"NotStartedCount\":0,\"InProgressCount\":0,\"CompletedCount\":0,\"ErrorCount\":0,\"CreatedBy\":\"00000000-0000-0000-0000-000000000000\",\"CreatedTime\":\"fixture-time\",\"CompletedTime\":null,\"OperationId\":\"00000000-0000-4000-8000-000000000004\"}")},
		{Name: "extended-values", Model: "ExtendedBatchStatus", Input: json.RawMessage("{\"OperationId\":\"00000000-0000-4000-8000-000000000004\",\"UserState\":[{\"Id\":\"fixture-user\",\"AlternateId\":null,\"State\":\"Future\",\"Result\":\"\"}]}"), Projection: json.RawMessage("{\"Region\":null,\"UserState\":[{\"Id\":\"fixture-user\",\"AlternateId\":null,\"State\":\"Future\",\"Result\":\"\"}],\"OperationName\":null,\"OverallStatus\":null,\"NotStartedCount\":0,\"InProgressCount\":0,\"CompletedCount\":0,\"ErrorCount\":0,\"CreatedBy\":\"00000000-0000-0000-0000-000000000000\",\"CreatedTime\":null,\"CompletedTime\":null,\"OperationId\":\"00000000-0000-4000-8000-000000000004\"}")},
		{Name: "null", Model: "GroupAssignment", Input: json.RawMessage("null"), Projection: json.RawMessage("null")},
	}
	for _, row := range cases {
		t.Run(row.Name, func(t *testing.T) {
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
				t.Fatal("native OCE model rejected model input")
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
			// Model defaults must not manufacture
			// fields absent from the wire in a native read snapshot.
			if !containsProjection(actual, receivedProjection(input, projection)) {
				t.Fatal("native typed fields differ from expected received values")
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

func TestUnitOCEReadModelsDoNotMaterializeDefaultsOrLeakData(t *testing.T) {
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
