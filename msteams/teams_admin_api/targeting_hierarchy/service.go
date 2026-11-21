package targetinghierarchy

import (
	"context"
	"errors"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Service composes version selection and the v2 existence/create/import flow.
// Native primitives remain independently accessible without implicit preparation.
type Service struct {
	session *client.Session
	v1      *V1Service
	v2      *V2Service
}

// NewService binds both known profiles locally. An unavailable cloud map is
// rejected before credentials when that profile is used, not at root construction.
func NewService(session *client.Session) (*Service, error) {
	v1, err := client.NewScope(client.Targeting, session)
	if err != nil {
		return nil, err
	}
	v2, err := client.NewScope(client.Hierarchy, session)
	if err != nil {
		return nil, err
	}
	return &Service{session: session, v1: NewV1Service(v1), v2: NewV2Service(v2)}, nil
}

func (s *Service) NativeV1() *V1Service {
	if s == nil {
		return nil
	}
	return s.v1
}
func (s *Service) NativeV2() *V2Service {
	if s == nil {
		return nil
	}
	return s.v2
}

func (s *Service) begin(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s == nil || s.session == nil {
		return nil, nil, client.ErrClosed
	}
	return s.session.Begin(ctx)
}

func selectedVersion(value Version) (Version, error) {
	if value == "" {
		value = V1
	}
	if value != V1 && value != V2 {
		return value, ErrVersion
	}
	return value, nil
}

// UploadCSV implements Set-TeamTargetingHierarchy. Version zero selects v1.
// Body ownership is immediate, including invalid options and failed preparation.
// V2 creates only after a received 404; every other failed read stops. Successful
// create plus failed import remains explicit; no automatic cleanup/replay occurs.
func (s *Service) UploadCSV(ctx context.Context, input CSV, version Version) (result UploadResult, err error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return result, rejectCSV(ctx, input, err)
	}
	defer stop()
	version, err = selectedVersion(version)
	result.Version = version
	if err != nil {
		return result, rejectCSV(ctx, input, err)
	}
	if _, err := csvPart(input); err != nil {
		return result, rejectCSV(ctx, input, err)
	}
	if version == V1 {
		if err := s.v1.ready(ctx); err != nil {
			return result, rejectCSV(ctx, input, err)
		}
	}
	if version == V2 {
		result.LookupResponse, err = s.v2.GetHierarchy(ctx)
		result.Response = result.LookupResponse
		var apiError *client.APIError
		if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
			result.CreateAttempted = true
			result.CreationResponse, err = s.v2.CreateHierarchy(ctx)
			result.Response = result.CreationResponse
			if err != nil {
				return result, rejectCSV(ctx, input, err)
			}
			result.CreationHTTPAccepted = true
		} else if err != nil {
			return result, rejectCSV(ctx, input, err)
		} else {
			result.LookupHTTPAccepted = true
		}
	}
	result.ImportAttempted = true
	if version == V1 {
		result.Import, result.Response, err = s.v1.ImportCSV(ctx, input)
	} else {
		result.Import, result.Response, err = s.v2.ImportCSV(ctx, input)
	}
	result.ImportHTTPAccepted = result.Response != nil && result.Response.StatusCode() >= 200 && result.Response.StatusCode() < 300
	return result, err
}

// Delete dispatches Remove-TeamTargetingHierarchy to one native singleton delete.
// Version zero selects v1. It does not require a preceding read or accept a caller
// tenant ID, and never restores the tenant's prior hierarchy automatically.
func (s *Service) Delete(ctx context.Context, version Version) (*resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer stop()
	version, err = selectedVersion(version)
	if err != nil {
		return nil, err
	}
	if version == V1 {
		return s.v1.Delete(ctx)
	}
	return s.v2.Delete(ctx)
}

// GetStatus performs one read and the pinned public projection. It never polls,
// starts/replays an import or borrows Graph's operation polling interval.
// Projection errors retain the readable native operation and response metadata.
func (s *Service) GetStatus(ctx context.Context, requestID string, version Version) (*Status, *resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer stop()
	version, err = selectedVersion(version)
	if err != nil {
		return nil, nil, err
	}
	if version == V1 {
		value, response, err := s.v1.GetOperation(ctx, requestID)
		if err != nil {
			return nil, response, err
		}
		status, err := projectRetail(value)
		return status, response, err
	}
	value, response, err := s.v2.GetOperation(ctx, requestID)
	if err != nil {
		return nil, response, err
	}
	status, err := projectHierarchy(value)
	return status, response, err
}
