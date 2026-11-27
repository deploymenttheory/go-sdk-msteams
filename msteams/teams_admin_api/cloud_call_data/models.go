package cloudcalldata

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected cloud call data connection response")

// ErrCardinality means the unique-result convenience cannot represent the full
// returned array. It does not mean HTTP404 or that creation did not occur.
var ErrCardinality = errors.New("cloud call data returned no unique non-null connection")

// CreateResult retains the Modern action outcome without claiming returned
// connections are newly created/owned. The data can contain sensitive tokens.
type CreateResult struct {
	Returned                    []*Connection
	Response                    *resty.Response
	CreateInvoked, HTTPAccepted bool
}

func (CreateResult) String() string     { return "cloud call data create result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }

// CardinalityError keeps the full typed data and receipt when Get/Create cannot
// return one model. CreateOutcome is nonnil only for the action convenience.
// An accepted action is not safe to replay because this error was returned.
type CardinalityError struct {
	Returned      []*Connection
	Response      *resty.Response
	CreateOutcome *CreateResult
}

func (e *CardinalityError) Error() string   { return "cloud call data cardinality error (redacted)" }
func (e CardinalityError) String() string   { return "cloud call data cardinality error (redacted)" }
func (e CardinalityError) GoString() string { return e.String() }
func (e *CardinalityError) Unwrap() error   { return ErrCardinality }

// Connection is one returned cloud call data connection model. Token is the
// credential an on-premises deployment uses to upload call data; it is secret
// and never printed by String/GoString. RawFields keeps the complete object.
type Connection struct {
	Token     *string                    `json:"Token"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *Connection) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value := Connection{RawFields: fields}
	// Generated FromJson is case-sensitive and treats a non-string Token as null.
	// Keep the complete object as a native read snapshot rather than dropping it.
	if token, ok := fields["Token"]; ok {
		var text *string
		if json.Unmarshal(token, &text) == nil {
			value.Token = text
		}
	}
	*v = value
	return nil
}

func (Connection) String() string     { return "cloud call data connection (redacted)" }
func (v Connection) GoString() string { return v.String() }
