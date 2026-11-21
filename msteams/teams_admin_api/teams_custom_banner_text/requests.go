package teamscustombannertext

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings preserves omitted, explicit null and empty values. Text cannot bind null.
type Settings struct {
	Text        policy.Field[string] `json:"Text,omitzero"`
	Description policy.Field[string] `json:"Description,omitzero"`
}

// CreateRequest requires nonempty Text as the compiled New binder does.
type CreateRequest struct {
	Text        string
	Description policy.Field[string]
	policy.Controls
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
	// BodyID must be supplied to PatchByID from the read model, including null.
	BodyID policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Banner
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}
type CreateResult struct {
	ID, Identity string
	Value        *Banner
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "banner update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }
func (CreateResult) String() string     { return "banner create result (redacted)" }
func (p CreateResult) GoString() string { return p.String() }
func fields(input Settings) (map[string]json.RawMessage, error) {
	if input.Text.IsNull() {
		return nil, ErrInput
	}
	return policyrp.Fields(input)
}
func valueField(value *string) policy.Field[string] {
	if value == nil {
		return policy.Null[string]()
	}
	return policy.Set(*value)
}
func put(body map[string]json.RawMessage, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return ErrInput
	}
	body[key] = raw
	return nil
}
func force(body map[string]json.RawMessage, controls policy.Controls) error {
	if controls.Force == nil {
		return nil
	}
	value := "False"
	if *controls.Force {
		value = "True"
	}
	return put(body, "Force", value)
}
