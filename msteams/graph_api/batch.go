package graphapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrBatchRequest  = errors.New("invalid Graph batch request")
	ErrBatchResponse = errors.New("incomplete or invalid Graph batch response")
	ErrBatchItems    = errors.New("one or more Graph batch operations failed")
)

// BatchRequest is one Graph JSON batch operation. URL is relative to the selected
// Graph API version, such as /teams/id/channels; it cannot select another host or
// version. Body is explicit JSON, including a JSON base64url string for binary data.
// Content-Type is required with Body. IDs are unique, case-insensitively.
type BatchRequest struct {
	ID        string            `json:"id"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      json.RawMessage   `json:"body,omitempty"`
	DependsOn []string          `json:"dependsOn,omitempty"`
}

// BatchResponse retains item status, headers (including Retry-After) and opaque
// body. Order can differ from requests; correlate by case-insensitive ID. Bodies
// and explicit fields may contain tenant data and must be handled deliberately.
type BatchResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

func (BatchRequest) String() string      { return "msteams Graph batch request (redacted)" }
func (r BatchRequest) GoString() string  { return r.String() }
func (BatchResponse) String() string     { return "msteams Graph batch response (redacted)" }
func (r BatchResponse) GoString() string { return r.String() }

// ExecuteBatch sends 1–20 operations once. It never retries the outer batch or
// automatically resubmits throttled/dependent items, because other items may have
// committed. It returns all valid item outcomes alongside ErrBatchItems for mixed
// failures, or ErrBatchResponse for missing/duplicate/unknown/malformed outcomes.
// Resource services own operation permissions and any safe item-specific retry.
func ExecuteBatch(ctx context.Context, scope *client.Scope, version string, requests []BatchRequest) ([]BatchResponse, *resty.Response, error) {
	if scope == nil || scope.Backend() != client.Graph || (version != "v1.0" && version != "beta") || len(requests) == 0 || len(requests) > 20 {
		return nil, nil, ErrBatchRequest
	}
	environment, err := scope.Session().Environment()
	if err != nil {
		return nil, nil, err
	}
	ids := make(map[string]int, len(requests))
	for i, request := range requests {
		id := strings.ToLower(request.ID)
		if strings.TrimSpace(id) == "" || strings.ContainsFunc(id, unicode.IsControl) {
			return nil, nil, ErrBatchRequest
		}
		if _, exists := ids[id]; exists {
			return nil, nil, ErrBatchRequest
		}
		ids[id] = i
		u, err := url.Parse(request.URL)
		if err != nil || u.IsAbs() || u.Host != "" || u.User != nil || u.Fragment != "" || u.Opaque != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
			return nil, nil, ErrBatchRequest
		}
		first := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")[0]
		if first == "v1.0" || first == "beta" || strings.Contains(u.Path, "$batch") {
			return nil, nil, ErrBatchRequest
		}
		headers := make(http.Header, len(request.Headers))
		seen := make(map[string]bool, len(request.Headers))
		for name, value := range request.Headers {
			key := http.CanonicalHeaderKey(name)
			if seen[key] {
				return nil, nil, ErrBatchRequest
			}
			seen[key] = true
			headers.Set(key, value)
		}
		if len(request.Body) > 0 && (!json.Valid(request.Body) || headers.Get("Content-Type") == "") {
			return nil, nil, ErrBatchRequest
		}
		if err := scope.ValidateRequest(client.Request{Method: request.Method, URL: environment.Endpoints.Graph + "/" + version + request.URL, Header: headers}); err != nil {
			return nil, nil, errors.Join(ErrBatchRequest, err)
		}
	}
	if !validBatchDependencies(requests, ids) {
		return nil, nil, ErrBatchRequest
	}
	response, err := scope.Do(ctx, client.Request{Method: http.MethodPost, Path: []string{version, "$batch"}, Body: struct {
		Requests []BatchRequest `json:"requests"`
	}{requests}, Replay: client.ReplayNever}, nil)
	if err != nil {
		return nil, response, err
	}
	var envelope struct {
		Responses []json.RawMessage `json:"responses"`
	}
	if json.Unmarshal(response.Body(), &envelope) != nil || envelope.Responses == nil {
		return nil, response, ErrBatchResponse
	}
	results := make([]BatchResponse, 0, len(envelope.Responses))
	seen := make(map[string]bool, len(ids))
	var outcome error
	for _, raw := range envelope.Responses {
		var item BatchResponse
		if json.Unmarshal(raw, &item) != nil {
			outcome = errors.Join(outcome, ErrBatchResponse)
			continue
		}
		id := strings.ToLower(item.ID)
		_, known := ids[id]
		if !known || seen[id] || item.Status < 200 || item.Status > 599 {
			outcome = errors.Join(outcome, ErrBatchResponse)
			continue
		}
		seen[id] = true
		results = append(results, item)
		if item.Status >= 300 {
			outcome = errors.Join(outcome, ErrBatchItems)
		}
	}
	if len(seen) != len(ids) {
		outcome = errors.Join(outcome, ErrBatchResponse)
	}
	return results, response, outcome
}

func validBatchDependencies(requests []BatchRequest, ids map[string]int) bool {
	state := make([]uint8, len(requests))
	var visit func(int) bool
	visit = func(i int) bool {
		if state[i] == 1 {
			return false
		}
		if state[i] == 2 {
			return true
		}
		state[i] = 1
		seen := make(map[string]bool, len(requests[i].DependsOn))
		for _, id := range requests[i].DependsOn {
			id = strings.ToLower(id)
			index, exists := ids[id]
			if !exists || seen[id] || !visit(index) {
				return false
			}
			seen[id] = true
		}
		state[i] = 2
		return true
	}
	for i := range requests {
		if !visit(i) {
			return false
		}
	}
	return true
}
