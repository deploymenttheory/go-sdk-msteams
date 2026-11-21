package client_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	teamsadminapi "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api"
	"github.com/go-resty/resty/v2"
)

type operationReply struct {
	status int
	body   string
	retry  string
}

func operationFixture(t *testing.T, replies []operationReply, decode client.OperationDecoder) (*client.Operation, *int) {
	t.Helper()
	calls := 0
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		i := calls
		calls++
		if i >= len(replies) {
			i = len(replies) - 1
		}
		reply := replies[i]
		header := http.Header{}
		if reply.retry != "" {
			header.Set("Retry-After", reply.retry)
		}
		return &http.Response{StatusCode: reply.status, Header: header, Body: io.NopCloser(strings.NewReader(reply.body))}, nil
	})}, "commercial")
	op, err := client.NewOperation(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture", "operation"}, Replay: client.ReplaySafe}, "fixture-id", decode, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	return op, &calls
}

func fastWait() client.WaitOptions {
	return client.WaitOptions{Timeout: time.Second, Interval: time.Millisecond, MaxPolls: 10}
}

func TestUnitOperationWaitSequencesAndTerminalDiagnostics(t *testing.T) {
	t.Parallel()
	for _, terminal := range []string{"succeeded", "failed"} {
		t.Run(terminal, func(t *testing.T) {
			t.Parallel()
			op, calls := operationFixture(t, []operationReply{
				{200, `{"id":"fixture-id","status":"notStarted"}`, ""},
				{200, `{"id":"fixture-id","status":"inProgress"}`, ""},
				{200, `{"id":"fixture-id","status":"` + terminal + `","error":{"code":"FixtureFailure","message":"private-server-message"},"targetResourceLocation":"https://foreign.invalid/private"}`, ""},
			}, graphapi.DecodeTeamOperation)
			if op.Snapshot().Status.State != client.OperationAccepted {
				t.Fatal("initial state was completion")
			}
			result, err := op.Wait(t.Context(), fastWait())
			if (terminal == "failed") != errors.Is(err, client.ErrOperationFailed) {
				t.Fatalf("wrong terminal outcome: %v", err)
			}
			if *calls != 3 || result.Handle.ID != "fixture-id" || result.Response == nil {
				t.Fatal("lost handle or response")
			}
			if terminal == "failed" && (result.Diagnostics == nil || len(result.Diagnostics.Codes) != 1 || result.Diagnostics.Codes[0] != "FixtureFailure") {
				t.Fatal("lost backend failure diagnostics")
			}
			_, again := op.Poll(t.Context())
			if *calls != 3 || !errors.Is(again, err) {
				t.Fatal("terminal operation polled again")
			}
			for _, value := range []any{op, result, result.Handle, result.Status} {
				if text := fmt.Sprintf("%+v %#v", value, value); strings.Contains(text, "fixture-id") || strings.Contains(text, "private") {
					t.Fatal("formatted operation leaked identifiers")
				}
			}
		})
	}
}

