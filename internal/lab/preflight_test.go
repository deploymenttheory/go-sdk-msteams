package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type fakeProbe func(context.Context, client.Request, any) (*resty.Response, error)

func (f fakeProbe) Do(ctx context.Context, r client.Request, v any) (*resty.Response, error) {
	return f(ctx, r, v)
}

func TestPreflightReportsReadAccessWithoutRetainingTenantData(t *testing.T) {
	t.Parallel()
	calls := 0
	probe := fakeProbe(func(ctx context.Context, r client.Request, v any) (*resty.Response, error) {
		calls++
		if r.Method != http.MethodGet || len(r.Path) != 2 || r.Path[0] != "v1.0" || r.Query.Get("$select") != "id" || v != nil {
			t.Fatal("preflight request escaped read-only contract")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("missing probe deadline")
		}
		status := 200
		var err error
		if r.Path[1] == "users" {
			status = 403
			err = auth.ErrForbidden
			if r.Query.Get("$top") != "1" {
				t.Fatal("unbounded user sample")
			}
		} else if r.Path[1] != "organization" {
			t.Fatal("unknown probe")
		}
		return &resty.Response{RawResponse: &http.Response{StatusCode: status, Header: http.Header{"Request-Id": {"private-tenant-id"}}}}, err
	})
	observations, err := lab.Preflight(t.Context(), lab.Config{ReadOnly: true, Timeout: time.Second}, probe)
	if err != nil || calls != 2 || observations[0].Outcome != "available" || observations[1].Outcome != "forbidden" {
		t.Fatal("capability inventory lost authorization outcome")
	}
	raw, err := json.Marshal(observations)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "private") {
		t.Fatal("preflight retained tenant identifiers")
	}
	for _, tc := range []struct {
		name    string
		status  int
		outcome string
	}{{"missing", 404, "unavailable"}, {"throttle", 429, "transient"}, {"service", 503, "transient"}, {"auth", 401, "unauthenticated"}, {"transport", 0, "error"}} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := lab.Preflight(t.Context(), lab.Config{ReadOnly: true, Timeout: time.Second}, fakeProbe(func(context.Context, client.Request, any) (*resty.Response, error) {
				return &resty.Response{RawResponse: &http.Response{StatusCode: tc.status}}, errors.New("private provider error")
			}))
			if err != nil || result[0].Outcome != tc.outcome {
				t.Fatal("unavailable and transient observations conflated")
			}
		})
	}
}

func TestPreflightDisabledCanceledAndTypedNilNeverSend(t *testing.T) {
	t.Parallel()
	calls := 0
	probe := fakeProbe(func(context.Context, client.Request, any) (*resty.Response, error) { calls++; return nil, nil })
	if _, err := lab.Preflight(t.Context(), lab.Config{}, probe); !errors.Is(err, lab.ErrDisabled) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := lab.Preflight(ctx, lab.Config{ReadOnly: true, Timeout: time.Second}, probe); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var nilProbe fakeProbe
	if _, err := lab.Preflight(t.Context(), lab.Config{ReadOnly: true, Timeout: time.Second}, nilProbe); !errors.Is(err, lab.ErrConfig) {
		t.Fatal("typed nil probe accepted")
	}
	if calls != 0 {
		t.Fatal("disabled/canceled preflight made requests")
	}
}

func TestPreflightDeadlineStopsInFlightProbe(t *testing.T) {
	t.Parallel()
	calls := 0
	probe := fakeProbe(func(ctx context.Context, _ client.Request, _ any) (*resty.Response, error) {
		calls++
		<-ctx.Done()
		return nil, ctx.Err()
	})
	result, err := lab.Preflight(t.Context(), lab.Config{ReadOnly: true, Timeout: time.Millisecond}, probe)
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 || len(result) != 0 {
		t.Fatal("probe deadline failed to stop traversal")
	}
}
