package onlinevoicemailvalidationconfiguration

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// UpdateByIdentity mirrors Set's read/modify/PATCH composition, retaining
// preparation if read, decode or submission fails. Only bound fields are sent.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}
	fields, err := policyrp.Fields(input.Settings)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 {
		return nil, nil, ErrInput
	}
	before, read, err := s.resource.Get(ctx, id, "Set-CsOnlineVoicemailValidationConfiguration")
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	patch := PatchRequest{Settings: input.Settings, Controls: input.Controls, BodyIdentity: policy.Null[string]()}
	if before.Identity != nil {
		patch.BodyIdentity = policy.Set(*before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchByIdentity(ctx, id, patch)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}
