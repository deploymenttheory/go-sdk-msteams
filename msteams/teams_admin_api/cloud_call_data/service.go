// Package cloudcalldata reads and creates the tenant's cloud call data
// connection models, which let on-premises Skype for Business Server deployments
// upload call data to the cloud. Modern responses are arrays; returned records
// are not presumed newly created or owned. The module has no remove cmdlet.
package cloudcalldata

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsCloudCallDataConnectionModern_Get"
const newCommand = "New-CsCloudCallDataConnectionModern_New"

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

func (s *Service) call(ctx context.Context, req client.Request) ([]*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response == nil || response.StatusCode() != http.StatusOK {
		return nil, response, ErrResponse
	}
	values, err := decodeConnections(response.Body())
	return values, response, err
}

// decodeConnections follows the pinned Modern callbacks: empty text is a null
// array, an array can contain null/nonobject models, and object/null roots fail.
func decodeConnections(raw []byte) ([]*Connection, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[0] != '[' {
		return nil, client.ErrDecode
	}
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil || items == nil {
		return nil, client.ErrDecode
	}
	values := make([]*Connection, 0, len(items))
	for _, item := range items {
		item = bytes.TrimSpace(item)
		if len(item) == 0 || item[0] != '{' {
			values = append(values, nil)
			continue
		}
		var value Connection
		if json.Unmarshal(item, &value) != nil {
			return nil, client.ErrDecode
		}
		values = append(values, &value)
	}
	return values, nil
}

// List reads all returned connections using the Modern Get route. It preserves
// source cardinality and null entries without selecting an arbitrary item.
func (s *Service) List(ctx context.Context) ([]*Connection, *resty.Response, error) {
	req := request(http.MethodGet, getCommand)
	req.Replay = client.ReplaySafe
	return s.call(ctx, req)
}

// Get preserves the existing singleton convenience. Exactly one non-null model
// is required; use List or inspect CardinalityError for the complete read outcome.
func (s *Service) Get(ctx context.Context) (*Connection, *resty.Response, error) {
	values, response, err := s.List(ctx)
	if err != nil {
		return nil, response, err
	}
	if len(values) != 1 || values[0] == nil {
		return nil, response, &CardinalityError{Returned: values, Response: response}
	}
	return values[0], response, nil
}

// CreateWithResult sends the Modern POST once and retains all returned
// connections, which are not assumed newly created or owned. It never replays;
// HTTPAccepted means only a 2xx receipt, not a supported status, parsed result or
// persisted creation. Only source-declared 200 is decoded as a successful array.
func (s *Service) CreateWithResult(ctx context.Context) (*CreateResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	result := &CreateResult{CreateInvoked: true}
	var err error
	result.Returned, result.Response, err = s.call(ctx, request(http.MethodPost, newCommand))
	result.HTTPAccepted = result.Response != nil && result.Response.StatusCode() >= 200 && result.Response.StatusCode() < 300
	return result, result.Response, err
}

// Create preserves the existing unique-result convenience. On accepted empty,
// multiple or null-model outcomes, CardinalityError retains the full typed action
// outcome/receipt. Never replay blindly or compensate returned pre-existing data.
func (s *Service) Create(ctx context.Context) (*Connection, *resty.Response, error) {
	result, response, err := s.CreateWithResult(ctx)
	if err != nil {
		return nil, response, err
	}
	if len(result.Returned) != 1 || result.Returned[0] == nil {
		return nil, response, &CardinalityError{Returned: result.Returned, Response: response, CreateOutcome: result}
	}
	return result.Returned[0], response, nil
}
