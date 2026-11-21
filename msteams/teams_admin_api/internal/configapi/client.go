// Package configapi adds the evidenced forest-routing header to generated ConfigAPI calls.
package configapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// ErrDiscovery means the service did not return a usable routing endpoint.
var ErrDiscovery = errors.New("configapi discovery response is invalid")

// Client caches discovery only for its injected, already identity-bound client.
// The discovered forest is a header value, never a credential-bearing destination.
type Client struct {
	base   client.Client
	gate   chan struct{}
	mu     sync.RWMutex
	target string
}

// New performs no I/O. The collection owns one instance per session.
func New(base client.Client) *Client {
	if existing, ok := base.(*Client); ok && existing != nil {
		return existing
	}
	return &Client{base: base, gate: make(chan struct{}, 1)}
}

// Do discovers on first use and delegates all requests, errors and replay policy.
func (c *Client) Do(ctx context.Context, request client.Request, result any) (*resty.Response, error) {
	if ctx == nil {
		return nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c == nil || c.base == nil {
		return nil, client.ErrClosed
	}
	target, metadata, err := c.discover(ctx)
	if err != nil {
		return metadata, err
	}
	request.Header = request.Header.Clone()
	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set("X-MS-Target-Uri", target)
	return c.base.Do(ctx, request, result)
}

func (c *Client) discover(ctx context.Context) (string, *resty.Response, error) {
	c.mu.RLock()
	target := c.target
	c.mu.RUnlock()
	if target != "" {
		return target, nil, nil
	}
	select {
	case c.gate <- struct{}{}:
	case <-ctx.Done():
		return "", nil, ctx.Err()
	}
	defer func() { <-c.gate }()
	if err := ctx.Err(); err != nil {
		return "", nil, err
	}
	c.mu.RLock()
	target = c.target
	c.mu.RUnlock()
	if target != "" {
		return target, nil, nil
	}
	var response struct {
		Endpoints struct{ AdminServiceEndpoint string }
	}
	metadata, err := c.base.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"Teams.Tenant", "serviceDiscovery"}, Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}, &response)
	if err != nil {
		return "", metadata, err
	}
	if metadata == nil || metadata.StatusCode() != http.StatusOK {
		return "", metadata, ErrDiscovery
	}
	endpoint := response.Endpoints.AdminServiceEndpoint
	if endpoint == "" || strings.Contains(endpoint, "://") {
		return "", metadata, ErrDiscovery
	}
	parsed, err := url.Parse("https://" + endpoint)
	if err != nil || parsed.User != nil || parsed.Hostname() == "" {
		return "", metadata, ErrDiscovery
	}
	// Match the pinned root normalization: drop port, path, query and fragment.
	host := parsed.Hostname()
	if strings.ContainsAny(host, "\r\n\t :/\\") {
		return "", metadata, ErrDiscovery
	}
	target = (&url.URL{Scheme: "https", Host: host, Path: "/"}).String()
	c.mu.Lock()
	c.target = target
	c.mu.Unlock()
	return target, metadata, nil
}
