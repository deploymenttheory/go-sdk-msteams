package licensereport

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var planIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service binds documented Graph reads to one session. It does not provision
// subscriptions, acquire extra roles or change the caller's authentication mode.
type Service struct{ scope *client.Scope }

func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }
func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, e := s.scope.Session().Info()
	return e
}
func (s *Service) licenseMode(ctx context.Context) error {
	if e := s.ready(ctx); e != nil {
		return e
	}
	info, e := s.scope.Session().Info()
	if e != nil {
		return e
	}
	if info.Mode != auth.Delegated {
		return ErrLicenseMode
	}
	return nil
}
func readRequest(path ...string) client.Request {
	return client.Request{Method: http.MethodGet, Path: path, Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}
}
func paging(options ListOptions) client.PagingOptions {
	return client.PagingOptions{Style: client.LinkPaging, MaxPages: options.MaxPages, MaxItems: options.MaxItems, ExpectedStatus: http.StatusOK}
}

func collect[T any](ctx context.Context, scope *client.Scope, request client.Request, options ListOptions, guard func(string) error) (*client.ListResult[T], error) {
	var continuationError error
	decoder := func(raw []byte) (client.PageContent[T], error) {
		page, e := client.DecodeCollection[T](raw, "value", "@odata.nextLink")
		if e != nil {
			return page, e
		}
		if page.Next != "" && guard != nil {
			if e = guard(page.Next); e != nil {
				continuationError = e
				page.Next = ""
			}
		}
		return page, nil
	}
	pager, e := client.NewPager(scope, request, decoder, paging(options))
	if e != nil {
		return nil, e
	}
	result, e := pager.Collect(ctx)
	return &result, errors.Join(e, continuationError)
}

// ListActivity reads only the documented beta JSON report. Its data and source
// plan predicate are not subscription entitlement or effective access evidence.
// Reports.Read.All can support application or delegated report access; the later
// license-details stage has a different documented mode/permission requirement.
func (s *Service) ListActivity(ctx context.Context, period Period, options ListOptions) (*client.ListResult[Activity], error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	if !validPeriod(period) {
		return nil, ErrInput
	}
	request := readRequest("beta", "reports", "getTeamsUserActivityUserDetail(period='D"+strconv.Itoa(int(period))+"')")
	request.Query = url.Values{"$format": {"application/json"}, "$top": {"2000"}}
	guard, e := s.resourceContinuation("/beta/reports/getTeamsUserActivityUserDetail(period='D" + strconv.Itoa(int(period)) + "')")
	if e != nil {
		return nil, e
	}
	return collect[Activity](ctx, s.scope, request, options, guard)
}

// ListUsers reads the documented directory ID/UPN/display-name projection. IDs
// are retained separately from UPNs; no user lifecycle operation is exposed.
func (s *Service) ListUsers(ctx context.Context, options ListOptions) (*client.ListResult[DirectoryUser], error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	request := readRequest("v1.0", "users")
	request.Query = url.Values{"$top": {"999"}, "$select": {"id,displayName,userPrincipalName"}}
	guard, e := s.resourceContinuation("/v1.0/users")
	if e != nil {
		return nil, e
	}
	return collect[DirectoryUser](ctx, s.scope, request, options, guard)
}

// ListLicenseDetailsByUserID uses only the documented directory-ID path. It
// never treats a UPN/guest alias as an interchangeable URL identifier and never
// falls back from application authentication to a delegated identity.
func (s *Service) ListLicenseDetailsByUserID(ctx context.Context, id string, options ListOptions) (*client.ListResult[LicenseDetail], error) {
	if e := s.licenseMode(ctx); e != nil {
		return nil, e
	}
	if !validUserID(id) {
		return nil, ErrInput
	}
	guard, e := s.licenseContinuation(id)
	if e != nil {
		return nil, e
	}
	return collect[LicenseDetail](ctx, s.scope, readRequest("v1.0", "users", id, "licenseDetails"), options, guard)
}

func (s *Service) licenseContinuation(id string) (func(string) error, error) {
	environment, e := s.scope.Session().Environment()
	if e != nil {
		return nil, e
	}
	origin, e := url.Parse(environment.Endpoints.Graph)
	if e != nil {
		return nil, client.ErrDestination
	}
	return func(next string) error {
		value, e := url.Parse(next)
		if e != nil || value.Scheme != "https" || !strings.EqualFold(value.Host, origin.Host) || value.User != nil || value.Fragment != "" || value.Opaque != "" {
			return client.ErrDestination
		}
		// Keep the same user and license resource as well as the same cloud. Query
		// tokens remain opaque; the shared pager supplies cycle/page/item guards.
		parts := strings.Split(strings.TrimPrefix(value.EscapedPath(), "/"), "/")
		subject := ""
		if len(parts) == 4 {
			subject, e = url.PathUnescape(parts[2])
		}
		if e != nil || len(parts) != 4 || parts[0] != "v1.0" || parts[1] != "users" || subject != id || parts[3] != "licenseDetails" {
			return client.ErrDestination
		}
		return nil
	}, nil
}

func (s *Service) continueLicenses(ctx context.Context, id, next string, options ListOptions) (*client.ListResult[LicenseDetail], error) {
	guard, e := s.licenseContinuation(id)
	if e != nil {
		return nil, e
	}
	if e = guard(next); e != nil {
		return nil, e
	}
	request := readRequest()
	request.Path = nil
	request.URL = next
	return collect[LicenseDetail](ctx, s.scope, request, options, guard)
}

// An explicit response is retained by primitives through ListResult.Response;
// this helper is used only to preserve outer batch receipts in composed results.
func responseStatus(response *resty.Response) int {
	if response == nil {
		return 0
	}
	return response.StatusCode()
}

func (s *Service) resourceContinuation(path string) (func(string) error, error) {
	environment, e := s.scope.Session().Environment()
	if e != nil {
		return nil, e
	}
	origin, e := url.Parse(environment.Endpoints.Graph)
	if e != nil {
		return nil, client.ErrDestination
	}
	return func(next string) error {
		value, e := url.Parse(next)
		if e != nil || value.Scheme != "https" || !strings.EqualFold(value.Host, origin.Host) || value.User != nil || value.Fragment != "" || value.Opaque != "" || value.Path != path {
			return client.ErrDestination
		}
		return nil
	}, nil
}

// IDs are opaque Graph strings. These checks match the transport's segment
// constraints, not a GUID schema or a UPN-to-ID conversion.
func validUserID(id string) bool {
	return id != "" && id != "." && id != ".." && !strings.ContainsRune(id, '\\') && !strings.ContainsFunc(id, unicode.IsControl)
}
