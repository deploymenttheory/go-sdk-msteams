package teamsmessagingpolicy

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
func (GroupAssignment) String() string     { return "Messaging group assignment (redacted)" }
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

func (UserAssignment) String() string     { return "messaging policy assignment (redacted)" }
func (v UserAssignment) GoString() string { return v.String() }

func (TenantAssignment) String() string     { return "messaging policy assignment (redacted)" }
func (v TenantAssignment) GoString() string { return v.String() }

func (GroupGrant) String() string     { return "messaging policy assignment (redacted)" }
func (v GroupGrant) GoString() string { return v.String() }

func (Assignment) String() string     { return "messaging policy assignment (redacted)" }
func (v Assignment) GoString() string { return v.String() }
