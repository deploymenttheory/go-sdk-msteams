// Package resourceaccounts manages Teams resource accounts (application
// instances): the accounts that give auto attendants and call queues their
// identity, and their associations with those configurations.
//
// Resource accounts are created with New-CsOnlineApplicationInstance and are
// Entra ID users; the module has no remove cmdlet, so deletion is a directory
// operation outside this package.
package resourceaccounts

import (
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
	platform     = "Teams.PlatformService"
	voiceApps    = "Teams.VoiceApps"
	instances    = "ApplicationInstances"
	associations = "applicationinstanceassociations"
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

// optionalGUID normalizes a non-nil GUID pointer.
func optionalGUID(value *string) (*string, bool) {
	if value == nil {
		return nil, true
	}
	normalized, ok := guid(*value)
	return &normalized, ok
}

// encodeSipURI mirrors the module's EncodeSipUri helper used for association
// path identities.
func encodeSipURI(value string) string {
	return strings.NewReplacer(":", "[COLON]", ".", "[DOT]").Replace(value)
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

func decodeInstance(response *resty.Response) (*ApplicationInstance, error) {
	var value ApplicationInstance
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}

func decodeInstances(raw []byte) ([]*ApplicationInstance, error) {
	var values []*ApplicationInstance
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

// envelope decodes a required object property from a voice apps response.
func envelope(raw []byte, name string, out any) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value, ok := fields[name]
	if !ok || json.Unmarshal(value, out) != nil {
		return client.ErrDecode
	}
	return nil
}

func decodeResults(raw []byte) ([]*AssociationOperationResult, error) {
	var values []*AssociationOperationResult
	if envelope(raw, "AssociationOperationResults", &values) != nil || values == nil {
		return nil, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, client.ErrDecode
		}
	}
	return values, nil
}

func pageQuery(options *ListOptions) url.Values {
	query := url.Values{}
	if options == nil {
		return query
	}
	if options.ResultSize != nil {
		query.Set("resultSize", strconv.FormatInt(int64(*options.ResultSize), 10))
	}
	if options.Skip != nil {
		query.Set("skip", strconv.FormatInt(int64(*options.Skip), 10))
	}
	return query
}

// List reads resource accounts (Get-CsOnlineApplicationInstance without
// Identity). Nil options read the service's default page.
func (s *Service) List(ctx context.Context, options *ListOptions) ([]*ApplicationInstance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{platform, "v2", "Legacy", instances}, Query: pageQuery(options),
		Header: headers("Get-CsOnlineApplicationInstance"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeInstances(response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetByID reads one resource account by object ID or user principal name
// (Get-CsOnlineApplicationInstance -Identity). A missing account is an HTTP
// 404 API error. The service may answer with the account or a one-item array;
// both are accepted.
func (s *Service) GetByID(ctx context.Context, identity string) (*ApplicationInstance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(identity) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{platform, "v2", "Legacy", instances}, Query: url.Values{"identity": {identity}},
		Header: headers("Get-CsOnlineApplicationInstance"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	if values, err := decodeInstances(response.Body()); err == nil {
		if len(values) != 1 {
			return nil, response, client.ErrDecode
		}
		return values[0], response, nil
	}
	value, err := decodeInstance(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// Search finds resource accounts by name (Find-CsOnlineApplicationInstance).
// No match is a non-nil empty slice.
func (s *Service) Search(ctx context.Context, request SearchRequest) ([]*SearchResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.SearchQuery) || (request.MaxResults != nil && *request.MaxResults < 0) {
		return nil, nil, ErrInput
	}
	query := url.Values{"SearchQuery": {request.SearchQuery}}
	if request.MaxResults != nil {
		query.Set("MaxResults", strconv.FormatInt(int64(*request.MaxResults), 10))
	}
	for name, value := range map[string]*bool{"ExactMatchOnly": request.ExactMatchOnly, "AssociatedOnly": request.AssociatedOnly, "UnAssociatedOnly": request.UnAssociatedOnly} {
		if value != nil {
			// .NET Boolean.ToString spelling.
			query.Set(name, map[bool]string{true: "True", false: "False"}[*value])
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, "applicationinstances"}, Query: query,
		Header: headers("Find-CsOnlineApplicationInstance"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*SearchResult
	if envelope(response.Body(), "ApplicationInstances", &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// Create creates a resource account (New-CsOnlineApplicationInstance). It is
// never replayed automatically; reconcile a lost response with GetByID on the
// user principal name. The service rejects application (app-only) callers
// without directory write rights with HTTP 401.
func (s *Service) Create(ctx context.Context, request CreateRequest) (*ApplicationInstance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, ok := optionalGUID(request.ApplicationID)
	if !ok || !validText(request.UserPrincipalName) || (request.DisplayName != nil && strings.ContainsFunc(*request.DisplayName, unicode.IsControl)) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"userPrincipalName": request.UserPrincipalName}
	if app != nil {
		body["applicationId"] = *app
	}
	if request.DisplayName != nil {
		body["displayName"] = *request.DisplayName
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{platform, "v2", instances},
		Header: headers("New-CsOnlineApplicationInstance"), Body: body}, http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeInstance(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// optionalResult decodes an account when the service returned a body.
func optionalResult(response *resty.Response) (*ApplicationInstance, error) {
	if len(strings.TrimSpace(string(response.Body()))) == 0 {
		return nil, nil
	}
	return decodeInstance(response)
}

// UpdateByID changes a resource account (Set-CsOnlineApplicationInstance).
// identity is the object ID or user principal name. The result is nil when the
// service returns no body.
func (s *Service) UpdateByID(ctx context.Context, identity string, request UpdateRequest) (*ApplicationInstance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	app, appOK := optionalGUID(request.ApplicationID)
	acs, acsOK := optionalGUID(request.AcsResourceID)
	if !appOK || !acsOK || !validPart(identity) {
		return nil, nil, ErrInput
	}
	body := map[string]string{}
	if app != nil {
		body["applicationId"] = *app
	}
	if request.DisplayName != nil {
		body["displayName"] = *request.DisplayName
	}
	if request.OnpremPhoneNumber != nil {
		body["onpremPhoneNumber"] = *request.OnpremPhoneNumber
	}
	if acs != nil {
		body["acsResourceId"] = *acs
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPatch, Path: []string{platform, "v2", instances, identity},
		Header: headers("Set-CsOnlineApplicationInstance"), Body: body}, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	value, err := optionalResult(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// UpdateTelephoneNumberByID is Set-CsOnlineVoiceApplicationInstance. A nil
// telephoneNumber omits the property. The service now rejects this operation
// as deprecated in favour of phone number assignment; the method exists for
// parity and surfaces that HTTP 400 as an API error.
func (s *Service) UpdateTelephoneNumberByID(ctx context.Context, identity string, telephoneNumber *string) (*ApplicationInstance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	body := map[string]string{}
	if telephoneNumber != nil {
		body["telephoneNumber"] = *telephoneNumber
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPatch, Path: []string{platform, "v2", instances, identity, "telephoneNumber"},
		Header: headers("Set-CsOnlineVoiceApplicationInstance"), Body: body}, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	value, err := optionalResult(response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// SyncByID asks the service to synchronize a resource account from the
// directory into the voice platform (Sync-CsOnlineApplicationInstance).
// objectID is the account's object ID GUID.
func (s *Service) SyncByID(ctx context.Context, objectID string, request SyncRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	id, idOK := guid(objectID)
	app, appOK := optionalGUID(request.ApplicationID)
	acs, acsOK := optionalGUID(request.AcsResourceID)
	if !idOK || !appOK || !acsOK {
		return nil, ErrInput
	}
	query := url.Values{}
	if app != nil {
		query.Set("applicationId", *app)
	}
	if acs != nil {
		query.Set("acsResourceId", *acs)
	}
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{platform, "v2", instances, id, "sync"}, Query: query,
		Header: headers("Sync-CsOnlineApplicationInstance")}, http.StatusOK, http.StatusNoContent)
}

// GetAssociationByID reads the association of one resource account
// (Get-CsOnlineApplicationInstanceAssociation). A resource account without an
// association is an HTTP 404 API error.
func (s *Service) GetAssociationByID(ctx context.Context, identity string) (*Association, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, associations, encodeSipURI(identity)},
		Header: headers("Get-CsOnlineApplicationInstanceAssociation"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value *Association
	if envelope(response.Body(), "ApplicationInstanceAssociation", &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// GetAssociationStatusByID reads the provisioning status of a resource
// account's association (Get-CsOnlineApplicationInstanceAssociationStatus).
// An unknown account is reported by the service as HTTP 200 with a failed
// StatusRecord, not as an HTTP error.
func (s *Service) GetAssociationStatusByID(ctx context.Context, identity string) (*StatusRecord, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, associations, identity, "status"},
		Header: headers("Get-CsOnlineApplicationInstanceAssociationStatus"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value *StatusRecord
	if envelope(response.Body(), "StatusRecord", &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// CreateAssociations associates resource accounts with an auto attendant or
// call queue (New-CsOnlineApplicationInstanceAssociation). It returns one result
// per account; per-account failures are reported in each StatusRecord with
// HTTP 200, not as an error. It is never replayed automatically.
func (s *Service) CreateAssociations(ctx context.Context, request AssociationRequest) ([]*AssociationOperationResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if len(request.EndpointIDs) == 0 || !validText(request.ConfigurationID) || !validText(request.ConfigurationType) ||
		(request.CallPriority != nil && (*request.CallPriority < 0 || *request.CallPriority > 5)) {
		return nil, nil, ErrInput
	}
	for _, id := range request.EndpointIDs {
		if !validText(id) {
			return nil, nil, ErrInput
		}
	}
	body := map[string]any{"EndpointsIds": request.EndpointIDs, "ConfigurationId": request.ConfigurationID, "ConfigurationType": request.ConfigurationType}
	if request.CallPriority != nil {
		body["CallPriority"] = int32(*request.CallPriority)
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, associations},
		Header: headers("New-CsOnlineApplicationInstanceAssociation"), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeResults(response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// DeleteAssociationByID removes one resource account's association. The
// module's Remove-CsOnlineApplicationInstanceAssociation sends one such request
// per identity. A missing association is reported in the StatusRecord with
// HTTP 200.
func (s *Service) DeleteAssociationByID(ctx context.Context, identity string) ([]*AssociationOperationResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(identity) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{voiceApps, associations, encodeSipURI(identity)},
		Header: headers("Remove-CsOnlineApplicationInstanceAssociation")}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeResults(response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}
