package policyassignments

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPinnedGeneratedAssignmentModelProjections(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-assignment-models.json")
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
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 14 {
		t.Fatal("incomplete pinned model observations")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			if row.Rejected {
				t.Fatal("unexpected pinned rejection")
			}
			var value any
			switch row.Model {
			case "GroupAssignment":
				value = &GroupAssignment{}
			case "EffectivePolicy":
				value = &EffectivePolicy{}
			case "SourceEntry":
				value = &SourceEntry{}
			case "BatchJobStatus", "SimpleBatchJobStatus":
				value = &BatchStatus{}
			case "GrantActionStatus":
				value = &UserOutcome{}
			default:
				t.Fatal("unknown captured model")
			}
			if json.Unmarshal(row.Input, value) != nil {
				t.Fatal("native model rejected pinned input")
			}
			// Remove root snapshot marshaling to check the decoded typed fields,
			// independently of the lossless received-object round trip below.
			var view any
			switch v := value.(type) {
			case *GroupAssignment:
				type plain GroupAssignment
				view = plain(*v)
			case *EffectivePolicy:
				type plain EffectivePolicy
				view = plain(*v)
			case *SourceEntry:
				type plain SourceEntry
				view = plain(*v)
			case *BatchStatus:
				type plain BatchStatus
				view = plain(*v)
			case *UserOutcome:
				type plain UserOutcome
				view = plain(*v)
			}
			encoded, e := json.Marshal(view)
			if e != nil {
				t.Fatal(e)
			}
			var got, want any
			if json.Unmarshal(encoded, &got) != nil || json.Unmarshal(row.Projection, &want) != nil {
				t.Fatal("invalid model projection")
			}
			if !containsProjection(got, want) {
				t.Fatal("typed fields differ from pinned projection")
			}
			encoded, e = json.Marshal(value)
			if e != nil {
				t.Fatal(e)
			}
			var original any
			if json.Unmarshal(encoded, &got) != nil || json.Unmarshal(row.Input, &original) != nil || !reflect.DeepEqual(got, original) {
				t.Fatal("snapshot lost received null/absence/future fields")
			}
		})
	}
}

func containsProjection(actual, expected any) bool {
	switch want := expected.(type) {
	case map[string]any:
		got, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for key, value := range want {
			received, present := got[key]
			if !present || !containsProjection(received, value) {
				return false
			}
		}
		return true
	case []any:
		got, ok := actual.([]any)
		if !ok || len(got) != len(want) {
			return false
		}
		for i, value := range want {
			if !containsProjection(got[i], value) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(actual, expected)
	}
}

func TestAssignmentSnapshotPresenceAndRedaction(t *testing.T) {
	t.Parallel()
	var value BatchStatus
	if json.Unmarshal([]byte(`{"operationId":"fixture-private","completedCount":0,"errorCount":null,"userState":[{"id":"fixture-private","result":""}],"future":false}`), &value) != nil {
		t.Fatal("snapshot rejected")
	}
	if value.CompletedCount == nil || *value.CompletedCount != 0 || value.ErrorCount != nil || value.NotStartedCount != nil || len(value.UserState) != 1 || value.UserState[0].Result == nil || *value.UserState[0].Result != "" {
		t.Fatal("lost zero/null/missing/nested fields")
	}
	if string(value.RawFields["errorCount"]) != "null" || string(value.RawFields["future"]) != "false" {
		t.Fatal("lost raw presence")
	}
	for _, v := range []any{value, value.UserState[0], GroupAssignment{RawFields: value.RawFields}, SourceEntry{RawFields: value.RawFields}, EffectivePolicy{RawFields: value.RawFields}, OperationResult{RawFields: value.RawFields}} {
		if strings.Contains(fmt.Sprintf("%v %#v", v, v), "fixture-private") {
			t.Fatal("diagnostic formatting leaked a snapshot")
		}
	}
	for _, raw := range []string{"null", "[]", "not-json", `{"completedCount":"0"}`, `{"userState":{}}`} {
		if json.Unmarshal([]byte(raw), &value) == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
}

func TestPinnedConditionalAssignmentMetadataMatchesTargets(t *testing.T) {
	t.Parallel()
	read := func(raw []byte, err error) []json.RawMessage {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct{ Commands []json.RawMessage }
		if json.Unmarshal(raw, &fixture) != nil || len(fixture.Commands) != 8 {
			t.Fatal("incomplete OCE metadata")
		}
		return fixture.Commands
	}
	first := read(os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-assignment-oce-net472-metadata.json"))
	second := read(os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-assignment-oce-netcoreapp3.1-metadata.json"))
	sets := 0
	for i, command := range first {
		var a, b any
		if json.Unmarshal(command, &a) != nil || json.Unmarshal(second[i], &b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("conditional signatures differ by assembly target")
		}
		var metadata struct {
			Sets []json.RawMessage `json:"parameter_sets"`
		}
		if json.Unmarshal(command, &metadata) != nil {
			t.Fatal("invalid OCE command")
		}
		sets += len(metadata.Sets)
	}
	if sets != 15 {
		t.Fatal("conditional parameter sets were dropped")
	}
}
