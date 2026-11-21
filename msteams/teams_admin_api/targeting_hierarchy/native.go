package targetinghierarchy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var tenantID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// V1Service owns targeting v1 singleton import/status/delete primitives.
type V1Service struct{ scope *client.Scope }

// V2Service owns hierarchy v2 primitives. Tenant paths come only from the session.
type V2Service struct{ scope *client.Scope }

func NewV1Service(scope *client.Scope) *V1Service { return &V1Service{scope: scope} }
func NewV2Service(scope *client.Scope) *V2Service { return &V2Service{scope: scope} }

func ready(ctx context.Context, scope *client.Scope, backend client.Backend) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if scope == nil || scope.Session() == nil {
		return client.ErrClosed
	}
	if scope.Backend() != backend {
		return client.ErrDestination
	}
	_, err := scope.Session().Info()
	return err
}
func (s *V1Service) ready(ctx context.Context) error {
	if s == nil {
		return ready(ctx, nil, client.Targeting)
	}
	if err := ready(ctx, s.scope, client.Targeting); err != nil {
		return err
	}
	return s.scope.ValidateRequest(request(http.MethodGet, "api", "rest", "v1", "nodes", "upload", "status"))
}
func (s *V2Service) tenant(ctx context.Context) (string, error) {
	var scope *client.Scope
	if s != nil {
		scope = s.scope
	}
	if err := ready(ctx, scope, client.Hierarchy); err != nil {
		return "", err
	}
	info, err := scope.Session().Info()
	if err != nil {
		return "", err
	}
	if !tenantID.MatchString(info.TenantID) {
		return "", ErrTenantIdentity
	}
	return info.TenantID, nil
}
func request(method string, parts ...string) client.Request {
	r := client.Request{Method: method, Path: parts, Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		r.Replay = client.ReplaySafe
	}
	return r
}
func v2Path(tenant string, parts ...string) []string {
	return append([]string{"api", "v1", "hierarchies", tenant}, parts...)
}
func validOperationID(value string) bool {
	return value != "." && value != ".." && !strings.ContainsRune(value, '\\') && !strings.ContainsFunc(value, unicode.IsControl)
}

// GetOperation performs one native status read. Blank requestID selects latest;
// a supplied selector is encoded as a query value, never interpolated into a URL.
func (s *V1Service) GetOperation(ctx context.Context, requestID string) (*RetailOperation, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req := request(http.MethodGet, "api", "rest", "v1", "nodes", "upload", "status")
	if strings.TrimSpace(requestID) != "" {
		req.Query = url.Values{"requestId": {requestID}}
	}
	response, err := s.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	var value RetailOperation
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// Delete sends one native singleton delete. It does not pre-read, replay or
// manufacture an inverse operation for this tenant-wide mutation.
func (s *V1Service) Delete(ctx context.Context) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	return s.scope.Do(ctx, request(http.MethodDelete, "api", "rest", "v1", "nodes", "hierarchy"), nil)
}

// GetHierarchy returns status-bearing opaque response data. No complete hierarchy
// response schema is established by the pinned upload's existence probe.
func (s *V2Service) GetHierarchy(ctx context.Context) (*resty.Response, error) {
	tenant, err := s.tenant(ctx)
	if err != nil {
		return nil, err
	}
	return s.scope.Do(ctx, request(http.MethodGet, v2Path(tenant)...), nil)
}

// CreateHierarchy sends the evidenced singleton body with the bound tenant ID
// and Operational Hierarchy name. Response data remains opaque/status-bearing.
func (s *V2Service) CreateHierarchy(ctx context.Context) (*resty.Response, error) {
	tenant, err := s.tenant(ctx)
	if err != nil {
		return nil, err
	}
	req := request(http.MethodPost, "api", "v1", "hierarchies")
	req.Body = struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	}{tenant, "Operational Hierarchy"}
	return s.scope.Do(ctx, req, nil)
}

func (s *V2Service) GetOperation(ctx context.Context, operationID string) (*HierarchyOperation, *resty.Response, error) {
	tenant, err := s.tenant(ctx)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(operationID) == "" {
		operationID = "latest"
	}
	if !validOperationID(operationID) {
		return nil, nil, client.ErrRequest
	}
	response, err := s.scope.Do(ctx, request(http.MethodGet, v2Path(tenant, "operations", operationID)...), nil)
	if err != nil {
		return nil, response, err
	}
	var value HierarchyOperation
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

func (s *V2Service) Delete(ctx context.Context) (*resty.Response, error) {
	tenant, err := s.tenant(ctx)
	if err != nil {
		return nil, err
	}
	return s.scope.Do(ctx, request(http.MethodDelete, v2Path(tenant)...), nil)
}

func csvPart(input CSV) (client.MultipartPart, error) {
	if input.Body == nil {
		return client.MultipartPart{}, client.ErrRequest
	}
	body := reflect.ValueOf(input.Body)
	switch body.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if body.IsNil() {
			return client.MultipartPart{}, client.ErrRequest
		}
	}
	name := path.Base(strings.ReplaceAll(input.Filename, "\\", "/"))
	if input.Filename == "" || name == "." || name == ".." || name == "/" || strings.ContainsFunc(name, unicode.IsControl) || input.MaxBytes <= 0 {
		return client.MultipartPart{}, client.ErrRequest
	}
	return client.MultipartPart{Name: "csvFile", Filename: name, ContentType: "application/json", Body: input.Body}, nil
}

func rejectCSV(ctx context.Context, input CSV, reason error) error {
	// Reuse shared typed-nil/stream ownership handling without a valid method,
	// session, token or HTTP destination. Never expose a source Close error text.
	var scope *client.Scope
	_, err := scope.Upload(ctx, client.Request{}, input.Body, "application/json", input.MaxBytes, nil)
	if errors.Is(err, client.ErrStreamClose) {
		return errors.Join(reason, client.ErrStreamClose)
	}
	return reason
}

func decodeImport(response *resty.Response, version Version) (*ImportResult, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	result := &ImportResult{RawFields: fields}
	key := "requestId"
	if version == V2 {
		key = "id"
	}
	value, present := fields[key]
	if present && json.Unmarshal(value, &result.RequestID) != nil {
		return result, client.ErrDecode
	}
	if version == V2 && (!present || result.RequestID == nil || strings.TrimSpace(*result.RequestID) == "") {
		return result, client.ErrDecode
	}
	return result, nil
}

// ImportCSV streams one multipart request and always owns/closes Body, including
// validation failure. It never parses/converts CSV, creates a hierarchy or polls.
func (s *V1Service) ImportCSV(ctx context.Context, input CSV) (*ImportResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, rejectCSV(ctx, input, err)
	}
	part, err := csvPart(input)
	if err != nil {
		return nil, nil, rejectCSV(ctx, input, err)
	}
	response, err := s.scope.UploadMultipart(ctx, request(http.MethodPost, "api", "rest", "v1", "nodes", "upload"), []client.MultipartPart{part}, input.MaxBytes, nil)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeImport(response, V1)
	return value, response, err
}
func (s *V2Service) ImportCSV(ctx context.Context, input CSV) (*ImportResult, *resty.Response, error) {
	tenant, err := s.tenant(ctx)
	if err != nil {
		return nil, nil, rejectCSV(ctx, input, err)
	}
	part, err := csvPart(input)
	if err != nil {
		return nil, nil, rejectCSV(ctx, input, err)
	}
	response, err := s.scope.UploadMultipart(ctx, request(http.MethodPost, v2Path(tenant, "importCsv")...), []client.MultipartPart{part}, input.MaxBytes, nil)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeImport(response, V2)
	return value, response, err
}
