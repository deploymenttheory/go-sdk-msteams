package recordingmessages

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// UpdateByIdentity explicitly composes Set's read, local edits and sparse PATCH.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	return s.update(ctx, identity, false, input)
}

// UpdateByID retains the read model's Id even when it differs from the route Id.
func (s *Service) UpdateByID(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.update(ctx, identity, true, input)
}
func (s *Service) update(ctx context.Context, identity string, byID bool, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !input.Description.IsSet() && input.Localizations == nil {
		return nil, nil, ErrInput
	}
	if input.Localizations != nil {
		if err := input.Localizations.Validate(); err != nil {
			return nil, nil, err
		}
	}
	before, read, err := s.resource.Get(ctx, identity, setCommand)
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	settings, err := resolve(input, before.Localizations)
	if err != nil {
		return result, read, err
	}
	patch := PatchRequest{Settings: settings, Controls: input.Controls, BodyIdentity: valueField(before.Identity)}
	if byID {
		patch.BodyID = valueField(before.ID)
	}
	result.PatchInvoked = true
	result.After, read, err = s.patch(ctx, identity, byID, patch)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}
