package policypackages

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/capabilities"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

// Service binds package operations to one identity-bound session. The documented
// application-mode restriction is checked before discovery or credentials. Unknown
// documentation profiles remain unqualified rather than inheriting a support claim.
type Service struct {
	scope   *client.Scope
	backend *configapi.Client
	catalog *capabilities.Catalog
	initErr error
}

func NewService(scope *client.Scope) *Service {
	catalog, err := capabilities.NewCatalog()
	return &Service{scope: scope, backend: configapi.New(scope), catalog: catalog, initErr: err}
}
func (s *Service) ready(ctx context.Context) error {
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
	if s.initErr != nil {
		return client.ErrRequest
	}
	return nil
}

func (s *Service) request(ctx context.Context, command string, request client.Request, statuses ...int) (*resty.Response, bool, error) {
	if err := s.ready(ctx); err != nil {
		return nil, false, err
	}
	info, err := s.scope.Session().Info()
	if err != nil {
		return nil, false, err
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, false, err
	}
	report, err := s.catalog.Assess(command, info.Mode, environment.Name)
	if err != nil && !errors.Is(err, capabilities.ErrUnknownProfile) {
		return nil, false, client.ErrRequest
	}
	if err == nil && report.Support == capabilities.ModeDocumentedUnsupported {
		return nil, false, client.ErrAuthenticationMode
	}
	parts := make([]string, len(request.Path))
	for i, value := range request.Path {
		parts[i] = strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
	}
	request.URL = environment.Endpoints.TeamsAdmin + "/" + strings.Join(parts, "/")
	request.Path = nil
	request.Header = request.Header.Clone()
	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-MS-CmdletName", command)
	// Validate dynamic user headers and the complete route before discovery can
	// acquire credentials. Business-error decoding must exclude discovery replies.
	if err := s.scope.ValidateRequest(request); err != nil {
		return nil, false, err
	}
	r, err := s.backend.Do(ctx, request, nil)
	operation := r != nil && r.Request != nil && r.Request.Method == request.Method && r.Request.URL == request.URL
	if err != nil {
		return r, operation, err
	}
	if !operation {
		return r, false, ErrResponse
	}
	for _, status := range statuses {
		if r.StatusCode() == status {
			return r, true, nil
		}
	}
	return r, true, ErrResponse
}