func TestUnitOperationCancellationAndPollingLimitAreResumable(t *testing.T) {
	t.Parallel()
	op, calls := operationFixture(t, []operationReply{
		{200, `{"id":"fixture-id","status":"inProgress"}`, ""},
		{200, `{"id":"fixture-id","status":"succeeded"}`, ""},
	}, graphapi.DecodeTeamOperation)
	options := fastWait()
	options.MaxPolls = 1
	result, err := op.Wait(t.Context(), options)
	if !errors.Is(err, client.ErrOperationPollLimit) || result.Handle.ID == "" {
		t.Fatal("poll limit lost recoverable handle")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = op.Wait(ctx, fastWait())
	if !errors.Is(err, context.Canceled) || *calls != 1 || result.Status.State != client.OperationRunning {
		t.Fatal("local cancellation changed remote state")
	}
	result, err = op.Wait(t.Context(), fastWait())
	if err != nil || result.Status.State != client.OperationSucceeded || *calls != 2 {
		t.Fatal("cancelled wait could not resume")
	}
	remote, _ := operationFixture(t, []operationReply{{200, `{}`, ""}}, func([]byte) (client.OperationStatus, error) {
		return client.OperationStatus{State: client.OperationCancelled}, nil
	})
	result, err = remote.Wait(t.Context(), fastWait())
	if !errors.Is(err, client.ErrOperationCancelled) || errors.Is(err, context.Canceled) || result.Handle.ID != "fixture-id" {
		t.Fatal("remote cancellation conflated with local cancellation")
	}
}

func TestUnitOperationRetryAfterAndStalledWait(t *testing.T) {
	t.Parallel()
	for _, header := range []string{"1", time.Now().Add(time.Minute).UTC().Format(http.TimeFormat)} {
		t.Run(header, func(t *testing.T) {
			op, calls := operationFixture(t, []operationReply{{200, `{"id":"fixture-id","status":"inProgress"}`, header}}, graphapi.DecodeTeamOperation)
			options := fastWait()
			options.Timeout = 20 * time.Millisecond
			result, err := op.Wait(t.Context(), options)
			if !errors.Is(err, context.DeadlineExceeded) || *calls != 1 || result.Handle.ID == "" {
				t.Fatal("server delay shortened or timeout lost handle")
			}
		})
	}
	stalled, calls := operationFixture(t, []operationReply{{200, `{"id":"fixture-id","status":"inProgress"}`, ""}}, graphapi.DecodeTeamOperation)
	options := fastWait()
	options.MaxPolls = 3
	if _, err := stalled.Wait(t.Context(), options); !errors.Is(err, client.ErrOperationPollLimit) || *calls != 3 {
		t.Fatal("stalled operation was not bounded")
	}
}

func TestUnitOperationRejectsUnknownStateIdentityChangeAnd202Success(t *testing.T) {
	t.Parallel()
	for _, reply := range []operationReply{
		{200, `{"id":"fixture-id","status":"unknownFutureValue"}`, ""},
		{200, `{"id":"changed","status":"succeeded"}`, ""},
		{202, `{"id":"fixture-id","status":"succeeded"}`, ""},
	} {
		op, calls := operationFixture(t, []operationReply{reply}, graphapi.DecodeTeamOperation)
		result, err := op.Wait(t.Context(), fastWait())
		if !errors.Is(err, client.ErrOperationState) || *calls != 1 || result.Status.State == client.OperationSucceeded || result.Handle.ID != "fixture-id" {
			t.Fatal("unknown/inconsistent state treated as success")
		}
	}
}

func TestUnitOperationPropagationRequiresExactBoundedEvidence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		code        string
		status      int
		rules       bool
		window      time.Duration
		wantSuccess bool
	}{
		{"explicit", "FixturePending", 404, true, time.Second, true},
		{"no-rule", "FixturePending", 404, false, 0, false},
		{"other-code", "NotFound", 404, true, time.Second, false},
		{"forbidden", "FixturePending", 403, true, time.Second, false},
		{"expired", "FixturePending", 404, true, time.Nanosecond, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			op, calls := operationFixture(t, []operationReply{
				{tc.status, `{"error":{"code":"` + tc.code + `"}}`, ""},
				{200, `{"id":"fixture-id","status":"succeeded"}`, ""},
			}, graphapi.DecodeTeamOperation)
			options := fastWait()
			if tc.rules {
				options.PropagationWindow = tc.window
				options.PropagationRules = []client.PropagationRule{{StatusCode: 404, Code: "FixturePending"}}
			}
			_, err := op.Wait(t.Context(), options)
			if (err == nil) != tc.wantSuccess {
				t.Fatalf("wrong propagation result: %v", err)
			}
			if !tc.wantSuccess && *calls != 1 {
				t.Fatal("unqualified propagation error retried")
			}
		})
	}
}

