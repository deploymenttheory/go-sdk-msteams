// Package aiagents manages Teams AI agents: the voice apps agent
// configurations that let a resource account hand calls to an external
// conversational provider (Get/New/Set/Remove-CsAgent), the platform catalog of
// AI agents (Get-CsAiAgents), and the AI-generated meeting knowledge kept by the
// MCPS service (Get-AIGeneratedKnowledgeContainer, Remove-AIGeneratedKnowledge).
//
// Agent configurations and the catalog use ConfigAPI. Knowledge requests use the
// fixed MCPS origin and application ID resource the module hardcodes; only the
// commercial cloud is mapped.
package aiagents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/internal/serviceprofile"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	voiceApps = "Teams.VoiceApps"
	agents    = "aiagents"
	platform  = "Teams.PlatformService"
	catalog   = "AiAgents"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery and
// the MCPS knowledge scope.
type Service struct {
	client    client.Client
	knowledge *client.Scope
}

// NewService performs no I/O. base is the backend bound to ConfigAPI's
// audience; knowledge is a scope for the client.AIGeneratedKnowledge backend
// and may be nil when the knowledge methods are not used.
func NewService(base client.Client, knowledge *client.Scope) *Service {
	return &Service{client: configapi.New(base), knowledge: knowledge}
}

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

func validText(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsFunc(value, unicode.IsControl)
}

