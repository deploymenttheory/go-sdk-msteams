package onlinevoicemailpolicy

import (
	"context"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Create submits one sparse native POST. It does not fetch Global, copy service
// defaults, perform post-create read-back or equate HTTP acceptance with propagation.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := policyrp.Fields(input.Settings)
	if err != nil {
		return nil, nil, err
	}
	if err := policyrp.Identity(body, input.Identity, policy.Field[string]{}, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, input.Identity, "New-CsOnlineVoicemailPolicy", http.MethodPost, body)
}

// CreateWithGlobalComparison provides the pinned New command's comparison
// read explicitly. A failed comparison is retained as an observation and does
// not prevent creation. The Global snapshot is never cloned into the new body.
func (s *Service) CreateWithGlobalComparison(ctx context.Context, input CreateRequest) (*CreateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(input.Identity) {
		return nil, nil, ErrInput
	}
	if _, err := policyrp.Fields(input.Settings); err != nil {
		return nil, nil, err
	}
	result := &CreateResult{}
	if !strings.EqualFold(input.Identity, "Global") {
		result.Comparison, result.ComparisonResponse, result.ComparisonError = s.resource.Get(ctx, "Global", "New-CsOnlineVoicemailPolicy")
	}
	result.CreateInvoked = true
	value, response, err := s.Create(ctx, input)
	result.Created = value
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	return result, response, err
}

// DeleteByIdentity sends one DELETE with bound Identity and optional string Force
// in its JSON body. It performs no GET or automatic mutation replay.
func (s *Service) DeleteByIdentity(ctx context.Context, id string, controls policy.Controls) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	body, err := policyrp.Fields(struct{}{})
	if err != nil {
		return nil, err
	}
	if err := policyrp.Identity(body, id, policy.Field[string]{}, controls); err != nil {
		return nil, err
	}
	_, response, err := s.resource.Write(ctx, id, "Remove-CsOnlineVoicemailPolicy", http.MethodDelete, body)
	return response, err
}
