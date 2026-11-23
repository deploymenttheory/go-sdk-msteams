package cloudcalldata

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected cloud call data connection response")

// Connection is the tenant's cloud call data connection. Token is the
// credential an on-premises deployment uses to upload call data; it is secret
// and never printed by String/GoString. RawFields keeps the complete object.
type Connection struct {
	Token     *string                    `json:"Token"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *Connection) UnmarshalJSON(raw []byte) error {
	type plain Connection
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Connection(value)
	return nil
}

func (Connection) String() string     { return "cloud call data connection (redacted)" }
func (v Connection) GoString() string { return v.String() }