func TestUnitGraphOperationLocationValidationAndMinimumDelay(t *testing.T) {
	t.Parallel()
	scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Error("no immediate poll allowed")
		return nil, errors.New("unexpected request")
	})}, "commercial")
	for _, location := range []string{"", "//foreign.invalid/operations", "https://foreign.invalid/v1.0/operation", "/teams/../users", "/teams/%2e%2e/users"} {
		response := &resty.Response{RawResponse: &http.Response{StatusCode: 202, Header: http.Header{"Location": {location}}}}
		if _, err := graphapi.NewTeamOperation(scope, response); err == nil {
			t.Fatalf("invalid operation location accepted: %q", location)
		}
	}
	response := &resty.Response{RawResponse: &http.Response{StatusCode: 202, Header: http.Header{"Location": {`/teams('fixture')/operations('job')`}}}}
	op, err := graphapi.NewTeamOperation(scope, response)
	if err != nil {
		t.Fatal(err)
	}
	options := fastWait()
	options.Timeout = 10 * time.Millisecond
	result, err := op.Wait(t.Context(), options)
	if !errors.Is(err, context.DeadlineExceeded) || result.Handle.Location != `https://graph.microsoft.com/v1.0/teams('fixture')/operations('job')` || len(provider.requests) != 0 {
		t.Fatal("minimum poll delay or handle preservation failed")
	}
}

func TestUnitPropagationWindowCannotBeExtendedByServerDelay(t *testing.T) {
	t.Parallel()
	op, calls := operationFixture(t, []operationReply{
		{404, `{"error":{"code":"FixturePending"}}`, "1"},
		{200, `{"id":"fixture-id","status":"succeeded"}`, ""},
	}, graphapi.DecodeTeamOperation)
	options := fastWait()
	options.PropagationWindow = 50 * time.Millisecond
	options.PropagationRules = []client.PropagationRule{{StatusCode: 404, Code: "FixturePending"}}
	_, err := op.Wait(t.Context(), options)
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 || *calls != 1 {
		t.Fatal("server delay extended propagation window")
	}
}

func TestUnitBatchCompleteDoesNotHidePartialFailure(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body  string
		state client.OperationState
		fail  bool
	}{
		{`{"operationId":"job","overallStatus":"NotStarted"}`, client.OperationAccepted, false},
		{`{"operationId":"job","overallStatus":"InProgress"}`, client.OperationRunning, false},
		{`{"operationId":"job","overallStatus":"Complete","errorCount":0}`, client.OperationSucceeded, false},
		{`{"operationId":"job","overallStatus":"Complete","errorCount":2}`, client.OperationFailed, false},
		{`{"operationId":"job","overallStatus":"Complete","errorCount":0,"userState":[{"state":"Completed","result":"User not found"}]}`, client.OperationFailed, false},
		{`{"operationId":"job","overallStatus":"Complete"}`, "", true},
		{`{"operationId":"job","overallStatus":"Complete","errorCount":0,"inProgressCount":1}`, "", true},
		{`{"operationId":"job","overallStatus":"FutureStatus"}`, "", true},
	} {
		status, err := teamsadminapi.DecodeBatchPolicyOperation([]byte(tc.body))
		if (err != nil) != tc.fail || status.State != tc.state {
			t.Fatalf("wrong batch completion result: %v", err)
		}
	}
}

func TestUnitOperationCloseCancelsSleepWithoutLosingHandle(t *testing.T) {
	t.Parallel()
	scope, session, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Error("closed sleeping operation should not acquire a response")
		return nil, errors.New("unexpected request")
	})}, "commercial")
	accepted := &resty.Response{RawResponse: &http.Response{StatusCode: 202, Header: http.Header{"Retry-After": {"3600"}}}}
	op, err := client.NewOperation(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: client.ReplaySafe}, "recoverable", graphapi.DecodeTeamOperation, 0, accepted)
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { _, err := op.Wait(t.Context(), client.WaitOptions{}); result <- err }()
	session.Close()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) && !errors.Is(err, client.ErrClosed) {
			t.Fatalf("wrong close result: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not interrupt poll delay")
	}
	if op.Snapshot().Handle.ID != "recoverable" {
		t.Fatal("close discarded remote identity")
	}
}
