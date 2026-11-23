package compliancerecording

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

// AssignToUser requires a policy name argument; empty sends the captured clear form.
func (s *Service) AssignToUser(ctx context.Context, user, policyName string) (*resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	if !policyrp.ValidIdentity(user) {
		return nil, ErrInput
	}
	return s.grantDo(ctx, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "users", user, "policies", policyResource}, Body: map[string]any{"policyName": policyName}, Replay: client.ReplayNever}, nil, http.StatusOK)
}

// AssignTenantDefault submits the native tenant grant. The source mapped helper
// can send forceSwitchPresent; the public family wrapper does not expose Force.
func (s *Service) AssignTenantDefault(ctx context.Context, policyName string, force bool) (*resty.Response, error) {
	return s.grantDo(ctx, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "tenants", "policies", policyResource}, Body: map[string]any{"policyName": policyName, "forceSwitchPresent": force}, Replay: client.ReplayNever}, nil, http.StatusOK)
}

// AssignToGroup retains 202 acceptance separately from 204, without polling.
// A nonnil priority explicitly sends zero; Assign applies source Rank-zero omission.
func (s *Service) AssignToGroup(ctx context.Context, group, policyName string, priority *int32) (*GroupAssignment, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(group) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"policyName": policyName}
	if priority != nil {
		body["priority"] = *priority
	}
	var raw json.RawMessage
	r, e := s.grantDo(ctx, client.Request{Method: http.MethodPost, Path: []string{"Skype.Policy", "groupPolicyAssignments", group, "policyTypes", policyResource}, Body: body, Replay: client.ReplayNever}, &raw, http.StatusAccepted, http.StatusNoContent)
	if e != nil {
		// A malformed accepted body must not hide the successful HTTP submission.
		// ConfigAPI can return GET discovery metadata on error. Acceptance
		// requires metadata from the actual POST, even when body decoding fails.
		if r != nil && r.StatusCode() == http.StatusAccepted && r.Request != nil && r.Request.Method == http.MethodPost {
			return &GroupAssignment{Accepted: true}, r, e
		}
		return nil, r, e
	}
	out := &GroupAssignment{Accepted: r.StatusCode() == http.StatusAccepted}
	if out.Accepted {
		if len(raw) == 0 {
			return out, r, client.ErrDecode
		}
		out.RawJSON = append(json.RawMessage(nil), raw...)
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return out, r, client.ErrDecode
		}
		if v, ok := fields["OperationId"]; ok {
			_ = json.Unmarshal(v, &out.OperationID)
		}
	}
	return out, r, nil
}

// Assign selects one source scope. It omits zero group Rank like the wrapper.
func (s *Service) Assign(ctx context.Context, input Assignment) (*GroupAssignment, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
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
	if count != 1 {
		return nil, nil, ErrInput
	}
	if v := input.User; v != nil {
		r, e := s.AssignToUser(ctx, v.Identity, v.PolicyName)
		return nil, r, e
	}
	if v := input.Tenant; v != nil {
		r, e := s.AssignTenantDefault(ctx, v.PolicyName, false)
		return nil, r, e
	}
	v := input.Group
	priority := v.Rank
	if priority != nil && *priority == 0 {
		priority = nil
	}
	return s.AssignToGroup(ctx, v.ID, v.PolicyName, priority)
}

// grantDo uses generated forest discovery only for exported modern Grant calls.
func (s *Service) grantDo(ctx context.Context, request client.Request, result any, statuses ...int) (*resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	if s.modern == nil {
		return nil, client.ErrClosed
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {grantCommand}}
	r, e := s.modern.Do(ctx, request, result)
	if e != nil {
		return r, e
	}
	for _, status := range statuses {
		if r != nil && r.StatusCode() == status {
			return r, nil
		}
	}
	return r, ErrResponse
}

// UserAssignment grants a policy to one user; an empty PolicyName clears it.
type UserAssignment struct{ Identity, PolicyName string }

// TenantAssignment sets the tenant default policy.
type TenantAssignment struct{ PolicyName string }

// GroupGrant grants a policy to a group with an optional rank.
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}

// Assignment selects exactly one Grant-CsTeamsComplianceRecordingPolicy scope.
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
