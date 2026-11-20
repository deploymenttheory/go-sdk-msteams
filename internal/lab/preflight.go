package lab

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type requestClient interface {
	Do(context.Context, client.Request, any) (*resty.Response, error)
}

// Observation records only a fixed probe name, status and access outcome. A failed
// sample never infers licensing or the availability of an unrelated API.
type Observation struct {
	Probe   string `json:"probe"`
	Status  int    `json:"status"`
	Outcome string `json:"outcome"`
}

// Preflight makes only the two documented read-only Graph requests. The caller
// supplies the Graph scope; no raw response, token, identifier or error is retained.
// A denied read is an inventory result; cancellation aborts the remaining probes.
func Preflight(ctx context.Context, cfg Config, graph requestClient) ([]Observation, error) {
	if err := cfg.Require(ReadOnly); err != nil {
		return nil, err
	}
	if ctx == nil || nilDependency(graph) || cfg.Timeout <= 0 || cfg.Timeout > 5*time.Minute {
		return nil, ErrConfig
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	results := make([]Observation, 0, 2)
	for _, probe := range []struct {
		name string
		path string
		top  bool
	}{{"organization-read", "organization", false}, {"users-read", "users", true}} {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		query := map[string][]string{"$select": {"id"}}
		if probe.top {
			query["$top"] = []string{"1"}
		}
		response, err := graph.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"v1.0", probe.path}, Query: query}, nil)
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		result := Observation{Probe: probe.name, Outcome: "error"}
		if response != nil {
			result.Status = response.StatusCode()
		}
		switch {
		case err == nil && result.Status == http.StatusOK:
			result.Outcome = "available"
		case errors.Is(err, auth.ErrUnauthenticated) || result.Status == http.StatusUnauthorized:
			result.Outcome = "unauthenticated"
		case errors.Is(err, auth.ErrForbidden) || result.Status == http.StatusForbidden:
			result.Outcome = "forbidden"
		case result.Status == http.StatusNotFound:
			result.Outcome = "unavailable"
		case result.Status == http.StatusTooManyRequests || result.Status >= 500:
			result.Outcome = "transient"
		}
		results = append(results, result)
	}
	return results, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
