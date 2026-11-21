// Package region reads Multi-Geo user/group location observations. These are
// independently scoped delegated reads, not Graph directory mutations.
package region

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Location retains nullable/future data. Timestamp text is preserved without
// imposing an unproven wire timezone format. Region alone is not resource CRUD.
type Location struct {
	Region                *string                    `json:"Region"`
	LastMovementTimestamp *string                    `json:"LastMovementTimeStamp"`
	PreferredDataLocation *string                    `json:"PreferredDataLocation"`
	RawFields             map[string]json.RawMessage `json:"-"`
}

func (v *Location) UnmarshalJSON(data []byte) error {
	type snapshot Location
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Location(value)
	return nil
}
func (Location) String() string     { return "multi-geo location observation (redacted)" }
func (v Location) GoString() string { return v.String() }

type Service struct{ scope *client.Scope }

func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) get(ctx context.Context, kind, id string) (*Location, *resty.Response, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return nil, nil, client.ErrClosed
	}
	if s.scope.Backend() != client.Regions {
		return nil, nil, client.ErrDestination
	}
	if strings.TrimSpace(id) == "" || id == "." || id == ".." || strings.ContainsRune(id, '\\') || strings.ContainsFunc(id, unicode.IsControl) {
		return nil, nil, client.ErrRequest
	}
	response, err := s.scope.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"api", "tms", "prod", kind, id, "region"}, Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}, nil)
	if err != nil {
		return nil, response, err
	}
	var value Location
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// GetForUser addresses one supplied identity without Graph lookup. The region
// profile enforces delegated mode and exact Region.Read.All before credentials.
func (s *Service) GetForUser(ctx context.Context, id string) (*Location, *resty.Response, error) {
	return s.get(ctx, "users", id)
}
func (s *Service) GetForGroup(ctx context.Context, id string) (*Location, *resty.Response, error) {
	return s.get(ctx, "groups", id)
}
