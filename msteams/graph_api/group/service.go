package group

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("group operation returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	extensionName       = regexp.MustCompile(`^extension_[0-9a-fA-F]{32}_[A-Za-z0-9_]+$`)
)

// Service manages backing-group lifecycle independently of teamification.
// It owns no credentials or worker; all requests use the client's Graph scope.
type Service struct{ scope *client.Scope }

// NewService constructs a lazy resource without authentication or HTTP.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) local() error {
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.local()
}

func request(method string, path ...string) client.Request {
	req := client.Request{Method: method, Path: append([]string{"v1.0", "groups"}, path...), Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		req.Replay = client.ReplaySafe
	}
	return req
}

func (s *Service) resource(ctx context.Context, req client.Request, status int) (*Group, *resty.Response, error) {
	response, err := s.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != status {
		return nil, response, ErrUnexpectedStatus
	}
	var value Group
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

type createPayload struct {
	Properties
	GroupTypes              []string  `json:"groupTypes"`
	MailEnabled             bool      `json:"mailEnabled"`
	SecurityEnabled         bool      `json:"securityEnabled"`
	Owners                  *[]string `json:"owners@odata.bind,omitempty"`
	Members                 *[]string `json:"members@odata.bind,omitempty"`
	CreationOptions         *[]string `json:"creationOptions,omitempty"`
	ResourceBehaviorOptions *[]string `json:"resourceBehaviorOptions,omitempty"`
}

func (s *Service) createBody(value CreateRequest) (map[string]json.RawMessage, error) {
	if value.DisplayName == nil || strings.TrimSpace(*value.DisplayName) == "" || value.MailNickname == nil || strings.TrimSpace(*value.MailNickname) == "" {
		return nil, client.ErrRequest
	}
	if len(value.OwnerIDs)+len(value.MemberIDs) > 20 {
		return nil, client.ErrRequest
	}
	for _, ids := range [][]string{value.OwnerIDs, value.MemberIDs} {
		for _, id := range ids {
			if !directoryID.MatchString(id) {
				return nil, client.ErrRequest
			}
		}
	}
	for name, raw := range value.Extensions {
		if !extensionName.MatchString(name) || !json.Valid(raw) {
			return nil, client.ErrRequest
		}
	}
	env, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	bind := func(ids []string) *[]string {
		if ids == nil {
			return nil
		}
		result := make([]string, 0, len(ids))
		for _, id := range ids {
			result = append(result, env.Endpoints.Graph+"/v1.0/users/"+id)
		}
		return &result
	}
	body := createPayload{Properties: value.Properties, GroupTypes: value.GroupTypes, MailEnabled: value.MailEnabled, SecurityEnabled: value.SecurityEnabled,
		Owners: bind(value.OwnerIDs), Members: bind(value.MemberIDs), CreationOptions: value.CreationOptions, ResourceBehaviorOptions: value.ResourceBehaviorOptions}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, client.ErrRequest
	}
	var payload map[string]json.RawMessage
	if json.Unmarshal(encoded, &payload) != nil {
		return nil, client.ErrRequest
	}
	for name, raw := range value.Extensions {
		payload[name] = raw
	}
	return payload, nil
}

// Create submits one group create and returns the 201 snapshot. It does not
// resolve owners, choose visibility/provisioning flags, create a team, wait or
// compensate. Owners/members are user directory IDs; at most 20 relationships
// can be included. Graph enforces type/field/consent restrictions. App-only
// ownerless creation has special provisioning/manageability limitations.
func (s *Service) Create(ctx context.Context, value CreateRequest) (*Group, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := s.createBody(value)
	if err != nil {
		return nil, nil, err
	}
	req := request(http.MethodPost)
	req.Body = body
	return s.resource(ctx, req, http.StatusCreated)
}

// ReadOptions selects group fields, including extension fields not in defaults.
type ReadOptions struct{ Select []string }

// GetByID reads a native group, without hydrating its team or relationships.
func (s *Service) GetByID(ctx context.Context, id string, options ReadOptions) (*Group, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, nil, client.ErrRequest
	}
	req := request(http.MethodGet, id)
	query, err := selectQuery(options.Select)
	if err != nil {
		return nil, nil, err
	}
	req.Query = query
	return s.resource(ctx, req, http.StatusOK)
}

// UpdateByID patches the five writable backing-group properties used by the
// pinned team command. Nil fields stay omitted; explicit empty strings survive.
// Team settings are a separate request with an independent outcome.
func (s *Service) UpdateByID(ctx context.Context, id string, value UpdateRequest) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, client.ErrRequest
	}
	req := request(http.MethodPatch, id)
	req.Body = value
	return s.empty(ctx, req)
}

// DeleteByID deletes the group and associated resources, including a backing
// team when present. It never infers ownership or attempts automatic recovery.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, client.ErrRequest
	}
	return s.empty(ctx, request(http.MethodDelete, id))
}

func (s *Service) empty(ctx context.Context, req client.Request) (*resty.Response, error) {
	response, err := s.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != http.StatusNoContent {
		err = ErrUnexpectedStatus
	}
	return response, err
}
