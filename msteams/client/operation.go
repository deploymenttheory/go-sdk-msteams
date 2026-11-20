package client

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	ErrOperation          = errors.New("invalid operation handle or polling contract")
	ErrOperationState     = errors.New("unknown or inconsistent operation state")
	ErrOperationFailed    = errors.New("remote operation failed")
	ErrOperationCancelled = errors.New("remote operation was cancelled")
	ErrOperationPollLimit = errors.New("operation polling limit reached")
)

// OperationState is normalized by a backend-specific decoder. HTTP acceptance
// alone is never a successful terminal state.
type OperationState string

const (
	OperationAccepted  OperationState = "accepted"
	OperationRunning   OperationState = "running"
	OperationSucceeded OperationState = "succeeded"
	OperationFailed    OperationState = "failed"
	OperationCancelled OperationState = "cancelled"
)

// OperationHandle preserves a resumable remote identity independently of a local
// wait's context. Location is validated before credentials are acquired. It may
// contain tenant data and must only be inspected explicitly.
type OperationHandle struct {
	Backend  Backend
	ID       string
	Location string
}

func (OperationHandle) String() string     { return "msteams operation handle (redacted)" }
func (h OperationHandle) GoString() string { return h.String() }

// OperationStatus is the decoder's exact interpretation of an evidenced envelope.
// ID, when present, must remain stable. TargetLocation is opaque; never followed
// automatically. A completed batch with per-item failures must return Failed.
type OperationStatus struct {
	ID             string
	State          OperationState
	TargetLocation string
}

func (OperationStatus) String() string     { return "msteams operation status (redacted)" }
func (s OperationStatus) GoString() string { return s.String() }

// OperationDecoder must not retain or mutate body. Decoder errors are redacted.
type OperationDecoder func(body []byte) (OperationStatus, error)

// OperationSnapshot retains the last observation and response, including on local
// timeout/cancellation. Diagnostics expose server codes without logging messages.
type OperationSnapshot struct {
	Handle      OperationHandle
	Status      OperationStatus
	Response    *resty.Response
	Diagnostics *APIError
}

func (OperationSnapshot) String() string     { return "msteams operation snapshot (redacted)" }
func (s OperationSnapshot) GoString() string { return s.String() }

// PropagationRule allows only an exact backend code/status pair during a bounded
// propagation window. Resource services must supply evidence for each rule; none
// are enabled by default. Authentication/authorization failures are never eligible.
type PropagationRule struct {
	StatusCode int
	Code       string
}

// WaitOptions bounds one local wait, without cancelling the remote job. Zero values
// select a five-minute timeout, five-second interval and 1,000 polls. The backend's
// minimum interval and Retry-After may increase Interval. PropagationWindow and
// PropagationRules must be specified together; the window begins with this wait.
type WaitOptions struct {
	Timeout           time.Duration
	Interval          time.Duration
	MaxPolls          int
	PropagationWindow time.Duration
	PropagationRules  []PropagationRule
}

// Operation is a serial polling handle. Calls to Poll and Wait must not overlap;
// separate operations are independent. No context or worker goroutine is retained.
type Operation struct {
	scope           *Scope
	request         Request
	decode          OperationDecoder
	minimumInterval time.Duration
	nextPoll        time.Time
	snapshot        OperationSnapshot
	terminal        error
	done            bool
}

// NewOperation constructs a handle for an explicitly replay-safe GET status route.
// accepted may be nil when resuming a saved handle, otherwise it must be a 2xx
// response. It is recorded as Accepted even for HTTP 200: the status envelope must
// establish completion. Resource code owns extraction of the ID and status route.
func NewOperation(scope *Scope, request Request, id string, decode OperationDecoder, minimumInterval time.Duration, accepted *resty.Response) (*Operation, error) {
	if scope == nil || scope.session == nil || decode == nil || request.Method != http.MethodGet || request.Replay != ReplaySafe || request.Body != nil || minimumInterval < 0 || request.MinimumRetryDelay < 0 {
		return nil, ErrOperation
	}
	if accepted != nil && (accepted.StatusCode() < 200 || accepted.StatusCode() >= 300) {
		return nil, ErrOperation
	}
	profile, err := scope.profile()
	if err != nil {
		return nil, err
	}
	destination, err := profile.destination(request)
	if err != nil {
		return nil, err
	}
	headers, err := requestHeaders(request.Header)
	if err != nil {
		return nil, err
	}
	request.URL, request.Path, request.Query, request.Header = destination.String(), nil, nil, headers
	request.MinimumRetryDelay = max(request.MinimumRetryDelay, minimumInterval)
	op := &Operation{scope: scope, request: request, decode: decode, minimumInterval: minimumInterval,
		snapshot: OperationSnapshot{Handle: OperationHandle{Backend: scope.Backend(), ID: id, Location: destination.String()},
			Status: OperationStatus{ID: id, State: OperationAccepted}, Response: accepted}}
	if accepted != nil {
		op.schedule(accepted, time.Now())
	}
	return op, nil
}

func (*Operation) String() string     { return "msteams operation (redacted)" }
func (o *Operation) GoString() string { return o.String() }

// Snapshot returns the last observation. Treat response/diagnostic pointers as
// read-only. Identifiers survive failed or cancelled waits and can be persisted
// privately with the resource's polling contract to resume under a new context.
func (o *Operation) Snapshot() OperationSnapshot {
	if o == nil {
		return OperationSnapshot{}
	}
	return o.snapshot
}

