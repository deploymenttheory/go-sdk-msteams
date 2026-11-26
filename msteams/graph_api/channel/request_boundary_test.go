package channel_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

// requestBoundary retains a controlled denial for unreviewed success replies.
// Individually sourced authored replies execute their intended protocol path;
// neither path establishes live Graph qualification.
func requestBoundary(t *testing.T, c contracttest.Case) contracttest.Case {
	t.Helper()
	if c.Scope == "local" || len(c.Steps) == 0 {
		return c
	}
	// Deliberately malformed and adversarial single-response cases remain
	// negative protocol tests, with no successful response qualification claim.
	if len(c.Steps) == 1 && c.Want.Error != nil {
		return c
	}
	for index, step := range c.Steps {
		// Reviewed authored replies execute the actual success/result path.
		// Sources are validated by contracttest and are never live evidence.
		if len(step.Reply.ProtocolSources) != 0 {
			continue
		}
		if step.Reply.Status < 200 || step.Reply.Status >= 300 {
			continue
		}
		c.Steps = append([]contracttest.Exchange(nil), c.Steps[:index+1]...)
		c.Steps[index].Reply = contracttest.Reply{Status: http.StatusForbidden, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"error":{"code":"UnitRequestDenied"}}`)}
		invoke := c.Invoke
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			observed := invoke(context.WithValue(ctx, requestBoundaryContextKey{}, true), sdk)
			if !errors.Is(observed.Err, auth.ErrForbidden) {
				t.Error("request boundary did not retain controlled authorization denial")
			}
			if observed.Response != nil && observed.Response.StatusCode() != http.StatusForbidden {
				t.Error("request boundary returned an unrelated HTTP response")
			}
			return contracttest.Observation{}
		}
		c.Want = contracttest.Expected{}
		return c
	}
	return c
}

type requestBoundaryContextKey struct{}

func isRequestBoundary(ctx context.Context) bool {
	return ctx != nil && ctx.Value(requestBoundaryContextKey{}) == true
}
