package teammanagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const operationID = "fixture-operation"
const operationPath = "/v1.0/teams/" + groupID + "/operations/" + operationID

func pairedRead(t *testing.T, archived string) []contracttest.Exchange {
	t.Helper()
	steps := fixture(t, "team-get-enriched")
	steps[1].Reply.Body = []byte(`{"id":"` + groupID + `","isArchived":` + archived + `}`)
	return steps
}

func submission(t *testing.T, archived bool) contracttest.Exchange {
	t.Helper()
	if archived {
		steps := fixture(t, "team-archive-native-async")
		return steps[len(steps)-1]
	}
	return fixture(t, "team-unarchive-native-async")[0]
}

func poll(state string) contracttest.Exchange {
	return wire("GET", operationPath, "", 200, `{"id":"`+operationID+`","status":"`+state+`","targetResourceLocation":"https://foreign.invalid/opaque-never-follow"}`)
}

func archiveObserved(result management.ArchiveStateResult, err error) contracttest.Observation {
	stages := []string{}
	for _, stage := range result.Stages {
		stages = append(stages, string(stage.Action)+":"+string(stage.State))
	}
	var before, after *bool
	if result.Before != nil {
		before = result.Before.IsArchived
	}
	if result.After != nil {
		after = result.After.IsArchived
	}
	accepted := result.Accepted.Response != nil && result.Accepted.Response.StatusCode() == 202
	return contracttest.Observation{Value: map[string]any{"requested": result.RequestedArchived, "before": before, "after": after, "accepted": accepted, "operation_id": result.Snapshot.Handle.ID, "operation_state": result.Snapshot.Status.State, "confirmed": result.Confirmed, "stages": stages}, Response: result.Response, Err: err}
}

