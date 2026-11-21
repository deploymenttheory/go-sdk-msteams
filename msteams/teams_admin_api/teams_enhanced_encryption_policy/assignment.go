package teamsenhancedencryptionpolicy

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// AssignToUser requires a policy name argument; empty sends the captured clear form.
func (s *Service) AssignToUser(ctx context.Context, user, policyName string) (*resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	if !validIdentity(user) {
		return nil, ErrInput
	}
	return s.do(ctx, grantCommand, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "users", user, "policies", resourceName}, Body: map[string]any{"policyName": policyName}, Replay: client.ReplayNever}, nil, http.StatusOK)
}

// AssignTenantDefault submits the native tenant grant. The source mapped helper
// can send forceSwitchPresent; the public family wrapper does not expose Force.
func (s *Service) AssignTenantDefault(ctx context.Context, policyName string, force bool) (*resty.Response, error) {
	return s.do(ctx, grantCommand, client.Request{Method: http.MethodPatch, Path: []string{"Skype.Policy", "tenants", "policies", resourceName}, Body: map[string]any{"policyName": policyName, "forceSwitchPresent": force}, Replay: client.ReplayNever}, nil, http.StatusOK)
}

// AssignToGroup retains 202 acceptance separately from 204, without polling.
// A nonnil priority explicitly sends zero; Assign applies source Rank-zero omission.
func (s *Service) AssignToGroup(ctx context.Context, group, policyName string, priority *int32) (*GroupAssignment, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !validIdentity(group) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"policyName": policyName}
	if priority != nil {
		body["priority"] = *priority
	}
	var raw json.RawMessage
	r, e := s.do(ctx, grantCommand, client.Request{Method: http.MethodPost, Path: []string{"Skype.Policy", "groupPolicyAssignments", group, "policyTypes", resourceName}, Body: body, Replay: client.ReplayNever}, &raw, http.StatusAccepted, http.StatusNoContent)
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
