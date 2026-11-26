package conversationrolessetting

import (
	"encoding/json"
	"fmt"

	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// ProjectListValue applies only the exact local transition and projected list
// serialization. It performs no authentication or HTTP. Returned bytes/paths are
// owned copies; source permission projection loss remains visible.
func ProjectListValue(current []*roles.Role, modifier *ListModifier) (json.RawMessage, []string, error) {
	if current == nil {
		return nil, nil, ErrInput
	}
	values := make([]*roles.Role, len(current))
	copy(values, current)
	switch {
	case modifier == nil || len(modifier.Add)+len(modifier.Remove)+len(modifier.Replace) == 0:
		values = make([]*roles.Role, 0)
	case len(modifier.Replace) > 0:
		values = append([]*roles.Role(nil), modifier.Replace...)
	default:
		// The compiled predicate throws whenever a comparison is performed. It is
		// never called for an empty current collection, regardless of Remove syntax.
		if len(modifier.Remove) > 0 && len(values) > 0 {
			return nil, nil, ErrInput
		}
		values = append(values, modifier.Add...)
	}
	body := make([]json.RawMessage, 0, len(values))
	var omitted []string
	for index, value := range values {
		raw, paths, e := roles.ProjectRoleValue(value)
		if e != nil {
			return nil, nil, e
		}
		body = append(body, raw)
		for _, path := range paths {
			omitted = append(omitted, fmt.Sprintf("ConversationRoles[%d].%s", index, path))
		}
	}
	raw, e := json.Marshal(body)
	if e != nil {
		return nil, nil, ErrInput
	}
	return raw, omitted, nil
}
func project(before *Configuration, identity string, input UpdateParameters) (roles.WriteProjection, error) {
	modifier, _ := input.ConversationRoles.Value()
	raw, omitted, e := ProjectListValue(before.ConversationRoles, modifier)
	if e != nil {
		return roles.WriteProjection{}, e
	}
	body := map[string]json.RawMessage{"ConversationRoles": raw}
	override := policy.Null[string]()
	if before.Identity != nil {
		override = policy.Set(*before.Identity)
	}
	if e = policyrp.Identity(body, identity, override, input.Controls); e != nil {
		return roles.WriteProjection{}, e
	}
	encoded, e := json.Marshal(body)
	if e != nil {
		return roles.WriteProjection{}, ErrInput
	}
	return roles.WriteProjection{Command: setCommand, ParameterSet: roles.IdentitySet, Body: encoded, OmittedPaths: omitted}, nil
}

// ResourceIdentity returns the frozen management target, distinct from model Id.
func (v UpdateRequest) ResourceIdentity() string { return v.identity }
