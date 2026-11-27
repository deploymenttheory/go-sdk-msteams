package batchdeployment_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	deployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
)

const configOrigin = "https://api.interfaces.records.teams.microsoft.com"
const deploymentBase = "/Teams.FLWOrchestrator/orchestration/v1.0"

func discover(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: body}}
}

func batchWire(method, path string, body json.RawMessage, status int, text string) testutil.Exchange {
	command := "Get-CsBatchTeamsDeploymentStatus"
	if method == "POST" {
		command = "New-CsBatchTeamsDeployment"
	}
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
	if body != nil {
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + deploymentBase + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(text)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("batch deployment %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
	return r
}
func testArchive(t *testing.T) deployment.Archive {
	t.Helper()
	value, err := deployment.BuildArchive(t.Context(), strings.NewReader("original teams\r\n"), strings.NewReader("original users\n"), budget)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func createBody(t *testing.T, archive deployment.Archive) json.RawMessage {
	t.Helper()
	value, err := json.Marshal(map[string]string{"DeploymentCsvs": archive.Encoded(), "UsersToNotify": "fixture@example.invalid"})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestUnitBatchSubmissionRequestAndDenial(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	run(t, testutil.Case{Name: "submit-request", Steps: []testutil.Exchange{discover(t), batchWire("POST", "/batchflwdeployment", createBody(t, archive), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
		return testutil.Observation{Value: v, Response: r, Err: err}
	}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestUnitBatchStatusIsOpaqueAndIdentityInputIsNotAURI(t *testing.T) {
	t.Parallel()

	for _, identity := range []bool{false, true} {
		set := "Get"
		if identity {
			set = "GetViaIdentity"
		}
		run(t, testutil.Case{Name: "status-" + strings.ToLower(set), Steps: []testutil.Exchange{discover(t), batchWire("GET", "/scenarios/"+url.PathEscape("fixture/instance"), nil, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			if identity {
				v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByIdentity(ctx, deployment.Identity{OrchestrationID: "fixture/instance"})
				var value any
				if v != nil {
					value = v.Text
				}
				return testutil.Observation{Value: value, Response: r, Err: err}
			}
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture/instance")
			var value any
			if v != nil {
				value = v.Text
			}
			return testutil.Observation{Value: value, Response: r, Err: err}
		}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden}})
	}
}

func TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	for _, status := range []int{400, 404, 429, 503} {
		run(t, testutil.Case{Name: "submit-failure", Steps: []testutil.Exchange{discover(t), batchWire("POST", "/batchflwdeployment", createBody(t, archive), status, "private raw error")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Status: status, Body: []byte("private raw error"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
	for _, status := range []int{400, 404} {
		run(t, testutil.Case{Name: "status-failure", Steps: []testutil.Exchange{discover(t), batchWire("GET", "/scenarios/fixture", nil, status, "raw status failure")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture")
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Status: status, Body: []byte("raw status failure"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
}

func TestUnitBatchRejectsInvalidInputBeforeDiscovery(t *testing.T) {
	t.Parallel()
	for _, input := range []deployment.CreateRequest{{}, {Archive: testArchive(t)}, {UsersToNotify: "fixture@example.invalid"}} {
		run(t, testutil.Case{Name: "invalid-submit", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, input)
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Error: deployment.ErrInput}})
	}
	run(t, testutil.Case{Name: "cancelled-status", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture")
		return testutil.Observation{Value: v, Response: r, Err: err}
	}, Want: testutil.Expected{Error: context.Canceled}})
}

func TestUnitBatchSubmissionLocalModelKeepsOptionalIdentities(t *testing.T) {
	t.Parallel()
	var v deployment.Submission
	if err := json.Unmarshal([]byte(`{"executionId":"fixture-execution"}`), &v); err != nil || v.InstanceID != nil || v.ExecutionID == nil || *v.ExecutionID != "fixture-execution" {
		t.Fatal("local submission decoding lost optional identity")
	}
}

// These are authored HTTP vectors from pinned 8.1.0 response handlers, not a
// successful tenant deployment. Acceptance only supplies a status identity;
// Get preserves arbitrary status text without deciding workflow completion.
func TestUnitBatchAcceptedSubmissionAndOpaqueStatus(t *testing.T) {
	t.Parallel()
	// Exact ordered columns follow the linked frontline deployment guide.
	// Identity/name values are authored, and no deployment is submitted live.
	teams := "Team Name,Existing Team ID,Visibility,Team Template ID\r\nOffline team,,Private,\r\n"
	users := "User Full Name,User UPN or ID,Team Name,ActionType,Owner or Member\r\nOffline owner,offline-owner@example.invalid,Offline team,AddMember,Owner\r\n"
	archive, buildErr := deployment.BuildArchive(t.Context(), strings.NewReader(teams), strings.NewReader(users), budget)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	accepted := batchWire("POST", "/batchflwdeployment", createBody(t, archive), 202, `{"instanceId":"offline/instance","executionId":"offline-execution","future":false}`)

	for _, identity := range []bool{false, true} {
		set := "Get"
		if identity {
			set = "GetViaIdentity"
		}
		status := batchWire("GET", "/scenarios/"+url.PathEscape("offline/instance"), nil, 200, "Completed\r\npartial result; not JSON")

		run(t, testutil.Case{Name: "accepted-then-" + strings.ToLower(set), Steps: []testutil.Exchange{discover(t), accepted, status}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			submission, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
			if err != nil {
				return testutil.Observation{Response: r, Err: err}
			}
			if submission == nil || submission.InstanceID == nil || *submission.InstanceID != "offline/instance" || submission.ExecutionID == nil || *submission.ExecutionID != "offline-execution" || string(submission.RawFields["future"]) != "false" {
				t.Fatal("accepted identities or raw metadata lost")
			}
			var value *deployment.Status
			if identity {
				value, r, err = sdk.TeamsAdminAPI.BatchDeployments.GetByIdentity(ctx, deployment.Identity{OrchestrationID: *submission.InstanceID})
			} else {
				value, r, err = sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, *submission.InstanceID)
			}
			var text any
			if value != nil {
				text = value.Text
			}
			return testutil.Observation{Value: text, Response: r, Err: err}
		}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`"Completed\r\npartial result; not JSON"`)}})
	}
}
