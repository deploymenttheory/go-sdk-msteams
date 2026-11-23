package teamsupgradepolicy

import "encoding/json"

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
func (GroupAssignment) String() string     { return "upgrade group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct {
	PolicyName string
	Force      bool
}
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}

// Assignment selects a public Grant parameter set. A nonnil migration binding
// returns the pinned modern wrapper's parameter-binding failure before HTTP.
type Assignment struct {
	MigrateMeetingsToTeams *bool
	User                   *UserAssignment
	Tenant                 *TenantAssignment
	Group                  *GroupGrant
}
