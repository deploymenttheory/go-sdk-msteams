package group

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// APIVersion selects the Graph collection contract, not a PowerShell version.
type APIVersion string

const (
	V1   APIVersion = "v1.0"
	Beta APIVersion = "beta"
)

// ListOptions maps the documented groups query. Version defaults to v1.0; beta
// is explicit for the pinned tenant-search composition. Query expressions remain
// caller-supplied OData; the SDK does not parse them or guess advanced-query use.
type ListOptions struct {
	Version             APIVersion
	Filter              string
	Search              string
	Select              []string
	Expand              []string
	OrderBy             []string
	Top                 int
	Count               *bool
	EventualConsistency bool
}

// ListPage is one native page with opaque continuation and full raw metadata.
type ListPage struct {
	Items     []Group
	NextLink  string
	RawFields map[string]json.RawMessage
}

func selectQuery(fields []string) (url.Values, error) {
	query := make(url.Values)
	if len(fields) == 0 {
		return query, nil
	}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" || strings.ContainsAny(field, ",\r\n") {
			return nil, client.ErrRequest
		}
	}
	query.Set("$select", strings.Join(fields, ","))
	return query, nil
}

func listRequest(options ListOptions) (client.Request, error) {
	version := options.Version
	if version == "" {
		version = V1
	}
	if version != V1 && version != Beta || options.Top < 0 || options.Top > 999 {
		return client.Request{}, client.ErrRequest
	}
	query, err := selectQuery(options.Select)
	if err != nil {
		return client.Request{}, err
	}
	for key, value := range map[string]string{"$filter": options.Filter, "$search": options.Search} {
		if value != "" {
			query.Set(key, value)
		}
	}
	for key, values := range map[string][]string{"$expand": options.Expand, "$orderby": options.OrderBy} {
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
	req := request(http.MethodGet)
	req.Path[0] = string(version)
	req.Query = query
	if options.EventualConsistency {
		req.Header.Set("ConsistencyLevel", "eventual")
	}
	return req, nil
}

func decodePage(body []byte) (client.PageContent[Group], error) {
	return client.DecodeCollection[Group](body, "value", "@odata.nextLink")
}

// List returns one page. Defaults, null properties and eventual consistency are
// not evidence of absence; use explicit traversal for a complete collection.
func (s *Service) List(ctx context.Context, options ListOptions) (*ListPage, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := listRequest(options)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.scope.Do(ctx, req, nil)
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

// NewPager constructs a lazy same-origin traversal. Only MaxPages/MaxItems apply;
// Style must be zero/LinkPaging, Parameter empty and ExpectedStatus zero/200.
// HTTP 200 is required for every page. MaxItems limits Collect aggregation.
func (s *Service) NewPager(options ListOptions, limits client.PagingOptions) (*client.Pager[Group], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	req, err := listRequest(options)
	if err != nil {
		return nil, err
	}
	if limits.Style != 0 && limits.Style != client.LinkPaging || limits.Parameter != "" || limits.ExpectedStatus != 0 && limits.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	limits.Style, limits.ExpectedStatus = client.LinkPaging, http.StatusOK
	return client.NewPager(s.scope, req, decodePage, limits)
}

// ListAll explicitly collects pages within shared bounds. Its partial prefix and
// last response survive errors; a non-nil error always means incomplete results.
func (s *Service) ListAll(ctx context.Context, options ListOptions, limits client.PagingOptions) (client.ListResult[Group], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Group]{}, err
	}
	pager, err := s.NewPager(options, limits)
	if err != nil {
		return client.ListResult[Group]{}, err
	}
	return pager.Collect(ctx)
}
