package teammanagement

import (
	"context"
	"errors"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

var (
	ErrArchivedStateUnknown     = errors.New("team archive state is not present in the read response")
	ErrArchivedStateUnchanged   = errors.New("team already has the requested archive state")
	ErrArchivedStateUnconfirmed = errors.New("team readback does not confirm the requested archive state")
	ErrUnarchiveOption          = errors.New("SharePoint archive option is not valid for unarchive")
)

// ArchiveStateOptions selects the native archive option and bounds the wait.
// A zero Wait uses the shared defaults. Propagation allowances are not accepted:
// this composition has no evidenced status/code exceptions for operation reads.
type ArchiveStateOptions struct {
	Archive team.ArchiveOptions
	Wait    client.WaitOptions
}

// ArchiveStateResult retains each distinct observation. Before is never reused
// as After. Accepted/Snapshot survive failed waits for private recovery; they do
// not imply completion. After can be non-nil with Confirmed false when a paired
// read succeeds but its observed state is missing or differs from the request.
type ArchiveStateResult struct {
	ID                string
	RequestedArchived bool
	Before            *ManagedTeam
	After             *ManagedTeam
	Accepted          team.AcceptedOperation
	Snapshot          client.OperationSnapshot
	Stages            []Stage
	Response          *resty.Response
	Confirmed         bool
}

func (ArchiveStateResult) String() string     { return "team archive-state result (redacted)" }
func (r ArchiveStateResult) GoString() string { return r.String() }

func (s *Service) archiveOptions(archived bool, options ArchiveStateOptions) error {
	if err := options.Wait.Validate(); err != nil {
		return err
	}
	if options.Wait.PropagationWindow != 0 || len(options.Wait.PropagationRules) != 0 {
		return client.ErrOperation
	}
	if options.Archive.ShouldSetSpoSiteReadOnlyForMembers == nil {
		return nil
	}
	if !archived {
		return ErrUnarchiveOption
	}
	info, err := s.scope.Session().Info()
	if err != nil {
		return err
	}
	if info.Mode == auth.Application {
		return team.ErrArchiveOption
	}
	return nil
}

// SetArchivedState implements the pinned precheck with native asynchronous
// completion: paired read, reject unknown/unchanged state, submit once, wait,
// paired readback, confirm the requested state. It never follows target URLs,
// replays an action, rolls back, or treats HTTP 202 as completed archiving.
func (s *Service) SetArchivedState(ctx context.Context, id string, archived bool, options ArchiveStateOptions) (ArchiveStateResult, error) {
	result := ArchiveStateResult{ID: id, RequestedArchived: archived}
	ctx, stop, err := s.begin(ctx, id)
	if err != nil {
		return result, err
	}
	defer stop()
	if err := s.archiveOptions(archived, options); err != nil {
		return result, err
	}
	trace := Result{ID: id}
	finish := func(err error) (ArchiveStateResult, error) {
		result.Stages, result.Response = trace.Stages, trace.Response
		return result, err
	}
	if err := s.read(ctx, &trace); err != nil {
		return finish(err)
	}
	result.Before = trace.Value
	if result.Before.IsArchived == nil {
		err = ErrArchivedStateUnknown
	} else if *result.Before.IsArchived == archived {
		err = ErrArchivedStateUnchanged
	}
	trace.record(CheckArchivedState, nil, err)
	if err != nil {
		trace.Stages[len(trace.Stages)-1].State = Blocked
		return finish(err)
	}
	action := ArchiveTeam
	if archived {
		result.Accepted, err = s.teams.Archive(ctx, id, options.Archive)
	} else {
		action = UnarchiveTeam
		result.Accepted, err = s.teams.Unarchive(ctx, id)
	}
	trace.record(action, result.Accepted.Response, err)
	if result.Accepted.Response != nil && result.Accepted.Response.StatusCode() == http.StatusAccepted {
		trace.Stages[len(trace.Stages)-1].State = Accepted
	}
	if result.Accepted.Operation != nil {
		result.Snapshot = result.Accepted.Operation.Snapshot()
	}
	if err != nil {
		return finish(err)
	}
	result.Snapshot, err = result.Accepted.Operation.Wait(ctx, options.Wait)
	trace.record(WaitArchivedState, result.Snapshot.Response, err)
	if err != nil {
		return finish(err)
	}
	trace.Value = nil
	if err := s.read(ctx, &trace); err != nil {
		return finish(err)
	}
	result.After = trace.Value
	if result.After.IsArchived == nil || *result.After.IsArchived != archived {
		err = ErrArchivedStateUnconfirmed
	}
	trace.record(ConfirmArchivedState, nil, err)
	result.Confirmed = err == nil
	return finish(err)
}
