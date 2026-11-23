// Package cloudcalldata reads and creates the tenant's cloud call data
// connection, which lets on-premises Skype for Business Server deployments
// upload call data to the cloud. The connection is a tenant singleton and the
// module has no cmdlet to remove it.
package cloudcalldata

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func request(method, command string) client.Request {
	return client.Request{Method: method, Path: []string{"Teams.InternalSupport", "CloudCallDataConnection"},
		Header: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}}
}

func (s *Service) call(ctx context.Context, req client.Request, statuses ...int) (*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	accepted := false
	for _, status := range statuses {
		accepted = accepted || (response != nil && response.StatusCode() == status)
	}
	if !accepted {
		return nil, response, ErrResponse
	}
	var value Connection
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// Get reads the existing connection and its token
// (Get-CsCloudCallDataConnection).
func (s *Service) Get(ctx context.Context) (*Connection, *resty.Response, error) {
	req := request(http.MethodGet, "Get-CsCloudCallDataConnection")
	req.Replay = client.ReplaySafe
	return s.call(ctx, req, http.StatusOK)
}

// Create creates the tenant's connection and returns its token
// (New-CsCloudCallDataConnection). The request has no body. It is never
// replayed automatically; reconcile a lost response with Get.
func (s *Service) Create(ctx context.Context) (*Connection, *resty.Response, error) {
	return s.call(ctx, request(http.MethodPost, "New-CsCloudCallDataConnection"), http.StatusOK, http.StatusCreated)
}
