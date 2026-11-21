// Package teamscallholdpolicy manages call hold policies and their assignments.
package teamscallholdpolicy

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput
var ErrResponse = errors.New("unexpected call hold assignment response")

// Policy preserves nullable source strings and the complete received snapshot.
type Policy struct {
	Identity                *string                    `json:"Identity"`
	Description             *string                    `json:"Description"`
	AudioFileID             *string                    `json:"AudioFileId"`
	StreamingSourceURL      *string                    `json:"StreamingSourceUrl"`
	StreamingSourceAuthType *string                    `json:"StreamingSourceAuthType"`
	RawFields               map[string]json.RawMessage `json:"-"`
	RawJSON                 json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var v Policy
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return client.ErrDecode
	}
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return client.ErrDecode
		}
		name, ok := key.(string)
		if !ok {
			return client.ErrDecode
		}
		var body json.RawMessage
		if d.Decode(&body) != nil {
			return client.ErrDecode
		}
		var target **string
		switch strings.ToLower(name) {
		case "identity":
			target = &v.Identity
		case "description":
			target = &v.Description
		case "audiofileid":
			target = &v.AudioFileID
		case "streamingsourceurl":
			target = &v.StreamingSourceURL
		case "streamingsourceauthtype":
			target = &v.StreamingSourceAuthType
		}
		if target != nil && json.Unmarshal(body, target) != nil {
			return client.ErrDecode
		}
	}
	v.RawFields = fields
	v.RawJSON = append(json.RawMessage(nil), raw...)
	*p = v
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(p))
}
func (Policy) String() string     { return "call hold policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings distinguishes omitted, explicit null and empty strings.
type Settings struct {
	Description             policy.Field[string] `json:"Description,omitzero"`
	AudioFileID             policy.Field[string] `json:"AudioFileId,omitzero"`
	StreamingSourceURL      policy.Field[string] `json:"StreamingSourceUrl,omitzero"`
	StreamingSourceAuthType policy.Field[string] `json:"StreamingSourceAuthType,omitzero"`
}
type CreateRequest struct {
	Settings
	policy.Controls
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "call hold update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

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
func (GroupAssignment) String() string     { return "call hold group assignment (redacted)" }
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
