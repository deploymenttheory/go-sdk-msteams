package teammanagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const ownerID = "00000000-0000-4000-8000-000000000001"
const emptyCreationSettings = `{"memberSettings":{},"guestSettings":{},"messagingSettings":{},"funSettings":{},"discoverySettings":{}}`

func creationRequest() management.CreateRequest {
	alias := "fixture"
	return management.CreateRequest{DisplayName: "Fixture", MailNickname: &alias, Owner: ownerID}
}

func creationCase(t *testing.T, set string, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Mappings = []contracttest.Mapping{{Command: "New-Team", ParameterSet: set}}
	return runCase(t, "New-Team", c)
}

func creationObserved(result management.CreationResult, err error) contracttest.Observation {
	var name *string
	if result.Value != nil {
		name = result.Value.DisplayName
	}
	stages := []string{}
	for _, stage := range result.Stages {
		stages = append(stages, string(stage.Action)+":"+string(stage.State))
	}
	primary, cleanup := 0, 0
	if result.PrimaryResponse != nil {
		primary = result.PrimaryResponse.StatusCode()
	}
	if result.Cleanup.Response != nil {
		cleanup = result.Cleanup.Response.StatusCode()
	}
	return contracttest.Observation{Value: map[string]any{"id": result.ID, "owned": result.CreatedGroupOwned, "owner_id": result.OwnerID, "owner_source": result.OwnerSource, "alias": result.MailNickname, "generated": result.GeneratedMailNickname, "name": name, "cleanup": result.Cleanup.Disposition, "primary_status": primary, "cleanup_status": cleanup, "stages": stages}, Response: result.Response, Err: err}
}

func expectedCreation(owned bool, name, disposition string, primary, cleanup int, stages string) json.RawMessage {
	owner, alias, resolution := "", "", ""
	if owned {
		owner, alias, resolution = ownerID, "fixture", "identifier"
	}
	return json.RawMessage(fmt.Sprintf(`{"id":%q,"owned":%t,"owner_id":%q,"owner_source":%q,"alias":%q,"generated":false,"name":%s,"cleanup":%q,"primary_status":%d,"cleanup_status":%d,"stages":%s}`, groupID, owned, owner, resolution, alias, name, disposition, primary, cleanup, stages))
}

func TestCreationAndMigrationUsePinnedSourceDerivedSequences(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, name := range []string{"team-create-owned-group", "team-migrate-existing-success", "team-migrate-existing-failure"} {
		t.Run(name, func(t *testing.T) {
			set := "MigrateGroup"
			want := contracttest.Expected{Status: 200, JSON: expectedCreation(false, `"Group name"`, "supplied-group-retained", 0, 0, `["group-availability:succeeded","teamify-group:succeeded","read-group:succeeded","read-team:succeeded"]`)}
			if name == "team-create-owned-group" {
				set = "CreateTeam"
				want.JSON = expectedCreation(true, `"Group name"`, "not-required", 0, 0, `["resolve-owner:succeeded","create-group:succeeded","group-availability:succeeded","teamify-group:succeeded","read-group:succeeded","read-team:succeeded"]`)
			}
			if name == "team-migrate-existing-failure" {
				want = contracttest.Expected{Status: 403, Error: auth.ErrForbidden, JSON: expectedCreation(false, `null`, "supplied-group-retained", 403, 0, `["group-availability:succeeded","teamify-group:unconfirmed"]`)}
			}
			results = append(results, creationCase(t, set, contracttest.Case{Name: name, Steps: fixture(t, name), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var result management.CreationResult
				var err error
				if set == "CreateTeam" {
					result, err = sdk.GraphAPI.TeamManagement.Create(ctx, creationRequest())
				} else {
					result, err = sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, groupID, management.CreationOptions{})
				}
				for _, value := range []any{result, result.Cleanup} {
					if rendered := fmt.Sprintf("%+v %#v", value, value); strings.Contains(rendered, groupID) || strings.Contains(rendered, ownerID) {
						t.Fatal("creation formatting exposed private recovery metadata")
					}
				}
				return creationObserved(result, err)
			}, Want: want}))
		})
	}
	evidence(t, "team-create-service.json", results)
}

