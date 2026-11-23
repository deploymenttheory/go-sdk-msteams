// Package callqueues manages Teams call queues (Get/New/Set/Remove-CsCallQueue).
package callqueues

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	root  = "Teams.VoiceApps"
	queue = "callqueues"
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

// envelope decodes the required CallQueue property of a response.
func envelope(raw []byte) (*CallQueue, error) {
	var body struct {
		CallQueue *CallQueue `json:"CallQueue"`
	}
	if json.Unmarshal(raw, &body) != nil || body.CallQueue == nil {
		return nil, client.ErrDecode
	}
	return body.CallQueue, nil
}

// List reads call queues (Get-CsCallQueue without Identity). The module always
// sends FilterInvalidObos=False. No queue is a non-nil empty slice.
func (s *Service) List(ctx context.Context, options *ListOptions) ([]*CallQueue, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query := url.Values{"FilterInvalidObos": {"False"}}
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
		if options.ExcludeContent != nil {
			query.Set("ExcludeContent", boolText(*options.ExcludeContent))
		}
		if options.NameFilter != nil {
			query.Set("NameFilter", *options.NameFilter)
		}
		if options.SortBy != nil {
			query.Set("SortBy", *options.SortBy)
		}
		if options.Descending != nil {
			query.Set("Descending", boolText(*options.Descending))
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, queue}, Query: query,
		Header: headers("Get-CsCallQueue"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var body struct {
		CallQueues []*CallQueue `json:"CallQueues"`
	}
	if json.Unmarshal(response.Body(), &body) != nil || body.CallQueues == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range body.CallQueues {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return body.CallQueues, response, nil
}

// GetByID reads one call queue (Get-CsCallQueue -Identity). The service
// answers an unknown or deleted identity with HTTP 400, not 404.
func (s *Service) GetByID(ctx context.Context, identity string) (*CallQueue, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{root, queue, identity}, Query: url.Values{"FilterInvalidObos": {"False"}},
		Header: headers("Get-CsCallQueue"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := envelope(response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// normalize applies the module's local rules shared by New and Set: blank
// strings are omitted, call priorities apply only to Forward actions (0-5),
// and AutoRecordingTemplateID requires ConferenceMode not to be false.
func normalize(settings Settings) (Settings, bool) {
	value := reflect.ValueOf(&settings).Elem()
	for i := range value.NumField() {
		if field, ok := value.Field(i).Interface().(*string); ok && field != nil {
			if strings.ContainsFunc(*field, unicode.IsControl) {
				return settings, false
			}
			if strings.TrimSpace(*field) == "" {
				value.Field(i).Set(reflect.Zero(value.Field(i).Type()))
			}
		}
	}
	for _, rule := range []struct {
		action   *int32
		forward  int32
		priority *int32
	}{
		{settings.OverflowAction, OverflowForward, settings.OverflowActionCallPriority},
		{settings.TimeoutAction, TimeoutForward, settings.TimeoutActionCallPriority},
		{settings.NoAgentAction, NoAgentForward, settings.NoAgentActionCallPriority},
	} {
		if rule.priority == nil {
			continue
		}
		forward := rule.action != nil && *rule.action == rule.forward
		if (forward && (*rule.priority < 0 || *rule.priority > 5)) || (!forward && *rule.priority != 0) {
			return settings, false
		}
	}
	if settings.AutoRecordingTemplateID != nil && settings.ConferenceMode != nil && !*settings.ConferenceMode {
		return settings, false
	}
	return settings, true
}

func writeQuery(options *WriteOptions) (url.Values, bool) {
	if options == nil || options.ChannelUserObjectID == nil {
		return nil, true
	}
	if !validPart(*options.ChannelUserObjectID) {
		return nil, false
	}
	return url.Values{"channelUserObjectId": {*options.ChannelUserObjectID}}, true
}

// Create creates a call queue (New-CsCallQueue). Name is required. As in the
// module, RoutingLongestIdle forces PresenceAwareRouting to false. The service
// answers with the created queue. Create is never replayed automatically;
// reconcile a lost response with List and NameFilter.
func (s *Service) Create(ctx context.Context, settings Settings, options *WriteOptions) (*CallQueue, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	settings, ok := normalize(settings)
	query, queryOK := writeQuery(options)
	if !ok || !queryOK || settings.Name == nil {
		return nil, nil, ErrInput
	}
	if settings.RoutingMethod != nil && *settings.RoutingMethod == RoutingLongestIdle {
		off := false
		settings.PresenceAwareRouting = &off
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{root, queue}, Query: query,
		Header: headers("New-CsCallQueue"), Body: settings}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := envelope(response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// UpdateByID replaces a call queue's settings (Set-CsCallQueue). The service
// treats the body as the complete queue: start from CallQueue.Settings of a
// fresh read, change fields, and pass the result. Name is required. The
// returned queue is nil when the service sends no body; read it back with
// GetByID as the module does.
func (s *Service) UpdateByID(ctx context.Context, identity string, settings Settings, options *WriteOptions) (*CallQueue, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	settings, ok := normalize(settings)
	query, queryOK := writeQuery(options)
	if !ok || !queryOK || !validPart(identity) || settings.Name == nil {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: []string{root, queue, identity}, Query: query,
		Header: headers("Set-CsCallQueue"), Body: settings}, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	if strings.TrimSpace(string(response.Body())) == "" {
		return nil, response, nil
	}
	var body struct {
		CallQueue *CallQueue `json:"CallQueue"`
	}
	if json.Unmarshal(response.Body(), &body) != nil {
		return nil, response, client.ErrDecode
	}
	return body.CallQueue, response, nil
}

// DeleteByID removes a call queue (Remove-CsCallQueue). The module reads the
// queue first only to echo it; a missing queue is an HTTP 400 API error.
func (s *Service) DeleteByID(ctx context.Context, identity string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(identity) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{root, queue, identity},
		Header: headers("Remove-CsCallQueue")}, http.StatusOK, http.StatusNoContent)
}
