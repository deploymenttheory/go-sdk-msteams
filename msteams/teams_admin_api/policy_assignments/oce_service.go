package policyassignments

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

type oceCall struct {
	scope   *client.Scope
	profile oceProfile
	headers http.Header
}

func (s *Service) prepareOCE(ctx context.Context, input OCEContext) (*oceCall, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	tenant, ok := canonicalGUID(input.TenantID)
	if !ok {
		return nil, ErrInput
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	var selected *oceProfile
	for _, profile := range oceProfiles() {
		if input.Environment != "" && !strings.EqualFold(string(input.Environment), string(profile.environment)) {
			continue
		}
		if profile.origin == environment.Endpoints.TeamsAdmin {
			selected = &profile
			break
		}
		if input.Environment != "" {
			return nil, client.ErrDestination
		}
	}
	if selected == nil {
		if input.Environment != "" {
			for _, profile := range oceProfiles() {
				if strings.EqualFold(string(input.Environment), string(profile.environment)) {
					return nil, client.ErrDestination
				}
			}
			return nil, ErrInput
		}
		return nil, client.ErrDestination
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, client.ErrRequest
	}
	id[6], id[8] = id[6]&0x0f|0x40, id[8]&0x3f|0x80
	correlation := fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
	return &oceCall{scope: s.scope, profile: *selected, headers: http.Header{"X-MS-Tenant-Id": {tenant}, "X-MS-Correlation-Id": {correlation}}}, nil
}

func (c *oceCall) do(ctx context.Context, method string, path []string, query url.Values, statuses ...int) (*resty.Response, error) {
	parts := make([]string, len(path))
	for i, value := range path {
		parts[i] = strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
	}
	destination := c.profile.origin + "/" + strings.Join(parts, "/")
	if encoded := query.Encode(); encoded != "" {
		destination += "?" + strings.ReplaceAll(encoded, "+", "%20")
	}
	// The pinned direct OCE sender makes one attempt, including for GET. Regional
	// search owns fallback; shared retry defaults must not hide failed attempts.
	r, err := c.scope.Do(ctx, client.Request{Method: method, URL: destination, Header: c.headers, Replay: client.ReplayNever}, nil)
	if err != nil {
		return r, err
	}
	if len(statuses) == 0 {
		statuses = []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent}
	}
	for _, status := range statuses {
		if r != nil && r.StatusCode() == status {
			return r, nil
		}
	}
	return r, ErrResponse
}

// GPA's Newtonsoft decoder permits empty and JSON-null reference responses.
// This is separate from generated assignment reads, which reject null envelopes.
func oceObject[T any](r *resty.Response) (*T, error) {
	if r == nil {
		return nil, client.ErrDecode
	}
	var value *T
	if len(bytes.TrimSpace(r.Body())) != 0 && json.Unmarshal(r.Body(), &value) != nil {
		return nil, client.ErrDecode
	}
	return value, nil
}
func oceArray[T any](r *resty.Response) ([]*T, error) {
	if r == nil {
		return nil, client.ErrDecode
	}
	var values []*T
	if len(bytes.TrimSpace(r.Body())) != 0 && json.Unmarshal(r.Body(), &values) != nil {
		return nil, client.ErrDecode
	}
	return values, nil
}

func oceGroupPath(groupID, policyType string) []string {
	parts := []string{"Skype.Policy", "groupPolicyAssignments", "oce"}
	if groupID != "" {
		parts = append(parts, groupID, "policyTypes")
	}
	if policyType != "" {
		parts = append(parts, policyType)
	}
	return parts
}

func (s *Service) GetOCEGroup(ctx context.Context, input OCEContext, groupID, policyType string) (*GroupAssignment, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, ok := canonicalGUID(groupID)
	if !ok || !policyrp.ValidIdentity(policyType) {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodGet, oceGroupPath(id, policyType), nil)
	if err != nil {
		return nil, r, err
	}
	v, err := oceObject[GroupAssignment](r)
	return v, r, err
}

func (s *Service) ListOCEGroup(ctx context.Context, input OCEContext, groupID string) ([]*GroupAssignment, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, ok := canonicalGUID(groupID)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodGet, oceGroupPath(id, ""), nil)
	if err != nil {
		return nil, r, err
	}
	v, err := oceArray[GroupAssignment](r)
	return v, r, err
}