func TestGroupCreationBindingsUseSelectedCloudAndConfiguredMode(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			env, err := auth.ResolveEnvironment(cloud, nil)
			if err != nil {
				t.Fatal(err)
			}
			steps := fixture(t, "team-create-owned-group")
			for i := range steps {
				steps[i].Request.URL = strings.ReplaceAll(steps[i].Request.URL, auth.GraphAudience, env.Endpoints.Graph)
				steps[i].Request.Audience = env.Endpoints.Graph
				if steps[i].Request.JSON != nil {
					steps[i].Request.JSON = json.RawMessage(strings.ReplaceAll(string(steps[i].Request.JSON), auth.GraphAudience, env.Endpoints.Graph))
				}
			}
			creationCase(t, "CreateTeam", contracttest.Case{Name: "cloud-binding", Cloud: cloud, Mode: mode, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.TeamManagement.Create(ctx, creationRequest())
				return contracttest.Observation{Response: result.Response, Err: err}
			}, Want: contracttest.Expected{Status: 200}})
		}
	}
}

func TestOwnerResolutionPrecedesAnyGroupWrite(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"delegated-current", "direct-upn", "denied-upn"} {
		selector, path := "", "/v1.0/me?$select=id"
		if name != "delegated-current" {
			selector = "owner@example.invalid"
			path = "/v1.0/users/owner@example.invalid?$select=id"
		}
		steps := []contracttest.Exchange{wire("GET", path, "", 200, `{"id":"`+ownerID+`"}`)}
		want := contracttest.Expected{Status: 200}
		if name == "denied-upn" {
			steps[0].Reply.Status = 403
			steps[0].Reply.Body = []byte(`{"error":{"code":"Forbidden"}}`)
			want = contracttest.Expected{Status: 403, Error: auth.ErrForbidden}
		} else {
			steps = append(steps, fixture(t, "team-create-owned-group")...)
		}
		creationCase(t, "CreateTeam", contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			value.Owner = selector
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: want})
	}
}

func TestEducationPresetsMatchActualPinnedModelCaptures(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var captured struct {
		Cases []struct {
			Name string
			Wire map[string]json.RawMessage
		}
	}
	if json.Unmarshal(raw, &captured) != nil {
		t.Fatal("pinned education capture unavailable")
	}
	for _, tc := range []struct {
		name     string
		template management.GroupTemplate
	}{{"group-template-EDU_Class", management.EducationClass}, {"group-template-EDU_PLC", management.EducationPLC}} {
		steps := fixture(t, "team-create-owned-group")
		var body map[string]json.RawMessage
		if json.Unmarshal(steps[0].Request.JSON, &body) != nil {
			t.Fatal("ordinary source fixture malformed")
		}
		found := false
		for _, c := range captured.Cases {
			if c.Name == tc.name {
				for key, value := range c.Wire {
					body[key] = value
				}
				found = true
			}
		}
		if !found {
			t.Fatal("education model capture missing")
		}
		steps[0].Request.JSON, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		creationCase(t, "CreateTeam", contracttest.Case{Name: strings.ToLower(strings.ReplaceAll(tc.name, "_", "-")), Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			value.Template = tc.template
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200}})
	}
}

