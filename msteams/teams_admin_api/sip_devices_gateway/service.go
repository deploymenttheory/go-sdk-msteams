// Package sipdevicesgateway drives the SIP devices gateway (SDG): bulk
// sign-in of SIP phones, batch status and summaries, and the support-only
// device tagging and transfer requests.
package sipdevicesgateway

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	root          = "Teams.SipDevicesGatewayManagement"
	byteOrderMark = string(rune(0xFEFF))
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

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func call[T any](ctx context.Context, s *Service, req client.Request) (*T, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response == nil || response.StatusCode() != http.StatusOK {
		return nil, response, ErrResponse
	}
	var value *T
	if json.Unmarshal(response.Body(), &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// ListBulkSignInRequests reads the tenant's bulk sign-in batches
// (Get-CsSdgBulkSignInRequestsSummary).
func (s *Service) ListBulkSignInRequests(ctx context.Context) (*BulkSignInRequestsSummary, *resty.Response, error) {
	return call[BulkSignInRequestsSummary](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, "v1", "batchsummaries"},
		Header: headers("Get-CsSdgBulkSignInRequestsSummary"), Replay: client.ReplaySafe})
}

// GetBulkSignInRequestStatus reads one batch and its per-device results
// (Get-CsSdgBulkSignInRequestStatus -Batchid).
func (s *Service) GetBulkSignInRequestStatus(ctx context.Context, batchID string) (*BulkSignInRequestStatus, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(batchID) == "" {
		return nil, nil, ErrInput
	}
	return call[BulkSignInRequestStatus](ctx, s, client.Request{Method: http.MethodGet, Path: []string{root, "v1", "batchstatus"},
		Query: url.Values{"batchid": {batchID}}, Header: headers("Get-CsSdgBulkSignInRequestStatus"), Replay: client.ReplaySafe})
}

// CreateBulkSignInRequest submits one batch of devices to sign in within the
// region (New-CsSdgBulkSignInRequest). The POST is never replayed; reconcile
// a lost response with ListBulkSignInRequests.
func (s *Service) CreateBulkSignInRequest(ctx context.Context, region string, items []BulkSignInItem) (*BulkSignInResponse, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(region) == "" || len(items) == 0 {
		return nil, nil, ErrInput
	}
	return call[BulkSignInResponse](ctx, s, client.Request{Method: http.MethodPost, Path: []string{root, "v1", "createbatch"},
		Query: url.Values{"TargetRegion": {region}}, Header: headers("New-CsSdgBulkSignInRequest"), Body: items})
}

// ReadBulkSignInItems parses the CSV that New-CsSdgBulkSignInRequest's
// -DeviceDetailsFilePath names: a header row with Username and HardwareId
// columns (any case, any order, extra columns ignored). A missing column
// leaves that field nil, as Import-Csv yields null.
func ReadBulkSignInItems(r io.Reader) ([]BulkSignInItem, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, errors.Join(ErrInput, err)
	}
	user, hardware := -1, -1
	for index, name := range header {
		switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, byteOrderMark))) {
		case "username":
			if user < 0 {
				user = index
			}
		case "hardwareid":
			if hardware < 0 {
				hardware = index
			}
		}
	}
	var items []BulkSignInItem
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return items, nil
		}
		if err != nil {
			return nil, errors.Join(ErrInput, err)
		}
		items = append(items, BulkSignInItem{UserName: column(record, user), HardwareID: column(record, hardware)})
	}
}

func column(record []string, index int) *string {
	if index < 0 || index >= len(record) {
		return nil
	}
	value := record[index]
	return &value
}

// CreateDeviceTaggingRequest tags a new device for support
// (New-CsSdgDeviceTaggingRequest). The module exports it only to support
// environments; tenant administrators are expected to be denied.
func (s *Service) CreateDeviceTaggingRequest(ctx context.Context, targetRegion string, input DeviceTaggingRequest) (*DeviceTaggingResponse, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(targetRegion) == "" {
		return nil, nil, ErrInput
	}
	return call[DeviceTaggingResponse](ctx, s, client.Request{Method: http.MethodPost, Path: []string{root, "devicetagging", "tagnewdevice"},
		Query: url.Values{"TargetRegion": {targetRegion}}, Header: headers("New-CsSdgDeviceTaggingRequest"), Body: input})
}

// CreateDeviceTransferRequest moves a device between regions
// (New-CsSdgDeviceTransferRequest). Support-only, like tagging. It has no body.
func (s *Service) CreateDeviceTransferRequest(ctx context.Context, input DeviceTransferRequest) (*DeviceTransferResponse, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(input.SdhDeviceID) == "" || strings.TrimSpace(input.SourceRegion) == "" || strings.TrimSpace(input.TargetRegion) == "" {
		return nil, nil, ErrInput
	}
	return call[DeviceTransferResponse](ctx, s, client.Request{Method: http.MethodPost, Path: []string{root, "devicetransfer", "transfer"},
		Query:  url.Values{"SourceRegion": {input.SourceRegion}, "TargetRegion": {input.TargetRegion}, "SdhDeviceId": {input.SdhDeviceID}},
		Header: headers("New-CsSdgDeviceTransferRequest")})
}
