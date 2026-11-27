package assignmentmigration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policytransport"
	"github.com/go-resty/resty/v2"
)

var groupGUID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// Service binds direct ConfigAPI GPA operations. It never performs generated
// forest discovery, resolves group IDs through Graph or guesses a feature flight.
type Service struct{ client *policytransport.Client }

func NewService(scope *client.Scope) *Service { return &Service{client: policytransport.New(scope)} }
func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return client.ErrClosed
	}
	return s.client.Ready(ctx)
}
func validID(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." && !strings.ContainsRune(value, '\\') && !strings.ContainsFunc(value, unicode.IsControl)
}
func path(parts ...string) []string {
	return append([]string{"Skype.Policy", "groupPolicyAssignments"}, parts...)
}

func decode[T any](response *resty.Response) (*T, error) {
	if response == nil {
		return nil, client.ErrDecode
	}
	var value *T
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, client.ErrDecode
	}
	return value, nil
}
func decodeList[T any](response *resty.Response) ([]*T, error) {
	if response == nil {
		return nil, client.ErrDecode
	}
	var values []*T
	if json.Unmarshal(response.Body(), &values) != nil {
		return nil, client.ErrDecode
	}
	return values, nil
}

// GetByID reads one opaque migration ID. Empty is not a list alias; callers
// choose List explicitly. Returned state never triggers polling or completion.
func (s *Service) GetByID(ctx context.Context, id string) (*Job, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(id) {
		return nil, nil, ErrInput
	}
	r, err := s.client.Do(ctx, "Get-DirectToGroupAssignmentsMigrationStatus", http.MethodGet, path("directAssignmentsMigration", id), nil, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := decode[Job](r)
	return v, r, err
}

// List reads one bare migration array without a fabricated cursor or envelope.
// Non-success/malformed responses remain errors instead of PowerShell messages.
func (s *Service) List(ctx context.Context) ([]*Job, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	r, err := s.client.Do(ctx, "Get-DirectToGroupAssignmentsMigrationStatus", http.MethodGet, path("directAssignmentsMigration"), nil, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := decodeList[Job](r)
	return v, r, err
}

func threshold(options RecommendationOptions) url.Values {
	value := "500"
	if options.GroupThreshold != nil {
		value = *options.GroupThreshold
	}
	return url.Values{"groupThreshold": {value}}
}

func (s *Service) RecommendationsByPolicyType(ctx context.Context, entityType string, options RecommendationOptions) ([]*PolicyTypeRecommendation, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(entityType) {
		return nil, nil, ErrInput
	}
	r, err := s.client.Do(ctx, "Get-GroupAssignmentRecommendationsPerPolicyType", http.MethodGet, path(entityType, "recommendations"), threshold(options), nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := decodeList[PolicyTypeRecommendation](r)
	return v, r, err
}
func (s *Service) RecommendationsByPolicyName(ctx context.Context, entityType, policyType string, options RecommendationOptions) ([]*PolicyNameRecommendation, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(entityType) || !validID(policyType) {
		return nil, nil, ErrInput
	}
	r, err := s.client.Do(ctx, "Get-GroupAssignmentRecommendationsPerPolicyName", http.MethodGet, path(entityType, "policyTypes", policyType, "recommendations"), threshold(options), nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := decodeList[PolicyNameRecommendation](r)
	return v, r, err
}
func (s *Service) GetConflicts(ctx context.Context, groupID, policyType string) (*Conflict, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validID(groupID) || !validID(policyType) {
		return nil, nil, ErrInput
	}
	r, err := s.client.Do(ctx, "Get-GroupPolicyAssignmentConflict", http.MethodGet, path(groupID, "policyTypes", policyType, "conflict"), nil, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := decode[Conflict](r)
	return v, r, err
}

func (s *Service) submit(ctx context.Context, operation, command string, input CreateRequest) (*Job, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.Authority == "" || input.PolicyType == "" || input.PolicyName == "" || !groupGUID.MatchString(input.GroupID) {
		return nil, nil, ErrInput
	}
	input.GroupID = strings.ToLower(input.GroupID)
	allow := "False"
	if input.AllowNonEmptyGroup {
		allow = "True"
	}
	r, err := s.client.Do(ctx, command, http.MethodPost, path("directAssignmentsMigration", operation), url.Values{"allowNonEmptyGroup": {allow}}, input, client.ReplayNever)
	if err != nil {
		return nil, r, err
	}
	v, err := decode[Job](r)
	return v, r, err
}

// Create submits a migration once. HTTP acceptance does not prove migration or
// tenant convergence; no automatic retry, wait or assignment rollback occurs.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*Job, *resty.Response, error) {
	return s.submit(ctx, "migrate", "Invoke-StartDirectToGroupAssignmentMigration", input)
}

// ClearDirectAssignments submits the consequential cleanup action once. It is
// not migration-job Delete and never deletes a group or implicitly compensates.
func (s *Service) ClearDirectAssignments(ctx context.Context, input CreateRequest) (*Job, *resty.Response, error) {
	return s.submit(ctx, "cleanup", "Invoke-ClearDirectToGroupAssignmentMigration", input)
}
