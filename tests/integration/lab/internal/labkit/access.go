package labkit

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

func LogAccessResponse(t *testing.T, response *resty.Response, err error) {
	t.Helper()
	status := 0
	if response != nil {
		status = response.StatusCode()
	}
	var api *client.APIError
	var codes []int
	if errors.As(err, &api) {
		codes = api.ServiceErrorCodes
	}
	t.Logf("SDK access outcome: HTTP %d; PolicyRP codes %v; error %T", status, codes, err)
}

func IsHTTP(err error, status int) bool {
	var apiError *client.APIError
	return errors.As(err, &apiError) && apiError != nil && apiError.StatusCode == status
}

var DirectoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Observation polling is test-owned, read-only and explicitly bounded by ctx.
// Mutations are never retried; received read errors fail immediately.
func ObserveUntil(ctx context.Context, read func(context.Context) (bool, error)) error {
	for {
		ok, err := read(ctx)
		if err != nil || ok {
			return err
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func ReadOnlyEnabled(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("targeting reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid targeting lab enablement")
	}
	return cfg
}
