package applicationaccesspolicy

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/go-resty/resty/v2"
)

func compilePattern(pattern string) (policyfilter.Pattern, error) {
	parsed, err := policyfilter.Compile(pattern)
	if err != nil {
		return policyfilter.Pattern{}, ErrInput
	}
	return parsed, nil
}

// ListFiltered composes List with the pinned helper's local Identity/Tag:Identity
// wildcard filtering. It supports *, ?, bracket sets/ranges and backtick escapes;
// matching uses portable Unicode case folding and UTF-16 character positions.
// Locale-specific PowerShell collation is not inferred from the tenant locale.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]Policy, *resty.Response, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(pattern) == "" {
		return s.List(ctx)
	}
	plain, err := compilePattern(pattern)
	if err != nil {
		return nil, nil, err
	}
	tagged, err := compilePattern("Tag:" + pattern)
	if err != nil {
		return nil, nil, err
	}
	items, response, err := s.List(ctx)
	if err != nil {
		return nil, response, err
	}
	filtered := make([]Policy, 0)
	for _, item := range items {
		matches, err := plain.Match(ctx, item.Identity)
		if err != nil {
			return nil, response, err
		}
		if !matches {
			matches, err = tagged.Match(ctx, item.Identity)
			if err != nil {
				return nil, response, err
			}
		}
		if matches {
			filtered = append(filtered, item)
		}
	}
	return filtered, response, nil
}
