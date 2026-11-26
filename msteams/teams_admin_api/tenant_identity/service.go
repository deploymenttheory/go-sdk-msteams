// Package tenantidentity reads tenant and directory identity data exposed by the
// Teams admin backend: the extended tenant record (Get-CsTenant), voice
// application group search (Find-CsGroup), hosting provider and tenant
// licensing configurations, and the support-only directory reads
// (Get-CsAadTenant/User, Get-CsCloudTenant/User). It also carries the
// licensing-impacting Remove-CsUserLicenseGracePeriod request.
package tenantidentity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/go-resty/resty/v2"
)

const (
	hostingProvider      = "HostingProvider"
	tenantLicensing      = "TenantLicensingConfiguration"
	hostingProviderCmd   = "Get-CsHostingProvider"
	tenantLicensingCmd   = "Get-CsTenantLicensingConfiguration"
	gracePeriodRemoveCmd = "Remove-CsUserLicenseGracePeriod"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func validText(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsFunc(value, unicode.IsControl)
}

func validSegment(value string) bool {
	return validText(value) && value != "." && value != ".." && !strings.ContainsRune(value, '\\')
}

func (s *Service) do(ctx context.Context, command string, request client.Request, status int, result any) (*resty.Response, error) {
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	response, err := s.client.Do(ctx, request, result)
	if err == nil && (response == nil || response.StatusCode() != status) {
		err = ErrResponse
	}
	return response, err
}

func (s *Service) getObject(ctx context.Context, command string, path []string, query url.Values) (*DirectoryObject, *resty.Response, error) {
	var value DirectoryObject
	response, err := s.do(ctx, command, client.Request{Method: http.MethodGet, Path: path, Query: query, Replay: client.ReplaySafe}, http.StatusOK, &value)
	if err != nil {
		return nil, response, err
	}
	if value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// SearchGroups is Find-CsGroup: GET Teams.VoiceApps/groups. The response
// Diagnostics array is not returned (the module writes it to its diagnostic
// stream). A missing/null/wrong-type Groups property is an empty result; a
// null generated root returns nil with the response. IDs follow the source
// public GUID projection while original properties remain in RawFields.
func (s *Service) SearchGroups(ctx context.Context, request GroupSearchRequest) ([]*Group, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if request.SearchQuery == "" || request.MaxResults != nil && *request.MaxResults > 1<<31-1 {
		return nil, nil, ErrInput
	}
	query := url.Values{"SearchQuery": {request.SearchQuery}}
	if request.MaxResults != nil {
		query.Set("MaxResults", strconv.FormatUint(uint64(*request.MaxResults), 10))
	}
	if request.ExactMatchOnly != nil {
		query.Set("ExactMatchOnly", dotnetBool(*request.ExactMatchOnly))
	}
	if request.MailEnabledOnly != nil {
		query.Set("MailEnabledOnly", dotnetBool(*request.MailEnabledOnly))
	}
	response, err := s.do(ctx, "Find-CsGroup_Find", client.Request{Method: http.MethodGet, Path: []string{"Teams.VoiceApps", "groups"}, Query: query, Replay: client.ReplaySafe}, http.StatusOK, nil)
	if err != nil {
		return nil, response, err
	}
	fields, err := groupNode(response.Body())
	if err != nil || fields == nil {
		return nil, response, err
	}
	var envelope struct {
		Groups []*Group `json:"Groups"`
	}
	if raw := bytes.TrimSpace(fields["Groups"]); len(raw) > 0 && raw[0] == '[' {
		if json.Unmarshal(raw, &envelope.Groups) != nil {
			return nil, response, client.ErrDecode
		}
	}
	groups := make([]*Group, 0, len(envelope.Groups))
	for _, group := range envelope.Groups {
		if group == nil {
			return nil, response, client.ErrDecode
		}
		groups = append(groups, group)
	}
	return groups, response, nil
}

func dotnetBool(value bool) string {
	if value {
		return "True"
	}
	return "False"
}

// GetTenant reads the source-default Modern Get-CsTenant route with the Extended
// property set. It returns the native wire view. Public CLR projection and
// legacy/flight-selected transport remain separate qualification boundaries.
// Filter, Identity and ResultSize are not parameters of the Modern helper.
func (s *Service) GetTenant(ctx context.Context) (*Tenant, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	var value Tenant
	response, err := s.do(ctx, "Get-CsTenantObou_Get", client.Request{Method: http.MethodGet, Path: []string{"Teams.Tenant", "tenants"}, Query: url.Values{"defaultpropertyset": {"Extended"}}, Replay: client.ReplaySafe}, http.StatusOK, &value)
	if err != nil {
		return nil, response, err
	}
	if value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// GetOnlineDirectoryTenant is Get-CsOnlineDirectoryTenant. The pinned module
// dispatches it to Invoke-CsDeprecatedError, so it always fails locally with
// ErrOnlineDirectoryTenantDeprecated and performs no I/O.
func (s *Service) GetOnlineDirectoryTenant(ctx context.Context) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	return ErrOnlineDirectoryTenantDeprecated
}

// GetAadTenant is Get-CsAadTenant (support-only): GET Teams.Tenant/AadTenant.
func (s *Service) GetAadTenant(ctx context.Context) (*DirectoryObject, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.getObject(ctx, "Get-CsAadTenant", []string{"Teams.Tenant", "AadTenant"}, nil)
}

// GetAadUser is Get-CsAadUser (support-only): GET Teams.User/AadUser/{identity}.
func (s *Service) GetAadUser(ctx context.Context, identity string) (*DirectoryObject, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validSegment(identity) {
		return nil, nil, ErrInput
	}
	return s.getObject(ctx, "Get-CsAadUser", []string{"Teams.User", "AadUser", identity}, nil)
}

// GetCloudTenant is Get-CsCloudTenant (support-only): GET
// Teams.InternalSupport/GetAADTenantData.
func (s *Service) GetCloudTenant(ctx context.Context) (*DirectoryObject, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.getObject(ctx, "Get-CsCloudTenant", []string{"Teams.InternalSupport", "GetAADTenantData"}, nil)
}

// GetCloudUser is Get-CsCloudUser (support-only): GET
// Teams.InternalSupport/GetAADUserData?userId=.
func (s *Service) GetCloudUser(ctx context.Context, userID string) (*DirectoryObject, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(userID) {
		return nil, nil, ErrInput
	}
	return s.getObject(ctx, "Get-CsCloudUser", []string{"Teams.InternalSupport", "GetAADUserData"}, url.Values{"userId": {userID}})
}

func configurationPath(kind, identity string) []string {
	p := []string{"Skype.Policy", "configurations", kind}
	if identity != "" {
		p = append(p, "configuration", identity)
	}
	return p
}

func (s *Service) getConfiguration(ctx context.Context, kind, command, identity string) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validSegment(identity) {
		return nil, nil, ErrInput
	}
	var value Configuration
	response, err := s.do(ctx, command, client.Request{Method: http.MethodGet, Path: configurationPath(kind, identity), Replay: client.ReplaySafe}, http.StatusOK, &value)
	if err != nil {
		return nil, response, err
	}
	if value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

func isGlobal(value *Configuration) bool {
	return value.Identity != nil && strings.EqualFold(*value.Identity, "Global")
}

func (s *Service) listConfigurations(ctx context.Context, kind, command string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	var raw json.RawMessage
	response, err := s.do(ctx, command, client.Request{Method: http.MethodGet, Path: configurationPath(kind, ""), Replay: client.ReplaySafe}, http.StatusOK, &raw)
	if err != nil {
		return nil, response, err
	}
	var values []*Configuration
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	slices.SortStableFunc(values, func(a, b *Configuration) int {
		left, right := isGlobal(a), isGlobal(b)
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	return values, response, nil
}

// listFiltered applies the module's local Identity or Tag:Identity wildcard
// match to the full list. Blank input selects the full list.
func (s *Service) listFiltered(ctx context.Context, kind, command, pattern string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(pattern) == "" {
		return s.listConfigurations(ctx, kind, command)
	}
	plain, err := policyfilter.Compile(pattern)
	if err != nil {
		return nil, nil, ErrInput
	}
	tagged, err := policyfilter.Compile("Tag:" + pattern)
	if err != nil {
		return nil, nil, ErrInput
	}
	items, response, err := s.listConfigurations(ctx, kind, command)
	if err != nil {
		return nil, response, err
	}
	filtered := make([]*Configuration, 0)
	for _, item := range items {
		identity := ""
		if item.Identity != nil {
			identity = *item.Identity
		}
		match, err := plain.Match(ctx, identity)
		if err == nil && !match {
			match, err = tagged.Match(ctx, identity)
		}
		if err != nil {
			return nil, response, err
		}
		if match {
			filtered = append(filtered, item)
		}
	}
	return filtered, response, nil
}

// ListHostingProviders is Get-CsHostingProvider without parameters, ordered
// Global first. The LocalStore switch has no service behaviour.
func (s *Service) ListHostingProviders(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	return s.listConfigurations(ctx, hostingProvider, hostingProviderCmd)
}

// ListHostingProvidersFiltered is Get-CsHostingProvider -Filter.
func (s *Service) ListHostingProvidersFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	return s.listFiltered(ctx, hostingProvider, hostingProviderCmd, pattern)
}

// GetHostingProviderByIdentity is Get-CsHostingProvider -Identity.
func (s *Service) GetHostingProviderByIdentity(ctx context.Context, identity string) (*Configuration, *resty.Response, error) {
	return s.getConfiguration(ctx, hostingProvider, hostingProviderCmd, identity)
}

// ListTenantLicensingConfigurations is Get-CsTenantLicensingConfiguration
// without parameters, ordered Global first.
func (s *Service) ListTenantLicensingConfigurations(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	return s.listConfigurations(ctx, tenantLicensing, tenantLicensingCmd)
}

// ListTenantLicensingConfigurationsFiltered is
// Get-CsTenantLicensingConfiguration -Filter.
func (s *Service) ListTenantLicensingConfigurationsFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	return s.listFiltered(ctx, tenantLicensing, tenantLicensingCmd, pattern)
}

// GetTenantLicensingConfigurationByIdentity is
// Get-CsTenantLicensingConfiguration -Identity.
func (s *Service) GetTenantLicensingConfigurationByIdentity(ctx context.Context, identity string) (*Configuration, *resty.Response, error) {
	return s.getConfiguration(ctx, tenantLicensing, tenantLicensingCmd, identity)
}

// RemoveUserLicenseGracePeriod is Remove-CsUserLicenseGracePeriod: PATCH
// Teams.User/users/{identity}/AssignedPlans, requiring HTTP 204. It ends a
// user's license grace period immediately and is never replayed. This changes
// licensing state; the SDK's acceptance suite does not call it.
func (s *Service) RemoveUserLicenseGracePeriod(ctx context.Context, identity string, request GracePeriodRemoval) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validSegment(identity) || (request.Action != nil && strings.ContainsFunc(*request.Action, unicode.IsControl)) {
		return nil, ErrInput
	}
	body := map[string]any{}
	if request.Capabilities != nil {
		body["capabilities"] = slices.Clone(request.Capabilities)
	}
	if request.Action != nil {
		body["action"] = *request.Action
	}
	return s.do(ctx, gracePeriodRemoveCmd, client.Request{Method: http.MethodPatch, Path: []string{"Teams.User", "users", identity, "AssignedPlans"}, Body: body, Replay: client.ReplayNever}, http.StatusNoContent, nil)
}
