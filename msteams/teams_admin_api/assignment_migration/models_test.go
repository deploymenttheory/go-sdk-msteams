package assignmentmigration_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

func TestUnitTypedMigrationRequestAndStateDTOObservations(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name  string          `json:"name"`
			Input string          `json:"input"`
			Wire  json.RawMessage `json:"wire"`
		} `json:"cases"`
	}{
		Cases: []struct {
			Name  string          `json:"name"`
			Input string          `json:"input"`
			Wire  json.RawMessage `json:"wire"`
		}{

			{
				Name: "appids-omitted",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-scalar",
				Wire: json.RawMessage("{\"AppIds\":[\"fixture-app\"],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-replacement",
				Wire: json.RawMessage("{\"AppIds\":[\"fixture-a\",\"fixture-b\"],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-empty-replacement",
				Wire: json.RawMessage("{\"AppIds\":[],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-null-replacement",
				Wire: json.RawMessage("{\"AppIds\":[],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-operators",
				Wire: json.RawMessage("{\"AppIds\":{\"Add\":[\"fixture-a\"],\"Remove\":[\"fixture-b\"]},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-empty-operators",
				Wire: json.RawMessage("{\"AppIds\":{},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-legacy-object-replacement",
				Wire: json.RawMessage("{\"AppIds\":[{\"AppId\":\"fixture-app\",\"Enabled\":false}],\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-legacy-object-operators",
				Wire: json.RawMessage("{\"AppIds\":{\"Add\":[{\"AppId\":\"fixture-app\",\"Enabled\":false}]},\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "appids-replace-operator-rejected",
			},

			{
				Name: "appids-numeric-item-rejected",
			},

			{
				Name: "appids-ps-list-modifier-rejected",
			},

			{
				Name: "migration-request-field-casing",
				Wire: json.RawMessage("{\"authority\":\"Tenant\",\"groupid\":\"00000000-0000-4000-8000-000000000001\",\"policyname\":\"Fixture\",\"policytype\":\"FixturePolicy\"}"),
			},

			{
				Name:  "migration-state-NotStarted",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"NotStarted\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"NotStarted\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-PopulationInProgress",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"PopulationInProgress\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"PopulationInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-GpaCreationInProgress",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"GpaCreationInProgress\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"GpaCreationInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-WaitingForGpaProcessed",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"WaitingForGpaProcessed\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"WaitingForGpaProcessed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-CleanUpInProgress",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"CleanUpInProgress\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"CleanUpInProgress\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-Completed",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"Completed\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"Completed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-Failed",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"Failed\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"Failed\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name:  "migration-state-ConnectionLost",
				Input: "{\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"state\":\"ConnectionLost\",\"directAssignmentsProcessed\":0,\"startTime\":\"2026-01-01T00:00:00Z\"}",
				Wire:  json.RawMessage("{\"completionTime\":\"0001-01-01T00:00:00\",\"directAssignmentsProcessed\":0,\"groupId\":\"00000000-0000-4000-8000-000000000001\",\"groupName\":null,\"migrationEventId\":\"00000000-0000-4000-8000-000000000002\",\"policyInstance\":null,\"policyType\":null,\"startTime\":\"2026-01-01T00:00:00Z\",\"state\":\"ConnectionLost\",\"tenantId\":\"00000000-0000-0000-0000-000000000000\"}"),
			},

			{
				Name: "provider-bound-omitted",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-bound-empty-and-false",
				Wire: json.RawMessage("{\"Description\":\"\",\"EnableFederationAccess\":false,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-OrganizationDefault",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":0,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-AllowAllExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":1,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-AllowSpecificExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":2,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-BlockSpecificExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":3,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-external-enum-BlockAllExternalDomains",
				Wire: json.RawMessage("{\"CommunicationWithExternalOrgs\":4,\"Identity\":\"tag:Fixture\"}"),
			},

			{
				Name: "provider-voicemail-timespan",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MaximumRecordingLength\":\"00:01:30\"}"),
			},

			{
				Name: "provider-voicemail-explicit-null",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MaximumRecordingLength\":null}"),
			},

			{
				Name: "provider-branding-empty-lists",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MeetingBackgroundImages\":[],\"MeetingBrandingThemes\":[],\"NdiAssuranceSlateImages\":[]}"),
			},

			{
				Name: "provider-branding-nested-lists",
				Wire: json.RawMessage("{\"Identity\":\"tag:Fixture\",\"MeetingBackgroundImages\":[{\"Id\":\"fixture-image\",\"Identity\":null,\"ImageUri\":\"https://example.invalid/background.png\",\"IsHidden\":null,\"IsRequired\":false,\"Name\":\"Fixture\",\"Order\":0}],\"MeetingBrandingThemes\":[{\"BackgroundImageDarkPreAuthUri\":null,\"BackgroundImageDarkUri\":null,\"BackgroundImageLightPreAuthUri\":null,\"BackgroundImageLightUri\":null,\"BrandAccentColor\":null,\"ChatBackgroundImageAlignment\":null,\"ChatBackgroundImageDarkUri\":null,\"ChatBackgroundImageLightUri\":null,\"DisableOriginalReactions\":null,\"DisplayName\":null,\"EnableChatBackground\":null,\"Enabled\":false,\"Identity\":\"fixture-theme\",\"LogoImageDarkPreAuthUri\":null,\"LogoImageDarkUri\":null,\"LogoImageLightPreAuthUri\":null,\"LogoImageLightUri\":null,\"MeetingInviteLogoImageDarkPreAuthUri\":null,\"MeetingInviteLogoImageLightPreAuthUri\":null,\"MeetingReactions\":[]}],\"NdiAssuranceSlateImages\":[{\"Id\":\"fixture-slate\",\"Identity\":null,\"IsDefault\":false,\"Name\":\"Fixture\",\"NdiImageUri\":\"https://example.invalid/slate.png\"}]}"),
			},

			{
				Name: "provider-user-clear",
				Wire: json.RawMessage("{\"additionalParameters\":null,\"policyName\":null}"),
			},

			{
				Name: "provider-tenant-clear",
				Wire: json.RawMessage("{\"PolicyName\":null,\"additionalParameters\":null,\"forceSwitchPresent\":false}"),
			},

			{
				Name: "provider-group-clear-zero-rank",
				Wire: json.RawMessage("{\"policyName\":null,\"priority\":0}"),
			},

			{
				Name:  "provider-identity-tagged",
				Input: "tag:Fixture / Policy",
				Wire:  json.RawMessage("\"tag%3AFixture%20/%20Policy\""),
			},

			{
				Name:  "provider-identity-nested",
				Input: "parent/Fixture / Child",
				Wire:  json.RawMessage("\"parent/Fixture%20%2F%20Child\""),
			},
		},
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
