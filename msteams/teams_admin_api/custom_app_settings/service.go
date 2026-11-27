// Package customappsettings reads and updates the tenant-wide Teams apps
// settings behind Get/Set-CsTeamsSettingsCustomApp, including whether users can
// interact with sideloaded (custom) apps.
package customappsettings

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	getCommand       = "Get-CsTeamsSettingsCustomApp"
	setCommand       = "Set-CsTeamsSettingsCustomApp"
	sourceGetCommand = "Get-CsTeamsSettingsCustomApp_Get"
	sourceSetCommand = "Set-CsTeamsSettingsCustomApp_SetExpanded"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) check(ctx context.Context) error {
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

func request(method, command string) client.Request {
	return client.Request{Method: method, Path: []string{"Teams.MiddletierService", "tenantWideAppsSettingsGlobal"},
		Header: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}}
}

func (s *Service) read(ctx context.Context, command string) (*resty.Response, error) {
	req := request(http.MethodGet, command)
	req.Replay = client.ReplaySafe
	response, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return response, err
	}
	if response == nil || response.StatusCode() != http.StatusOK {
		return response, ErrResponse
	}
	return response, nil
}
func (s *Service) get(ctx context.Context, command string) (*Settings, *resty.Response, error) {
	response, err := s.read(ctx, command)
	if err != nil {
		return nil, response, err
	}
	var value Settings
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// GetSideloadedAppsInteraction reads the pinned public command's one-property
// output. It uses the exact raw key and nullable bool projection, independent of
// the enhanced full document's stricter extension-field decoding. A false value
// with an error is not a successful disabled-setting observation.
func (s *Service) GetSideloadedAppsInteraction(ctx context.Context) (bool, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return false, nil, err
	}
	response, err := s.read(ctx, sourceGetCommand)
	if err != nil {
		return false, response, err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &fields) != nil || fields == nil {
		return false, response, client.ErrDecode
	}
	return projectSideloadedFlag(fields), response, nil
}

// Get reads the enhanced complete settings document with strict typed decoding.
// For every pinned public flag outcome, including malformed nullable scalars,
// use GetSideloadedAppsInteraction. The snapshot getter projects valid documents.
func (s *Service) Get(ctx context.Context) (*Settings, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	return s.get(ctx, getCommand)
}

// Update replaces the settings document with one PUT. A nil AppSettingsList is
// sent as an empty list, as the module does. It is never replayed
// automatically. The source declares only 204 as successful.
func (s *Service) Update(ctx context.Context, value UpdateRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if value.AppSettingsList == nil {
		value.AppSettingsList = []AppSettingUpdate{}
	}
	for _, item := range value.AppSettingsList {
		if item.ID == "" {
			return nil, ErrInput
		}
	}
	req := request(http.MethodPut, setCommand)
	req.Body = value
	response, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return response, err
	}
	if response == nil || response.StatusCode() != http.StatusNoContent {
		return response, ErrResponse
	}
	return response, nil
}

// SetSideloadedAppsInteraction is the existing enhanced read/modify/write
// convenience. It uses strict full-document decoding, preserves nested unknown
// config properties and omits branding. For the pinned source projection, use
// PrepareSideloadedAppsInteraction and UpdateSideloadedAppsInteraction. The read
// and write are not atomic, and PUT is never automatically replayed.
func (s *Service) SetSideloadedAppsInteraction(ctx context.Context, enabled bool) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	current, response, err := s.get(ctx, getCommand)
	if err != nil {
		return response, err
	}
	return s.Update(ctx, UpdateRequest{
		IsAppsEnabled:                      current.IsAppsEnabled,
		IsAppsPurchaseEnabled:              current.IsAppsPurchaseEnabled,
		IsExternalAppsEnabledByDefault:     current.IsExternalAppsEnabledByDefault,
		IsLicenseBasedPinnedAppsEnabled:    current.IsLicenseBasedPinnedAppsEnabled,
		IsSideloadedAppsInteractionEnabled: &enabled,
		IsTenantWideAutoInstallEnabled:     current.IsTenantWideAutoInstallEnabled,
		AppSettingsList:                    []AppSettingUpdate{},
		AppAccessRequestConfig:             current.AppAccessRequestConfig,
	})
}
