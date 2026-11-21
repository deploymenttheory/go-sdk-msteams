package batchdeployment_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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

func discover() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored"}}`)}}
}
func batchWire(method, path string, body json.RawMessage, status int, text string) contracttest.Exchange {
	command := "Get-CsBatchTeamsDeploymentStatus"
	if method == "POST" {
		command = "New-CsBatchTeamsDeployment"
	}
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://forest.fixture.invalid/"}}
	if body != nil {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + deploymentBase + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(text)}}
}
func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family, c.Source = "batch_deployment", "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("batch deployment %s differs: setup=%v differences=%+v", c.Name, err, r.Differences())
	}
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

func TestBatchSubmissionUsesSelectedConfigAPIAndRetainsBothIdentities(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	body := createBody(t, archive)
	run(t, contracttest.Case{Name: "submit-configapi", Mappings: []contracttest.Mapping{{Command: "New-CsBatchTeamsDeployment", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{discover(), batchWire("POST", "/batchflwdeployment", body, 202, `{"instanceId":"fixture/instance","executionId":"fixture-execution","future":{"v":8}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
		var value any
		if v != nil {
			value = map[string]any{"instance": v.InstanceID, "execution": v.ExecutionID, "raw": v.RawFields}
		}
		return contracttest.Observation{Value: value, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 202, JSON: json.RawMessage(`{"instance":"fixture/instance","execution":"fixture-execution","raw":{"instanceId":"fixture/instance","executionId":"fixture-execution","future":{"v":8}}}`)}})
}

func TestBatchStatusIsOpaqueAndIdentityInputIsNotAURI(t *testing.T) {
	t.Parallel()
	for _, identity := range []bool{false, true} {
		set := "Get"
		if identity {
			set = "GetViaIdentity"
		}
		run(t, contracttest.Case{Name: "status-" + strings.ToLower(set), Mappings: []contracttest.Mapping{{Command: "Get-CsBatchTeamsDeploymentStatus", ParameterSet: set}}, Steps: []contracttest.Exchange{discover(), batchWire("GET", "/scenarios/"+url.PathEscape("fixture/instance"), nil, 200, "Completed; raw unparsed state")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
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
		}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`"Completed; raw unparsed state"`)}})
	}
}

func TestBatchErrorsRemainErrorsAndSubmissionIsNotReplayed(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	for _, status := range []int{400, 404, 429, 503} {
		run(t, contracttest.Case{Name: "submit-failure", Steps: []contracttest.Exchange{discover(), batchWire("POST", "/batchflwdeployment", createBody(t, archive), status, "private raw error")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, Body: []byte("private raw error"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
	for _, status := range []int{400, 404} {
		run(t, contracttest.Case{Name: "status-failure", Steps: []contracttest.Exchange{discover(), batchWire("GET", "/scenarios/fixture", nil, status, "raw status failure")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, "fixture")
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, Body: []byte("raw status failure"), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
	}
}

func TestBatchRejectsInvalidInputBeforeDiscovery(t *testing.T) {
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

func TestBatchAcceptedWithoutUsableInstanceKeepsMetadata(t *testing.T) {
	t.Parallel()
	archive := testArchive(t)
	run(t, contracttest.Case{Name: "accepted-missing-instance", Steps: []contracttest.Exchange{discover(), batchWire("POST", "/batchflwdeployment", createBody(t, archive), 202, `{"executionId":"fixture-execution"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.BatchDeployments.Create(ctx, deployment.CreateRequest{Archive: archive, UsersToNotify: "fixture@example.invalid"})
		if v == nil || v.ExecutionID == nil || v.RawFields == nil {
			t.Error("accepted response lost partial identities")
		}
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 202, Error: client.ErrDecode}})
}
