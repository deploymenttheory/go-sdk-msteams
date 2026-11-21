// Package privatechannelmigration reads the tenant's private-channel migration
// status using the templates service's exact cloud origin, prefix and audience.
package privatechannelmigration

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/internal/serviceprofile"
	"github.com/go-resty/resty/v2"
)

// Status keeps nullable timestamps as received text rather than inventing a
// stricter date format. RawFields preserves null/omission and future data.
type Status struct {
	TenantID                     *string                    `json:"TenantId"`
	MigrationStatus              *string                    `json:"MigrationStatus"`
	Details                      *string                    `json:"Details"`
	LastChecked                  *string                    `json:"LastChecked"`
	MigrationStartTimestamp      *string                    `json:"MigrationStartTimeStamp"`
	MigrationCompletionTimestamp *string                    `json:"MigrationCompletionTimeStamp"`
	RawFields                    map[string]json.RawMessage `json:"-"`
}

func (s *Status) UnmarshalJSON(data []byte) error {
	type snapshot Status
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*s = Status(value)
	return nil
}
func (Status) String() string     { return "private-channel migration status (redacted)" }
func (s Status) GoString() string { return s.String() }

// DecodeDetails explicitly decodes the opaque JSON text into retained fields.
// A decode failure leaves the native status/text intact. No count schema or
// successful migration meaning is inferred from arbitrary detail properties.
func (s Status) DecodeDetails() (map[string]json.RawMessage, error) {
	if s.Details == nil || strings.TrimSpace(*s.Details) == "" {
		return nil, nil
	}
	var value map[string]json.RawMessage
	if json.Unmarshal([]byte(*s.Details), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

type Service struct{ scope *client.Scope }

func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

// Get performs one native singleton read with no caller tenant or destination
// parameter. No mutation/list/cancel contract is established by this command.
func (s *Service) Get(ctx context.Context) (*Status, *resty.Response, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return nil, nil, client.ErrClosed
	}
	if s.scope.Backend() != client.PrivateChannelMigration {
		return nil, nil, client.ErrDestination
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, nil, err
	}
	profile, ok := serviceprofile.Resolve(serviceprofile.PrivateChannelMigration, environment.Name)
	if !ok {
		return nil, nil, client.ErrDestination
	}
	path := append(strings.Split(strings.TrimPrefix(profile.PathPrefix, "/"), "/"), "channels", "privateChannelMigrationStatus")
	response, err := s.scope.Do(ctx, client.Request{Method: http.MethodGet, Path: path, Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}, nil)
	if err != nil {
		return nil, response, err
	}
	var value Status
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}
