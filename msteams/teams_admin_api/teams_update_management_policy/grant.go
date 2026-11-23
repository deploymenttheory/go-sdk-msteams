package teamsupdatemanagementpolicy

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

const resourceName = "TeamsUpdateManagementPolicy"
const grantCommand = "Grant-CsTeamsUpdateManagementPolicy"

func validGrantIdentity(v string) bool {
	return strings.TrimSpace(v) != "" && v != "." && v != ".." && !strings.ContainsRune(v, '\\') && !strings.ContainsFunc(v, unicode.IsControl)
}
func (s *Service) doGrant(ctx context.Context, command string, request client.Request, result any, statuses ...int) (*resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	if s.modern == nil {
		return nil, client.ErrClosed
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	r, e := s.modern.Do(ctx, request, result)
	if e != nil {
		return r, e
	}
	for _, status := range statuses {
		if r != nil && r.StatusCode() == status {
			return r, nil
		}
	}
	return r, client.ErrDecode
}

type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

func (v GroupAssignment) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain GroupAssignment
	return json.Marshal(plain(v))
}
func (GroupAssignment) String() string     { return "update management group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct{ PolicyName string }
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