func TestCreationPropagationIsBoundedAndOnlyRetriesReceived404(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		happy := fixture(t, "team-create-owned-group")
		missingGroup := wire("GET", "/v1.0/groups/"+groupID, "", 404, `{"error":{"code":"ResourceNotFound"}}`)
		missingTeam := wire("PUT", "/v1.0/groups/"+groupID+"/team", emptyCreationSettings, 404, `{"error":{"code":"ResourceNotFound"}}`)
		steps := []contracttest.Exchange{happy[0], missingGroup, missingGroup, missingGroup, happy[1], missingTeam, missingTeam, missingTeam, happy[2], happy[3], happy[4]}
		start := time.Now()
		creationCase(t, "CreateTeam", contracttest.Case{Name: "three-retries-per-stage", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, creationRequest())
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200}})
		if time.Since(start) != 60*time.Second {
			t.Fatal("propagation retry bound or interval differs")
		}
	})
	synctest.Test(t, func(t *testing.T) {
		happy := fixture(t, "team-create-owned-group")
		missing := wire("PUT", "/v1.0/groups/"+groupID+"/team", emptyCreationSettings, 404, `{"error":{"code":"ResourceNotFound"}}`)
		missing.Reply.Headers.Set("Retry-After", "20")
		steps := []contracttest.Exchange{happy[0], happy[1], missing, happy[2], happy[3], happy[4]}
		start := time.Now()
		creationCase(t, "CreateTeam", contracttest.Case{Name: "propagation-retry-after", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			value.Options.Interval = time.Second
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200}})
		if time.Since(start) != 20*time.Second {
			t.Fatal("server Retry-After was shortened")
		}
	})
}

func TestCleanupKeepsPrimaryAndCleanupFailuresDistinct(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"delete-succeeded", "delete-denied", "retain-requested", "ambiguous-404", "ambiguous-existing-team", "reconcile-denied", "rejected-existing-team"} {
		happy := fixture(t, "team-create-owned-group")
		put := wire("PUT", "/v1.0/groups/"+groupID+"/team", emptyCreationSettings, 400, `{"error":{"code":"FixtureRejected"}}`)
		primaryCode, primaryStatus := "FixtureRejected", 400
		if strings.HasPrefix(name, "ambiguous") {
			put.Reply.Status = 503
			put.Reply.Body = []byte(`{"error":{"code":"ServiceUnavailable"}}`)
			primaryCode, primaryStatus = "ServiceUnavailable", 503
		}
		steps := []contracttest.Exchange{happy[0], happy[1], put}
		disposition := management.CleanupDeleted
		cleanupStatus := 0
		reconcile := wire("GET", "/v1.0/teams/"+groupID, "", 404, `{"error":{"code":"ResourceNotFound"}}`)
		switch name {
		case "retain-requested":
			disposition = management.CleanupRetainedByRequest
		case "ambiguous-404":
			steps = append(steps, reconcile)
			disposition = management.CleanupUnresolved
		case "ambiguous-existing-team", "rejected-existing-team":
			steps = append(steps, happy[4])
			disposition = management.CleanupTeamExists
		case "reconcile-denied":
			reconcile.Reply.Status = 403
			reconcile.Reply.Body = []byte(`{"error":{"code":"Forbidden"}}`)
			steps = append(steps, reconcile)
			disposition = management.CleanupUnresolved
		default:
			steps = append(steps, reconcile)
			cleanupStatus = 204
			deletion := wire("DELETE", "/v1.0/groups/"+groupID, "", 204, "")
			if name == "delete-denied" {
				cleanupStatus = 403
				deletion.Reply.Status = 403
				deletion.Reply.Body = []byte(`{"error":{"code":"Forbidden"}}`)
				disposition = management.CleanupDeleteUnconfirmed
			}
			steps = append(steps, deletion)
		}
		creationCase(t, "CreateTeam", contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			value.RetainCreatedGroup = name == "retain-requested"
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			var primary *client.APIError
			if !errors.As(result.PrimaryError, &primary) || primary.StatusCode != primaryStatus || len(primary.Codes) != 1 || primary.Codes[0] != primaryCode {
				t.Fatal("primary teamification failure was lost")
			}
			if result.ID != groupID || !result.CreatedGroupOwned || result.Value != nil || result.Cleanup.Disposition != disposition {
				t.Fatal("known ownership, partial result or cleanup disposition differs")
			}
			if name == "delete-denied" && (!errors.Is(err, auth.ErrForbidden) || !errors.Is(result.Cleanup.Error, auth.ErrForbidden)) {
				t.Fatal("cleanup rejection was not joined with primary failure")
			}
			if disposition == management.CleanupUnresolved && !errors.Is(err, management.ErrCreationReconciliation) {
				t.Fatal("unresolved compensation was not explicit")
			}
			if cleanupStatus != 0 && (result.Cleanup.Response == nil || result.Cleanup.Response.StatusCode() != cleanupStatus) {
				t.Fatal("cleanup response was lost")
			}
			return contracttest.Observation{Err: err}
		}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.Graph, StatusCode: primaryStatus, Codes: []string{primaryCode}}}})
	}
}

