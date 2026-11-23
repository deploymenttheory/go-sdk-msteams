// Package unifiedapps manages the Microsoft 365 unified Teams app catalog: app
// availability and installation, custom apps pending review, and unified app
// tenant settings. It calls the Substrate AdminAppCatalog service, which needs a
// delegated token for the exact AdminAppCatalog.ReadWrite.All scope.
package unifiedapps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Service reads and updates the unified app catalog.
type Service struct{ scope *client.Scope }

// NewService binds the service to an AppCatalog backend scope. No I/O occurs.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

var admin = []string{"AdminAppCatalog", "ps", "v2", "admin", "unifiedApp"}

func route(parts ...string) []string { return append(slices.Clone(admin), parts...) }

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.AppCatalog {
		return client.ErrDestination
	}
	return nil
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.scope.Do(ctx, request, nil)
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

func (s *Service) read(ctx context.Context, command string, path []string, query url.Values, out any) (*resty.Response, error) {
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: path, Query: query, Header: headers(command), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return response, err
	}
	if json.Unmarshal(response.Body(), out) != nil {
		return response, client.ErrDecode
	}
	return response, nil
}

func validID(id string) bool {
	return strings.TrimSpace(id) != "" && id == strings.TrimSpace(id) && id != "." && id != ".." &&
		!strings.ContainsFunc(id, unicode.IsControl)
}

func validMembers(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" || strings.ContainsFunc(value, unicode.IsControl) {
			return false
		}
	}
	return true
}

func distinct(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}

// List returns every catalog app with availability, installation and metadata
// expanded (Get-AllM365TeamsApps).
func (s *Service) List(ctx context.Context) ([]App, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	var apps []App
	response, err := s.read(ctx, "Get-AllM365TeamsApps", admin, url.Values{"$expand": {"DeployedTo,AvailableTo,metadata"}}, &apps)
	if err != nil {
		return nil, response, err
	}
	if apps == nil {
		return nil, response, client.ErrDecode
	}
	return apps, response, nil
}

// Get returns one catalog app with its user and group assignments
// (Get-M365TeamsApp).
func (s *Service) Get(ctx context.Context, id string) (*App, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(id) {
		return nil, nil, ErrInput
	}
	var app *App
	response, err := s.read(ctx, "Get-M365TeamsApp", route(id), url.Values{"$expand": {"metadata"}}, &app)
	if err != nil {
		return nil, response, err
	}
	if app == nil {
		return nil, response, client.ErrDecode
	}
	return app, response, nil
}

// ListCustomPending returns custom apps awaiting review
// (Get-M365UnifiedCustomPendingApps).
func (s *Service) ListCustomPending(ctx context.Context) ([]PendingApp, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	return s.pending(ctx, "Get-M365UnifiedCustomPendingApps")
}

func (s *Service) pending(ctx context.Context, command string) ([]PendingApp, *resty.Response, error) {
	var apps []PendingApp
	response, err := s.read(ctx, command, route("stagedApps"), nil, &apps)
	if err != nil {
		return nil, response, err
	}
	if apps == nil {
		return nil, response, client.ErrDecode
	}
	return apps, response, nil
}

var settingNames = []string{SettingDefaultApp, SettingGlobalApp, SettingPrivateApp, SettingEnableCopilotExtensibility}

// GetTenantSettings returns unified app tenant settings, optionally limited to
// the named settings (Get-M365UnifiedTenantSettings).
func (s *Service) GetTenantSettings(ctx context.Context, names ...string) ([]TenantSetting, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	var query url.Values
	if len(names) > 0 {
		for _, name := range names {
			if !slices.Contains(settingNames, name) {
				return nil, nil, ErrInput
			}
		}
		query = url.Values{"SettingNames": {strings.Join(names, ",")}}
	}
	var settings []TenantSetting
	response, err := s.read(ctx, "Get-M365UnifiedTenantSettings", route("settings"), query, &settings)
	if err != nil {
		return nil, response, err
	}
	if settings == nil {
		return nil, response, client.ErrDecode
	}
	return settings, response, nil
}

var writeStatuses = []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}

func (s *Service) write(ctx context.Context, method, command string, path []string, body any) (*resty.Response, error) {
	h := headers(command)
	h.Set("Content-Type", "application/json")
	return s.do(ctx, client.Request{Method: method, Path: path, Header: h, Body: body}, writeStatuses...)
}

