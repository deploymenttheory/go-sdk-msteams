package policyassignments

import (
	"bytes"
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// UserIdentity is the native equivalent of the generated InputObject.Identity
// field used by Get-CsUserPolicyAssignment's GetViaIdentity parameter set.
// It is an opaque user identifier, not a caller-supplied request URL.
type UserIdentity struct{ Identity string }

func (UserIdentity) String() string     { return "policy assignment user identity (redacted)" }
func (v UserIdentity) GoString() string { return v.String() }

// GroupSettings follows the generated DTO: unset and null keys are omitted;
// bound empty names and zero/negative priorities remain explicit wire values.
// Create and Update share these settings but use distinct HTTP methods/statuses.
type GroupSettings struct {
	PolicyName policy.Field[string]
	Priority   policy.Field[int32]
}

func (v GroupSettings) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if value, ok := v.PolicyName.Value(); ok {
		body["policyName"] = value
	}
	if value, ok := v.Priority.Value(); ok {
		body["priority"] = value
	}
	return json.Marshal(body)
}

// BatchCreateRequest preserves identity order and duplicates. A non-nil empty
// slice represents the command's allowed empty collection. No help-text policy
// allowlist or batch-size restriction is turned into an unproven local validator.
type BatchCreateRequest struct {
	Identities           []string
	PolicyType           policy.Field[string]
	PolicyName           policy.Field[string]
	OperationName        string
	AdditionalParameters map[string]json.RawMessage
}

func (v BatchCreateRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if v.Identities != nil {
		body["identities"] = v.Identities
	}
	if v.PolicyType.IsSet() || v.PolicyName.IsSet() {
		definition := map[string]any{}
		if value, ok := v.PolicyType.Value(); ok {
			definition["policyType"] = value
		}
		if value, ok := v.PolicyName.Value(); ok {
			definition["policyName"] = value
		}
		body["policy"] = definition
	}
	if v.AdditionalParameters != nil {
		values := map[string]json.RawMessage{}
		for key, value := range v.AdditionalParameters {
			projected, err := projectAdditional(value, false)
			if err != nil {
				return nil, err
			}
			if projected != nil {
				values[key] = projected
			}
		}
		body["additionalParameters"] = values
	}
	return json.Marshal(body)
}

// Match the generated JSON dictionary decoder/encoder, including its unusual
// array-null projection to the string "null". Objects omit null members at every
// depth. Raw numeric text avoids introducing Go float64 precision loss.
func projectAdditional(raw json.RawMessage, inArray bool) (json.RawMessage, error) {
	if !json.Valid(raw) {
		return nil, ErrInput
	}
	value := bytes.TrimSpace(raw)
	if bytes.Equal(value, []byte("null")) {
		if inArray {
			return json.RawMessage(`"null"`), nil
		}
		return nil, nil
	}
	switch value[0] {
	case '{':
		var source map[string]json.RawMessage
		if json.Unmarshal(value, &source) != nil {
			return nil, ErrInput
		}
		result := map[string]json.RawMessage{}
		for key, entry := range source {
			projected, err := projectAdditional(entry, false)
			if err != nil {
				return nil, err
			}
			if projected != nil {
				result[key] = projected
			}
		}
		return json.Marshal(result)
	case '[':
		var source []json.RawMessage
		if json.Unmarshal(value, &source) != nil {
			return nil, ErrInput
		}
		result := make([]json.RawMessage, 0, len(source))
		for _, entry := range source {
			projected, err := projectAdditional(entry, true)
			if err != nil {
				return nil, err
			}
			result = append(result, projected)
		}
		return json.Marshal(result)
	default:
		return value, nil
	}
}
func (GroupSettings) String() string          { return "policy assignment group settings (redacted)" }
func (v GroupSettings) GoString() string      { return v.String() }
func (BatchCreateRequest) String() string     { return "policy assignment batch request (redacted)" }
func (v BatchCreateRequest) GoString() string { return v.String() }
