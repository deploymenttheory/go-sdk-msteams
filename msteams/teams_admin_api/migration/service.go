// Package migration reads meeting migration status and history and exposes the
// module's meeting, tenant and AVS partition migration operations.
//
// Start-CsExMeetingMigration, Move-CsTenantCrossRegion,
// Move-CsTenantServiceInstance and Move-CsAvsTenantPartition change where a
// user's meetings or a tenant's data live. They cannot be undone through this
// SDK or the module, are never replayed automatically, and most are
// Microsoft-internal (SelfHost/Torus) operations that ordinary tenant
// administrators are denied. Read the package documentation before calling them.
package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	mms    = "Teams.Mms"
	tenant = "Teams.Tenant"
	// dateLayout is .NET's invariant DateTime.ToString() form, which PowerShell
	// uses when the remoting layer binds a DateTime to the string parameters.
	dateLayout = "01/02/2006 15:04:05"
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

func validOptional(value *string) bool { return value == nil || validText(*value) }

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

// decodeArray decodes a JSON array whose items must all be objects.
func decodeArray[T any](raw []byte) ([]*T, error) {
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

func setTime(query url.Values, name string, value *time.Time) {
	if value != nil {
		query.Set(name, value.Format(dateLayout))
	}
}

func statusQuery(options *StatusOptions) (url.Values, bool) {
	query := url.Values{}
	if options == nil {
		return query, true
	}
	if !validOptional(options.Identity) || !validOptional(options.State) || !validOptional(options.MigrationType) {
		return nil, false
	}
	for name, value := range map[string]*string{"identity": options.Identity, "state": options.State, "migrationType": options.MigrationType} {
		if value != nil {
			query.Set(name, *value)
		}
	}
	setTime(query, "startTime", options.StartTime)
	setTime(query, "endTime", options.EndTime)
	return query, true
}

// ListMeetingMigrationStatuses reads per-user meeting migration records
// (Get-CsMeetingMigrationStatus). Nil options read every record the service
// returns for the tenant. No match is a non-nil empty slice.
func (s *Service) ListMeetingMigrationStatuses(ctx context.Context, options *StatusOptions) ([]*MeetingMigrationStatus, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query, ok := statusQuery(options)
	if !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{mms, "GetMeetingMigrationStatus"}, Query: query,
		Header: headers("Get-CsMeetingMigrationStatus"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeArray[MeetingMigrationStatus](response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetMeetingMigrationStatusSummary reads user counts per migration state
// (Get-CsMeetingMigrationStatus -SummaryOnly). The module prints this object
// as a JSON string; the SDK returns it decoded.
func (s *Service) GetMeetingMigrationStatusSummary(ctx context.Context, options *StatusOptions) (*MeetingMigrationStatusSummary, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query, ok := statusQuery(options)
	if !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{mms, "GetMeetingMigrationStatusSummary"}, Query: query,
		Header: headers("Get-CsMeetingMigrationStatus"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeObject[MeetingMigrationStatusSummary](response.Body())
	if err != nil {
		return nil, response, err
	}
	for _, item := range value.SummaryOutput {
		if item == nil {
			return nil, response, client.ErrDecode
		}
	}
	return value, response, nil
}

// ListMeetingMigrationTransactionHistory reads one user's meeting migration
// history (Get-CsMeetingMigrationTransactionHistory). identity is the user's
// object ID, user principal name or SIP address.
func (s *Service) ListMeetingMigrationTransactionHistory(ctx context.Context, identity string, options *TransactionHistoryOptions) ([]*MeetingMigrationTransaction, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(identity) || (options != nil && !validOptional(options.CorrelationID)) {
		return nil, nil, ErrInput
	}
	query := url.Values{"userIdentity": {identity}}
	if options != nil {
		if options.CorrelationID != nil {
			query.Set("CorrelationId", *options.CorrelationID)
		}
		setTime(query, "startTime", options.StartTime)
		setTime(query, "endTime", options.EndTime)
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{mms, "GetMeetingMigrationTransactionHistory"}, Query: query,
		Header: headers("Get-CsMeetingMigrationTransactionHistory"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeArray[MeetingMigrationTransaction](response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// StartMeetingMigration queues a migration of one user's meetings
// (Start-CsExMeetingMigration). Migrated meetings are recreated and attendees
// receive updated invitations; the operation cannot be undone. It is never
// replayed automatically: check ListMeetingMigrationStatuses before retrying.
// The result is the HTTP 201 receipt, or nil for HTTP 204.
func (s *Service) StartMeetingMigration(ctx context.Context, request StartMeetingMigrationRequest) (*MeetingMigrationEnqueueResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Identity) || !validOptional(request.SourceMeetingType) || !validOptional(request.TargetMeetingType) {
		return nil, nil, ErrInput
	}
	body := map[string]string{"Identity": request.Identity}
	if request.SourceMeetingType != nil {
		body["SourceMeetingType"] = *request.SourceMeetingType
	}
	if request.TargetMeetingType != nil {
		body["TargetMeetingType"] = *request.TargetMeetingType
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{mms, "StartMeetingMigration"},
		Header: headers("Start-CsExMeetingMigration"), Body: body}, http.StatusCreated, http.StatusNoContent)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() == http.StatusNoContent || len(bytes.TrimSpace(response.Body())) == 0 {
		return nil, response, nil
	}
	value, err := decodeObject[MeetingMigrationEnqueueResult](response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// ListTenantMoveTaskStatuses reads tenant migration detail records
// (Get-CsMoveTenantServiceInstanceTaskStatus). Records are free-form.
func (s *Service) ListTenantMoveTaskStatuses(ctx context.Context) ([]*TenantMoveTaskStatus, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	// The generated route ends with a slash; the gateway routes the path with
	// and without it to the same resource (both HTTP 403 for the lab, against
	// 404 for an unknown route), and the transport sends no empty segments.
	response, err := s.do(ctx, client.Request{Method: http.MethodGet, Path: []string{tenant, "TenantMigrationDetail"},
		Header: headers("Get-CsMoveTenantServiceInstanceTaskStatus"), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeArray[TenantMoveTaskStatus](response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

func (s *Service) moveTenant(ctx context.Context, command string, body map[string]string) (*TenantMigrationResult, *resty.Response, error) {
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{tenant, "CrossRegionMoveTask"},
		Header: headers(command), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeObject[TenantMigrationResult](response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// MoveTenantCrossRegion starts moving the calling tenant to another region
// (Move-CsTenantCrossRegion). The request body is empty: the service moves the
// tenant the credential belongs to. This is an irreversible tenant-level
// operation reserved for Microsoft-internal environments. It is never replayed
// automatically.
func (s *Service) MoveTenantCrossRegion(ctx context.Context) (*TenantMigrationResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	return s.moveTenant(ctx, "Move-CsTenantCrossRegion", map[string]string{})
}

// MoveTenantServiceInstance runs one phase of moving the calling tenant to
// another service instance (Move-CsTenantServiceInstance). This is an
// irreversible tenant-level operation reserved for Microsoft-internal
// environments. It is never replayed automatically.
func (s *Service) MoveTenantServiceInstance(ctx context.Context, request MoveTenantServiceInstanceRequest) (*TenantMigrationResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.MoveOption) || (request.TargetServiceInstance != nil && strings.ContainsFunc(*request.TargetServiceInstance, unicode.IsControl)) {
		return nil, nil, ErrInput
	}
	target := ""
	if request.TargetServiceInstance != nil {
		target = *request.TargetServiceInstance
	}
	return s.moveTenant(ctx, "Move-CsTenantServiceInstance", map[string]string{"MoveOption": request.MoveOption, "TargetServiceInstance": target})
}

// MoveAVSTenantPartition moves documents between AVS storage partitions
// (Move-CsAvsTenantPartition). The service answers HTTP 207 with one result
// per document; check each IsSuccessful. Moves cannot be undone through the
// module and the operation is reserved for Microsoft-internal environments. It
// is never replayed automatically.
func (s *Service) MoveAVSTenantPartition(ctx context.Context, request PartitionMoveRequest) ([]*PartitionMoveResult, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.SourcePartitionKey) || !validText(request.TargetPartitionKey) ||
		!validOptional(request.Workload) || !validOptional(request.ContainerName) || !validOptional(request.BasePartitionKey) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"SourcePartitionKey": request.SourcePartitionKey, "TargetPartitionKey": request.TargetPartitionKey}
	for name, value := range map[string]*string{"Workload": request.Workload, "ContainerName": request.ContainerName, "BasePartitionKey": request.BasePartitionKey} {
		if value != nil {
			body[name] = *value
		}
	}
	for name, value := range map[string]*int64{"NumberOfDocuments": request.NumberOfDocuments, "PercentageOfPartition": request.PercentageOfPartition, "BatchSize": request.BatchSize} {
		if value != nil {
			body[name] = *value
		}
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{"Teams.AVS", "MovePartition"},
		Header: headers("Move-CsAvsTenantPartition"), Body: body}, http.StatusMultiStatus)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeArray[PartitionMoveResult](response.Body())
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}
