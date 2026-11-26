package conversationrole

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const IdentitySet = "Identity"
const RelativeIDSet = "ParentAndRelativeKey"

// WriteProjection exposes the exact pinned compatibility request before a write.
// OmittedPaths describes source projection loss, including nested permission
// objects. These fields are observations, not proof of server persistence.
type WriteProjection struct {
	Command      string
	ParameterSet string
	Body         json.RawMessage
	OmittedPaths []string
	ChangedPaths []string
	NoOp         bool
}

func (WriteProjection) String() string     { return "conversation write projection (redacted)" }
func (v WriteProjection) GoString() string { return v.String() }
func copyProjection(v WriteProjection) WriteProjection {
	v.Body = append(json.RawMessage(nil), v.Body...)
	v.OmittedPaths = append([]string(nil), v.OmittedPaths...)
	v.ChangedPaths = append([]string(nil), v.ChangedPaths...)
	return v
}

// CreateParameters supplies the upstream accepted DTO and scalar parameters.
// Chat/Calling flags are accepted; inspect ProjectCreate's resulting omissions.
type CreateParameters struct {
	Permissions          *Permissions
	CreatedAt, UpdatedAt DateTime
	Name                 policy.Field[string]
	IsAssignable         policy.Field[bool]
	policy.Controls
}

func (CreateParameters) String() string     { return "conversation create parameters (redacted)" }
func (v CreateParameters) GoString() string { return v.String() }

// CreateRequest freezes a source-observed request. Projection returns an owned
// view; changing inputs or that view cannot change the body sent by Create.
type CreateRequest struct {
	identity   string
	projection WriteProjection
}

func (v CreateRequest) Projection() WriteProjection { return copyProjection(v.projection) }
func (CreateRequest) String() string                { return "conversation create request (redacted)" }
func (v CreateRequest) GoString() string            { return v.String() }

// ProjectCreateByIdentity prepares the pinned Identity parameter-set body only.
// It never authenticates or writes. Nested permission projection is inspectable.
func ProjectCreateByIdentity(identity string, input CreateParameters) (*CreateRequest, error) {
	return projectCreate(identity, IdentitySet, input)
}

// ProjectCreateByID prepares the distinct relative-ID parameter-set body. The
// pinned pipeline leaves its bound model Id null; this helper preserves that.
func ProjectCreateByID(id string, input CreateParameters) (*CreateRequest, error) {
	if !policyrp.ValidIdentity(id) {
		return nil, ErrInput
	}
	return projectCreate("Global/"+id, RelativeIDSet, input)
}
func projectCreate(identity, set string, input CreateParameters) (*CreateRequest, error) {
	if !policyrp.ValidIdentity(identity) || input.Permissions == nil {
		return nil, ErrInput
	}
	permission, omitted, e := projectPermissions(input.Permissions)
	if e != nil {
		return nil, e
	}
	body, e := policyrp.Fields(struct {
		Name         policy.Field[string] `json:"Name,omitzero"`
		IsAssignable policy.Field[bool]   `json:"IsAssignable,omitzero"`
		Permissions  json.RawMessage      `json:"Permissions"`
		CreatedAt    DateTime             `json:"CreatedAt"`
		UpdatedAt    DateTime             `json:"UpdatedAt"`
	}{input.Name, input.IsAssignable, permission, input.CreatedAt, input.UpdatedAt})
	if e != nil {
		return nil, e
	}
	if e = policyrp.Identity(body, identity, policy.Field[string]{}, input.Controls); e != nil {
		return nil, e
	}
	changed := []string(nil)
	if set == RelativeIDSet {
		body["Id"] = json.RawMessage("null")
		changed = []string{"Id"}
	} else if strings.Contains(identity, "/") {
		body["Id"], e = json.Marshal(strings.Split(identity, "/")[1])
		if e != nil {
			return nil, e
		}
	}
	raw, e := json.Marshal(body)
	if e != nil {
		return nil, ErrInput
	}
	return &CreateRequest{identity: identity, projection: WriteProjection{Command: "New-CsConversationRole", ParameterSet: set, Body: raw, OmittedPaths: omitted, ChangedPaths: changed}}, nil
}

// UpdateParameters selects bound settings; nil dates are omitted, while selected
// Name/IsAssignable null values retain their CLR nullable representation.
type UpdateParameters struct {
	Permissions          policy.Field[*Permissions]
	Name                 policy.Field[string]
	IsAssignable         policy.Field[bool]
	CreatedAt, UpdatedAt *DateTime
	policy.Controls
}

func (UpdateParameters) String() string     { return "conversation update parameters (redacted)" }
func (v UpdateParameters) GoString() string { return v.String() }
func (v UpdateParameters) specified() bool {
	return v.Permissions.IsSet() || v.Name.IsSet() || v.IsAssignable.IsSet() || v.CreatedAt != nil || v.UpdatedAt != nil
}

// UpdateRequest binds an inspected frozen projection to its preparing service,
// resource identity and parameter set. Prepare performs a read; apply is separate.
type UpdateRequest struct {
	owner                   *Service
	identity, selector, set string
	before                  *Role
	read                    *resty.Response
	projection              WriteProjection
	prepared                bool
}

func (v UpdateRequest) Projection() WriteProjection   { return copyProjection(v.projection) }
func (v UpdateRequest) Before() *Role                 { return cloneRole(v.before) }
func (v UpdateRequest) ReadResponse() *resty.Response { return v.read }
func (UpdateRequest) String() string                  { return "conversation prepared update (redacted)" }
func (v UpdateRequest) GoString() string              { return v.String() }

