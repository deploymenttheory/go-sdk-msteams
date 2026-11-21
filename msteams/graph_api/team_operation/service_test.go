package teamoperation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	operation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
)

const teamID = "11111111-1111-4111-8111-111111111111"
const operationID = "opaque/operation+#雪"

func path() string { return "/beta/teams/" + teamID + "/operations/" + url.PathEscape(operationID) }

func wire(status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: auth.GraphAudience + path(), Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func body(status string) string { return `{"id":"` + operationID + `","status":"` + status + `"}` }

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "team_operations", "contracts/endpoints/8.1.0/core.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("operation contract differs: setup=%v differences=%+v", e, r.Differences())
	}
	return r
}

func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, e := os.OpenRoot(t.ArtifactDir())
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = root.Close() }()
	f, e := root.Create("team-operation-service.json")
	if e != nil {
		t.Fatal(e)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("operation structural evidence write failed")
	}
}

func TestUnitGetOperationPreservesNativeStatesAndFutureData(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, status := range []string{"notStarted", "inProgress", "succeeded", "failed", "futureStatus"} {
		t.Run(status, func(t *testing.T) {
			raw := `{"id":"` + operationID + `","status":"` + status + `","operationType":"futureType","attemptsCount":0,"createdDateTime":"2026-01-01T00:00:00Z","lastActionDateTime":null,"targetResourceId":"fixture-resource","targetResourceLocation":"https://foreign.invalid/opaque","error":{"code":"FixtureFailure","message":"private-message","future":7},"future":{"enabled":false}}`
			results = append(results, run(t, contracttest.Case{Name: "get-" + strings.ToLower(status), Mappings: []contracttest.Mapping{{Command: "Get-Operation", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{wire(200, raw)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
				var value any
				if v != nil {
					value = map[string]any{"id": v.ID, "status": v.Status, "type": v.OperationType, "attempts": v.AttemptsCount, "target": v.TargetResourceLocation, "raw": v.RawFields}
					if v.Error == nil || v.Error.Code == nil || *v.Error.Code != "FixtureFailure" || string(v.Error.RawFields["future"]) != "7" {
						t.Error("typed/future failure diagnostics were lost")
					}
				}
				return contracttest.Observation{Value: value, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + operationID + `","status":"` + status + `","type":"futureType","attempts":0,"target":"https://foreign.invalid/opaque","raw":` + raw + `}`)}}))
		})
	}
	evidence(t, results)
}

func TestUnitGetRejectsMalformedAndUnexpectedResponses(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"malformed", "{", 200, client.ErrDecode}, {"missing-id", `{"status":"succeeded"}`, 200, client.ErrDecode}, {"null", "null", 200, client.ErrDecode}, {"array", "[]", 200, client.ErrDecode}, {"invalid-time", `{"id":"x","createdDateTime":"invalid"}`, 200, client.ErrDecode}, {"invalid-error", `{"id":"x","error":"invalid"}`, 200, client.ErrDecode},
		{"accepted-status", body("inProgress"), 202, operation.ErrUnexpectedStatus}, {"partial-status", body("succeeded"), 206, operation.ErrUnexpectedStatus},
		{"forbidden", `{"error":{"code":"Denied"}}`, 403, auth.ErrForbidden}, {"unauthenticated", `{"error":{"code":"Unauthorized"}}`, 401, auth.ErrUnauthenticated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			results = append(results, run(t, contracttest.Case{Name: "get-" + tc.name, Steps: []contracttest.Exchange{wire(tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}}))
		})
	}
	evidence(t, results)
}

