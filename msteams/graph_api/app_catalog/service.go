// Package appcatalog manages the tenant's Teams app catalog through Graph
// /v1.0/appCatalogs/teamsApps, mirroring Get/New/Set/Remove-TeamsApp.
package appcatalog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

const zipType = "application/zip"

// Service owns no credentials; all requests use the client's Graph scope.
type Service struct{ scope *client.Scope }

// NewService constructs a lazy resource without authentication or HTTP.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) local() error {
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.local()
}

func validID(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." && !strings.ContainsFunc(value, unicode.IsControl) && !strings.ContainsAny(value, `/\`)
}

func path(extra ...string) []string {
	return append([]string{"v1.0", "appCatalogs", "teamsApps"}, extra...)
}

func headers() http.Header { return http.Header{"Accept": {"application/json"}} }

func quote(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }

// Filter returns the $filter value List sends, or "" when no option is set.
func (o ListOptions) Filter() string {
	var clauses []string
	for _, item := range []struct{ name, value string }{{"id", o.ID}, {"externalId", o.ExternalID}, {"displayName", o.DisplayName}, {"distributionMethod", o.DistributionMethod}} {
		if strings.TrimSpace(item.value) != "" {
			clauses = append(clauses, item.name+" eq "+quote(item.value))
		}
	}
	return strings.Join(clauses, " and ")
}

// appCollection reproduces JsonConvert's collection reuse for repeated value
// aliases: append arrays in wire order, null resets, a later array starts anew.
type appCollection struct{ items []*TeamsApp }

func (v *appCollection) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		v.items = nil
		return nil
	}
	var items []*TeamsApp
	if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &items) != nil {
		return client.ErrDecode
	}
	if v.items == nil {
		v.items = make([]*TeamsApp, 0, len(items))
	}
	v.items = append(v.items, items...)
	return nil
}

func decode(data []byte) (client.PageContent[*TeamsApp], error) {
	var envelope struct {
		Items appCollection `json:"value"`
		Next  coercedString `json:"@odata.nextLink"`
	}
	var result client.PageContent[*TeamsApp]
	if json.Unmarshal(data, &envelope) != nil {
		return result, client.ErrDecode
	}
	if envelope.Items.items == nil {
		return result, client.ErrDecode
	}
	result.Items = envelope.Items.items
	if envelope.Next.value != nil && strings.TrimSpace(*envelope.Next.value) != "" {
		result.Next = *envelope.Next.value
	}
	return result, nil
}

// NewPager builds a lazy Get-TeamsApp traversal following @odata.nextLink.
func (s *Service) NewPager(options ListOptions) (*client.Pager[*TeamsApp], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	bounds := options.Paging
	if bounds.Style != 0 && bounds.Style != client.LinkPaging || bounds.Parameter != "" {
		return nil, client.ErrPagination
	}
	bounds.Style = client.LinkPaging
	req := client.Request{Method: http.MethodGet, Path: path(), Replay: client.ReplaySafe, Header: headers()}
	if filter := options.Filter(); filter != "" {
		req.Query = url.Values{"$filter": {filter}}
	}
	return client.NewPager(s.scope, req, decode, bounds)
}

// List is Get-TeamsApp: every page of the filtered catalog. The result keeps
// partial items and the last response when traversal fails.
func (s *Service) List(ctx context.Context, options ListOptions) (client.ListResult[*TeamsApp], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[*TeamsApp]{}, err
	}
	pager, err := s.NewPager(options)
	if err != nil {
		return client.ListResult[*TeamsApp]{}, err
	}
	return pager.Collect(ctx)
}

func (s *Service) upload(ctx context.Context, method string, segments []string, pkg []byte) (*resty.Response, error) {
	if pkg == nil {
		return nil, ErrInput
	}
	req := client.Request{Method: method, Path: path(segments...), Header: headers(), Replay: client.ReplayNever}
	return s.scope.Upload(ctx, req, io.NopCloser(bytes.NewReader(pkg)), zipType, int64(max(1, len(pkg))), nil)
}

// Create is New-TeamsApp: it uploads an app package to the organization
// catalog and returns the nullable source teamsApp from any 2xx. An accepted
// empty/null response retains the receipt without inventing an identity.
// DistributionMethod must be
// "organization" in any letter case; other values fail locally, as in the
// module. Uploads are never retried; reconcile a lost response with List.
func (s *Service) Create(ctx context.Context, value CreateRequest) (*TeamsApp, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !strings.EqualFold(value.DistributionMethod, OrganizationDistribution) {
		return nil, nil, ErrInput
	}
	response, err := s.upload(ctx, http.MethodPost, nil, value.Package)
	if err != nil {
		return nil, response, err
	}
	if len(bytes.TrimSpace(response.Body())) == 0 {
		return nil, response, nil
	}
	var app *TeamsApp
	if json.Unmarshal(response.Body(), &app) != nil {
		return nil, response, client.ErrDecode
	}
	return app, response, nil
}

// UpdateByID is Set-TeamsApp: it PUTs a new app package to
// /appCatalogs/teamsApps/{id}. The module accepts any 2xx and reads no body;
// every 2xx is accepted. Never retried.
func (s *Service) UpdateByID(ctx context.Context, id string, pkg []byte) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	return s.upload(ctx, http.MethodPut, []string{id}, pkg)
}

// DeleteByID is Remove-TeamsApp: any 2xx succeeds. It is not replayed automatically.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	response, err := s.scope.Do(ctx, client.Request{Method: http.MethodDelete, Path: path(id), Header: headers()}, nil)
	if err != nil {
		return response, err
	}
	return response, nil
}
