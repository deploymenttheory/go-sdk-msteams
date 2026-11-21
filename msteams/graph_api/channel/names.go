package channel

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// ResolveByDisplayName is an explicit complete-list composition. Unlike the
// pinned Set/Remove commands' first-page lookup, it rejects ambiguity and never
// mutates after an incomplete traversal. Comparison is case-insensitive.
func (s *Service) ResolveByDisplayName(ctx context.Context, teamID, name string, limits client.PagingOptions) (*Channel, *resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer stop()
	if strings.TrimSpace(name) == "" {
		return nil, nil, client.ErrRequest
	}
	listed, err := s.ListComplete(ctx, teamID, ListOptions{}, limits)
	if err != nil {
		return nil, listed.Response, err
	}
	var found *Channel
	for _, value := range listed.Items {
		if value.DisplayName != nil && strings.EqualFold(*value.DisplayName, name) {
			if found != nil {
				return nil, listed.Response, ErrAmbiguous
			}
			candidate := value
			found = &candidate
		}
	}
	if found == nil {
		return nil, listed.Response, ErrNotFound
	}
	return found, listed.Response, nil
}

// UpdateByDisplayName resolves a unique hosted channel, then uses UpdateByID.
func (s *Service) UpdateByDisplayName(ctx context.Context, teamID, name string, value UpdateRequest, limits client.PagingOptions) (*resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer stop()
	found, response, err := s.ResolveByDisplayName(ctx, teamID, name, limits)
	if err != nil {
		return response, err
	}
	return s.UpdateByID(ctx, teamID, found.ID, value)
}

// DeleteByDisplayName resolves a unique hosted channel, then uses DeleteByID.
func (s *Service) DeleteByDisplayName(ctx context.Context, teamID, name string, limits client.PagingOptions) (*resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer stop()
	found, response, err := s.ResolveByDisplayName(ctx, teamID, name, limits)
	if err != nil {
		return response, err
	}
	return s.DeleteByID(ctx, teamID, found.ID)
}
