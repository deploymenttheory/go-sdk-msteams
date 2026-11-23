// Package provisioningdiagnostics wraps the module's internal provisioning,
// directory-sync, user-settings and hosted-migration diagnostics. These are
// Microsoft support (SelfHost, Torus and OCE) operations: tenant
// administrator credentials are denied, and several actions resynchronize,
// rehome or move users. Callers must own the decision to send them.
package provisioningdiagnostics

import (
	"bytes"
	"context"
	"encoding/hex"
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
	tenantRoute   = "Teams.Tenant"
	internalRoute = "Teams.InternalSupport"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
// OCE operations use the bound backend directly, without discovery.
type Service struct {
	client *configapi.Client
	base   client.Client
}

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service {
	generated := configapi.New(base)
	return &Service{client: generated, base: generated.Base()}
}

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil || s.base == nil {
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

// guid parses the forms .NET Guid accepts for these parameters (32 digits,
// optionally hyphenated and braced) and returns the lowercase hyphenated form
// the module sends after binding a System.Guid.
func guid(value string) (string, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}")
	if len(text) != len(value) && len(text) != len(value)-2 {
		return "", false
	}
	if len(text) == 36 {
		if text[8] != '-' || text[13] != '-' || text[18] != '-' || text[23] != '-' {
			return "", false
		}
		text = strings.ReplaceAll(text, "-", "")
	}
	if len(text) != 32 {
		return "", false
	}
	if _, err := hex.DecodeString(text); err != nil {
		return "", false
	}
	text = strings.ToLower(text)
	return text[:8] + "-" + text[8:12] + "-" + text[12:16] + "-" + text[16:20] + "-" + text[20:], true
}

// validBody accepts the JSON text the module forwards from its string Body
// parameters. The transport sends it compacted.
func validBody(body json.RawMessage) bool {
	return len(bytes.TrimSpace(body)) != 0 && json.Valid(body)
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func accepted(response *resty.Response, statuses []int) bool {
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return true
		}
	}
	return false
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	if !accepted(response, statuses) {
		return response, ErrResponse
	}
	return response, nil
}

// decodeList decodes a JSON array of objects, rejecting null and null items.
func decodeList[T any](raw []byte) ([]*T, error) {
	var values []*T
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, client.ErrDecode
		}
	}
	return values, nil
}

