package applicationmeetingconfiguration

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/go-resty/resty/v2"
)

// ListFiltered composes the bare configuration list with the pinned wrapper's local
// Identity or Tag:Identity wildcard matching. It sends no Filter query and
// preserves the stable Global-first order. Blank input selects the full list.
// UTF-16 wildcard positions and portable case folding match the established
// SDK policy; locale-specific PowerShell collation remains separately qualified.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(pattern) == "" {
		return s.List(ctx)
	}
	plain, err := policyfilter.Compile(pattern)
	if err != nil {
		return nil, nil, ErrInput
	}
	tagged, err := policyfilter.Compile("Tag:" + pattern)
	if err != nil {
		return nil, nil, ErrInput
	}
	items, response, err := s.List(ctx)
	if err != nil {
		return nil, response, err
	}
	filtered := make([]Configuration, 0)
	for _, item := range items {
		match, err := plain.Match(ctx, item.Identity)
		if err != nil {
			return nil, response, err
		}
		if !match {
			match, err = tagged.Match(ctx, item.Identity)
		}
		if err != nil {
			return nil, response, err
		}
		if match {
			filtered = append(filtered, item)
		}
	}
	return filtered, response, nil
}