// ListOCEGroupsByPolicy selects the query parameter set, retaining policyType=
// when empty. ListOCEGroups is the separate default set without that query.
func (s *Service) ListOCEGroupsByPolicy(ctx context.Context, input OCEContext, policyType string) ([]*GroupAssignment, *resty.Response, error) {
	return s.listOCEGroups(ctx, input, url.Values{"policyType": {policyType}})
}
func (s *Service) ListOCEGroups(ctx context.Context, input OCEContext) ([]*GroupAssignment, *resty.Response, error) {
	return s.listOCEGroups(ctx, input, nil)
}
func (s *Service) listOCEGroups(ctx context.Context, input OCEContext, query url.Values) ([]*GroupAssignment, *resty.Response, error) {
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodGet, oceGroupPath("", ""), query)
	if err != nil {
		return nil, r, err
	}
	v, err := oceArray[GroupAssignment](r)
	return v, r, err
}

func (s *Service) ListOCEGroupMembers(ctx context.Context, input OCEContext, groupID string, all bool) ([]*OCEGroupMember, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, ok := canonicalGUID(groupID)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	flag := "False"
	if all {
		flag = "True"
	}
	r, err := call.do(ctx, http.MethodGet, []string{"Teams.InternalSupport", "groups", id, "users"}, url.Values{"all": {flag}})
	if err != nil {
		return nil, r, err
	}
	v, err := oceArray[OCEGroupMember](r)
	return v, r, err
}

func (s *Service) ListOCEUserMembership(ctx context.Context, input OCEContext, userID string) ([]*OCEUserMembership, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, ok := canonicalGUID(userID)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodGet, []string{"Teams.InternalSupport", "users", id, "memberOf"}, nil)
	if err != nil {
		return nil, r, err
	}
	v, err := oceArray[OCEUserMembership](r)
	return v, r, err
}

type OCERefreshOptions struct {
	UserID     *string
	PolicyType *string
}

func (OCERefreshOptions) String() string     { return "OCE group refresh options (redacted)" }
func (v OCERefreshOptions) GoString() string { return v.String() }

// RefreshOCEGroupUsers is an explicit consequential action. An omitted UserID
// substitutes Guid.Empty; both optional query keys remain present. It never runs
// as an implicit follow-up to a group assignment or a successful status read.
func (s *Service) RefreshOCEGroupUsers(ctx context.Context, input OCEContext, groupID string, region Region, options OCERefreshOptions) (*OCEActionResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, validID := canonicalGUID(groupID)
	selected, validRegion := normalizeRegion(region)
	if !validID || !validRegion {
		return nil, nil, ErrInput
	}
	user := "00000000-0000-0000-0000-000000000000"
	if options.UserID != nil {
		var ok bool
		user, ok = canonicalGUID(*options.UserID)
		if !ok {
			return nil, nil, ErrInput
		}
	}
	policyType := ""
	if options.PolicyType != nil {
		policyType = *options.PolicyType
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodPost, []string{"Skype.Policy", "groupPolicyAssignments", id, "refresh", "users"}, url.Values{"tenantRegion": {string(selected)}, "userId": {user}, "policyType": {policyType}})
	if err != nil {
		return nil, r, err
	}
	v, err := oceAction(r)
	return v, r, err
}

func (s *Service) ReprocessOCEGroupAssignment(ctx context.Context, input OCEContext, groupID, policyType string, region Region) (*OCEActionResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	id, validID := canonicalGUID(groupID)
	selected, validRegion := normalizeRegion(region)
	if !validID || !validRegion || !policyrp.ValidIdentity(policyType) {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodPost, []string{"Skype.Policy", "groupPolicyAssignments", id, "policyTypes", policyType, "reprocess"}, url.Values{"tenantRegion": {string(selected)}})
	if err != nil {
		return nil, r, err
	}
	v, err := oceAction(r)
	return v, r, err
}

func oceAction(r *resty.Response) (*OCEActionResult, error) {
	result := &OCEActionResult{HTTPAccepted: true}
	value, err := oceObject[string](r)
	if err != nil {
		return result, err
	}
	result.Value = value
	return result, nil
}
