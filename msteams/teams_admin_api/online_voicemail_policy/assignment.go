package onlinevoicemailpolicy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrResponse = errors.New("unexpected voicemail assignment response")

type UserAssignment struct{ Identity string }
type TenantAssignment struct{ Force bool }
type GroupAssignment struct {
	ID       string
	Priority *int32
}

// Assignment selects exactly one explicit scope. PolicyName must be bound;
// Null omits the generated DTO key, Set("") sends the explicit empty string.
// Priority nil and explicit zero remain distinct in the native operation.
type Assignment struct {
	PolicyName policy.Field[string]
	User       *UserAssignment
	Tenant     *TenantAssignment
	Group      *GroupAssignment
}

// AssignmentResult separates HTTP acceptance from an optional opaque operation
// identifier. It never polls or infers completion/effective assignment state.
type AssignmentResult struct {
	Accepted    bool
	OperationID *string
	RawFields   map[string]json.RawMessage
}

func (AssignmentResult) String() string     { return "voicemail assignment result (redacted)" }
func (v AssignmentResult) GoString() string { return v.String() }

func (s *Service) generated(ctx context.Context, command string, request client.Request, statuses ...int) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}

	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	segments := make([]string, len(request.Path))
	for i, part := range request.Path {
		segments[i] = strings.ReplaceAll(url.QueryEscape(part), "+", "%20")
	}
	request.URL = environment.Endpoints.TeamsAdmin + "/" + strings.Join(segments, "/")
	request.Path = nil
	r, err := s.assignments.Do(ctx, request, nil)
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

// Assign uses the selected modern generated protocol, including forest discovery.
// It does not reuse compiled CRUD/Grant DTOs or replay a mutation automatically.
func (s *Service) Assign(ctx context.Context, input Assignment) (*AssignmentResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	count := 0
	if input.User != nil {
		count++
	}
	if input.Tenant != nil {
		count++
	}
	if input.Group != nil {
		count++
	}
	if count != 1 || !input.PolicyName.IsSet() {
		return nil, nil, ErrInput
	}
	body := map[string]any{}
	if value, ok := input.PolicyName.Value(); ok {
		body["policyName"] = value
	}
	request := client.Request{Method: http.MethodPatch, Body: body, Replay: client.ReplayNever}
	statuses := []int{http.StatusOK}
	switch {
	case input.User != nil:
		if !policyrp.ValidIdentity(input.User.Identity) {
			return nil, nil, ErrInput
		}
		request.Path = []string{"Skype.Policy", "users", input.User.Identity, "policies", "OnlineVoicemailPolicy"}
	case input.Tenant != nil:
		request.Path = []string{"Skype.Policy", "tenants", "policies", "OnlineVoicemailPolicy"}
		body["forceSwitchPresent"] = input.Tenant.Force
	case input.Group != nil:
		if !policyrp.ValidIdentity(input.Group.ID) {
			return nil, nil, ErrInput
		}
		request.Method = http.MethodPost
		request.Path = []string{"Skype.Policy", "groupPolicyAssignments", input.Group.ID, "policyTypes", "OnlineVoicemailPolicy"}
		if input.Group.Priority != nil {
			body["priority"] = *input.Group.Priority
		}
		statuses = []int{http.StatusAccepted, http.StatusNoContent}
	}
	r, err := s.generated(ctx, "Grant-CsOnlineVoicemailPolicy", request, statuses...)
	if err != nil {
		return nil, r, err
	}
	result := &AssignmentResult{Accepted: r.StatusCode() == http.StatusAccepted}
	if result.Accepted {
		if json.Unmarshal(r.Body(), &result.RawFields) != nil {
			return result, r, client.ErrDecode
		}
		if value, present := result.RawFields["OperationId"]; present && json.Unmarshal(value, &result.OperationID) != nil {
			return result, r, client.ErrDecode
		}
	}
	return result, r, nil
}

// AssignWithPowerShellBinding explicitly applies the mapped helper's string/rank
// coercions: bound null PolicyName becomes empty, and zero group Rank is omitted.
// Returned Go data/metadata replaces the optional PowerShell display switch.
func (s *Service) AssignWithPowerShellBinding(ctx context.Context, input Assignment) (*AssignmentResult, *resty.Response, error) {
	if input.PolicyName.IsNull() {
		input.PolicyName = policy.Set("")
	}
	if input.Group != nil {
		group := *input.Group
		if group.Priority != nil && *group.Priority == 0 {
			group.Priority = nil
		}
		input.Group = &group
	}
	return s.Assign(ctx, input)
}

func (s *Service) AssignUser(ctx context.Context, id string, name policy.Field[string]) (*AssignmentResult, *resty.Response, error) {
	return s.Assign(ctx, Assignment{PolicyName: name, User: &UserAssignment{Identity: id}})
}
func (s *Service) AssignTenantDefault(ctx context.Context, name policy.Field[string], force bool) (*AssignmentResult, *resty.Response, error) {
	return s.Assign(ctx, Assignment{PolicyName: name, Tenant: &TenantAssignment{Force: force}})
}
func (s *Service) AssignGroup(ctx context.Context, id string, name policy.Field[string], priority *int32) (*AssignmentResult, *resty.Response, error) {
	return s.Assign(ctx, Assignment{PolicyName: name, Group: &GroupAssignment{ID: id, Priority: priority}})
}
