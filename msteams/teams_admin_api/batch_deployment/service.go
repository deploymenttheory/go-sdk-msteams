package batchdeployment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

var (
	ErrInput    = errors.New("invalid batch deployment input")
	ErrResponse = errors.New("batch deployment returned an unexpected successful HTTP status")
)

type CreateRequest struct {
	Archive       Archive
	UsersToNotify string
}

// Submission preserves both service identities and future response fields.
// An accepted response is not workflow completion; InstanceID addresses status.
type Submission struct {
	InstanceID  *string                    `json:"instanceId"`
	ExecutionID *string                    `json:"executionId"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (Submission) String() string     { return "batch deployment submission (redacted)" }
func (s Submission) GoString() string { return s.String() }

// Status retains the original opaque status text. No terminal interpretation is
// inferred from Completed, which can include success, partial success or failure.
type Status struct{ Text string }

func (Status) String() string     { return "batch deployment status (redacted)" }
func (s Status) GoString() string { return s.String() }

// Identity represents the exported InputObject convenience. It is an operation
// selector, never a URI or permission to forward a credential elsewhere.
type Identity struct{ OrchestrationID string }

type Service struct{ client client.Client }

// NewService injects the selected ConfigAPI backend/discovery adapter locally.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) ready(ctx context.Context) error {
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
func (s *Service) do(ctx context.Context, command string, request client.Request) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	return s.client.Do(ctx, request, nil)
}

// Create sends one Base64 ZIP submission and requires 202. It does not replay,
// poll or synthesize a successful deployment. Decoding failure after acceptance
// preserves partial response fields and HTTP metadata for reconciliation.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*Submission, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if len(input.Archive.data) == 0 || strings.TrimSpace(input.UsersToNotify) == "" {
		return nil, nil, ErrInput
	}
	body := struct {
		DeploymentCSVs string `json:"DeploymentCsvs"`
		UsersToNotify  string `json:"UsersToNotify"`
	}{input.Archive.Encoded(), input.UsersToNotify}
	response, err := s.do(ctx, "New-CsBatchTeamsDeployment", client.Request{Method: http.MethodPost, Path: []string{"Teams.FLWOrchestrator", "orchestration", "v1.0", "batchflwdeployment"}, Body: body})
	if err != nil {
		return nil, response, err
	}
	if response == nil || response.StatusCode() != http.StatusAccepted {
		return nil, response, ErrResponse
	}
	var value Submission
	if json.Unmarshal(response.Body(), &value) != nil || json.Unmarshal(response.Body(), &value.RawFields) != nil || value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	if value.InstanceID == nil || strings.TrimSpace(*value.InstanceID) == "" {
		return &value, response, client.ErrDecode
	}
	return &value, response, nil
}

// GetByID performs one raw status read. Named 400/404/500 responses remain Go
// errors, with original text/status available in response.Body/metadata. No JSON
// status grammar, list, mutation or waiter is invented from display strings.
func (s *Service) GetByID(ctx context.Context, orchestrationID string) (*Status, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(orchestrationID) == "" || orchestrationID == "." || orchestrationID == ".." || strings.ContainsRune(orchestrationID, '\\') || strings.ContainsFunc(orchestrationID, unicode.IsControl) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, "Get-CsBatchTeamsDeploymentStatus", client.Request{Method: http.MethodGet, Path: []string{"Teams.FLWOrchestrator", "orchestration", "v1.0", "scenarios", orchestrationID}, Replay: client.ReplaySafe})
	if err != nil {
		return nil, response, err
	}
	if response == nil || response.StatusCode() != http.StatusOK {
		return nil, response, ErrResponse
	}
	return &Status{Text: string(response.Body())}, response, nil
}

func (s *Service) GetByIdentity(ctx context.Context, identity Identity) (*Status, *resty.Response, error) {
	return s.GetByID(ctx, identity.OrchestrationID)
}
