// Package conversationrolessetting manages the conditional conversation-role list.
package conversationrolessetting

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration preserves the complete response. Its CLR collection defaults
// to empty for both an omitted and null property; RawFields retains that wire
// distinction, while typed ConversationRoles follows the pinned projection.
type Configuration struct {
	Identity          *string                    `json:"Identity"`
	ConversationRoles []*roles.Role              `json:"ConversationRoles"`
	RawFields         map[string]json.RawMessage `json:"-"`
	RawJSON           json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	if next.ConversationRoles == nil {
		next.ConversationRoles = make([]*roles.Role, 0)
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(next)
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string     { return "conversation role settings (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// ListModifier follows the pinned helper: nonempty Replace wins; null/empty
// modifier clears; Remove rejects only if comparison occurs on a nonempty list.
type ListModifier struct{ Add, Remove, Replace []*roles.Role }

func (ListModifier) String() string     { return "conversation role list modifier (redacted)" }
func (v ListModifier) GoString() string { return v.String() }

type UpdateParameters struct {
	ConversationRoles policy.Field[*ListModifier]
	policy.Controls
}

func (UpdateParameters) String() string     { return "conversation role settings parameters (redacted)" }
func (v UpdateParameters) GoString() string { return v.String() }

// UpdateRequest freezes a projection for one preparing client/service/identity.
// It is not an assertion of complete permission-write compatibility.
type UpdateRequest struct {
	owner      *Service
	identity   string
	before     *Configuration
	read       *resty.Response
	projection roles.WriteProjection
	prepared   bool
}

func (v UpdateRequest) Projection() roles.WriteProjection { return copyProjection(v.projection) }
func (v UpdateRequest) Before() *Configuration            { return cloneConfiguration(v.before) }
func (v UpdateRequest) ReadResponse() *resty.Response     { return v.read }
func (UpdateRequest) String() string                      { return "conversation settings prepared update (redacted)" }
func (v UpdateRequest) GoString() string                  { return v.String() }

type UpdateResult struct {
	Before, After                    *Configuration
	ReadResponse                     *resty.Response
	Projection                       roles.WriteProjection
	PatchInvoked, HTTPAccepted, NoOp bool
}

func (UpdateResult) String() string     { return "conversation settings update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func cloneConfiguration(v *Configuration) *Configuration {
	if v == nil {
		return nil
	}
	raw, e := json.Marshal(v)
	if e != nil {
		return nil
	}
	var next Configuration
	if json.Unmarshal(raw, &next) != nil {
		return nil
	}
	return &next
}
func copyProjection(v roles.WriteProjection) roles.WriteProjection {
	v.Body = append(json.RawMessage(nil), v.Body...)
	v.OmittedPaths = append([]string(nil), v.OmittedPaths...)
	v.ChangedPaths = append([]string(nil), v.ChangedPaths...)
	return v
}