// UpdateResult retains preparation, projected loss and action stages on error.
// HTTPAccepted does not establish persisted or effective permissions.
type UpdateResult struct {
	Before, After                    *Role
	ReadResponse                     *resty.Response
	Projection                       WriteProjection
	PatchInvoked, HTTPAccepted, NoOp bool
}

func (UpdateResult) String() string     { return "conversation update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func cloneRole(v *Role) *Role {
	if v == nil {
		return nil
	}
	raw, e := json.Marshal(v)
	if e != nil {
		return nil
	}
	var next Role
	if json.Unmarshal(raw, &next) != nil {
		return nil
	}
	return &next
}

func projectUpdate(before *Role, identity, set string, input UpdateParameters) (WriteProjection, error) {
	body, e := policyrp.Fields(struct {
		Name         policy.Field[string] `json:"Name,omitzero"`
		IsAssignable policy.Field[bool]   `json:"IsAssignable,omitzero"`
		CreatedAt    *DateTime            `json:"CreatedAt,omitempty"`
		UpdatedAt    *DateTime            `json:"UpdatedAt,omitempty"`
	}{input.Name, input.IsAssignable, input.CreatedAt, input.UpdatedAt})
	if e != nil {
		return WriteProjection{}, e
	}
	var omitted []string
	if input.Permissions.IsSet() {
		value, _ := input.Permissions.Value()
		if input.Permissions.IsNull() || value == nil {
			return WriteProjection{}, ErrInput
		}
		body["Permissions"], omitted, e = projectPermissions(value)
		if e != nil {
			return WriteProjection{}, e
		}
	}
	if set == RelativeIDSet {
		body["Id"], e = json.Marshal(before.ID)
		if e != nil {
			return WriteProjection{}, ErrInput
		}
	}
	override := policy.Null[string]()
	if before.Identity != nil {
		override = policy.Set(*before.Identity)
	}
	if e = policyrp.Identity(body, identity, override, input.Controls); e != nil {
		return WriteProjection{}, e
	}
	raw, e := json.Marshal(body)
	if e != nil {
		return WriteProjection{}, ErrInput
	}
	return WriteProjection{Command: "Set-CsConversationRole", ParameterSet: set, Body: raw, OmittedPaths: omitted}, nil
}

func projectPermissions(v *Permissions) (json.RawMessage, []string, error) {
	if v == nil {
		return json.RawMessage("null"), nil, nil
	}
	body, e := json.Marshal(struct {
		Identity *string `json:"Identity"`
	}{v.Identity})
	if e != nil {
		return nil, nil, ErrInput
	}
	omitted := []string{"Permissions.Chat", "Permissions.Calling"}
	for key := range v.RawFields {
		if key != "Identity" && key != "Chat" && key != "Calling" {
			omitted = append(omitted, "Permissions."+key)
		}
	}
	sort.Strings(omitted)
	return body, omitted, nil
}

// ProjectRoleValue exposes the fully bound role value used by the pinned list
// composition. It preserves source permission loss rather than inventing a
// successful full-permission wire contract. Nil list elements remain null.
func ProjectRoleValue(v *Role) (json.RawMessage, []string, error) {
	if v == nil {
		return json.RawMessage("null"), nil, nil
	}
	if v.ID != nil && !validModelID(*v.ID) {
		return nil, nil, ErrInput
	}
	permissions, omitted, e := projectPermissions(v.Permissions)
	if e != nil {
		return nil, nil, e
	}
	raw, e := json.Marshal(struct {
		Identity     *string         `json:"Identity"`
		Permissions  json.RawMessage `json:"Permissions"`
		ID           *string         `json:"Id"`
		Name         *string         `json:"Name"`
		IsAssignable *bool           `json:"IsAssignable"`
		CreatedAt    DateTime        `json:"CreatedAt"`
		UpdatedAt    DateTime        `json:"UpdatedAt"`
	}{v.Identity, permissions, v.ID, v.Name, v.IsAssignable, v.CreatedAt, v.UpdatedAt})
	if e != nil {
		return nil, nil, ErrInput
	}
	for key := range v.RawFields {
		switch key {
		case "Identity", "Permissions", "Id", "Name", "IsAssignable", "CreatedAt", "UpdatedAt":
		default:
			omitted = append(omitted, key)
		}
	}
	sort.Strings(omitted)
	return raw, omitted, nil
}

// RemoveAsync starts from an empty body and adds original bound parameters;
// unlike CreateAsync it never applies RequiredNotBoundedParameters metadata.
func deleteBody(identity, set, id string, controls policy.Controls) (map[string]json.RawMessage, error) {
	if !policyrp.ValidIdentity(identity) {
		return nil, ErrInput
	}
	body := make(map[string]json.RawMessage)
	if e := policyrp.Identity(body, identity, policy.Field[string]{}, controls); e != nil {
		return nil, e
	}
	if set == RelativeIDSet {
		delete(body, "Identity")
		raw, e := json.Marshal(id)
		if e != nil {
			return nil, ErrInput
		}
		body["Id"] = raw
	}
	return body, nil
}

// ResourceIdentity returns the frozen management target, distinct from model Id.
func (v CreateRequest) ResourceIdentity() string { return v.identity }

// ResourceIdentity returns the frozen management target, distinct from model Id.
func (v UpdateRequest) ResourceIdentity() string { return v.identity }