func membershipFor(kind, operation string, users, groups []string, version string) (*membership, bool) {
	switch {
	case strings.EqualFold(kind, TypeUsersAndGroups):
		operation = strings.ToLower(operation)
		if operation != OperationAdd && operation != OperationRemove || len(users) == 0 && len(groups) == 0 || !validMembers(users) || !validMembers(groups) {
			return nil, false
		}
		return &membership{Type: kind, Users: distinct(users), Groups: distinct(groups), Operation: operation, Version: version}, true
	case strings.EqualFold(kind, TypeEveryone), strings.EqualFold(kind, TypeNoone):
		if users != nil || groups != nil {
			return nil, false
		}
		return &membership{Type: kind, Operation: OperationAdd, Version: version}, true
	}
	return nil, false
}

// Update changes an app's blocked state, availability or installation
// (Update-M365TeamsApp). Validation mirrors the module and runs before I/O.
func (s *Service) Update(ctx context.Context, id string, update AppUpdate) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validID(id) || update.IsBlocked == nil && update.AssignmentType == "" && update.InstallType == "" {
		return nil, ErrInput
	}
	body := appUpdateBody{ID: id, IsBlocked: update.IsBlocked}
	if update.AssignmentType != "" {
		m, ok := membershipFor(update.AssignmentType, update.Operation, update.Users, update.Groups, "")
		if !ok {
			return nil, ErrInput
		}
		body.AvailableTo = m
	} else if update.Operation != "" || update.Users != nil || update.Groups != nil {
		return nil, ErrInput
	}
	if update.InstallType != "" {
		m, ok := membershipFor(update.InstallType, update.InstallOperation, update.InstallUsers, update.InstallGroups, update.InstallVersion)
		if !ok {
			return nil, ErrInput
		}
		body.DeployTo = m
	} else if update.InstallOperation != "" || update.InstallUsers != nil || update.InstallGroups != nil || update.InstallVersion != "" {
		return nil, ErrInput
	}
	return s.write(ctx, http.MethodPatch, "Update-M365TeamsApp", admin, body)
}

// ReviewCustomPending publishes or rejects a custom pending app
// (Update-M365UnifiedCustomPendingApp). Like the module it first confirms the
// app is pending, then reads its staged definition etag, then sends the review.
// The status is case-insensitive and sent as Published or Rejected.
func (s *Service) ReviewCustomPending(ctx context.Context, id, status string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	const command = "Update-M365UnifiedCustomPendingApp"
	switch {
	case strings.EqualFold(status, ReviewPublished):
		status = ReviewPublished
	case strings.EqualFold(status, ReviewRejected):
		status = ReviewRejected
	default:
		return nil, ErrInput
	}
	if !validID(id) {
		return nil, ErrInput
	}
	apps, response, err := s.pending(ctx, command)
	if err != nil {
		return response, err
	}
	if !slices.ContainsFunc(apps, func(app PendingApp) bool { return app.ID != nil && *app.ID == id }) {
		return response, ErrPendingAppNotFound
	}
	var details pendingDetails
	response, err = s.read(ctx, command, []string{"AdminAppCatalog", "v2", "admin", "unifiedApps", "stagedApps", id, "details"}, nil, &details)
	if err != nil {
		return response, err
	}
	if details.StagedAppEntitlement == nil || details.StagedAppEntitlement.Iteration == nil || *details.StagedAppEntitlement.Iteration == "" {
		return response, ErrMissingIteration
	}
	return s.write(ctx, http.MethodPut, command, route("stagedApps", "review"), reviewBody{ID: id, ReviewStatus: status, StagedAppDefinitionEtag: *details.StagedAppEntitlement.Iteration})
}

// UpdateTenantSettings changes one unified app tenant setting
// (Update-M365UnifiedTenantSettings).
func (s *Service) UpdateTenantSettings(ctx context.Context, update TenantSettingsUpdate) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !slices.Contains(settingNames, update.SettingName) ||
		update.SettingValue != SettingValueAll && update.SettingValue != SettingValueNone && update.SettingValue != SettingValueSome ||
		update.Operation != "" && update.Operation != OperationAdd && update.Operation != OperationRemove ||
		!validMembers(update.Users) || !validMembers(update.Groups) {
		return nil, ErrInput
	}
	if update.SettingName != SettingEnableCopilotExtensibility &&
		(update.SettingValue == SettingValueSome || update.Users != nil || update.Groups != nil || update.Operation != "") {
		return nil, ErrInput
	}
	return s.write(ctx, http.MethodPatch, "Update-M365UnifiedTenantSettings", route("settings"), settingsBody(update))
}
