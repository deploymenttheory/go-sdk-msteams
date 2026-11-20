package applicationaccesspolicy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

// ErrResponse identifies a success response outside the evidenced operation contract.
var ErrResponse = errors.New("unexpected application access policy response")

// Service uses an injected backend client and adds ConfigAPI discovery/routing.
type Service struct{ client client.Client }

// NewService performs no I/O. The backend must be bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func path(identity string) []string {
	base := []string{"Skype.Policy", "configurations", "ApplicationAccessPolicy"}
	if identity != "" {
		base = append(base, "configuration", identity)
	}
	return base
}

func validIdentity(identity string) bool {
	return strings.TrimSpace(identity) != "" && identity != "." && identity != ".." && !strings.ContainsAny(identity, "\x00\r\n")
}

func (s *Service) do(ctx context.Context, request client.Request, result any, statuses ...int) (*resty.Response, error) {
	if ctx == nil {
		return nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil || s.client == nil {
		return nil, client.ErrClosed
	}
	response, err := s.client.Do(ctx, request, result)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// GetByIdentity reads one policy; Identity is escaped as one complete path segment.
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Policy, *resty.Response, error) {
	if !validIdentity(identity) {
		return nil, nil, ErrInput
	}
	var policy Policy
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: path(identity), Replay: client.ReplaySafe}, &policy, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	if policy.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &policy, response, nil
}

// List reads the service's bare array. No unlimited paging guarantee is inferred.
func (s *Service) List(ctx context.Context) ([]Policy, *resty.Response, error) {
	var raw json.RawMessage
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: path(""), Replay: client.ReplaySafe}, &raw, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var policies []Policy
	if len(raw) == 0 || json.Unmarshal(raw, &policies) != nil || policies == nil {
		return nil, response, client.ErrDecode
	}
	return policies, response, nil
}

// ReadBackError means creation succeeded, but the separate policy read did not.
// CreationResponse retains the successful POST metadata for reconciliation/cleanup.
type ReadBackError struct {
	CreationResponse *resty.Response
	Err              error
}

func (e *ReadBackError) Error() string { return "policy created but read-back failed" }
func (e *ReadBackError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
func (e ReadBackError) GoString() string { return "policy created but read-back failed" }

// String also protects copied error values from exposing response data in logs.
func (e ReadBackError) String() string { return "policy created but read-back failed" }

// Create posts a policy then reads it, mirroring New-CsApplicationAccessPolicy.
// A failed read returns the requested identity and ReadBackError; it never deletes
// or replays the successful mutation automatically.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*Policy, *resty.Response, error) {
	if !validIdentity(input.Identity) {
		return nil, nil, ErrInput
	}
	body, err := writeBody(input.Identity, input.AppIDs, input.Description)
	if err != nil {
		return nil, nil, err
	}
	created, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: path(""), Body: body}, nil, http.StatusCreated)
	if err != nil {
		return nil, created, err
	}
	policy, response, err := s.GetByIdentity(ctx, input.Identity)
	if err != nil {
		return &Policy{Identity: input.Identity}, response, &ReadBackError{CreationResponse: created, Err: err}
	}
	return policy, response, nil
}

// UpdateByIdentity sends the evidenced sparse PUT, not PATCH. Nil fields remain
// absent; empty Description and an explicit empty AppIDs replacement survive.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*resty.Response, error) {
	if !validIdentity(identity) {
		return nil, ErrInput
	}
	body, err := writeBody(identity, input.AppIDs, input.Description)
	if err != nil {
		return nil, err
	}
	return s.do(ctx, client.Request{Method: http.MethodPut, Path: path(identity), Body: body}, nil, http.StatusNoContent)
}

// DeleteByIdentity removes one policy and returns its HTTP metadata.
func (s *Service) DeleteByIdentity(ctx context.Context, identity string) (*resty.Response, error) {
	if !validIdentity(identity) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: path(identity)}, nil, http.StatusNoContent)
}

// AssignToUser assigns a policy. Empty policyName sends the evidenced clear form.
func (s *Service) AssignToUser(ctx context.Context, user, policyName string) (*resty.Response, error) {
	if !validIdentity(user) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "users", user, "policies", "ApplicationAccessPolicy"}, Body: map[string]any{"policyName": policyName}}, nil, http.StatusOK)
}

// AssignTenantDefault changes the tenant assignment, separately from policy CRUD.
func (s *Service) AssignTenantDefault(ctx context.Context, policyName string, force bool) (*resty.Response, error) {
	return s.do(ctx, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "tenants", "policies", "ApplicationAccessPolicy"}, Body: map[string]any{"policyName": policyName, "forceSwitchPresent": force}}, nil, http.StatusOK)
}

// AssignToGroup submits an assignment and retains 202 acceptance versus 204.
// An explicit priority, including zero, is sent. The PowerShell composition omits
// zero Rank; callers mirroring that adapter should pass nil for zero.
func (s *Service) AssignToGroup(ctx context.Context, group, policyName string, priority *int) (*GroupAssignment, *resty.Response, error) {
	if !validIdentity(group) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"policyName": policyName}
	if priority != nil {
		body["priority"] = *priority
	}
	var raw json.RawMessage
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{"Skype.Policy", "groupPolicyAssignments", group, "policyTypes", "ApplicationAccessPolicy"}, Body: body}, &raw, http.StatusAccepted, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	result := &GroupAssignment{Accepted: response.StatusCode() == http.StatusAccepted}
	if result.Accepted && (len(raw) == 0 || json.Unmarshal(raw, result) != nil) {
		return result, response, fmt.Errorf("assignment response: %w", client.ErrDecode)
	}
	return result, response, nil
}