func archiveExpected(requested bool, before, after string, accepted, confirmed bool, operationID, state, stages string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"requested":%t,"before":%s,"after":%s,"accepted":%t,"operation_id":%q,"operation_state":%q,"confirmed":%t,"stages":%s}`, requested, before, after, accepted, operationID, state, confirmed, stages))
}

const beforeStages = `"read-group:succeeded","read-team:succeeded","check-archived-state:succeeded"`

func TestArchivedStateWaitsAndConfirmsReadback(t *testing.T) {
	var results []contracttest.Result
	for _, name := range []string{"archive-explicit-false", "archive-omitted", "unarchive"} {
		synctest.Test(t, func(t *testing.T) {
			archived := name != "unarchive"
			before, after, action := "false", "true", "archive-team"
			if !archived {
				before, after, action = "true", "false", "unarchive-team"
			}
			steps := pairedRead(t, before)
			submit := submission(t, archived)
			options := management.ArchiveStateOptions{}
			if name == "archive-explicit-false" {
				disabled := false
				options.Archive.ShouldSetSpoSiteReadOnlyForMembers = &disabled
			}
			if name == "archive-omitted" {
				submit.Request.JSON = nil
				submit.Request.Headers.Del("Content-Type")
			}
			steps = append(steps, submit, poll("succeeded"))
			steps = append(steps, pairedRead(t, after)...)
			start := time.Now()
			results = append(results, runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return archiveObserved(sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, archived, options))
			}, Want: contracttest.Expected{Status: 200, JSON: archiveExpected(archived, before, after, true, true, operationID, "succeeded", `[`+beforeStages+`,"`+action+`:accepted","wait-archived-state:succeeded","read-group:succeeded","read-team:succeeded","confirm-archived-state:succeeded"]`)}}))
			if time.Since(start) != 31*time.Second {
				t.Fatal("operation was checked before the Graph minimum cadence")
			}
		})
	}
	evidence(t, "team-archive-service.json", results)
}

func TestArchivePrechecksRejectUnchangedOrUnknownState(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, before string
		archived     bool
		want         error
	}{
		{"already-archived", "true", true, management.ErrArchivedStateUnchanged},
		{"already-unarchived", "false", false, management.ErrArchivedStateUnchanged},
		{"unknown-is-not-false", "null", true, management.ErrArchivedStateUnknown},
	} {
		runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: tc.name, Steps: pairedRead(t, tc.before), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return archiveObserved(sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, tc.archived, management.ArchiveStateOptions{}))
		}, Want: contracttest.Expected{Status: 200, Error: tc.want, JSON: archiveExpected(tc.archived, tc.before, "null", false, false, "", "", `["read-group:succeeded","read-team:succeeded","check-archived-state:blocked"]`)}})
	}
}

func TestArchiveTimeoutAndFailureRetainRemoteOperation(t *testing.T) {
	for _, tc := range []struct {
		name, state string
		timeout     time.Duration
		want        error
		status      int
	}{
		{"wait-timeout", "accepted", time.Second, context.DeadlineExceeded, 202},
		{"remote-failure", "failed", 0, client.ErrOperationFailed, 200},
		{"unknown-future-state", "futureState", 0, client.ErrOperationState, 200},
		{"accepted-response-is-not-success", "succeeded", 0, client.ErrOperationState, 202},
		{"poll-limit", "inProgress", 0, client.ErrOperationPollLimit, 200},
	} {
		synctest.Test(t, func(t *testing.T) {
			submit := submission(t, true)
			submit.Request.JSON = nil
			submit.Request.Headers.Del("Content-Type")
			steps := append(pairedRead(t, "false"), submit)
			state := "accepted"
			if tc.timeout == 0 {
				statusRead := poll(tc.state)
				statusRead.Reply.Status = tc.status
				steps = append(steps, statusRead)
				switch tc.state {
				case "failed":
					state = "failed"
				case "inProgress":
					state = "running"
				}
			}
			runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{Wait: client.WaitOptions{Timeout: tc.timeout, MaxPolls: 1}})
				if result.Accepted.Operation == nil || result.Snapshot.Handle.ID != operationID {
					t.Fatal("recoverable operation was lost")
				}
				for _, value := range []any{result, result.Accepted, result.Snapshot} {
					if rendered := fmt.Sprintf("%+v %#v", value, value); strings.Contains(rendered, groupID) || strings.Contains(rendered, operationID) {
						t.Fatal("formatting exposed private recovery identity")
					}
				}
				return archiveObserved(result, err)
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: archiveExpected(true, "false", "null", true, false, operationID, state, `[`+beforeStages+`,"archive-team:accepted","wait-archived-state:unconfirmed"]`)}})
		})
	}
}

func TestSuccessfulOperationDoesNotFabricateArchiveReadback(t *testing.T) {
	for _, state := range []string{"false", "null"} {
		synctest.Test(t, func(t *testing.T) {
			submit := submission(t, true)
			submit.Request.JSON = nil
			submit.Request.Headers.Del("Content-Type")
			steps := append(pairedRead(t, "false"), submit, poll("succeeded"))
			steps = append(steps, pairedRead(t, state)...)
			runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: "unconfirmed-readback-" + state, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return archiveObserved(sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{}))
			}, Want: contracttest.Expected{Status: 200, Error: management.ErrArchivedStateUnconfirmed, JSON: archiveExpected(true, "false", state, true, false, operationID, "succeeded", `[`+beforeStages+`,"archive-team:accepted","wait-archived-state:succeeded","read-group:succeeded","read-team:succeeded","confirm-archived-state:unconfirmed"]`)}})
		})
	}
}

func TestArchiveWaitCanResumeUnderNewContextWithoutRepeatingAction(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		submit := submission(t, true)
		submit.Request.JSON = nil
		submit.Request.Headers.Del("Content-Type")
		steps := append(pairedRead(t, "false"), submit, poll("succeeded"))
		steps = append(steps, pairedRead(t, "true")...)
		start := time.Now()
		runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: "resume-after-timeout", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{Wait: client.WaitOptions{Timeout: time.Second}})
			if !errors.Is(err, context.DeadlineExceeded) || result.Accepted.Operation == nil {
				t.Fatal("accepted action was not retained on timeout")
			}
			snapshot, err := result.Accepted.Operation.Wait(ctx, client.WaitOptions{})
			if err != nil || snapshot.Status.State != client.OperationSucceeded {
				t.Fatal("same operation could not resume with a new wait context")
			}
			team, err := sdk.GraphAPI.TeamManagement.GetByID(ctx, groupID)
			if err != nil || team.Value.IsArchived == nil || !*team.Value.IsArchived {
				t.Fatal("recovery readback did not confirm state")
			}
			return contracttest.Observation{Response: team.Response}
		}, Want: contracttest.Expected{Status: 200}})
		if time.Since(start) != 31*time.Second {
			t.Fatal("resumed handle lost original polling cadence")
		}
	})
}

func TestArchiveCancellationAndClosePreserveAcceptance(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			submit := submission(t, true)
			submit.Request.JSON = nil
			submit.Request.Headers.Del("Content-Type")
			steps := append(pairedRead(t, "false"), submit)
			name := "cancel-during-wait"
			if shutdown {
				name = "close-during-wait"
			}
			runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				done := make(chan error, 1)
				go func() {
					time.Sleep(time.Second)
					if shutdown {
						done <- sdk.Close()
					} else {
						cancel()
						done <- nil
					}
				}()
				result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{})
				if closeErr := <-done; closeErr != nil {
					t.Fatal("client close failed")
				}
				if result.Accepted.Operation == nil {
					t.Fatal("cancellation lost accepted operation")
				}
				return archiveObserved(result, err)
			}, Want: contracttest.Expected{Status: 202, Error: context.Canceled, JSON: archiveExpected(true, "false", "null", true, false, operationID, "accepted", `[`+beforeStages+`,"archive-team:accepted","wait-archived-state:unconfirmed"]`)}})
		})
	}
}

func TestArchiveValidationStopsBeforeAnyReadOrWrite(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unarchive-spo-false", "application-spo-false", "negative-timeout", "negative-interval", "negative-polls", "propagation-not-evidenced"} {
		mode := auth.Delegated
		want := client.ErrOperation
		if name == "unarchive-spo-false" {
			want = management.ErrUnarchiveOption
		}
		if name == "application-spo-false" {
			mode = auth.Application
			want = team.ErrArchiveOption
		}
		runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			options := management.ArchiveStateOptions{}
			archived := true
			switch name {
			case "unarchive-spo-false":
				archived = false
				disabled := false
				options.Archive.ShouldSetSpoSiteReadOnlyForMembers = &disabled
			case "application-spo-false":
				disabled := false
				options.Archive.ShouldSetSpoSiteReadOnlyForMembers = &disabled
			case "negative-timeout":
				options.Wait.Timeout = -1
			case "negative-interval":
				options.Wait.Interval = -1
			case "negative-polls":
				options.Wait.MaxPolls = -1
			case "propagation-not-evidenced":
				options.Wait.PropagationWindow = time.Minute
				options.Wait.PropagationRules = []client.PropagationRule{{StatusCode: 404, Code: "notFound"}}
			}
			result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, archived, options)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestArchiveFailureStopsLaterStagesAndNeverResubmits(t *testing.T) {
	for _, phase := range []string{"group-preread", "team-preread", "submit-forbidden", "unsafe-handle", "poll-not-found", "group-readback", "team-readback"} {
		synctest.Test(t, func(t *testing.T) {
			submit := submission(t, true)
			submit.Request.JSON = nil
			submit.Request.Headers.Del("Content-Type")
			steps := append(pairedRead(t, "false"), submit, poll("succeeded"))
			steps = append(steps, pairedRead(t, "true")...)
			index := map[string]int{"group-preread": 0, "team-preread": 1, "submit-forbidden": 2, "unsafe-handle": 2, "poll-not-found": 3, "group-readback": 4, "team-readback": 5}[phase]
			status := 403
			want := auth.ErrForbidden
			var apiError *client.APIError
			switch phase {
			case "unsafe-handle":
				steps[index].Reply.Headers.Set("Location", "https://foreign.invalid/private")
				status = 202
				want = client.ErrDestination
			case "poll-not-found":
				status = 404
				want = nil
				apiError = &client.APIError{Backend: client.Graph, StatusCode: 404, Codes: []string{"FixtureDenied"}}
			}
			if phase != "unsafe-handle" {
				steps[index].Reply.Status = status
				steps[index].Reply.Body = []byte(`{"error":{"code":"FixtureDenied"}}`)
			}
			steps = steps[:index+1]
			runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: phase, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{})
				if result.Confirmed || result.After != nil {
					t.Fatal("failed workflow exposed a completed archive/readback")
				}
				if phase == "unsafe-handle" && (result.Accepted.Response == nil || result.Accepted.Operation != nil || result.Stages[len(result.Stages)-1].State != management.Accepted) {
					t.Fatal("unsafe accepted handle lost submission metadata")
				}
				return contracttest.Observation{Response: result.Response, Err: err}
			}, Want: contracttest.Expected{Status: status, Error: want, APIError: apiError}})
		})
	}
}
