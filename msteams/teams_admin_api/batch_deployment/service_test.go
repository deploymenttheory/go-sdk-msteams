package batchdeployment_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	deployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
)

const configOrigin = "https://api.interfaces.records.teams.microsoft.com"
const deploymentBase = "/Teams.FLWOrchestrator/orchestration/v1.0"

func discover(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: body}}
}

func batchWire(method, path string, body json.RawMessage, status int, text string) contracttest.Exchange {
	command := "Get-CsBatchTeamsDeploymentStatus"
	if method == "POST" {
		command = "New-CsBatchTeamsDeployment"
	}
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
	if body != nil {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + deploymentBase + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(text)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "batch_deployment", "msteams/teams_admin_api/batch_deployment/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
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
	run(t, contracttest.Case{Name: "submit-request", Mappings: []contracttest.Mapping{{Command: "New-CsBatchTeamsDeployment", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{discover(t), batchWire("POST", "/batchflwdeployment", createBody(t, archive), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestUnitBatchStatusIsOpaqueAndIdentityInputIsNotAURI(t *testing.T) {
	t.Parallel()

	for _, identity := range []bool{false, true} {
		set := "Get"
		if identity {
			set = "GetViaIdentity"
		}
		run(t, contracttest.Case{Name: "status-" + strings.ToLower(set), Mappings: []contracttest.Mapping{{Command: "Get-CsBatchTeamsDeploymentStatus", ParameterSet: set}}, Steps: []contracttest.Exchange{discover(t), batchWire("GET", "/scenarios/"+url.PathEscape("fixture/instance"), nil, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			if identity {
				v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByIdentity(ctx, deployment.Identity{OrchestrationID: "fixture/instance"})
				var value any
				if v != nil {
					value = v.Text
				}
				return contracttest.Observation{Value: value, Response: r, Err: err}
			}
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture/instance")
			var value any
			if v != nil {
				value = v.Text
			}
			return contracttest.Observation{Value: value, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	}
}

func TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	for _, status := range []int{400, 404, 429, 503} {
		run(t, contracttest.Case{Name: "submit-failure", Steps: []contracttest.Exchange{discover(t), batchWire("POST", "/batchflwdeployment", createBody(t, archive), status, "private raw error")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, Body: []byte("private raw error"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
	for _, status := range []int{400, 404} {
		run(t, contracttest.Case{Name: "status-failure", Steps: []contracttest.Exchange{discover(t), batchWire("GET", "/scenarios/fixture", nil, status, "raw status failure")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture")
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, Body: []byte("raw status failure"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
}

func TestUnitBatchRejectsInvalidInputBeforeDiscovery(t *testing.T) {
	t.Parallel()
	for _, input := range []deployment.CreateRequest{{}, {Archive: testArchive(t)}, {UsersToNotify: "fixture@example.invalid"}} {
		run(t, contracttest.Case{Name: "invalid-submit", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, input)
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Error: deployment.ErrInput}})
	}
	run(t, contracttest.Case{Name: "cancelled-status", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
}

func TestUnitBatchSubmissionLocalModelKeepsOptionalIdentities(t *testing.T) {
	t.Parallel()
	var v deployment.Submission
	if err := json.Unmarshal([]byte(`{"executionId":"fixture-execution"}`), &v); err != nil || v.InstanceID != nil || v.ExecutionID == nil || *v.ExecutionID != "fixture-execution" {
		t.Fatal("local submission decoding lost optional identity")
	}
}
