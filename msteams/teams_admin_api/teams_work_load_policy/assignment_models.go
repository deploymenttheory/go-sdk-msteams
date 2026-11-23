package teamsworkloadpolicy

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
func (GroupAssignment) String() string     { return "workload group assignment (redacted)" }
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
