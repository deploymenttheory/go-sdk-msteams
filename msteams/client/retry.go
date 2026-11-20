package client

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ReplaySafety is an operation contract, independent of its HTTP method. Services
// must opt in only when repeating the operation after an ambiguous failure is safe.
type ReplaySafety uint8

const (
	ReplayNever ReplaySafety = iota
	ReplaySafe
)

// RequestPolicy bounds each logical operation, including authentication, queueing,
// attempts and backoff. Zero fields select defaults. MaxAttempts includes the first
// attempt and any auth replay; set it to one to disable retries. MaxConcurrent is
// shared by both backend groups, and covers in-flight API attempts only.
type RequestPolicy struct {
	MaxAttempts   int
	Timeout       time.Duration
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	MaxConcurrent int
}

var (
	ErrRequestPolicy = errors.New("invalid request policy")
	ErrRetryBudget   = errors.New("request retry budget exhausted")
)

// SessionOption configures immutable shared request controls at construction.
type SessionOption func(*sessionOptions) error

// WithRequestPolicy configures request controls without performing any I/O.
func WithRequestPolicy(policy RequestPolicy) SessionOption {
	return func(o *sessionOptions) error { o.policy = policy; return nil }
}

func normalizePolicy(p RequestPolicy) (RequestPolicy, error) {
	if p.MaxAttempts == 0 {
		p.MaxAttempts = 3
	}
	if p.Timeout == 0 {
		p.Timeout = 2 * time.Minute
	}
	if p.BaseDelay == 0 {
		p.BaseDelay = 500 * time.Millisecond
	}
	if p.MaxDelay == 0 {
		p.MaxDelay = 30 * time.Second
	}
	if p.MaxConcurrent == 0 {
		p.MaxConcurrent = 16
	}
	if p.MaxAttempts < 1 || p.MaxAttempts > 100 || p.Timeout < 0 || p.BaseDelay < 0 || p.MaxDelay < p.BaseDelay || p.MaxConcurrent < 1 || p.MaxConcurrent > 1024 {
		return RequestPolicy{}, ErrRequestPolicy
	}
	return p, nil
}

type requestController struct {
	settings sessionOptions
	policy   RequestPolicy
	slots    chan struct{}
	now      func() time.Time
	sleep    func(context.Context, time.Duration) error
	jitter   func(time.Duration) time.Duration
}

func newRequestController(options []SessionOption) (*requestController, error) {
	o := defaultSessionOptions()
	for _, option := range options {
		if option == nil {
			return nil, ErrRequestPolicy
		}
		if err := option(&o); err != nil {
			return nil, err
		}
	}
	p, err := normalizePolicy(o.policy)
	if err != nil {
		return nil, err
	}
	return &requestController{settings: o, policy: p, slots: make(chan struct{}, p.MaxConcurrent), now: time.Now, sleep: waitForRetry,
		//nolint:gosec // G404: retry timing jitter carries no cryptographic or security-sensitive randomness.
		jitter: func(cap time.Duration) time.Duration { return time.Duration(rand.Int64N(int64(cap))) }}, nil
}

func (c *requestController) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case c.slots <- struct{}{}:
		if err := ctx.Err(); err != nil {
			c.release()
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (c *requestController) release() { <-c.slots }

func (c *requestController) delay(header string, retry int) time.Duration {
	if delay, valid := retryAfter(header, c.now()); valid {
		return delay
	}
	cap := c.policy.BaseDelay
	for i := 1; i < retry && cap < c.policy.MaxDelay; i++ {
		if cap > c.policy.MaxDelay/2 {
			cap = c.policy.MaxDelay
			break
		}
		cap *= 2
	}
	// Equal jitter retains a nonzero minimum and avoids synchronized retry waves.
	return cap/2 + c.jitter(cap-cap/2)
}

// A valid server delay is never shortened to MaxDelay: if it exceeds the remaining
// operation budget, the caller receives ErrRetryBudget instead of an early retry.
func retryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	digits := true
	for _, r := range value {
		if r < '0' || r > '9' {
			digits = false
			break
		}
	}
	if digits {
		seconds, err := strconv.ParseUint(value, 10, 64)
		const maxDuration = time.Duration(1<<63 - 1)
		if err != nil || seconds > uint64(maxDuration/time.Second) {
			return maxDuration, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	if date, err := http.ParseTime(value); err == nil {
		return max(time.Duration(0), date.Sub(now)), true
	}
	return 0, false
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
