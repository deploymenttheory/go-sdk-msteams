package policyassignments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

// Service owns generated assignment operations. Regional OCE operations have a
// separate direct protocol; neither protocol is a Graph directory update.
type Service struct {
	scope     *client.Scope
	generated *configapi.Client
}

func NewService(scope *client.Scope) *Service {
	return &Service{scope: scope, generated: configapi.New(scope)}
}
func (s *Service) ready(ctx context.Context) error {
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

func (s *Service) request(ctx context.Context, command string, request client.Request, statuses ...int) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	parts := make([]string, len(request.Path))
	for i, value := range request.Path {
		parts[i] = strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
	}
	request.URL = environment.Endpoints.TeamsAdmin + "/" + strings.Join(parts, "/")
	if query := request.Query.Encode(); query != "" {
		request.URL += "?" + strings.ReplaceAll(query, "+", "%20")
	}
	request.Path, request.Query = nil, nil
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	r, err := s.generated.Do(ctx, request, nil)
	if err != nil {
		return r, err
	}
	for _, status := range statuses {
		if r != nil && r.StatusCode() == status {
			return r, nil
		}
	}
	return r, ErrResponse
}

func groupPath(id, policyType string) []string {
	parts := []string{"Skype.Policy", "groupPolicyAssignments", id, "policyTypes"}
	if policyType != "" {
		parts = append(parts, policyType)
	}
	return parts
}
func optionalQuery(key, value string) url.Values {
	if value == "" {
		return nil
	}
	return url.Values{key: {value}}
}
func object[T any](r *resty.Response) (*T, error) {
	var value *T
	if r == nil || json.Unmarshal(r.Body(), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}
func array[T any](r *resty.Response) ([]*T, error) {
	var value []*T
	if r == nil || json.Unmarshal(r.Body(), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

// GetGroup reads one group/policy-type assignment, including its server rank.
func (s *Service) GetGroup(ctx context.Context, groupID, policyType string) (*GroupAssignment, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(groupID) || !policyrp.ValidIdentity(policyType) {
		return nil, nil, ErrInput
	}
	r, err := s.request(ctx, "Get-CsGroupPolicyAssignment", client.Request{Method: http.MethodGet, Path: groupPath(groupID, policyType), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := object[GroupAssignment](r)
	return value, r, err
}

// ListGroup reads assignments for one group. It returns the bare response array
// without fabricated pagination, sorting or effective-policy calculations.
func (s *Service) ListGroup(ctx context.Context, groupID string) ([]*GroupAssignment, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(groupID) {
		return nil, nil, ErrInput
	}
	r, err := s.request(ctx, "Get-CsGroupPolicyAssignment", client.Request{Method: http.MethodGet, Path: groupPath(groupID, ""), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := array[GroupAssignment](r)
	return value, r, err
}

// ListGroupsByPolicy reads the policy-type query variant. Empty omits the query
// and requests all group assignments, matching the generated optional parameter.
func (s *Service) ListGroupsByPolicy(ctx context.Context, policyType string) ([]*GroupAssignment, *resty.Response, error) {
	r, err := s.request(ctx, "Get-CsGroupPolicyAssignment", client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "groupPolicyAssignments"}, Query: optionalQuery("policyType", policyType), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := array[GroupAssignment](r)
	return value, r, err
}

// GetUser reads server-reported effective policies and ordered source entries.
// PolicyType empty omits the query. It never infers precedence from source order.
func (s *Service) GetUser(ctx context.Context, identity, policyType string) ([]*EffectivePolicy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(identity) {
		return nil, nil, ErrInput
	}
	r, err := s.request(ctx, "Get-CsUserPolicyAssignment", client.Request{Method: http.MethodGet, Path: []string{"Skype.User", "users", identity, "policies"}, Query: optionalQuery("policyType", policyType), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := array[EffectivePolicy](r)
	return value, r, err
}

// GetUserByIdentity maps the exported InputObject variant, whose command passes
// InputObject.Identity to the same generated UserGetPolicyAssignment operation.
func (s *Service) GetUserByIdentity(ctx context.Context, input UserIdentity, policyType string) ([]*EffectivePolicy, *resty.Response, error) {
	return s.GetUser(ctx, input.Identity, policyType)
}

func (s *Service) groupWrite(ctx context.Context, method, command, groupID, policyType string, settings *GroupSettings, status int) (*OperationResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(groupID) || !policyrp.ValidIdentity(policyType) {
		return nil, nil, ErrInput
	}
	request := client.Request{Method: method, Path: groupPath(groupID, policyType), Replay: client.ReplayNever}
	if settings != nil {
		body, err := json.Marshal(settings)
		if err != nil {
			return nil, nil, ErrInput
		}
		request.Body = json.RawMessage(body)
	}
	r, err := s.request(ctx, command, request, status, http.StatusNoContent)
	if err != nil {
		return nil, r, err
	}
	result, err := operation(r)
	return result, r, err
}

// CreateGroup uses the evidenced PUT protocol. A 201 may carry an operation ID;
// 204 has no operation body. No read-back or propagation wait runs implicitly.
func (s *Service) CreateGroup(ctx context.Context, groupID, policyType string, settings GroupSettings) (*OperationResult, *resty.Response, error) {
	return s.groupWrite(ctx, http.MethodPut, "New-CsGroupPolicyAssignment", groupID, policyType, &settings, http.StatusCreated)
}

// UpdateGroup is an additional native operation backed by the pinned generated
// GroupSetPolicyAssignment method. It has no exported baseline command claim.
func (s *Service) UpdateGroup(ctx context.Context, groupID, policyType string, settings GroupSettings) (*OperationResult, *resty.Response, error) {
	return s.groupWrite(ctx, http.MethodPatch, "Set-CsGroupPolicyAssignment", groupID, policyType, &settings, http.StatusAccepted)
}

func (s *Service) DeleteGroup(ctx context.Context, groupID, policyType string) (*OperationResult, *resty.Response, error) {
	return s.groupWrite(ctx, http.MethodDelete, "Remove-CsGroupPolicyAssignment", groupID, policyType, nil, http.StatusAccepted)
}

func operation(r *resty.Response) (*OperationResult, error) {
	result := &OperationResult{HTTPAccepted: true}
	if r.StatusCode() == http.StatusNoContent {
		return result, nil
	}
	if json.Unmarshal(r.Body(), &result.RawFields) != nil || result.RawFields == nil {
		return result, client.ErrDecode
	}
	if value, ok := result.RawFields["OperationId"]; ok {
		var id *string
		if json.Unmarshal(value, &id) != nil {
			return result, client.ErrDecode
		}
		result.OperationID = id
	}
	return result, nil
}

func (s *Service) GetBatch(ctx context.Context, id string) (*BatchStatus, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(id) {
		return nil, nil, ErrInput
	}
	r, err := s.request(ctx, "Get-CsBatchPolicyAssignmentOperation", client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "assignments", "operations", id}, Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := object[BatchStatus](r)
	return value, r, err
}

func (s *Service) ListBatches(ctx context.Context, status string) ([]*BatchStatus, *resty.Response, error) {
	r, err := s.request(ctx, "Get-CsBatchPolicyAssignmentOperation", client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "assignments", "operations"}, Query: optionalQuery("status", status), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	value, err := array[BatchStatus](r)
	return value, r, err
}

// CreateBatch submits once, preserving identity order and duplicates. Required
// Identities must be supplied; a non-nil empty collection is allowed by the
// pinned command. Successful submission does not imply per-user success.
func (s *Service) CreateBatch(ctx context.Context, input BatchCreateRequest) (*OperationResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.Identities == nil {
		return nil, nil, ErrInput
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, nil, ErrInput
	}
	r, err := s.request(ctx, "New-CsBatchPolicyAssignmentOperation", client.Request{Method: http.MethodPost, Path: []string{"Skype.Policy", "assignments", "operations"}, Query: optionalQuery("operationName", input.OperationName), Body: json.RawMessage(body), Replay: client.ReplayNever}, http.StatusAccepted)
	if err != nil {
		return nil, r, err
	}
	result, err := operation(r)
	return result, r, err
}
