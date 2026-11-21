package teamtemplate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

// ErrResponse identifies an undeclared successful HTTP status. Error responses
// preserve their actual status through Error and the shared client.APIError.
var ErrResponse = errors.New("unexpected team template response")

// Error adds the template error envelope to the shared classified HTTP error.
// Envelope is nil for an empty/malformed/non-object error body. Raw bytes remain
// in the returned Resty response; formatting this error never prints them.
type Error struct {
	APIError *client.APIError
	Envelope *ErrorEnvelope
}

func (e *Error) Error() string {
	if e == nil || e.APIError == nil {
		return "team template request failed"
	}
	return e.APIError.Error()
}

func (e *Error) Unwrap() error {
	if e == nil || e.APIError == nil {
		return nil
	}
	return e.APIError
}

// Value receivers also redact copied error values, as in client.APIError.
func (e Error) String() string   { return (&e).Error() }
func (e Error) GoString() string { return (&e).Error() }

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
// The root collection supplies a discovery adapter shared by its resources.
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
	if value == "" || strings.ContainsRune(value, '\\') || strings.ContainsFunc(value, unicode.IsControl) {
		return false
	}
	// Match the shared route guard before discovery can acquire credentials.
	for _, part := range strings.Split(value, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	return true
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		var apiError *client.APIError
		if response != nil && errors.As(err, &apiError) && apiError != nil {
			templateError := &Error{APIError: apiError}
			var envelope ErrorEnvelope
			if json.Unmarshal(response.Body(), &envelope) == nil {
				templateError.Envelope = &envelope
			}
			return response, templateError
		}
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// GetByID treats the returned @odata.id as opaque data and escapes it once as a
// complete path segment. It never follows a URL from the template object.
func (s *Service) GetByID(ctx context.Context, odataID string) (*Template, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(odataID) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{"Teams.Templates", odataID},
		Header: headers("Get-CsTeamTemplate"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	fields, ok := object(response.Body())
	if !ok {
		return nil, response, client.ErrDecode
	}
	var value Template
	if json.Unmarshal(fields["value"], &value) != nil || value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// List reads the immediate value envelope. Nil options/locale use en-US; an
// explicitly blank locale is invalid rather than silently replacing a bound key.
// Locale selects Microsoft templates; tenant custom templates remain included.
// Check IsIncomplete and Diagnostics before treating the collection as complete.
func (s *Service) List(ctx context.Context, options *ListOptions) (*ListResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	locale := "en-US"
	if options != nil && options.PublicTemplateLocale != nil {
		locale = *options.PublicTemplateLocale
	}
	if strings.TrimSpace(locale) == "" || !validPart(locale) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet,
		Path:   []string{"Teams.Templates", "api", "TeamTemplates", "v1.0", locale},
		Header: headers("Get-CsTeamTemplateList"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	fields, ok := object(response.Body())
	if !ok {
		return nil, response, client.ErrDecode
	}
	items := []TemplateSummary{}
	if json.Unmarshal(fields["value"], &items) != nil || items == nil {
		return nil, response, client.ErrDecode
	}
	diagnostics := response.Header().Values("X-Teams-TemplateSvc-Diagnostics")
	incomplete := false
	for _, value := range diagnostics {
		if strings.Contains(strings.ToLower(value), "templatesretrievalfailure") {
			incomplete = true
		}
	}
	return &ListResult{Items: items, IsIncomplete: incomplete, Diagnostics: diagnostics, RawFields: fields}, response, nil
}

// Create adds the evidenced General channel on a copy, then returns the
// immediate create receipt. It does not invent a polling/read-back operation.
// Use a returned @odata.id (or reconcile the owned fixture via List) for GetByID.
func (s *Service) Create(ctx context.Context, locale string, definition Definition) (*TemplateSummary, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(locale) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost,
		Path:   []string{"Teams.Templates", "api", "TeamTemplates", "v1.0", "Tenant", locale},
		Header: headers("New-CsTeamTemplate"), Body: withGeneral(definition)}, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	return decodeSummary(response)
}

// UpdateByID sends the supplied flat definition as PUT. It performs no General
// channel insertion and never implicitly retries the mutation or polls a task.
func (s *Service) UpdateByID(ctx context.Context, odataID string, definition Definition) (*TemplateSummary, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(odataID) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut,
		Path: []string{"Teams.Templates", odataID}, Header: headers("Update-CsTeamTemplate"), Body: definition}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	return decodeSummary(response)
}

// DeleteByID accepts 200/204. A nonempty 200 Any payload remains in response.Body;
// a genuinely empty 204 succeeds. Malformed nonempty JSON preserves metadata and
// returns ErrDecode. This method never replays a deletion automatically.
func (s *Service) DeleteByID(ctx context.Context, odataID string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(odataID) {
		return nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodDelete,
		Path: []string{"Teams.Templates", odataID}, Header: headers("Remove-CsTeamTemplate")}, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return response, err
	}
	if response.StatusCode() == http.StatusOK && len(bytes.TrimSpace(response.Body())) != 0 && !json.Valid(response.Body()) {
		return response, client.ErrDecode
	}
	return response, nil
}

func decodeSummary(response *resty.Response) (*TemplateSummary, *resty.Response, error) {
	var value TemplateSummary
	if json.Unmarshal(response.Body(), &value) != nil || value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

func withGeneral(definition Definition) Definition {
	for _, channel := range definition.Channels {
		if channel != nil && channel.DisplayName != nil && strings.EqualFold(*channel.DisplayName, "General") {
			return definition
		}
	}
	name, favorite := "General", true
	channels := make([]*Channel, len(definition.Channels), len(definition.Channels)+1)
	copy(channels, definition.Channels)
	definition.Channels = append(channels, &Channel{ID: &name, DisplayName: &name, IsFavoriteByDefault: &favorite})
	return definition
}
