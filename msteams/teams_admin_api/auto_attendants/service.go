// Package autoattendants manages Teams auto attendants
// (Get/New/Set/Remove/Update-CsAutoAttendant, Get-CsAutoAttendantStatus, the
// supported language, time zone and tenant information reads, and the
// server-validated New-CsAutoAttendant* builders).
package autoattendants

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	root      = "Teams.VoiceApps"
	attendant = "auto-attendants"
	draft     = "draft"
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

func validPart(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." &&
		!strings.ContainsAny(value, `/\`) && !strings.ContainsFunc(value, unicode.IsControl)
}

func blank(value *string) bool { return value == nil || strings.TrimSpace(*value) == "" }

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

// boolText is .NET Boolean.ToString, the module's query spelling.
func boolText(value bool) string {
	if value {
		return "True"
	}
	return "False"
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

// field decodes one required, non-null property of a response object.
func field[T any](raw []byte, name string) (*T, error) {
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil {
		return nil, client.ErrDecode
	}
	value, ok := body[name]
	if !ok || string(value) == "null" {
		return nil, client.ErrDecode
	}
	out := new(T)
	if json.Unmarshal(value, out) != nil {
		return nil, client.ErrDecode
	}
	return out, nil
}

// fetch sends one request and decodes the named response property.
func fetch[T any](ctx context.Context, s *Service, request client.Request, name string, statuses ...int) (*T, *resty.Response, error) {
	response, err := s.do(ctx, request, statuses...)
	if err != nil {
		return nil, response, err
	}
	value, err := field[T](response.Body(), name)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// List reads auto attendants (Get-CsAutoAttendant without Identity). No
// attendant is a non-nil empty slice.
func (s *Service) List(ctx context.Context, options *ListOptions) ([]*AutoAttendant, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if options != nil {
		if (options.First != nil && *options.First < 0) || (options.Skip != nil && *options.Skip < 0) {
			return nil, nil, ErrInput
		}
		if options.First != nil {
			query.Set("First", strconv.FormatInt(int64(*options.First), 10))
		}
		if options.Skip != nil {
			query.Set("Skip", strconv.FormatInt(int64(*options.Skip), 10))
		}
		for name, value := range map[string]*bool{"ExcludeContent": options.ExcludeContent, "Descending": options.Descending, "IncludeStatus": options.IncludeStatus} {
			if value != nil {
				query.Set(name, boolText(*value))
			}
		}
		if options.NameFilter != nil {
			query.Set("NameFilter", *options.NameFilter)
		}
		if options.SortBy != nil {
			query.Set("SortBy", *options.SortBy)
		}
	}
	values, response, err := fetch[[]*AutoAttendant](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, attendant}, Query: query,
		Header: headers("Get-CsAutoAttendant"), Replay: client.ReplaySafe}, "AutoAttendants", http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	for _, value := range *values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return *values, response, nil
}

// GetByID reads one auto attendant (Get-CsAutoAttendant -Identity). An
// unknown identity is HTTP 404 with code 100004. The module also reads the
// status; use GetStatusByID for that.
func (s *Service) GetByID(ctx context.Context, identity string) (*AutoAttendant, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	return fetch[AutoAttendant](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, attendant, identity},
		Header: headers("Get-CsAutoAttendant"), Replay: client.ReplaySafe}, "AutoAttendant", http.StatusOK)
}

// GetStatusByID reads an auto attendant's provisioning status records
// (Get-CsAutoAttendantStatus). A healthy attendant has none: the result is a
// non-nil empty slice. The pinned module's wrapper parses but never emits the
// records; the SDK returns them.
func (s *Service) GetStatusByID(ctx context.Context, identity string, options *StatusOptions) ([]*StatusRecord, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	var query url.Values
	if options != nil && len(options.IncludeResources) > 0 {
		for _, value := range options.IncludeResources {
			if !validPart(value) || strings.Contains(value, ",") {
				return nil, nil, ErrInput
			}
		}
		query = url.Values{"includeResources": {strings.Join(options.IncludeResources, ",")}}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, attendant, "status", identity}, Query: query,
		Header: headers("Get-CsAutoAttendantStatus"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*StatusRecord
	if json.Unmarshal(response.Body(), &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// Create creates an auto attendant (New-CsAutoAttendant). Name, LanguageID,
// TimeZoneID and DefaultCallFlow are required. As in the module,
// AutoRecordingTemplateID requires MainlineAttendantEnabled, and enabling the
// mainline attendant turns voice response on. No resource account is needed.
// Create is never replayed automatically; reconcile a lost response with
// List and NameFilter.
func (s *Service) Create(ctx context.Context, settings Settings) (*AutoAttendant, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(settings.Name) || blank(settings.LanguageID) || blank(settings.TimeZoneID) || settings.DefaultCallFlow == nil ||
		strings.ContainsFunc(*settings.Name, unicode.IsControl) {
		return nil, nil, ErrInput
	}
	mainline := settings.MainlineAttendantEnabled != nil && *settings.MainlineAttendantEnabled
	if !blank(settings.AutoRecordingTemplateID) && !mainline {
		return nil, nil, ErrInput
	}
	if mainline {
		on := true
		settings.VoiceResponseEnabled = &on
	}
	return fetch[AutoAttendant](ctx, s, client.Request{Method: http.MethodPost, Path: []string{root, attendant},
		Header: headers("New-CsAutoAttendant"), Body: settings}, "AutoAttendant", http.StatusCreated, http.StatusOK)
}

// UpdateByID replaces an auto attendant (Set-CsAutoAttendant -Instance). The
// service treats the body as the complete attendant: start from
// AutoAttendant.Settings of a fresh read, change fields, and pass the result.
// The service answers with the updated attendant.
func (s *Service) UpdateByID(ctx context.Context, identity string, settings Settings) (*AutoAttendant, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) || blank(settings.Name) || strings.ContainsFunc(*settings.Name, unicode.IsControl) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: []string{root, attendant, identity},
		Header: headers("Set-CsAutoAttendant"), Body: settings}, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	if strings.TrimSpace(string(response.Body())) == "" {
		return nil, response, nil
	}
	value, err := field[AutoAttendant](response.Body(), "AutoAttendant")
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// RefreshByID asks the service to re-resolve an auto attendant's
// dependencies, such as dial-by-name and audio resources
// (Update-CsAutoAttendant). It answers with the attendant.
func (s *Service) RefreshByID(ctx context.Context, identity string) (*AutoAttendant, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	return fetch[AutoAttendant](ctx, s, client.Request{Method: http.MethodPost, Path: []string{root, attendant, identity, "refresh"},
		Header: headers("Update-CsAutoAttendant")}, "AutoAttendant", http.StatusOK)
}

// DeleteByID removes an auto attendant (Remove-CsAutoAttendant). A missing
// attendant is an HTTP 404 API error.
func (s *Service) DeleteByID(ctx context.Context, identity string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(identity) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{root, attendant, identity},
		Header: headers("Remove-CsAutoAttendant")}, http.StatusOK, http.StatusNoContent)
}

// GetTenantInformation reads the tenant's auto attendant defaults and the
// supported languages and time zones (Get-CsAutoAttendantTenantInformation).
func (s *Service) GetTenantInformation(ctx context.Context) (*TenantInformation, *resty.Response, error) {
	return s.information(ctx, "Get-CsAutoAttendantTenantInformation")
}

func (s *Service) information(ctx context.Context, command string) (*TenantInformation, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	return fetch[TenantInformation](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, "information"},
		Header: headers(command), Replay: client.ReplaySafe}, "TenantInformation", http.StatusOK)
}

// ListSupportedLanguages reads the supported languages
// (Get-CsAutoAttendantSupportedLanguage without Identity). The module reads
// them from the tenant information.
func (s *Service) ListSupportedLanguages(ctx context.Context) ([]*Language, *resty.Response, error) {
	value, response, err := s.information(ctx, "Get-CsAutoAttendantTenantInformation")
	if err != nil {
		return nil, response, err
	}
	if value.SupportedLanguages == nil {
		return nil, response, client.ErrDecode
	}
	return value.SupportedLanguages, response, nil
}

// ListSupportedTimeZones reads the supported time zones
// (Get-CsAutoAttendantSupportedTimeZone without Identity) from the tenant
// information.
func (s *Service) ListSupportedTimeZones(ctx context.Context) ([]*TimeZone, *resty.Response, error) {
	value, response, err := s.information(ctx, "Get-CsAutoAttendantTenantInformation")
	if err != nil {
		return nil, response, err
	}
	if value.SupportedTimeZones == nil {
		return nil, response, client.ErrDecode
	}
	return value.SupportedTimeZones, response, nil
}

// GetSupportedLanguageByID reads one supported language
// (Get-CsAutoAttendantSupportedLanguage -Identity). An unsupported language
// is HTTP 404 with code 100004.
func (s *Service) GetSupportedLanguageByID(ctx context.Context, identity string) (*Language, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	return fetch[Language](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, "supported-languages", identity},
		Header: headers("Get-CsAutoAttendantSupportedLanguage"), Replay: client.ReplaySafe}, "Language", http.StatusOK)
}

// GetSupportedTimeZoneByID reads one supported time zone
// (Get-CsAutoAttendantSupportedTimeZone -Identity). An unknown time zone is
// HTTP 404 with code 100004.
func (s *Service) GetSupportedTimeZoneByID(ctx context.Context, identity string) (*TimeZone, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	return fetch[TimeZone](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, "supported-timezones", identity},
		Header: headers("Get-CsAutoAttendantSupportedTimeZone"), Replay: client.ReplaySafe}, "TimeZone", http.StatusOK)
}

// draftOf posts a builder value to its draft route. The service validates it
// and answers with the normalized value; nothing is stored.
func draftOf[T any](ctx context.Context, s *Service, path []string, command, name string, body any) (*T, *resty.Response, error) {
	return fetch[T](ctx, s, client.Request{Method: http.MethodPost, Path: append([]string{root, attendant}, append(path, draft)...),
		Header: headers(command), Body: body, Replay: client.ReplaySafe}, name, http.StatusOK)
}

// DraftPrompt validates a prompt (New-CsAutoAttendantPrompt). A prompt needs
// a text-to-speech or audio file value matching ActiveType.
func (s *Service) DraftPrompt(ctx context.Context, value Prompt) (*Prompt, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.ActiveType) {
		return nil, nil, ErrInput
	}
	return draftOf[Prompt](ctx, s, []string{"prompts"}, "New-CsAutoAttendantPrompt", "Prompt", value)
}

// DraftMenuOption validates a menu option (New-CsAutoAttendantMenuOption).
// Action and DtmfResponse are required.
func (s *Service) DraftMenuOption(ctx context.Context, value MenuOption) (*MenuOption, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.Action) || blank(value.DtmfResponse) {
		return nil, nil, ErrInput
	}
	return draftOf[MenuOption](ctx, s, []string{"menus", "menu-options"}, "New-CsAutoAttendantMenuOption", "MenuOption", value)
}

// DraftMenu validates a menu (New-CsAutoAttendantMenu). Name is required.
func (s *Service) DraftMenu(ctx context.Context, value Menu) (*Menu, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.Name) {
		return nil, nil, ErrInput
	}
	return draftOf[Menu](ctx, s, []string{"menus"}, "New-CsAutoAttendantMenu", "Menu", value)
}

// DraftCallFlow validates a call flow and assigns its ID
// (New-CsAutoAttendantCallFlow). Name and Menu are required.
func (s *Service) DraftCallFlow(ctx context.Context, value CallFlow) (*CallFlow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.Name) || value.Menu == nil {
		return nil, nil, ErrInput
	}
	return draftOf[CallFlow](ctx, s, []string{"call-flows"}, "New-CsAutoAttendantCallFlow", "CallFlow", value)
}

// DraftCallableEntity validates a transfer or operator target
// (New-CsAutoAttendantCallableEntity). ID and Type are required.
func (s *Service) DraftCallableEntity(ctx context.Context, value CallableEntity) (*CallableEntity, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.ID) || blank(value.Type) {
		return nil, nil, ErrInput
	}
	return draftOf[CallableEntity](ctx, s, []string{"callable-entities"}, "New-CsAutoAttendantCallableEntity", "CallableEntity", value)
}

// DraftGroupDialScope validates a group dial scope and returns it as a
// DialScope (New-CsAutoAttendantDialScope -GroupScope). The service checks
// each group exists in Microsoft Entra ID (HTTP 404, code 100004).
func (s *Service) DraftGroupDialScope(ctx context.Context, groupIDs []string) (*DialScope, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if len(groupIDs) == 0 {
		return nil, nil, ErrInput
	}
	for _, id := range groupIDs {
		if blank(&id) {
			return nil, nil, ErrInput
		}
	}
	return draftOf[DialScope](ctx, s, []string{"group-dial-scopes"}, "New-CsAutoAttendantDialScope", "DialScope", GroupDialScope{GroupIDs: groupIDs})
}

// DraftCallHandlingAssociation validates a call handling association
// (New-CsAutoAttendantCallHandlingAssociation). Type, ScheduleID and
// CallFlowID are required; as in the module, Enabled defaults to true.
func (s *Service) DraftCallHandlingAssociation(ctx context.Context, value CallHandlingAssociation) (*CallHandlingAssociation, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if blank(value.Type) || blank(value.ScheduleID) || blank(value.CallFlowID) {
		return nil, nil, ErrInput
	}
	if value.Enabled == nil {
		on := true
		value.Enabled = &on
	}
	return draftOf[CallHandlingAssociation](ctx, s, []string{"call-handling-associations"}, "New-CsAutoAttendantCallHandlingAssociation", "CallHandlingAssociation", value)
}
