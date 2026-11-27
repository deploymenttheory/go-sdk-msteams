package policies_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	assignment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	packages "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
)

func assignmentCompositionWire(method, path, command string, body json.RawMessage, status int, reply string, generated bool) testutil.Exchange {
	headers := http.Header{"X-MS-CmdletName": {command}}
	if generated {
		headers.Set("Accept", "application/json")
		headers.Set("X-MS-Target-Uri", "https://admin.example.invalid/")
		headers[http.CanonicalHeaderKey("MPACmdlet")] = nil
		if body != nil {
			headers.Set("Content-Type", "application/json")
		}
	} else {
		headers.Set("MPACmdlet", "true")
		headers[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
		if body != nil {
			headers.Set("Content-Type", "application/json; charset=utf-8")
		}
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func TestUnitPolicyCompositionKeepsGroupMigrationAndCleanupIdentitiesDistinct(t *testing.T) {
	t.Parallel()
	const group = "00000000-0000-4000-8000-0000000000AB"
	const canonical = "00000000-0000-4000-8000-0000000000ab"
	const jobID = "00000000-0000-4000-8000-000000000002"
	const policyType = "TeamsMeetingPolicy"
	const groupPath = "/Skype.Policy/groupPolicyAssignments/" + group + "/policyTypes/" + policyType
	const migrationPath = "/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration"
	const job = `{"MigrationEventId":"` + jobID + `","GroupId":"` + canonical + `","PolicyType":"` + policyType + `","PolicyInstance":"fixture-policy","State":"PopulationInProgress","DirectAssignmentsProcessed":0,"Future":false}`
	const input = `{"authority":"Tenant","policytype":"` + policyType + `","policyname":"fixture-policy","groupid":"` + canonical + `"}`
	discovery := policyCompositionFixture(t, "internal/configapi", "service_discovery.json")
	steps := []testutil.Exchange{
		{Request: testutil.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: discovery}},
		assignmentCompositionWire("PUT", groupPath, "New-CsGroupPolicyAssignment", json.RawMessage(`{"policyName":"fixture-policy","priority":0}`), 204, "", true),
		assignmentCompositionWire("GET", "/Skype.Policy/groupPolicyAssignments/groups/policyTypes/"+policyType+"/recommendations?groupThreshold=500", "Get-GroupAssignmentRecommendationsPerPolicyName", nil, 200, `[{"policyName":"fixture-policy","daCount":2}]`, false),
		assignmentCompositionWire("GET", "/Skype.Policy/groupPolicyAssignments/"+group+"/policyTypes/"+policyType+"/conflict", "Get-GroupPolicyAssignmentConflict", nil, 200, `{"TotalCountUsersWithConflict":0,"UsersWithConflictDueDirectAssignmentsCount":0,"UsersWithConflictDueDirectAssignmentsDetail":[],"UsersWithConflictDueGroupAssignmentsCount":0,"UsersWithConflictDueGroupAssignmentsDetail":[]}`, false),
		assignmentCompositionWire("POST", migrationPath+"/migrate?allowNonEmptyGroup=False", "Invoke-StartDirectToGroupAssignmentMigration", json.RawMessage(input), 200, job, false),
		assignmentCompositionWire("POST", migrationPath+"/cleanup?allowNonEmptyGroup=False", "Invoke-ClearDirectToGroupAssignmentMigration", json.RawMessage(input), 403, "", false),
		assignmentCompositionWire("DELETE", groupPath, "Remove-CsGroupPolicyAssignment", nil, 204, "", true),
		assignmentCompositionWire("GET", groupPath, "Get-CsGroupPolicyAssignment", nil, 404, "", true),
		assignmentCompositionWire("GET", migrationPath+"/"+jobID, "Get-DirectToGroupAssignmentsMigrationStatus", nil, 200, job, false),
	}
	result, err := testutil.Run(t.Context(), testutil.Case{Name: "group-migration-composition",

		Steps: steps, Want: testutil.Expected{Status: 200, Body: []byte(job), JSON: json.RawMessage(job)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			a := sdk.TeamsAdminAPI.PolicyAssignments
			receipt, response, err := a.CreateGroup(ctx, group, policyType, assignment.GroupSettings{PolicyName: policy.Set("fixture-policy"), Priority: policy.Set(int32(0))})
			if err != nil || response == nil || response.StatusCode() != 204 || receipt == nil || !receipt.HTTPAccepted || receipt.OperationID != nil {
				t.Fatal("group assignment receipt was lost or mistaken for a migration handle")
			}
			m := sdk.TeamsAdminAPI.AssignmentMigrations
			recommendations, _, err := m.RecommendationsByPolicyName(ctx, "groups", policyType, migration.RecommendationOptions{})
			if err != nil || len(recommendations) != 1 || recommendations[0] == nil || recommendations[0].PolicyName == nil || *recommendations[0].PolicyName != "fixture-policy" || recommendations[0].DirectAssignmentsCount == nil || *recommendations[0].DirectAssignmentsCount != 2 {
				t.Fatal("recommendation lost the policy-name/type distinction or typed count")
			}
			conflict, _, err := m.GetConflicts(ctx, group, policyType)
			if err != nil || conflict == nil || conflict.TotalCountUsersWithConflict == nil || *conflict.TotalCountUsersWithConflict != 0 {
				t.Fatal("conflict zero was lost or fabricated from an empty response")
			}
			request := migration.CreateRequest{Authority: "Tenant", PolicyType: policyType, PolicyName: *recommendations[0].PolicyName, GroupID: group}
			created, response, err := m.Create(ctx, request)
			if err != nil || response == nil || created == nil || created.MigrationEventID == nil || *created.MigrationEventID != jobID || created.GroupID == nil || *created.GroupID != canonical || created.DirectAssignmentsProcessed == nil || *created.DirectAssignmentsProcessed != 0 {
				t.Fatal("migration lost canonical group versus distinct job identity or zero progress")
			}
			if state, ok := created.State.Text(); !ok || state != migration.PopulationInProgress {
				t.Fatal("accepted migration was incorrectly promoted to completion")
			}
			failed, response, err := m.ClearDirectAssignments(ctx, request)
			if failed != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) || !receipt.HTTPAccepted || string(created.RawFields["Future"]) != "false" {
				t.Fatal("denied consequential clear erased earlier results")
			}
			removed, response, err := a.DeleteGroup(ctx, group, policyType)
			if err != nil || response == nil || response.StatusCode() != 204 || removed == nil || !removed.HTTPAccepted || removed.OperationID != nil {
				t.Fatal("explicit assignment-only cleanup lost its receipt")
			}
			absent, response, err := a.GetGroup(ctx, group, policyType)
			var api *client.APIError
			if absent != nil || response == nil || !errors.As(err, &api) || api.StatusCode != 404 {
				t.Fatal("assignment deletion was not independently confirmed")
			}
			last, response, err := m.GetByID(ctx, *created.MigrationEventID)
			if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, created.RawFields)) {
				t.Fatal("assignment cleanup deleted or altered the distinct migration record")
			}
			return testutil.Observation{Value: last, Response: response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("group migration composition differs: %v %+v", err, result.Differences())
	}
}

func TestUnitPolicyCompositionPackageModeRefusalDoesNotPoisonIndependentService(t *testing.T) {
	t.Parallel()
	before := policyCompositionFixture(t, "meeting_branding", "get_global.json")
	step := policyCompositionWire("GET", "TeamsMeetingBrandingPolicy", "Global", "Get-CsTeamsMeetingBrandingPolicy", nil, 200, before)
	result, err := testutil.Run(t.Context(), testutil.Case{Name: "package-mode-isolation", Mode: auth.Application,

		Steps: []testutil.Exchange{step, step}, Want: testutil.Expected{Status: 200, Body: before, JSON: before}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			first, _, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.GetByIdentity(ctx, "Global")
			if err != nil || first == nil || first.Identity == nil || *first.Identity != "Global" {
				t.Fatal("independent permitted primitive failed")
			}
			packages, response, err := sdk.TeamsAdminAPI.PolicyPackages.List(ctx)
			if packages != nil || response != nil || !errors.Is(err, client.ErrAuthenticationMode) {
				t.Fatal("documented package application-mode refusal sent a request or lost its classification")
			}
			last, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.GetByIdentity(ctx, "Global")
			if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, first.RawFields)) {
				t.Fatal("one service's local mode refusal poisoned the independent root service")
			}
			return testutil.Observation{Value: last, Response: response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("package mode isolation differs: %v %+v", err, result.Differences())
	}
}

// Primary-authored GPA recommendation and package integration, not captured
// tenant acceptance. TeamsMeetingPolicy is a documented package type/example;
// branding support is not inferred merely from the Package PolicyType string.
// Primary documentation: https://learn.microsoft.com/en-us/microsoftteams/manage-policy-packages
// and https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-cscustompolicypackage
func TestUnitPolicyCompositionBuildsPackageFromRecommendationAndRetainsPartialFailure(t *testing.T) {
	t.Parallel()
	discovery := policyCompositionFixture(t, "internal/configapi", "service_discovery.json")
	const recommendations = `[{"policyName":"FixtureMeeting","daCount":2,"Future":false}]`
	const definition = `{"Name":"fixture-package","Description":"","PackageType":"Fixture","Policies":{"TeamsMeetingPolicy":"FixtureMeeting"}}`
	const partial = `{"ResponseMessage":"received","FailedUserIds":["fixture/b"],"FailedUserErrorMessages":{"LegacyUser":["fixture/b"]},"FailedAssignPolicies":[],"FailedCreatePolicies":[]}`
	read := assignmentCompositionWire("GET", "/Skype.Policy/groupPolicyAssignments/groups/policyTypes/TeamsMeetingPolicy/recommendations?groupThreshold=500", "Get-GroupAssignmentRecommendationsPerPolicyName", nil, 200, recommendations, false)
	steps := []testutil.Exchange{
		read,
		{Request: testutil.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: discovery}},
		assignmentCompositionWire("POST", "/Teams.VerticalPackaging/package/custompackage", "New-CsCustomPolicyPackage", json.RawMessage(`{"PackageName":"fixture-package","PolicyList":[{"PolicyType":"TeamsMeetingPolicy","PolicyName":"FixtureMeeting"}]}`), 201, `{"ResponseMessage":"received"}`, true),
		assignmentCompositionWire("GET", "/Teams.VerticalPackaging/package/fixture-package", "Get-CsPolicyPackage", nil, 200, definition, true),
		assignmentCompositionWire("POST", "/Teams.VerticalPackaging/package/batch", "Grant-CsUserPolicyPackage", json.RawMessage(`{"PackageType":"fixture-package","UserIds":["fixture/a","fixture/a","fixture/b"]}`), 201, partial, true),
		read,
	}
	result, err := testutil.Run(t.Context(), testutil.Case{Name: "recommendation-package-composition", Mode: auth.Delegated,

		Steps: steps, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(recommendations), Body: []byte(recommendations)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			first, _, err := sdk.TeamsAdminAPI.AssignmentMigrations.RecommendationsByPolicyName(ctx, "groups", "TeamsMeetingPolicy", migration.RecommendationOptions{})
			if err != nil || len(first) != 1 || first[0] == nil || first[0].PolicyName == nil || *first[0].PolicyName != "FixtureMeeting" || first[0].DirectAssignmentsCount == nil || *first[0].DirectAssignmentsCount != 2 {
				t.Fatal("package definition lacks its typed policy-name recommendation")
			}
			s := sdk.TeamsAdminAPI.PolicyPackages
			created, response, err := s.Create(ctx, packages.CustomPackageRequest{Name: new("fixture-package"), Policies: []packages.PolicyPair{{PolicyType: new("TeamsMeetingPolicy"), PolicyName: first[0].PolicyName}}})
			if err != nil || response == nil || response.StatusCode() != 201 || created == nil || !created.HTTPAccepted || created.ResponseMessage == nil || *created.ResponseMessage != "received" {
				t.Fatal("package create lost its independent HTTP receipt")
			}
			value, _, err := s.GetByName(ctx, "fixture-package")
			if err != nil || value == nil || value.Name == nil || *value.Name != "fixture-package" || value.Description == nil || *value.Description != "" || string(value.Policies["TeamsMeetingPolicy"]) != `"FixtureMeeting"` {
				t.Fatal("package dictionary confused package and recommended policy name")
			}
			assigned, response, err := s.AssignUsers(ctx, packages.UserAssignmentRequest{PackageName: value.Name, UserIDs: []string{"fixture/a", "fixture/a", "fixture/b"}})
			if err != nil || response == nil || response.StatusCode() != 201 || assigned == nil || !assigned.HTTPAccepted || len(assigned.FailedUserIDs) != 1 || assigned.FailedUserIDs[0] == nil || *assigned.FailedUserIDs[0] != "fixture/b" || assigned.FailedUserErrorMessages == nil || len(assigned.FailedUserErrorMessages.LegacyUser) != 1 || assigned.FailedUserErrorMessages.LegacyUser[0] == nil || *assigned.FailedUserErrorMessages.LegacyUser[0] != "fixture/b" || assigned.OperationID != nil || !created.HTTPAccepted || *first[0].PolicyName != "FixtureMeeting" {
				t.Fatal("synchronous accepted assignment erased a typed per-user failure or invented a batch handle")
			}
			// No unassign/reset or package deletion is inferred from a partial
			// receipt. Those server and tenant contracts remain separately scoped.
			last, response, err := sdk.TeamsAdminAPI.AssignmentMigrations.RecommendationsByPolicyName(ctx, "groups", "TeamsMeetingPolicy", migration.RecommendationOptions{})
			if err == nil && (len(last) != 1 || last[0] == nil || !reflect.DeepEqual(last[0].RawFields, first[0].RawFields)) {
				t.Fatal("package partial failure mutated the independent recommendation snapshot")
			}
			return testutil.Observation{Value: last, Response: response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("recommendation/package composition differs: %v %+v", err, result.Differences())
	}
}