func TestUnitResumedWaitUsesBetaCadenceAndRetainsFailure(t *testing.T) {
	for _, terminal := range []string{"succeeded", "failed", "futureStatus"} {
		t.Run(terminal, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				started := time.Now()
				var results []contracttest.Result
				var expected error
				if terminal == "failed" {
					expected = client.ErrOperationFailed
				}
				if terminal == "futureStatus" {
					expected = client.ErrOperationState
				}
				raw := `{"id":"` + operationID + `","status":"` + terminal + `","targetResourceLocation":"https://foreign.invalid/never-follow","error":{"code":"FixtureFailure","message":"private-message"}}`
				results = append(results, run(t, contracttest.Case{Name: "wait-" + strings.ToLower(terminal), Steps: []contracttest.Exchange{wire(200, body("inProgress")), wire(200, raw)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					value, e := sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: 2, Interval: time.Millisecond, Timeout: time.Minute})
					if value.Handle.ID != operationID || value.Handle.Location != auth.GraphAudience+path() {
						t.Error("resumable beta identity was lost")
					}
					if terminal == "failed" && (value.Diagnostics == nil || value.Diagnostics.StatusCode != 200 || len(value.Diagnostics.Codes) != 1 || value.Diagnostics.Codes[0] != "FixtureFailure") {
						t.Error("failed operation diagnostics were lost")
					}
					if terminal == "futureStatus" && value.Status.State != client.OperationRunning {
						t.Error("unknown state changed the prior observation")
					}
					return contracttest.Observation{Response: value.Response, Err: e}
				}, Want: contracttest.Expected{Status: 200, Error: expected}}))
				if time.Since(started) < 31*time.Second {
					t.Fatal("beta status checks ignored the Graph minimum")
				}
				evidence(t, results)
			})
		})
	}
}

func TestUnitReadRetryAndCancellationRespectMinimumCadence(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		started := time.Now()
		r := run(t, contracttest.Case{Name: "get-retry-floor", Steps: []contracttest.Exchange{wire(503, `{"error":{"code":"Busy"}}`), wire(200, body("succeeded"))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, Want: contracttest.Expected{Status: 200}})
		if time.Since(started) < 31*time.Second {
			t.Fatal("read retry polled too early")
		}
		evidence(t, []contracttest.Result{r})
	})
	synctest.Test(t, func(t *testing.T) {
		r := run(t, contracttest.Case{Name: "wait-cancel-before-next-read", Steps: []contracttest.Exchange{wire(200, body("inProgress"))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			op, e := sdk.GraphAPI.TeamOperations.ResumeByID(teamID, operationID)
			if e != nil {
				return contracttest.Observation{Err: e}
			}
			short, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			v, e := op.Wait(short, client.WaitOptions{MaxPolls: 2, Interval: time.Millisecond, Timeout: time.Minute})
			if v.Handle.ID != operationID || v.Status.State != client.OperationRunning {
				t.Error("cancelled wait lost its resumable state")
			}
			return contracttest.Observation{Response: v.Response, Err: e}
		}, Want: contracttest.Expected{Status: 200, Error: context.DeadlineExceeded}})
		evidence(t, []contracttest.Result{r})
	})
}

func TestUnitLocalValidationAndResumeDoNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	for index, id := range []string{"", ".", "..", " ", "line\n", "back\\slash"} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			run(t, contracttest.Case{Name: "invalid-id-" + strconv.Itoa(index), Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, id)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Error: client.ErrRequest}})
		})
	}
	run(t, contracttest.Case{Name: "resume-local", Scope: "local", Invoke: func(_ context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.TeamOperations.ResumeByID(teamID, operationID)
		if e == nil {
			snapshot := v.Snapshot()
			if snapshot.Status.State != client.OperationAccepted || snapshot.Handle.Location != auth.GraphAudience+path() {
				t.Error("local handle did not retain beta route")
			}
		}
		return contracttest.Observation{Err: e}
	}, Want: contracttest.Expected{}})
	run(t, contracttest.Case{Name: "cancelled", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		v, r, e := sdk.GraphAPI.TeamOperations.GetByID(cancelled, teamID, operationID)
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Error: context.Canceled}})
	run(t, contracttest.Case{Name: "invalid-wait", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, e := sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: -1})
		return contracttest.Observation{Err: e}
	}, Want: contracttest.Expected{Error: client.ErrOperation}})
	value := operation.Operation{ID: operationID, Error: &operation.OperationError{Message: stringPointer("private-message")}}
	for _, v := range []any{value, *value.Error} {
		if text := fmt.Sprintf("%+v %#v", v, v); strings.Contains(text, operationID) || strings.Contains(text, "private-message") {
			t.Fatal("snapshot formatting leaked server data")
		}
	}
}

func stringPointer(value string) *string { return &value }
