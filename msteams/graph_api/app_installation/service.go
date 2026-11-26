package appinstallation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Service owns no credentials and uses only the bound Graph backend.
type Service struct{ scope *client.Scope }

// NewService constructs a lazy resource without network activity.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func validID(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." && !strings.ContainsAny(value, `/\`) && !strings.ContainsFunc(value, unicode.IsControl)
}

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
func (target Target) path(extra ...string) ([]string, error) {
	var segments []string
	switch {
	case validID(target.TeamID) && target.UserID == "":
		segments = []string{"v1.0", "teams", target.TeamID, "installedApps"}
	case validID(target.UserID) && target.TeamID == "":
		segments = []string{"v1.0", "users", target.UserID, "teamwork", "installedApps"}
	default:
		return nil, ErrInput
	}
	return append(segments, extra...), nil
}
func headers() http.Header { return http.Header{"Accept": {"application/json"}} }
func decode(raw []byte) (client.PageContent[Installation], error) {
	return client.DecodeCollection[Installation](raw, "value", "@odata.nextLink")
}

// NewPager builds a lazy Get-TeamsAppInstallation collection traversal. CatalogID
// filters teamsAppDefinition/teamsAppId; quotes are doubled for OData.
func (s *Service) NewPager(target Target, options ListOptions) (*client.Pager[Installation], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	path, err := target.path()
	if err != nil {
		return nil, err
	}
	if options.CatalogID != "" && !validID(options.CatalogID) {
		return nil, ErrInput
	}
	bounds := options.Paging
	if bounds.Style != 0 && bounds.Style != client.LinkPaging || bounds.Parameter != "" || bounds.ExpectedStatus != 0 && bounds.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	bounds.Style, bounds.ExpectedStatus = client.LinkPaging, http.StatusOK
	query := url.Values{"$expand": {"teamsAppDefinition"}}
	if options.CatalogID != "" {
		query.Set("$filter", "teamsAppDefinition/teamsAppId eq '"+strings.ReplaceAll(options.CatalogID, "'", "''")+"'")
	}
	return client.NewPager(s.scope, client.Request{Method: http.MethodGet, Path: path, Query: query, Header: headers(), Replay: client.ReplaySafe}, decode, bounds)
}

// List collects every page, retaining partial items and the last response on
// failure. App IDs must never be substituted for installation IDs.
func (s *Service) List(ctx context.Context, target Target, options ListOptions) (client.ListResult[Installation], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Installation]{}, err
	}
	pager, err := s.NewPager(target, options)
	if err != nil {
		return client.ListResult[Installation]{}, err
	}
	return pager.Collect(ctx)
}

// GetByID reads one installation and its expanded app definition.
func (s *Service) GetByID(ctx context.Context, target Target, id string) (*Installation, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(id) {
		return nil, nil, ErrInput
	}
	path, err := target.path(id)
	if err != nil {
		return nil, nil, err
	}
	var raw json.RawMessage
	response, err := s.scope.Do(ctx, client.Request{Method: http.MethodGet, Path: path, Query: url.Values{"$expand": {"teamsAppDefinition"}}, Header: headers(), Replay: client.ReplaySafe}, &raw)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, response, ErrResponse
	}
	var value Installation
	if json.Unmarshal(raw, &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

func (s *Service) write(ctx context.Context, target Target, method string, extra []string, body any, statuses ...int) (*resty.Response, error) {
	path, err := target.path(extra...)
	if err != nil {
		return nil, err
	}
	response, err := s.scope.Do(ctx, client.Request{Method: method, Path: path, Header: headers(), Body: body, Replay: client.ReplayNever}, nil)
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

// Create adds an existing catalog app. A lost response remains ambiguous; the
// mutation is never automatically replayed. Reconcile with List by catalog ID.
func (s *Service) Create(ctx context.Context, target Target, input CreateRequest) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(input.CatalogID) {
		return nil, ErrInput
	}
	// Resolve the bound cloud's Graph origin instead of hardcoding commercial.
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	graph := environment.Endpoints.Graph
	body := struct {
		App     string         `json:"teamsApp@odata.bind"`
		Consent *PermissionSet `json:"consentedPermissionSet,omitempty"`
	}{App: strings.TrimRight(graph, "/") + "/v1.0/appCatalogs/teamsApps/" + url.PathEscape(input.CatalogID), Consent: input.Consent}
	if !validConsent(input.Consent) {
		return nil, ErrInput
	}
	statuses := []int{http.StatusCreated}
	if target.TeamID != "" {
		statuses = append(statuses, http.StatusOK)
	}
	return s.write(ctx, target, http.MethodPost, nil, body, statuses...)
}

// UpgradeByID upgrades one installation. Optional permissions use the
// native Graph v1.0 consent schema; no automatic replay occurs.
func (s *Service) UpgradeByID(ctx context.Context, target Target, id string, consent *PermissionSet) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	if !validConsent(consent) {
		return nil, ErrInput
	}
	return s.write(ctx, target, http.MethodPost, []string{id, "upgrade"}, struct {
		Consent *PermissionSet `json:"consentedPermissionSet,omitempty"`
	}{consent}, http.StatusNoContent)
}

