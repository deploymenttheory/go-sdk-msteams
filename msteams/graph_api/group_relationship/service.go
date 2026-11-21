package grouprelationship

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("group reference returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

type relationship struct {
	scope       *client.Scope
	name        string
	defaultKind ReferenceKind
}

func (r *relationship) local() error {
	if r == nil || r.scope == nil || r.scope.Session() == nil {
		return client.ErrClosed
	}
	if r.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := r.scope.Session().Info()
	return err
}

func (r *relationship) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.local()
}

func (r *relationship) request(method, groupID string) (client.Request, error) {
	if !directoryID.MatchString(groupID) {
		return client.Request{}, client.ErrRequest
	}
	req := client.Request{Method: method, Path: []string{"v1.0", "groups", groupID, r.name}, Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		req.Replay = client.ReplaySafe
	}
	return req, nil
}

func (r *relationship) empty(ctx context.Context, req client.Request) (*resty.Response, error) {
	response, err := r.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != http.StatusNoContent {
		err = ErrUnexpectedStatus
	}
	return response, err
}

func (r *relationship) addReference(ctx context.Context, groupID string, value Reference) (*resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	req, err := r.request(http.MethodPost, groupID)
	if err != nil {
		return nil, err
	}
	if !directoryID.MatchString(string(value.ID)) {
		return nil, client.ErrRequest
	}
	kind := value.Kind
	if kind == "" {
		kind = r.defaultKind
	}
	switch kind {
	case DirectoryObjects, Users, ServicePrincipals:
	default:
		return nil, client.ErrRequest
	}
	env, err := r.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	req.Path = append(req.Path, "$ref")
	req.Body = struct {
		ID string `json:"@odata.id"`
	}{ID: env.Endpoints.Graph + "/v1.0/" + string(kind) + "/" + string(value.ID)}
	return r.empty(ctx, req)
}

func (r *relationship) removeReference(ctx context.Context, groupID string, id DirectoryObjectID) (*resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	req, err := r.request(http.MethodDelete, groupID)
	if err != nil {
		return nil, err
	}
	if !directoryID.MatchString(string(id)) {
		return nil, client.ErrRequest
	}
	// This suffix is part of the operation, never caller configuration. Omitting
	// it could delete the directory object rather than only the relationship.
	req.Path = append(req.Path, string(id), "$ref")
	return r.empty(ctx, req)
}

func (r *relationship) listRequest(groupID string, options ListOptions) (client.Request, error) {
	req, err := r.request(http.MethodGet, groupID)
	if err != nil {
		return client.Request{}, err
	}
	if options.Top < 0 || r.name == "members" && options.Top > 999 {
		return client.Request{}, client.ErrRequest
	}
	query := make(url.Values)
	for key, value := range map[string]string{"$filter": options.Filter, "$search": options.Search} {
		if value != "" {
			query.Set(key, value)
		}
	}
	for key, values := range map[string][]string{"$select": options.Select, "$expand": options.Expand} {
		for _, value := range values {
			if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n") {
				return client.Request{}, client.ErrRequest
			}
		}
		if len(values) != 0 {
			query.Set(key, strings.Join(values, ","))
		}
	}
	if options.Top != 0 {
		query.Set("$top", strconv.Itoa(options.Top))
	}
	if options.Count != nil {
		query.Set("$count", strconv.FormatBool(*options.Count))
	}
	req.Query = query
	if options.EventualConsistency {
		req.Header.Set("ConsistencyLevel", "eventual")
	}
	return req, nil
}

func decodePage(body []byte) (client.PageContent[DirectoryObject], error) {
	return client.DecodeCollection[DirectoryObject](body, "value", "@odata.nextLink")
}

func (r *relationship) list(ctx context.Context, groupID string, options ListOptions) (*ListPage, *resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := r.listRequest(groupID, options)
	if err != nil {
		return nil, nil, err
	}
	response, err := r.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, response, ErrUnexpectedStatus
	}
	content, err := decodePage(response.Body())
	if err != nil {
		return nil, response, err
	}
	page := &ListPage{Items: content.Items, NextLink: content.Next}
	if json.Unmarshal(response.Body(), &page.RawFields) != nil {
		return nil, response, client.ErrDecode
	}
	return page, response, nil
}

func (r *relationship) newPager(groupID string, options ListOptions, limits client.PagingOptions) (*client.Pager[DirectoryObject], error) {
	if err := r.local(); err != nil {
		return nil, err
	}
	req, err := r.listRequest(groupID, options)
	if err != nil {
		return nil, err
	}
	if limits.Style != 0 && limits.Style != client.LinkPaging || limits.Parameter != "" || limits.ExpectedStatus != 0 && limits.ExpectedStatus != 200 {
		return nil, client.ErrPagination
	}
	limits.Style, limits.ExpectedStatus = client.LinkPaging, http.StatusOK
	return client.NewPager(r.scope, req, decodePage, limits)
}

func (r *relationship) listAll(ctx context.Context, groupID string, options ListOptions, limits client.PagingOptions) (client.ListResult[DirectoryObject], error) {
	if err := r.ready(ctx); err != nil {
		return client.ListResult[DirectoryObject]{}, err
	}
	pager, err := r.newPager(groupID, options, limits)
	if err != nil {
		return client.ListResult[DirectoryObject]{}, err
	}
	return pager.Collect(ctx)
}