func validPart(value string) bool {
	return validText(value) && value != "." && value != ".." && !strings.ContainsAny(value, `/\`)
}

func validOptional(value *string) bool {
	return value == nil || validText(*value)
}

// dotnetBool is the .NET Boolean.ToString spelling the module sends.
func dotnetBool(value bool) string {
	if value {
		return "True"
	}
	return "False"
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func accept(response *resty.Response, err error, statuses ...int) (*resty.Response, error) {
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

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	return accept(response, err, statuses...)
}

// decodeAgent decodes the required AIAgent envelope property.
func decodeAgent(raw []byte) (*Agent, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	var value *Agent
	if json.Unmarshal(fields["AIAgent"], &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

func agentBody(request AgentRequest) (map[string]string, bool) {
	if !validText(request.Name) || !validText(request.AgentID) || !validText(request.AgentType) || !validOptional(request.AgentTargetTagTemplateID) {
		return nil, false
	}
	body := map[string]string{"Name": request.Name, "AgentId": request.AgentID, "AgentType": request.AgentType}
	if request.AgentTargetTagTemplateID != nil {
		body["AgentTargetTagTemplateId"] = *request.AgentTargetTagTemplateID
	}
	return body, true
}

// List reads AI agent configurations (Get-CsAgent without Id). Nil options
// send no query, as the module does. No configuration is a non-nil empty list.
func (s *Service) List(ctx context.Context, options *ListOptions) (*AgentList, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if options != nil {
		if (options.First != nil && *options.First < 0) || (options.Skip != nil && *options.Skip < 0) ||
			!validOptional(options.NameFilter) || !validOptional(options.TypeFilter) || !validOptional(options.SortBy) {
			return nil, nil, ErrInput
		}
		for name, value := range map[string]*int32{"First": options.First, "Skip": options.Skip} {
			if value != nil {
				query.Set(name, strconv.FormatInt(int64(*value), 10))
			}
		}
		for name, value := range map[string]*bool{"ExcludeContent": options.ExcludeContent, "Descending": options.Descending} {
			if value != nil {
				query.Set(name, dotnetBool(*value))
			}
		}
		for name, value := range map[string]*string{"NameFilter": options.NameFilter, "TypeFilter": options.TypeFilter, "SortBy": options.SortBy} {
			if value != nil {
				query.Set(name, *value)
			}
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, agents}, Query: query,
		Header: headers("Get-CsAgent"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value AgentList
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// GetByID reads one AI agent configuration (Get-CsAgent -Id). A missing
// configuration is an HTTP 404 API error.
func (s *Service) GetByID(ctx context.Context, id string) (*Agent, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, agents, id},
		Header: headers("Get-CsAgent"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeAgent(response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// Create creates an AI agent configuration (New-CsAgent). The service answers
// HTTP 201 with the new configuration. It is never replayed automatically;
// reconcile a lost response with List.
func (s *Service) Create(ctx context.Context, request AgentRequest) (*Agent, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	body, ok := agentBody(request)
	if !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, agents},
		Header: headers("New-CsAgent"), Body: body}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeAgent(response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// UpdateByID replaces an AI agent configuration (Set-CsAgent). The request is a
// full replacement: the module sends every field of the instance it was given.
func (s *Service) UpdateByID(ctx context.Context, id string, request AgentRequest) (*Agent, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	body, ok := agentBody(request)
	if !ok || !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: []string{voiceApps, agents, id},
		Header: headers("Set-CsAgent"), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeAgent(response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// DeleteByID removes an AI agent configuration (Remove-CsAgent). A missing
// configuration is an HTTP 404 API error.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{voiceApps, agents, id},
		Header: headers("Remove-CsAgent")}, http.StatusOK, http.StatusNoContent)
}

// QueryPlatformAgents reads the platform AI agent catalog for one provider
// (Get-CsAiAgents). Pass QueryResult.ContinuationToken back to read the next page.
func (s *Service) QueryPlatformAgents(ctx context.Context, request QueryRequest) (*QueryResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.ProviderID) || (request.MaxResults != nil && *request.MaxResults < 0) {
		return nil, nil, ErrInput
	}
	query := url.Values{"providerId": {request.ProviderID}}
	for name, value := range map[string]*string{"agentId": request.AgentID, "agentIds": request.AgentIDs, "channel": request.Channel,
		"continuationToken": request.ContinuationToken, "displayNameContains": request.DisplayNameContains, "displayNamePrefix": request.DisplayNamePrefix} {
		if !validOptional(value) {
			return nil, nil, ErrInput
		}
		if value != nil {
			query.Set(name, *value)
		}
	}
	for name, value := range map[string]*bool{"isTeamsIvrEnabled": request.IsTeamsIvrEnabled, "showCount": request.ShowCount} {
		if value != nil {
			query.Set(name, dotnetBool(*value))
		}
	}
	if request.MaxResults != nil {
		query.Set("maxResults", strconv.FormatInt(int64(*request.MaxResults), 10))
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{platform, catalog}, Query: query,
		Header: headers("Get-CsAiAgents"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value QueryResult
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// knowledgeRequest builds an MCPS request under the profile's fixed prefix.
func (s *Service) knowledgeRequest(ctx context.Context, request KnowledgeRequest, method string, parts ...string) (client.Request, error) {
	if err := s.check(ctx); err != nil {
		return client.Request{}, err
	}
	if s.knowledge == nil || s.knowledge.Session() == nil {
		return client.Request{}, client.ErrClosed
	}
	if s.knowledge.Backend() != client.AIGeneratedKnowledge {
		return client.Request{}, client.ErrDestination
	}
	if !validText(request.ThreadID) || !validOptional(request.CallID) {
		return client.Request{}, ErrInput
	}
	environment, err := s.knowledge.Session().Environment()
	if err != nil {
		return client.Request{}, err
	}
	profile, ok := serviceprofile.Resolve(serviceprofile.AIGeneratedKnowledge, environment.Name)
	if !ok {
		return client.Request{}, client.ErrDestination
	}
	query := url.Values{"threadId": {request.ThreadID}}
	if request.CallID != nil {
		query.Set("callId", *request.CallID)
	}
	path := append(strings.Split(strings.TrimPrefix(profile.PathPrefix, "/"), "/"), parts...)
	return client.Request{Method: method, Path: path, Query: query, Header: http.Header{"Accept": {"application/json"}}}, nil
}

// ListKnowledgeContainers reads the storage container URLs that hold a
// meeting's or call's AI-generated knowledge (Get-AIGeneratedKnowledgeContainer).
// The module prints them in batches for Purview import; this returns them all.
func (s *Service) ListKnowledgeContainers(ctx context.Context, request KnowledgeRequest) ([]string, *resty.Response, error) {
	req, err := s.knowledgeRequest(ctx, request, http.MethodGet, "mkb", "crux", "container-urls")
	if err != nil {
		return nil, nil, err
	}
	req.Replay = client.ReplaySafe
	response, err := s.knowledge.Do(ctx, req, nil)
	if response, err = accept(response, err, http.StatusOK); err != nil {
		return nil, response, err
	}
	var values []string
	if json.Unmarshal(response.Body(), &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	return values, response, nil
}

// DeleteKnowledge permanently removes a meeting's or call's AI-generated
// knowledge (Remove-AIGeneratedKnowledge). It cannot be undone and is never
// replayed automatically.
func (s *Service) DeleteKnowledge(ctx context.Context, request KnowledgeRequest) (*resty.Response, error) {
	req, err := s.knowledgeRequest(ctx, request, http.MethodDelete, "mkb", "crux")
	if err != nil {
		return nil, err
	}
	response, err := s.knowledge.Do(ctx, req, nil)
	return accept(response, err, http.StatusOK, http.StatusAccepted, http.StatusNoContent)
}