// DeleteByID removes one installation, never the catalog app.
func (s *Service) DeleteByID(ctx context.Context, target Target, id string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	return s.write(ctx, target, http.MethodDelete, []string{id}, nil, http.StatusNoContent)
}

func (s *Service) resolve(ctx context.Context, target Target, catalogID string) (string, *resty.Response, bool, error) {
	if err := s.ready(ctx); err != nil {
		return "", nil, false, err
	}
	if !validID(catalogID) {
		return "", nil, false, ErrInput
	}
	result, err := s.List(ctx, target, ListOptions{CatalogID: catalogID})
	if err != nil {
		return "", result.Response, false, err
	}
	if len(result.Items) == 0 {
		return "", result.Response, true, ErrNotFound
	}
	if len(result.Items) != 1 {
		return "", result.Response, true, ErrAmbiguous
	}
	value := result.Items[0]
	if value.AppDefinition != nil && value.AppDefinition.TeamsAppID != nil && *value.AppDefinition.TeamsAppID != catalogID {
		return "", result.Response, true, ErrIdentityMismatch
	}
	return value.ID, result.Response, true, nil
}

// DeleteByCatalogID resolves exactly one installation before deletion. The
// result retains its ID when a cancellation or ambiguous write loses a response.
func (s *Service) DeleteByCatalogID(ctx context.Context, target Target, catalogID string) (*CatalogMutationResult, error) {
	return s.mutateByCatalogID(ctx, target, catalogID, nil, false)
}

// UpgradeByCatalogID resolves exactly one installation before upgrading. Partial
// or duplicate lookup results never trigger a write; action failures retain ID.
func (s *Service) UpgradeByCatalogID(ctx context.Context, target Target, catalogID string, consent *PermissionSet) (*CatalogMutationResult, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validConsent(consent) {
		return nil, ErrInput
	}
	return s.mutateByCatalogID(ctx, target, catalogID, consent, true)
}

func (s *Service) mutateByCatalogID(ctx context.Context, target Target, catalogID string, consent *PermissionSet, upgrade bool) (*CatalogMutationResult, error) {
	result := &CatalogMutationResult{CatalogID: catalogID}
	id, response, complete, err := s.resolve(ctx, target, catalogID)
	result.LookupResponse = response
	result.LookupComplete = complete
	if err != nil {
		return result, err
	}
	result.InstallationID = id
	if err := ctx.Err(); err != nil {
		return result, err
	}
	result.ActionInvoked = true
	if upgrade {
		result.ActionResponse, err = s.UpgradeByID(ctx, target, id, consent)
	} else {
		result.ActionResponse, err = s.DeleteByID(ctx, target, id)
	}
	result.HTTPAccepted = err == nil
	return result, err
}

func validConsent(consent *PermissionSet) bool {
	if consent == nil {
		return true
	}
	if consent.ResourceSpecificPermissions == nil {
		return false
	}
	for _, permission := range consent.ResourceSpecificPermissions {
		if strings.TrimSpace(permission.PermissionValue) == "" || strings.ContainsFunc(permission.PermissionValue, unicode.IsControl) || permission.PermissionType != "application" && permission.PermissionType != "delegated" {
			return false
		}
	}
	return true
}
