package meetingbranding

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// UpdateByIdentity reads current, applies only bound scalar/object-list edits,
// then submits PATCH once. Comparator failures retain Before without PATCH.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	fields, err := policyrp.Fields(input.Fields)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 && input.NDIAssuranceSlateImages == nil && input.MeetingBackgroundImages == nil && input.MeetingBrandingThemes == nil {
		return nil, nil, ErrInput
	}
	if input.NDIAssuranceSlateImages != nil {
		if err := input.NDIAssuranceSlateImages.Validate(); err != nil {
			return nil, nil, err
		}
	}
	if input.MeetingBackgroundImages != nil {
		if err := input.MeetingBackgroundImages.Validate(); err != nil {
			return nil, nil, err
		}
	}
	if input.MeetingBrandingThemes != nil {
		if err := input.MeetingBrandingThemes.Validate(); err != nil {
			return nil, nil, err
		}
	}
	before, read, err := s.resource.Get(ctx, id, "Set-CsTeamsMeetingBrandingPolicy")
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	patch := PatchRequest{Settings: Settings{Fields: input.Fields}, Controls: input.Controls, BodyIdentity: policy.Null[string]()}
	if before.Identity != nil {
		patch.BodyIdentity = policy.Set(*before.Identity)
	}
	if input.NDIAssuranceSlateImages != nil {
		patch.NDIAssuranceSlateImages, err = input.NDIAssuranceSlateImages.Apply(before.NDIAssuranceSlateImages)
		if err != nil {
			return result, read, err
		}
	}
	if input.MeetingBackgroundImages != nil {
		patch.MeetingBackgroundImages, err = input.MeetingBackgroundImages.Apply(before.MeetingBackgroundImages)
		if err != nil {
			return result, read, err
		}
	}
	if input.MeetingBrandingThemes != nil {
		patch.MeetingBrandingThemes, err = input.MeetingBrandingThemes.Apply(before.MeetingBrandingThemes)
		if err != nil {
			return result, read, err
		}
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchByIdentity(ctx, id, patch)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}
