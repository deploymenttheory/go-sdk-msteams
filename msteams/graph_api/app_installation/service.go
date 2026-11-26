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

// NewPager builds a lazy Get-TeamsAppInstallation collection traversal. AppID
// filters teamsAppDefinition/teamsAppId; quotes are doubled for OData.
func (s *Service) NewPager(target Target, options ListOptions) (*client.Pager[Installation], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	path, err := target.path()
	if err != nil {
		return nil, err
	}
	if options.AppID != "" && !validID(options.AppID) {
		return nil, ErrInput
	}
	bounds := options.Paging
	if bounds.Style != 0 && bounds.Style != client.LinkPaging || bounds.Parameter != "" || bounds.ExpectedStatus != 0 && bounds.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	bounds.Style, bounds.ExpectedStatus = client.LinkPaging, http.StatusOK
	query := url.Values{"$expand": {"teamsAppDefinition"}}
	if options.AppID != "" {
		query.Set("$filter", "teamsAppDefinition/teamsAppId eq '"+strings.ReplaceAll(options.AppID, "'", "''")+"'")
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

func (s *Service) write(ctx context.Context, target Target, method string, extra []string, body any, status int) (*resty.Response, error) {
	path, err := target.path(extra...)
	if err != nil {
		return nil, err
	}
	response, err := s.scope.Do(ctx, client.Request{Method: method, Path: path, Header: headers(), Body: body, Replay: client.ReplayNever}, nil)
	if err != nil {
		return response, err
	}
	if response.StatusCode() != status {
		return response, ErrResponse
	}
	return response, nil
}

// Install adds an existing catalog app. A lost response remains ambiguous; the
// mutation is never automatically replayed. Reconcile with List by catalog ID.
func (s *Service) Install(ctx context.Context, target Target, input InstallRequest) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(input.AppID) {
		return nil, ErrInput
	}
	// Resolve the bound cloud's Graph origin instead of hardcoding commercial.
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	graph := environment.Endpoints.Graph
	body := struct {
		App string `json:"teamsApp@odata.bind"`
		Permissions
	}{App: strings.TrimRight(graph, "/") + "/v1.0/appCatalogs/teamsApps/" + url.PathEscape(input.AppID), Permissions: input.Permissions}
	return s.write(ctx, target, http.MethodPost, nil, body, http.StatusCreated)
}

// UpgradeByID upgrades one installation. Optional permissions use the pinned
// module's request property; no automatic replay occurs.
func (s *Service) UpgradeByID(ctx context.Context, target Target, id string, permissions Permissions) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validID(id) {
		return nil, ErrInput
	}
	return s.write(ctx, target, http.MethodPost, []string{id, "upgrade"}, permissions, http.StatusNoContent)
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

func (s *Service) resolve(ctx context.Context, target Target, appID string) (string, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return "", nil, err
	}
	if !validID(appID) {
		return "", nil, ErrInput
	}
	result, err := s.List(ctx, target, ListOptions{AppID: appID})
	if err != nil {
		return "", result.Response, err
	}
	if len(result.Items) == 0 {
		return "", result.Response, ErrNotFound
	}
	if len(result.Items) != 1 {
		return "", result.Response, ErrAmbiguous
	}
	return result.Items[0].ID, result.Response, nil
}

// DeleteByAppID resolves the catalog app to exactly one installation before
// deleting. Missing, duplicate or partial reads never send a delete.
func (s *Service) DeleteByAppID(ctx context.Context, target Target, appID string) (*resty.Response, error) {
	id, response, err := s.resolve(ctx, target, appID)
	if err != nil {
		return response, err
	}
	return s.DeleteByID(ctx, target, id)
}

// UpgradeByAppID resolves a catalog app to exactly one installation before
// upgrading. Missing, duplicate or partial reads never send an upgrade.
func (s *Service) UpgradeByAppID(ctx context.Context, target Target, appID string, permissions Permissions) (*resty.Response, error) {
	id, response, err := s.resolve(ctx, target, appID)
	if err != nil {
		return response, err
	}
	return s.UpgradeByID(ctx, target, id, permissions)
}