// Poll waits for any backend minimum/Retry-After delay, then reads one status.
// It never submits/replays the initiating mutation. A local failure is resumable;
// a decoded remote terminal state is sticky and causes no additional requests.
func (o *Operation) Poll(ctx context.Context) (OperationSnapshot, error) {
	if ctx == nil {
		return o.Snapshot(), ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return o.Snapshot(), err
	}
	if o == nil || o.scope == nil {
		return OperationSnapshot{}, ErrOperation
	}
	if o.done {
		return o.snapshot, o.terminal
	}
	ctx, stop, err := o.scope.session.Begin(ctx)
	if err != nil {
		return o.snapshot, err
	}
	defer stop()
	if err := waitForRetry(ctx, time.Until(o.nextPoll)); err != nil {
		return o.snapshot, err
	}
	response, err := o.scope.Do(ctx, o.request, nil)
	// A failed transport attempt also consumes the backend's polling cadence.
	o.nextPoll = time.Now().Add(o.minimumInterval)
	if response != nil {
		o.snapshot.Response = response
		o.schedule(response, time.Now())
	}
	if err != nil {
		return o.snapshot, err
	}
	status, err := o.decode(response.Body())
	if err != nil {
		if errors.Is(err, ErrOperationState) {
			return o.snapshot, ErrOperationState
		}
		return o.snapshot, ErrDecode
	}
	if status.ID != "" && o.snapshot.Handle.ID != "" && status.ID != o.snapshot.Handle.ID {
		return o.snapshot, ErrOperationState
	}
	if response.StatusCode() == http.StatusAccepted && status.State == OperationSucceeded {
		return o.snapshot, ErrOperationState
	}
	switch status.State {
	case OperationAccepted, OperationRunning:
	case OperationSucceeded:
		o.done = true
	case OperationFailed:
		o.done, o.terminal = true, ErrOperationFailed
		o.snapshot.Diagnostics = normalizeError(o.scope.Backend(), response)
	case OperationCancelled:
		o.done, o.terminal = true, ErrOperationCancelled
		o.snapshot.Diagnostics = normalizeError(o.scope.Backend(), response)
	default:
		return o.snapshot, ErrOperationState
	}
	if o.snapshot.Handle.ID == "" {
		o.snapshot.Handle.ID = status.ID
	}
	if status.ID == "" {
		status.ID = o.snapshot.Handle.ID
	}
	o.snapshot.Status = status
	return o.snapshot, o.terminal
}

// Wait polls until a decoded terminal state or a local bound. Its returned snapshot
// is recoverable even when err is context.DeadlineExceeded/context.Canceled. Unknown
// statuses and unlisted errors fail immediately, including ordinary 404 responses.
func (o *Operation) Wait(ctx context.Context, options WaitOptions) (OperationSnapshot, error) {
	if ctx == nil {
		return o.Snapshot(), ErrNilContext
	}
	if o == nil || o.scope == nil {
		return OperationSnapshot{}, ErrOperation
	}
	options, err := normalizeWaitOptions(options)
	if err != nil {
		return o.snapshot, err
	}
	ctx, stop, err := o.scope.session.Begin(ctx)
	if err != nil {
		return o.snapshot, err
	}
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	propagationEnd := time.Now().Add(options.PropagationWindow)
	var pendingPropagation error
	for poll := 0; poll < options.MaxPolls; poll++ {
		if pendingPropagation != nil && (!time.Now().Before(propagationEnd) || !o.nextPoll.Before(propagationEnd)) {
			return o.snapshot, pendingPropagation
		}
		if poll > 0 {
			delay := options.Interval
			if pendingPropagation != nil {
				delay = min(delay, time.Until(propagationEnd))
			}
			if err := waitForRetry(ctx, delay); err != nil {
				return o.snapshot, err
			}
		}
		if pendingPropagation != nil && !time.Now().Before(propagationEnd) {
			return o.snapshot, pendingPropagation
		}
		snapshot, err := o.Poll(ctx)
		if err != nil {
			if time.Now().Before(propagationEnd) && matchesPropagation(err, options.PropagationRules) {
				pendingPropagation = err
				continue
			}
			return snapshot, err
		}
		pendingPropagation = nil
		if snapshot.Status.State == OperationSucceeded {
			return snapshot, nil
		}
	}
	return o.snapshot, ErrOperationPollLimit
}

func (o *Operation) schedule(response *resty.Response, now time.Time) {
	delay := o.minimumInterval
	if server, valid := retryAfter(response.Header().Get("Retry-After"), now); valid {
		delay = max(delay, server)
	}
	o.nextPoll = now.Add(delay)
}

func normalizeWaitOptions(options WaitOptions) (WaitOptions, error) {
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Minute
	}
	if options.Interval == 0 {
		options.Interval = 5 * time.Second
	}
	if options.MaxPolls == 0 {
		options.MaxPolls = 1000
	}
	if options.Timeout < 0 || options.Interval < 0 || options.MaxPolls < 1 || options.PropagationWindow < 0 ||
		(options.PropagationWindow == 0) != (len(options.PropagationRules) == 0) {
		return WaitOptions{}, ErrOperation
	}
	options.PropagationRules = slices.Clone(options.PropagationRules)
	for _, rule := range options.PropagationRules {
		if rule.Code == "" || (rule.StatusCode != 400 && rule.StatusCode != 404 && rule.StatusCode != 409) {
			return WaitOptions{}, ErrOperation
		}
	}
	return options, nil
}

func matchesPropagation(err error, rules []PropagationRule) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, rule := range rules {
		if apiErr.StatusCode == rule.StatusCode && slices.Contains(apiErr.Codes, rule.Code) {
			return true
		}
	}
	return false
}
