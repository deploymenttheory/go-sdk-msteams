package configurationdiagnostics

import (
	"encoding/json"
	"errors"
	"maps"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput reports a request rejected locally before any network I/O.
var ErrInput = errors.New("configuration diagnostics input is invalid")

// ErrResponse reports a success status the module does not declare for the
// operation.
var ErrResponse = errors.New("unexpected configuration diagnostics response")

func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// ReadOptions selects the optional apiVersion and schemaVersion query values
// of Get-CsConfiguration. Nil fields are omitted from the request.
type ReadOptions struct {
	APIVersion    *string
	SchemaVersion *string
}

// WriteOptions selects the optional apiVersion and schemaVersion query values
// of New-CsConfiguration and Set-CsConfiguration. Nil fields are omitted.
type WriteOptions struct {
	APIVersion    *string
	SchemaVersion *string
}

// Configuration is one configuration or policy instance as the generic
// configuration route returns it. The schema depends on the configuration
// type, so only the identity fields are typed; RawFields keeps every
// property the service sent, including Identity and ConfigId.
type Configuration struct {
	Identity  *string                    `json:"Identity"`
	ConfigID  *string                    `json:"ConfigId"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Configuration(value)
	return nil
}

// ConfigurationRequest is the XdsConfiguration body of New-CsConfiguration and
// Set-CsConfiguration: the Identity plus the type's own properties. Settings
// must not contain an "Identity" key; Identity is always sent from the field.
type ConfigurationRequest struct {
	Identity string
	Settings map[string]any
}

func (r ConfigurationRequest) MarshalJSON() ([]byte, error) {
	body := make(map[string]any, len(r.Settings)+1)
	maps.Copy(body, r.Settings)
	body["Identity"] = r.Identity
	return json.Marshal(body)
}

// ClearCacheRequest is the CacheClearRequest body of Clear-CsCacheOperation.
// KeysToDelete is always sent and may be empty; a nil slice is sent as [].
// Region is omitted when nil.
type ClearCacheRequest struct {
	KeysToDelete []string
	Region       *string
}

func (r ClearCacheRequest) MarshalJSON() ([]byte, error) {
	keys := r.KeysToDelete
	if keys == nil {
		keys = []string{}
	}
	return json.Marshal(struct {
		KeysToDelete []string `json:"KeysToDelete"`
		Region       *string  `json:"Region,omitempty"`
	}{keys, r.Region})
}

// CacheClearResult is one CacheClearResponse element: the keys that could not
// be removed from one region's cache.
type CacheClearResult struct {
	FailedKeys []string                   `json:"FailedKeys"`
	Region     *string                    `json:"Region"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

func (v *CacheClearResult) UnmarshalJSON(raw []byte) error {
	type plain CacheClearResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = CacheClearResult(value)
	return nil
}
