// Package configurationdiagnostics reads and writes configuration and policy
// instances of any type through the generic configuration route, and sends
// the support commands that clear service caches and publish policy schema
// defaults.
//
// The generic route is the one behind Get/New/Set/Remove-CsConfiguration:
// /Skype.Policy/configurations/{configType}[/configuration/{name}]. The module
// exports these commands only in support (OCE) sessions; the same calls work
// with tenant administrator credentials. Typed per-configuration services in
// this module remain the preferred way to manage a known configuration type.
package configurationdiagnostics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policytransport"
	"github.com/go-resty/resty/v2"
)

// Service composes the bound ConfigAPI backend (with lazy forest discovery)
// for the generated commands and the direct policy protocol for
// Set-CsPublishPolicySchemaDefaults.
type Service struct {
	client client.Client
	policy *policytransport.Client
}

// NewService performs no I/O. Pass the backend and scope bound to ConfigAPI's
// audience.
func NewService(base client.Client, scope *client.Scope) *Service {
	return &Service{client: configapi.New(base), policy: policytransport.New(scope)}
}

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil || s.policy == nil {
		return client.ErrClosed
	}
	return nil
}

func validPart(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." &&
		!strings.ContainsAny(value, `/\`) && !strings.ContainsFunc(value, unicode.IsControl)
}

func validOptional(value *string) bool {
	return value == nil || (*value != "" && !strings.ContainsFunc(*value, unicode.IsControl))
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func versions(api, schema *string) (url.Values, bool) {
	if !validOptional(api) || !validOptional(schema) {
		return nil, false
	}
	query := url.Values{}
	if api != nil {
		query.Set("apiVersion", *api)
	}
	if schema != nil {
		query.Set("schemaVersion", *schema)
	}
	return query, true
}

func configurationPath(configType, name string) []string {
	path := []string{"Skype.Policy", "configurations", configType}
	if name != "" {
		path = append(path, "configuration", name)
	}
	return path
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// ListConfigurations returns every instance of a configuration type
// (Get-CsConfiguration -ConfigType). The order is the service's; the module's
// wrapper moves Global first, which callers can do on Identity.
func (s *Service) ListConfigurations(ctx context.Context, configType string, options *ReadOptions) ([]*Configuration, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if options == nil {
		options = &ReadOptions{}
	}
	query, ok := versions(options.APIVersion, options.SchemaVersion)
	if !validPart(configType) || !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: configurationPath(configType, ""), Query: query,
		Header: headers("Get-CsConfiguration"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values := []*Configuration{}
	if json.Unmarshal(response.Body(), &values) != nil || values == nil || slices.Contains(values, nil) {
		return nil, response, client.ErrDecode
	}
	return values, response, nil
}

// GetConfiguration reads one instance by name, such as "Global" or
// "Tag:Name" (Get-CsConfiguration -ConfigType -ConfigName). A missing instance
// is an *client.APIError with status 404.
func (s *Service) GetConfiguration(ctx context.Context, configType, name string, options *ReadOptions) (*Configuration, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if options == nil {
		options = &ReadOptions{}
	}
	query, ok := versions(options.APIVersion, options.SchemaVersion)
	if !validPart(configType) || !validPart(name) || !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: configurationPath(configType, name), Query: query,
		Header: headers("Get-CsConfiguration"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value *Configuration
	if json.Unmarshal(response.Body(), &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

func (s *Service) write(ctx context.Context, method, command, configType, name string, input ConfigurationRequest, options *WriteOptions, status int) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if options == nil {
		options = &WriteOptions{}
	}
	query, ok := versions(options.APIVersion, options.SchemaVersion)
	_, duplicate := input.Settings["Identity"]
	named := method == http.MethodPut
	if !validPart(configType) || !validPart(input.Identity) || duplicate || !ok || named != (name != "") || (named && !validPart(name)) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: method, Path: configurationPath(configType, name), Query: query,
		Header: headers(command), Body: input}, status)
}

// CreateConfiguration creates a new instance of a configuration type
// (New-CsConfiguration). The service answers 201 with no body; read the
// instance back with GetConfiguration. It is never replayed automatically.
func (s *Service) CreateConfiguration(ctx context.Context, configType string, input ConfigurationRequest, options *WriteOptions) (*resty.Response, error) {
	return s.write(ctx, http.MethodPost, "New-CsConfiguration", configType, "", input, options, http.StatusCreated)
}

// SetConfiguration replaces the named instance with the request body
// (Set-CsConfiguration, an HTTP PUT). Properties the body omits are reset to
// the type's defaults, so send the complete instance. The module's wrapper
// names the instance by input.Identity; name is the path instance and is
// normally the same value.
func (s *Service) SetConfiguration(ctx context.Context, configType, name string, input ConfigurationRequest, options *WriteOptions) (*resty.Response, error) {
	return s.write(ctx, http.MethodPut, "Set-CsConfiguration", configType, name, input, options, http.StatusNoContent)
}

// RemoveConfiguration deletes the named instance (Remove-CsConfiguration).
// It is never replayed automatically.
func (s *Service) RemoveConfiguration(ctx context.Context, configType, name string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(configType) || !validPart(name) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: configurationPath(configType, name),
		Header: headers("Remove-CsConfiguration")}, http.StatusNoContent)
}

// ClearCache removes keys from the service's caches (Clear-CsCacheOperation).
// The service answers 200 when every key was removed and 206 with the keys it
// could not remove, per region. This is a support operation that affects
// service-wide caches; it is never replayed automatically.
func (s *Service) ClearCache(ctx context.Context, input ClearCacheRequest) ([]*CacheClearResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validOptional(input.Region) {
		return nil, nil, ErrInput
	}
	for _, key := range input.KeysToDelete {
		if key == "" || strings.ContainsFunc(key, unicode.IsControl) {
			return nil, nil, ErrInput
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{"Teams.InternalSupport", "CacheClear"},
		Header: headers("Clear-CsCacheOperation"), Body: input}, http.StatusOK, http.StatusPartialContent)
	if err != nil {
		return nil, response, err
	}
	values := []*CacheClearResult{}
	if json.Unmarshal(response.Body(), &values) != nil || values == nil || slices.Contains(values, nil) {
		return nil, response, client.ErrDecode
	}
	return values, response, nil
}

// PublishPolicySchemaDefaults asks the policy service to publish the schema
// defaults of one policy type (Set-CsPublishPolicySchemaDefaults). The
// request has no body and goes directly to the ConfigAPI endpoint, without
// forest discovery. It returns the response text, which the cmdlet writes
// unchanged. Any 2xx status succeeds; it is never replayed automatically.
func (s *Service) PublishPolicySchemaDefaults(ctx context.Context, policyType string) (string, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return "", nil, err
	}
	if !validPart(policyType) {
		return "", nil, ErrInput
	}
	response, err := s.policy.Do(ctx, "Set-CsPublishPolicySchemaDefaults", http.MethodPost,
		[]string{"Skype.Policy", "schemadefaults", policyType}, nil, nil, client.ReplayNever)
	if err != nil {
		return "", response, err
	}
	if response == nil || response.StatusCode() < 200 || response.StatusCode() > 299 {
		return "", response, ErrResponse
	}
	return string(response.Body()), response, nil
}
