// Package policytransport sends direct compiled policy/GPA requests. Generated
// ConfigAPI forest discovery remains a separate adapter.
package policytransport

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type Client struct{ scope *client.Scope }

func New(scope *client.Scope) *Client { return &Client{scope: scope} }
func (s *Client) Ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.TeamsAdmin {
		return client.ErrDestination
	}
	return nil
}
func (s *Client) Do(ctx context.Context, command, method string, path []string, query url.Values, body any, replay client.ReplaySafety) (*resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, err
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, client.ErrRequest
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	correlation := fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
	h := http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {command}, "X-MS-Correlation-Id": {correlation}}
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	segments := make([]string, len(path))
	for i, segment := range path {
		// Uri.EscapeDataString uses the unreserved set. QueryEscape plus this
		// space conversion preserves that contract for colon/plus/at as well
		// as escaped slashes, unlike net/url's more permissive PathEscape.
		segments[i] = strings.ReplaceAll(url.QueryEscape(segment), "+", "%20")
	}
	destination := environment.Endpoints.TeamsAdmin + "/" + strings.Join(segments, "/")
	if encoded := query.Encode(); encoded != "" {
		destination += "?" + encoded
	}
	response, err := s.scope.Do(ctx, client.Request{Method: method, URL: destination, Header: h, Body: body, Replay: replay}, nil)
	if response != nil {
		err = classifyError(err, response.StatusCode(), response.Body())
	}
	return response, err
}