func packagePath(parts ...string) []string {
	return append([]string{"Teams.VerticalPackaging", "package"}, parts...)
}
func object[T any](r *resty.Response) (*T, error) {
	var value *T
	if r == nil || json.Unmarshal(r.Body(), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

func (s *Service) List(ctx context.Context) ([]*Package, *resty.Response, error) {
	r, _, err := s.request(ctx, "Get-CsPolicyPackage", client.Request{Method: http.MethodGet, Path: packagePath(), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	var values []*Package
	if json.Unmarshal(r.Body(), &values) != nil || values == nil {
		return nil, r, client.ErrDecode
	}
	return values, r, nil
}
func (s *Service) GetByName(ctx context.Context, name string) (*Package, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(name) {
		return nil, nil, ErrInput
	}
	r, _, err := s.request(ctx, "Get-CsPolicyPackage", client.Request{Method: http.MethodGet, Path: packagePath(name), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	v, err := object[Package](r)
	return v, r, err
}
func (s *Service) GetByIdentity(ctx context.Context, input Identity) (*Package, *resty.Response, error) {
	return s.GetByName(ctx, input.PackageName)
}

func (s *Service) userRead(ctx context.Context, identity string, recommendations bool) (*Package, []*Package, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, nil, err
	}
	if strings.TrimSpace(identity) == "" || strings.ContainsFunc(identity, unicode.IsControl) {
		return nil, nil, nil, ErrInput
	}
	command, path := "Get-CsUserPolicyPackage", packagePath("assigned")
	if recommendations {
		command, path = "Get-CsUserPolicyPackageRecommendation", packagePath()
	}
	r, _, err := s.request(ctx, command, client.Request{Method: http.MethodGet, Path: path, Header: http.Header{"user": {identity}}, Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, nil, r, err
	}
	if recommendations {
		var values []*Package
		if json.Unmarshal(r.Body(), &values) != nil || values == nil {
			return nil, nil, r, client.ErrDecode
		}
		return nil, values, r, nil
	}
	v, err := object[Package](r)
	return v, nil, r, err
}
func (s *Service) GetForUser(ctx context.Context, identity string) (*Package, *resty.Response, error) {
	v, _, r, err := s.userRead(ctx, identity, false)
	return v, r, err
}
func (s *Service) RecommendationsForUser(ctx context.Context, identity string) ([]*Package, *resty.Response, error) {
	_, v, r, err := s.userRead(ctx, identity, true)
	return v, r, err
}

func (s *Service) customWrite(ctx context.Context, method, command string, input CustomPackageRequest, status int) (*MessageResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.Name == nil || *input.Name == "" || input.Policies == nil {
		return nil, nil, ErrInput
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, nil, ErrInput
	}
	r, sent, err := s.request(ctx, command, client.Request{Method: method, Path: packagePath("custompackage"), Body: json.RawMessage(body), Replay: client.ReplayNever}, status)
	if !sent {
		return nil, r, err
	}
	result, decodeErr := message(r, err == nil)
	if err != nil {
		return result, r, err
	}
	return result, r, decodeErr
}
func (s *Service) Create(ctx context.Context, input CustomPackageRequest) (*MessageResult, *resty.Response, error) {
	return s.customWrite(ctx, http.MethodPost, "New-CsCustomPolicyPackage", input, http.StatusCreated)
}

// Update uses collection PUT with PackageName in the body, preserving complete
// list/empty/omission semantics instead of fabricating a sparse PATCH endpoint.
func (s *Service) Update(ctx context.Context, input CustomPackageRequest) (*MessageResult, *resty.Response, error) {
	return s.customWrite(ctx, http.MethodPut, "Update-CsCustomPolicyPackage", input, http.StatusOK)
}
func (s *Service) DeleteByName(ctx context.Context, name string) (*MessageResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !policyrp.ValidIdentity(name) {
		return nil, nil, ErrInput
	}
	r, sent, err := s.request(ctx, "Remove-CsCustomPolicyPackage", client.Request{Method: http.MethodDelete, Path: packagePath("custompackage", name), Replay: client.ReplayNever}, http.StatusNoContent)
	if !sent {
		return nil, r, err
	}
	result, decodeErr := message(r, err == nil)
	if err != nil {
		return result, r, err
	}
	return result, r, decodeErr
}
func (s *Service) DeleteByIdentity(ctx context.Context, input Identity) (*MessageResult, *resty.Response, error) {
	return s.DeleteByName(ctx, input.PackageName)
}

func message(r *resty.Response, accepted bool) (*MessageResult, error) {
	result := &MessageResult{HTTPAccepted: accepted}
	if accepted && r.StatusCode() == http.StatusNoContent {
		return result, nil
	}
	err := json.Unmarshal(r.Body(), result)
	result.HTTPAccepted = accepted
	if err != nil {
		return result, client.ErrDecode
	}
	return result, nil
}

func (s *Service) assignment(ctx context.Context, command string, path []string, input any, status int) (*AssignmentResult, *resty.Response, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, nil, ErrInput
	}
	r, sent, err := s.request(ctx, command, client.Request{Method: http.MethodPost, Path: path, Body: json.RawMessage(body), Replay: client.ReplayNever}, status)
	if !sent {
		return nil, r, err
	}
	result := &AssignmentResult{}
	decodeErr := json.Unmarshal(r.Body(), result)
	result.HTTPAccepted = err == nil
	if err != nil {
		return result, r, err
	}
	if decodeErr != nil {
		return result, r, client.ErrDecode
	}
	return result, r, nil
}
func (s *Service) AssignGroup(ctx context.Context, input GroupAssignmentRequest) (*AssignmentResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.GroupID == "" {
		return nil, nil, ErrInput
	}
	// The exported custom wrapper always binds an empty ranking list when no
	// rankings are provided. Preserve that default without dropping rank zero.
	if input.Rankings == nil {
		input.Rankings = []PolicyRanking{}
	}
	return s.assignment(ctx, "Grant-CsGroupPolicyPackageAssignment", packagePath("group"), input, http.StatusOK)
}

// AssignUsers invokes the synchronous route and retains per-user/policy failures
// even on a successful HTTP status. It does not infer success from message text.
func (s *Service) AssignUsers(ctx context.Context, input UserAssignmentRequest) (*AssignmentResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.UserIDs == nil {
		return nil, nil, ErrInput
	}
	return s.assignment(ctx, "Grant-CsUserPolicyPackage", packagePath("batch"), input, http.StatusCreated)
}
func (s *Service) SubmitBatch(ctx context.Context, input UserAssignmentRequest) (*AssignmentResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.UserIDs == nil {
		return nil, nil, ErrInput
	}
	return s.assignment(ctx, "New-CsBatchPolicyPackageAssignmentOperation", packagePath("operation", "batchpackage"), input, http.StatusAccepted)
}
