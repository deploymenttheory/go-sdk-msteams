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

func decode(data []byte) (client.PageContent[TeamsApp], error) {
	return client.DecodeCollection[TeamsApp](data, "value", "@odata.nextLink")
}

// NewPager builds a lazy Get-TeamsApp traversal following @odata.nextLink.
func (s *Service) NewPager(options ListOptions) (*client.Pager[TeamsApp], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	for _, value := range []string{options.ID, options.ExternalID, options.DisplayName, options.DistributionMethod} {
		if strings.ContainsFunc(value, unicode.IsControl) {
			return nil, ErrInput
		}
	}
	bounds := options.Paging
	if bounds.Style != 0 && bounds.Style != client.LinkPaging || bounds.Parameter != "" || bounds.ExpectedStatus != 0 && bounds.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	bounds.Style, bounds.ExpectedStatus = client.LinkPaging, http.StatusOK
	req := client.Request{Method: http.MethodGet, Path: path(), Replay: client.ReplaySafe, Header: headers()}
	if filter := options.Filter(); filter != "" {
		req.Query = url.Values{"$filter": {filter}}
	}
	return client.NewPager(s.scope, req, decode, bounds)
}

// List is Get-TeamsApp: every page of the filtered catalog. The result keeps
// partial items and the last response when traversal fails.
func (s *Service) List(ctx context.Context, options ListOptions) (client.ListResult[TeamsApp], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[TeamsApp]{}, err
	}
	pager, err := s.NewPager(options)
	if err != nil {
		return client.ListResult[TeamsApp]{}, err
	}
	return pager.Collect(ctx)
}

func (s *Service) upload(ctx context.Context, method string, segments []string, pkg []byte, statuses ...int) (*resty.Response, error) {
	if len(pkg) == 0 {
		return nil, ErrInput
	}
	req := client.Request{Method: method, Path: path(segments...), Header: headers(), Replay: client.ReplayNever}
	response, err := s.scope.Upload(ctx, req, io.NopCloser(bytes.NewReader(pkg)), zipType, int64(len(pkg)), nil)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// Create is New-TeamsApp: it uploads an app package to the organization
// catalog and returns the created teamsApp (201). DistributionMethod must be
// "organization" in any letter case; other values fail locally, as in the
// module. Uploads are never retried; reconcile a lost response with List.
func (s *Service) Create(ctx context.Context, value CreateRequest) (*TeamsApp, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !strings.EqualFold(value.DistributionMethod, OrganizationDistribution) {
		return nil, nil, ErrInput
	}
	response, err := s.upload(ctx, http.MethodPost, nil, value.Package, http.StatusCreated, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var app TeamsApp
	if json.Unmarshal(response.Body(), &app) != nil {
		return nil, response, client.ErrDecode
	}
	return &app, response, nil
}

// UpdateByID is Set-TeamsApp: it PUTs a new app package to
// /appCatalogs/teamsApps/{id}. The module accepts any 2xx and reads no body;
// declared successes are 200, 201 and 204. Never retried.
func (s *Service) UpdateByID(ctx context.Context, id string, pkg []byte) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	return s.upload(ctx, http.MethodPut, []string{id}, pkg, http.StatusOK, http.StatusCreated, http.StatusNoContent)
}

// DeleteByID is Remove-TeamsApp (204). It is not replayed automatically.
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
	if response.StatusCode() != http.StatusNoContent {
		return response, ErrResponse
	}
	return response, nil
}
