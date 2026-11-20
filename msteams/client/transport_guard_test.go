package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

type guardBody struct {
	io.Reader
	closed bool
}

func (b *guardBody) Close() error { b.closed = true; return nil }

type forbiddenTransport struct{ t *testing.T }

func (f forbiddenTransport) RoundTrip(*http.Request) (*http.Response, error) {
	f.t.Error("guard delegated rejected request")
	return nil, errors.New("unexpected")
}

func TestTransportGuardClosesRejectedRequestBodies(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, url, host, token string
		cancel                 bool
		want                   error
	}{
		{"destination", "https://attacker.invalid/v1.0/teams", "", "fixture-token", false, ErrDestination},
		{"host-override", "https://graph.microsoft.com/v1.0/teams", "attacker.invalid", "fixture-token", false, ErrDestination},
		{"cancelled", "https://graph.microsoft.com/v1.0/teams", "", "fixture-token", true, context.Canceled},
		{"cleared-token", "https://graph.microsoft.com/v1.0/teams", "", "", false, ErrClosed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			body := &guardBody{Reader: strings.NewReader("fixture-payload")}
			request, err := http.NewRequestWithContext(ctx, http.MethodPost, tc.url, body)
			if err != nil {
				t.Fatal(err)
			}
			if tc.host != "" {
				request.Host = tc.host
			}
			if tc.cancel {
				cancel()
			}
			transport := &authenticatedTransport{base: forbiddenTransport{t}, profile: routeProfile{"graph.microsoft.com", auth.GraphAudience}, token: tc.token}
			response, err := transport.RoundTrip(request)
			if response != nil && response.Body != nil {
				_ = response.Body.Close()
			}
			if !errors.Is(err, tc.want) || !body.closed {
				t.Fatalf("rejection did not close body: %v", err)
			}
		})
	}
}
