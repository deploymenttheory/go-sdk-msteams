package client_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitConcurrentCloseCleansOwnedResourcesOnce(t *testing.T) {
	t.Parallel()
	var cleaned atomic.Int32
	session, err := client.NewSession(&auth.Config{Method: auth.ManagedIdentity, TenantID: "fixture.example"},
		&http.Client{}, func() { cleaned.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 20 {
		group.Go(session.Close)
	}
	group.Wait()
	if cleaned.Load() != 1 {
		t.Fatal("owned cleanup did not run exactly once")
	}
	if _, err := session.HTTPClient(); !errors.Is(err, client.ErrClosed) {
		t.Fatal("closed session still supplied HTTP execution configuration")
	}
}

func TestUnitOperationContextPreservesDeadlineAndRejectsInvalidState(t *testing.T) {
	t.Parallel()
	session, err := client.NewSession(&auth.Config{Method: auth.ManagedIdentity, TenantID: "fixture.example"}, &http.Client{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	//nolint:staticcheck // Deliberately exercise the documented nil-context rejection boundary.
	if _, _, err := session.Begin(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil operation context must fail")
	}
	deadline := time.Now().Add(time.Minute)
	parent, cancelParent := context.WithDeadline(context.Background(), deadline)
	defer cancelParent()
	operation, cancel, err := session.Begin(parent)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	if got, ok := operation.Deadline(); !ok || !got.Equal(deadline) {
		t.Fatal("caller deadline was lost")
	}
	cancelParent()
	if _, _, err := session.Begin(parent); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled parent accepted new work")
	}
	var zero client.Session
	zero.Close()
	if _, _, err := zero.Begin(context.Background()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("uninitialized session accepted work")
	}
}
