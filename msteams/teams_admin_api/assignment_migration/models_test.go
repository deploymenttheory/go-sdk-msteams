package assignmentmigration_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

func TestUnitPinnedMigrationRequestAndStateDTOObservations(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-model-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name  string          `json:"name"`
			Input string          `json:"input"`
			Wire  json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixture) != nil {
		t.Fatal("invalid pinned model observations")
	}
	states, requests := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "migration-request-field-casing" {
			requests++
			encoded, err := json.Marshal(migration.CreateRequest{Authority: "Tenant", PolicyType: "FixturePolicy", PolicyName: "Fixture", GroupID: "00000000-0000-4000-8000-000000000001", AllowNonEmptyGroup: true})
			var actual, want any
			if err != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal(c.Wire, &want) != nil || !reflect.DeepEqual(actual, want) {
				t.Fatal("request keys or query/body separation differ")
			}
		}
		if strings.HasPrefix(c.Name, "migration-state-") {
			states++
			var job migration.Job
			if json.Unmarshal([]byte(c.Input), &job) != nil {
				t.Fatal("known state rejected")
			}
			name, ok := job.State.Text()
			if !ok || string(name) != strings.TrimPrefix(c.Name, "migration-state-") {
				t.Fatal("state name changed")
			}
			if job.TenantID != nil || job.CompletionTime != nil {
				t.Fatal("absent CLR defaults materialized")
			}
			encoded, err := json.Marshal(job)
			var actual, want any
			if err != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(c.Input), &want) != nil || !reflect.DeepEqual(actual, want) {
				t.Fatal("job snapshot did not preserve wire presence")
			}
		}
	}
	if states != 8 || requests != 1 {
		t.Fatal("pinned migration observation coverage differs")
	}
}

func TestUnitFutureStatesAndSnapshotFieldsRemainAvailable(t *testing.T) {
	t.Parallel()
	for _, value := range []string{`"FutureState"`, `17`, `null`} {
		input := `{"state":` + value + `,"directAssignmentsProcessed":0,"startTime":"original-timestamp","completionTime":null,"Future":{"enabled":false}}`
		var job migration.Job
		if json.Unmarshal([]byte(input), &job) != nil {
			t.Fatal("future/nullable state rejected")
		}
		encoded, err := json.Marshal(job)
		var actual, want any
		if err != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(input), &want) != nil || !reflect.DeepEqual(actual, want) {
			t.Fatal("future snapshot changed")
		}
		if value == `17` {
			number, ok := job.State.Number()
			if !ok || number != 17 {
				t.Fatal("future numeric enum lost")
			}
		}
		copy := job.State.RawJSON()
		clear(copy)
		if string(job.State.RawJSON()) != value {
			t.Fatal("state borrowed returned bytes")
		}
	}
	for _, value := range []string{`true`, `{}`, `[]`, `2147483648`} {
		var state migration.State
		if json.Unmarshal([]byte(value), &state) == nil {
			t.Fatal("invalid Int32/string enum representation accepted")
		}
	}
}

func TestUnitConflictAndRecommendationNullZeroAndFutureData(t *testing.T) {
	t.Parallel()
	input := `{"totalCountUsersWithConflict":0,"usersWithConflictDueDirectAssignmentsCount":0,"usersWithConflictDueDirectAssignmentsDetail":[null,{"userId":"fixture-user","policyInstance":null,"assignmentSource":"FutureSource","Future":false}],"usersWithConflictDueGroupAssignmentsDetail":null,"Future":{"x":null}}`
	var value migration.Conflict
	if json.Unmarshal([]byte(input), &value) != nil || value.TotalCountUsersWithConflict == nil || *value.TotalCountUsersWithConflict != 0 || value.UsersWithConflictDueGroupAssignmentsCount != nil {
		t.Fatal("conflict presence changed")
	}
	encoded, err := json.Marshal(value)
	var actual, want any
	if err != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(input), &want) != nil || !reflect.DeepEqual(actual, want) {
		t.Fatal("nested conflict snapshot changed")
	}
	var recommendation migration.PolicyTypeRecommendation
	if json.Unmarshal([]byte(`{"policyType":"FuturePolicy","daCount":0,"Future":null}`), &recommendation) != nil || recommendation.DirectAssignmentsCount == nil || *recommendation.DirectAssignmentsCount != 0 {
		t.Fatal("recommendation zero/count lost")
	}
}
