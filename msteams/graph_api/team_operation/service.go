package teamoperation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("team operation returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

const minimumPollInterval = 31 * time.Second

// Service owns team-operation status reads on the explicit beta route.
type Service struct{ scope *client.Scope }

// NewService constructs a local dependency without tokens, HTTP or discovery.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) local() error {
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.local()
}

func request(teamID, operationID string) (client.Request, error) {
	if !directoryID.MatchString(teamID) || strings.TrimSpace(operationID) == "" || operationID == "." || operationID == ".." || strings.ContainsRune(operationID, '\\') || strings.ContainsFunc(operationID, unicode.IsControl) {
		return client.Request{}, client.ErrRequest
	}
	return client.Request{Method: http.MethodGet, Path: []string{"beta", "teams", teamID, "operations", operationID}, Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe, MinimumRetryDelay: minimumPollInterval}, nil
}

// GetByID performs one status-resource read with bounded shared read retries.
// Retries observe the Graph polling minimum. Operation IDs remain opaque single
// segments; target URLs and response headers never redirect this request.
func (s *Service) GetByID(ctx context.Context, teamID, operationID string) (*Operation, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := request(teamID, operationID)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, response, ErrUnexpectedStatus
	}
	var value Operation
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

func decodeStatus(data []byte) (client.OperationStatus, error) {
	var value Operation
	if json.Unmarshal(data, &value) != nil {
		return client.OperationStatus{}, client.ErrDecode
	}
	state, err := value.State()
	if err != nil {
		return client.OperationStatus{}, err
	}
	status := client.OperationStatus{ID: value.ID, State: state}
	if value.TargetResourceLocation != nil {
		status.TargetLocation = *value.TargetResourceLocation
	}
	return status, nil
}

// ResumeByID constructs a local polling handle from explicit resource IDs in
// this client's tenant/cloud. The first observation is immediate; subsequent
// polls and transport retries wait at least 31 seconds. It never replays the
// initiating mutation or follows an arbitrary saved/target URL.
func (s *Service) ResumeByID(teamID, operationID string) (*client.Operation, error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	req, err := request(teamID, operationID)
	if err != nil {
		return nil, err
	}
	return client.NewOperation(s.scope, req, operationID, decodeStatus, minimumPollInterval, nil)
}

// WaitByID explicitly composes a beta status handle and the shared bounded
// waiter. A local cancellation/limit retains its snapshot for private recovery.
// GetByID itself never waits for terminal completion.
func (s *Service) WaitByID(ctx context.Context, teamID, operationID string, options client.WaitOptions) (client.OperationSnapshot, error) {
	if err := s.ready(ctx); err != nil {
		return client.OperationSnapshot{}, err
	}
	if err := options.Validate(); err != nil {
		return client.OperationSnapshot{}, err
	}
	op, err := s.ResumeByID(teamID, operationID)
	if err != nil {
		return client.OperationSnapshot{}, err
	}
	return op.Wait(ctx, options)
}
