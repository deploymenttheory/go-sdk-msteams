// Package audioprompts manages tenant audio files in Teams media storage, the
// prompts used by auto attendants, call queues and hold music. Custom prompt
// and prompt package builders for audio conferencing live in the
// audio_conferencing_prompts package, which owns the configuration they edit.
package audioprompts

import (
	"context"
	"encoding/base64"
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

const root = "Teams.MediaStorage"

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

// application applies the module's normalization: empty or a case-insensitive
// known name becomes the canonical spelling; other values pass through.
func application(value string) (string, bool) {
	if value == "" {
		return TenantGlobal, true
	}
	for _, known := range []string{TenantGlobal, OrgAutoAttendant, HuntGroup} {
		if strings.EqualFold(value, known) {
			return known, true
		}
	}
	return value, validPart(value)
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
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

func decodeFile(response *resty.Response) (*AudioFile, error) {
	var value AudioFile
	if err := json.Unmarshal(response.Body(), &value); err != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}

// List reads every audio file of one application, including files already
// marked for deletion. An empty applicationID means TenantGlobal.
func (s *Service) List(ctx context.Context, applicationID string) ([]*AudioFile, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, ok := application(applicationID)
	if !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, "audiofile", app},
		Header: headers("Get-CsOnlineAudioFile"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*AudioFile
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

// GetByID reads one audio file. A deleted file may still read successfully
// with MarkedForDeletion true until the service purges it.
func (s *Service) GetByID(ctx context.Context, applicationID, id string, options *GetOptions) (*AudioFile, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, ok := application(applicationID)
	if !ok || !validPart(id) {
		return nil, nil, ErrInput
	}
	var query url.Values
	if options != nil && options.DurationInMinutes != nil {
		query = url.Values{"durationInMins": {strconv.FormatInt(int64(*options.DurationInMinutes), 10)}}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, "audiofile", app, id}, Query: query,
		Header: headers("Get-CsOnlineAudioFile"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeFile(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// Create uploads an audio file (Import-CsOnlineAudioFile) and returns the
// immediate receipt. It is never replayed automatically; a lost response is
// reconciled by listing for the original file name.
func (s *Service) Create(ctx context.Context, request ImportRequest) (*AudioFile, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, ok := application(request.ApplicationID)
	if !ok || strings.TrimSpace(request.FileName) == "" || strings.ContainsFunc(request.FileName, unicode.IsControl) || len(request.Content) == 0 {
		return nil, nil, ErrInput
	}
	body := map[string]string{"ApplicationId": app, "OriginalFilename": request.FileName, "Content": base64.StdEncoding.EncodeToString(request.Content)}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{root, "audiofile"}, Header: headers("Import-CsOnlineAudioFile"), Body: body}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeFile(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// GetContentByID downloads the audio bytes (Export-CsOnlineAudioFile).
func (s *Service) GetContentByID(ctx context.Context, applicationID, id string) ([]byte, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, ok := application(applicationID)
	if !ok || !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, "audiofile", app, id, "content"},
		Header: headers("Export-CsOnlineAudioFile"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &fields) != nil || fields == nil {
		return nil, response, client.ErrDecode
	}
	content, err := decodeContent(fields["Content"])
	if err != nil || content == nil {
		return nil, response, client.ErrDecode
	}
	return content, response, nil
}

// DeleteByID marks a TenantGlobal audio file for deletion, as the module's
// Remove-CsOnlineAudioFile always does. The file stays readable with
// MarkedForDeletion true; deleting it again is rejected by the service.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{root, "audiofile", TenantGlobal, id},
		Header: headers("Remove-CsOnlineAudioFile")}, http.StatusOK, http.StatusNoContent)
}
