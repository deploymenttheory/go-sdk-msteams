package teamsaipolicy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policytransport"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// compiledResource binds the compiled PolicyRP protocol to the single
// TeamsAIPolicy configuration name. It sends through the shared
// policytransport client and reuses the exported policyrp identity, body and
// decode helpers, so the wire contract matches every other PolicyRP family.
type compiledResource struct{ scope *client.Scope }

func (s *compiledResource) Ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.TeamsAdmin {
		return client.ErrDestination
	}
	return nil
}

func (s *compiledResource) path(identity string) ([]string, error) {
	p := []string{"Skype.Policy", "configurations", resourceName}
	if identity == "" {
		return p, nil
	}
	if !policyrp.ValidIdentity(identity) {
		return nil, policy.ErrInput
	}
	p = append(p, "configuration")
	before, after, scoped := strings.Cut(identity, "/")
	if scoped {
		if before == "" || after == "" || before == "." || before == ".." || after == "." || after == ".." {
			return nil, policy.ErrInput
		}
		return append(p, before, after), nil
	}
	return append(p, identity), nil
}

func (s *compiledResource) Do(ctx context.Context, command, method string, path []string, body any, replay client.ReplaySafety) (*resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, err
	}
	return policytransport.New(s.scope).Do(ctx, command, method, path, nil, body, replay)
}

func (s *compiledResource) Get(ctx context.Context, identity, command string) (*Policy, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	p, err := s.path(identity)
	if err != nil || identity == "" {
		return nil, nil, policy.ErrInput
	}
	r, err := s.Do(ctx, command, http.MethodGet, p, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := policyrp.Decode[Policy](r)
	return v, r, err
}

func (s *compiledResource) List(ctx context.Context, command string) ([]*Policy, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	p, _ := s.path("")
	r, err := s.Do(ctx, command, http.MethodGet, p, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	var values []*Policy
	if len(bytes.TrimSpace(r.Body())) > 0 && json.Unmarshal(r.Body(), &values) != nil {
		return nil, r, client.ErrDecode
	}
	global := func(v *Policy) bool { return v != nil && v.Identity != nil && *v.Identity == "Global" }
	slices.SortStableFunc(values, func(a, b *Policy) int {
		left, right := global(a), global(b)
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	return values, r, nil
}

func (s *compiledResource) ListFiltered(ctx context.Context, pattern, command string) ([]*Policy, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	plain, err := policyfilter.Compile(pattern)
	if err != nil {
		return nil, nil, policy.ErrInput
	}
	tagged, err := policyfilter.Compile("Tag:" + pattern)
	if err != nil {
		return nil, nil, policy.ErrInput
	}
	items, r, err := s.List(ctx, command)
	if err != nil || items == nil {
		return nil, r, err
	}
	values := make([]*Policy, 0, len(items))
	for _, item := range items {
		if item == nil || item.Identity == nil {
			continue
		}
		match, err := plain.Match(ctx, *item.Identity)
		if err == nil && !match {
			match, err = tagged.Match(ctx, *item.Identity)
		}
		if err != nil {
			return nil, r, err
		}
		if match {
			values = append(values, item)
		}
	}
	return values, r, nil
}

func (s *compiledResource) Write(ctx context.Context, identity, command, method string, body map[string]json.RawMessage) (*Policy, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(identity) {
		return nil, nil, policy.ErrInput
	}
	target := identity
	if method == http.MethodPost {
		target = ""
	}
	p, err := s.path(target)
	if err != nil {
		return nil, nil, err
	}
	r, err := s.Do(ctx, command, method, p, body, client.ReplayNever)
	if err != nil {
		return nil, r, err
	}
	v, err := policyrp.Decode[Policy](r)
	return v, r, err
}