func decodeObject[T any](raw []byte) (*T, error) {
	var value *T
	if json.Unmarshal(raw, &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

// GetDirectoryDiagnosticData reads Business Voice Directory table entities
// (Get-CsBusinessVoiceDirectoryDiagnosticData).
func (s *Service) GetDirectoryDiagnosticData(ctx context.Context, query DirectoryDiagnosticQuery) ([]*DirectoryDiagnosticEntity, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(query.Table) || !validPart(query.Region) || !validPart(query.PartitionKey) || (query.RowKey != nil && !validText(*query.RowKey)) {
		return nil, nil, ErrInput
	}
	values := url.Values{}
	if query.RowKey != nil {
		values.Set("rowKey", *query.RowKey)
	}
	if query.ResultSize != nil {
		values.Set("resultSize", strconv.FormatInt(int64(*query.ResultSize), 10))
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{"Teams.Cpc", "bvdDiagnosticData", query.Table, query.Region, query.PartitionKey},
		Query: values, Header: headers("Get-CsBusinessVoiceDirectoryDiagnosticData"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	entities, err := decodeList[DirectoryDiagnosticEntity](response.Body())
	if err != nil {
		return nil, response, err
	}
	return entities, response, nil
}

func optionalQuery(values url.Values, name, value string) bool {
	if value == "" {
		return true
	}
	values.Set(name, value)
	return validText(value)
}

// ListObjectChangelog reads MAS object change history
// (Get-CsMasObjectChangelog).
func (s *Service) ListObjectChangelog(ctx context.Context, query ObjectChangelogQuery) ([]*ChangelogItem, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	values := url.Values{}
	if !optionalQuery(values, "Identity", query.Identity) || !optionalQuery(values, "SchemaNames", query.SchemaName) || !optionalQuery(values, "Versions", query.Version) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{tenantRoute, "GetMASObjectChangelog"}, Query: values,
		Header: headers("Get-CsMasObjectChangelog"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	items, err := decodeList[ChangelogItem](response.Body())
	if err != nil {
		return nil, response, err
	}
	return items, response, nil
}

// ListVersionedSchemaData reads versioned MAS schema documents
// (Get-CsMasVersionedSchemaData). Each item keeps every service field, as the
// module outputs each item's properties unchanged.
func (s *Service) ListVersionedSchemaData(ctx context.Context, query VersionedSchemaQuery) ([]map[string]json.RawMessage, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	values := url.Values{}
	if !validText(query.SchemaName) || !optionalQuery(values, "SchemaName", query.SchemaName) || !optionalQuery(values, "Identity", query.Identity) || !optionalQuery(values, "Versions", query.Version) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{tenantRoute, "GetSchemaDataFromMAS"}, Query: values,
		Header: headers("Get-CsMasVersionedSchemaData"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &items) != nil || items == nil {
		return nil, response, client.ErrDecode
	}
	for _, item := range items {
		if item == nil {
			return nil, response, client.ErrDecode
		}
	}
	return items, response, nil
}

func settingPath(tenantID, userID string, name ...string) ([]string, bool) {
	if !validPart(tenantID) || !validPart(userID) {
		return nil, false
	}
	return append([]string{"Teams.UserSettings", "api", "v1", "tenant", tenantID, "user", userID, "setting"}, name...), true
}

// GetUserSetting reads one user setting from the user settings service
// (Get-CsUssUserSettings) and returns the response text.
func (s *Service) GetUserSetting(ctx context.Context, tenantID, userID, name string) (string, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return "", nil, err
	}
	path, ok := settingPath(tenantID, userID)
	if !ok || !validText(name) {
		return "", nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: path, Query: url.Values{"name": {name}},
		Header: headers("Get-CsUssUserSettings"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return "", response, err
	}
	return string(response.Body()), response, nil
}

// SetUserSetting patches one user setting (Set-CsUssUserSettings) with a JSON
// body and returns the response text. It is never replayed automatically.
func (s *Service) SetUserSetting(ctx context.Context, tenantID, userID, name string, body json.RawMessage) (string, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return "", nil, err
	}
	path, ok := settingPath(tenantID, userID, name)
	if !ok || !validPart(name) || !validBody(body) {
		return "", nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPatch, Path: path, Body: body, Header: headers("Set-CsUssUserSettings")}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return "", response, err
	}
	return string(response.Body()), response, nil
}

// InvokeCustomHandlerCallback runs a management operation on a provisioning
// custom handler instance (Invoke-CsCustomHandlerCallBackNgtprov). The result
// keeps every service field.
func (s *Service) InvokeCustomHandlerCallback(ctx context.Context, request CustomHandlerCallbackRequest) (map[string]json.RawMessage, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	values := url.Values{"id": {request.ID}, "operation": {string(request.Operation)}}
	if !validText(request.ID) || !validText(string(request.Operation)) || !optionalQuery(values, "eventname", request.EventName) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{tenantRoute, "CustomHandlerCallback"}, Query: values,
		Header: headers("Invoke-CsCustomHandlerCallBackNgtprov")}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var result map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &result) != nil || result == nil {
		return nil, response, client.ErrDecode
	}
	return result, response, nil
}

// InvokeCustomHandler starts a provisioning custom handler
// (Invoke-CsCustomHandlerNgtprov). It is never replayed automatically.
func (s *Service) InvokeCustomHandler(ctx context.Context, payload CustomHandlerPayload) (*CustomHandlerResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{tenantRoute, "CustomHandler"}, Body: payload,
		Header: headers("Invoke-CsCustomHandlerNgtprov")}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	result, err := decodeObject[CustomHandlerResult](response.Body())
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}

// SyncDirectObjects asks provisioning to resynchronize directory objects
// (Invoke-CsDirectObjectSync). Set IsValidationRequest to validate without
// syncing. It is never replayed automatically.
func (s *Service) SyncDirectObjects(ctx context.Context, request DirectObjectSyncRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{internalRoute, "ProvisioningOperations", "DsRequest"}, Body: request,
		Header: headers("Invoke-CsDirectObjectSync")}, http.StatusOK)
}

// SyncMsodsObjects asks the tenant service to resynchronize objects from the
// directory (Invoke-CsMsodsSync). Set IsValidationRequest to validate without
// syncing. It is never replayed automatically.
func (s *Service) SyncMsodsObjects(ctx context.Context, request MsodsSyncRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validText(request.TenantID) || !validText(request.ObjectClass) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{tenantRoute, "MsodsSync"}, Body: request,
		Header: headers("Invoke-CsMsodsSync")}, http.StatusOK)
}

// RehomeUser moves a user to another home (Invoke-CsRehomeuser) with the JSON
// body the caller supplies. It is never replayed automatically.
func (s *Service) RehomeUser(ctx context.Context, userID string, body json.RawMessage) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(userID) || !validBody(body) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{"Teams.User", "RehomeUser", userID, "moveUser"}, Body: body,
		Header: headers("Invoke-CsRehomeuser")}, http.StatusNoContent)
}

// SetTenantUserBackfill publishes a tenant user backfill operation
// (Set-CsTenantUserBackfill). It is never replayed automatically.
func (s *Service) SetTenantUserBackfill(ctx context.Context, operation string, body json.RawMessage) (*OperationResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(operation) || !validBody(body) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{internalRoute, "SetBackfillToPubsub"}, Query: url.Values{"operation": {operation}},
		Body: body, Header: headers("Set-CsTenantUserBackfill")}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	result, err := decodeObject[OperationResult](response.Body())
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}