func TestCreatedTeamIsRetainedWhenEnrichmentFails(t *testing.T) {
	t.Parallel()
	steps := fixture(t, "team-create-owned-group")
	steps[3].Reply.Status = 403
	steps[3].Reply.Body = []byte(`{"error":{"code":"Forbidden"}}`)
	steps = steps[:4]
	creationCase(t, "CreateTeam", contracttest.Case{Name: "created-team-readback-denied", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.TeamManagement.Create(ctx, creationRequest())
		if result.Teamified == nil || result.Teamified.ID != groupID || result.Cleanup.Disposition != management.CleanupTeamExists || result.Value != nil {
			t.Fatal("successful team creation was lost after readback failure")
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestCreationTimeoutRetainsIDAndUsesBoundedFreshCleanupContext(t *testing.T) {
	for _, phase := range []string{"availability", "teamification"} {
		synctest.Test(t, func(t *testing.T) {
			happy := fixture(t, "team-create-owned-group")
			steps := []contracttest.Exchange{happy[0]}
			if phase == "availability" {
				steps = append(steps, wire("GET", "/v1.0/groups/"+groupID, "", 404, `{"error":{"code":"Missing"}}`))
			} else {
				steps = append(steps, happy[1], wire("PUT", "/v1.0/groups/"+groupID+"/team", emptyCreationSettings, 404, `{"error":{"code":"Missing"}}`), wire("GET", "/v1.0/teams/"+groupID, "", 404, `{"error":{"code":"Missing"}}`))
			}
			steps = append(steps, wire("DELETE", "/v1.0/groups/"+groupID, "", 204, ""))
			start := time.Now()
			creationCase(t, "CreateTeam", contracttest.Case{Name: "timeout-" + phase, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value := creationRequest()
				value.Options.Timeout = time.Second
				result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
				if result.ID != groupID || result.Cleanup.Disposition != management.CleanupDeleted {
					t.Fatal("context end lost known group identity or cleanup")
				}
				return contracttest.Observation{Response: result.Response, Err: err}
			}, Want: contracttest.Expected{Status: 204, Error: context.DeadlineExceeded}})
			if time.Since(start) != time.Second {
				t.Fatal("foreground or cleanup context bound differs")
			}
		})
	}
}

func TestCreationValidationPrecedesCredentialsAndWrites(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"empty-name", "empty-alias", "invalid-template", "class-explicit-visibility", "invalid-visibility", "negative-timeout", "negative-attempts", "negative-interval", "negative-cleanup-timeout", "application-without-owner"} {
		mode := auth.Delegated
		want := client.ErrRequest
		if name == "application-without-owner" {
			mode = auth.Application
			want = directoryuser.ErrCurrentUser
		}
		creationCase(t, "CreateTeam", contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			switch name {
			case "empty-name":
				value.DisplayName = " "
			case "empty-alias":
				empty := ""
				value.MailNickname = &empty
			case "invalid-template":
				value.Template = "standard"
			case "class-explicit-visibility":
				visibility := group.Private
				value.Template = management.EducationClass
				value.Visibility = &visibility
			case "invalid-visibility":
				visibility := group.Visibility("future")
				value.Visibility = &visibility
			case "negative-timeout":
				value.Options.Timeout = -1
			case "negative-attempts":
				value.Options.MaxAttempts = -1
			case "negative-interval":
				value.Options.Interval = -1
			case "negative-cleanup-timeout":
				value.Options.CleanupTimeout = -1
			case "application-without-owner":
				value.Owner = ""
			}
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}
