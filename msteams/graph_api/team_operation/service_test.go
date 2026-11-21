package teamoperation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
	c.Family, c.Source = "team_operations", "msteams/graph_api/team_operation/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, e := contracttest.Run(t.Context(), requestBoundary(t, c))
	if e != nil || !r.Passed() {
		t.Fatalf("operation contract differs: setup=%v differences=%+v", e, r.Differences())
	}
	return r
}

func TestUnitGetOperationPreservesNativeStatesAndFutureData(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"notStarted", "inProgress", "succeeded", "failed", "futureStatus"} {
		t.Run(status, func(t *testing.T) {
			raw := `{"id":"` + operationID + `","status":"` + status + `","operationType":"futureType","attemptsCount":0,"createdDateTime":"2026-01-01T00:00:00Z","lastActionDateTime":null,"targetResourceId":"fixture-resource","targetResourceLocation":"https://foreign.invalid/opaque","error":{"code":"FixtureFailure","message":"private-message","future":7},"future":{"enabled":false}}`
			run(t, contracttest.Case{Name: "get-" + strings.ToLower(status), Mappings: []contracttest.Mapping{{Command: "Get-Operation", ParameterSet: "__AllParameterSets"}}, Steps: []contracttest.Exchange{wire(200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				var value any
				if v != nil {
					value = map[string]any{"id": v.ID, "status": v.Status, "type": v.OperationType, "attempts": v.AttemptsCount, "target": v.TargetResourceLocation, "raw": v.RawFields}
					if v.Error == nil || v.Error.Code == nil || *v.Error.Code != "FixtureFailure" || string(v.Error.RawFields["future"]) != "7" {
						t.Error("typed/future failure diagnostics were lost")
					}
				}
				return contracttest.Observation{Value: value, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + operationID + `","status":"` + status + `","type":"futureType","attempts":0,"target":"https://foreign.invalid/opaque","raw":` + raw + `}`)}})
		})
	}
}

func TestUnitGetRejectsMalformedAndUnexpectedResponses(t *testing.T) {
	t.Parallel()

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
			run(t, contracttest.Case{Name: "get-" + tc.name, Steps: []contracttest.Exchange{wire(tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
		})
	}
}

func TestUnitWaitDenialRetainsResumableHandle(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "wait-denial", Steps: []contracttest.Exchange{wire(403, `{"error":{"code":"UnitRequestDenied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, err := sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: 2, Timeout: time.Minute})
		if value.Handle.ID != operationID || value.Handle.Location != auth.GraphAudience+path() {
			t.Error("denied operation read lost resumable identity")
		}
		return contracttest.Observation{Response: value.Response, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestUnitTransientReadThenDenialHonorsMinimumCadence(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		run(t, contracttest.Case{Name: "read-retry-denial", Steps: []contracttest.Exchange{wire(503, `{"error":{"code":"UnitBusy"}}`), wire(403, `{"error":{"code":"UnitRequestDenied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		if time.Since(start) < 31*time.Second {
			t.Error("transient retry ignored the Graph polling minimum")
		}
	})
}
func TestUnitCancelledOperationWaitIsLocal(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "cancelled-wait", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		op, err := sdk.GraphAPI.TeamOperations.ResumeByID(teamID, operationID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		value, err := op.Wait(cancelled, client.WaitOptions{MaxPolls: 2})
		if value.Handle.ID != operationID {
			t.Error("local cancellation lost the operation identity")
		}
		return contracttest.Observation{Response: value.Response, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
}

func TestUnitLocalValidationAndResumeDoNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	for index, id := range []string{"", ".", "..", " ", "line\n", "back\\slash"} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			run(t, contracttest.Case{Name: "invalid-id-" + strconv.Itoa(index), Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, id)
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: r, Err: e}
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Error: client.ErrRequest}})
		})
	}
	run(t, contracttest.Case{Name: "resume-local", Scope: "local", Invoke: func(_ context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.TeamOperations.ResumeByID(teamID, operationID)
		if e != nil {
			return contracttest.Observation{Err: e}
		}
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
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: r, Err: e}
		}
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Error: context.Canceled}})
	run(t, contracttest.Case{Name: "invalid-wait", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, e := sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: -1})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: nil, Err: e}
		}
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
