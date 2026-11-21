package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/go-resty/resty/v2"
)

// APIError preserves machine-readable service diagnostics for an unsuccessful
// HTTP response, including redirects. Codes and identifiers are untrusted server
// data intended for explicit inspection, not automatic logging. Error and GoString
// print only backend and status; response bodies/messages are never retained here.
type APIError struct {
	Backend            Backend
	StatusCode         int
	Codes              []string
	ServiceErrorCodes  []int
	RequestIDs         []string
	ClientRequestIDs   []string
	CorrelationIDs     []string
	OperationIDs       []string
	DiagnosticCodes    []int
	DiagnosticSubCodes []int
}

// Is classifies HTTP credential rejection and forbidden access. A 403 does not
// prove a missing role: consent, policy and resource access can also forbid a call.
// Other status codes never imply an unsupported mode or absent tenant feature.
func (e *APIError) Is(target error) bool {
	if e == nil {
		return false
	}
	return (target == auth.ErrUnauthenticated && e.StatusCode == http.StatusUnauthorized) ||
		(target == auth.ErrForbidden && e.StatusCode == http.StatusForbidden)
}

func (e *APIError) Error() string {
	if e == nil {
		return "msteams backend error"
	}
	// Do not print even Backend verbatim: this public struct may be caller-created.
	backend := "unknown"
	if knownBackend(e.Backend) {
		backend = string(e.Backend)
	}
	return fmt.Sprintf("msteams %s request failed (HTTP %d)", backend, e.StatusCode)
}

// String also redacts copied APIError values that no longer implement error.
func (e APIError) String() string { return (&e).Error() }

// GoString keeps %#v from exposing server-provided identifiers or codes.
func (e APIError) GoString() string { return (&e).Error() }

func normalizeError(backend Backend, response *resty.Response) *APIError {
	return NewAPIError(backend, response.StatusCode(), response.Header(), response.Body())
}

// NewAPIError normalizes a received HTTP or batch-item error without retaining
// headers/body or printing service messages. Backend selects a known native
// service family. Codes and identifiers remain untrusted for explicit inspection.
func NewAPIError(backend Backend, status int, headers http.Header, body []byte) *APIError {
	err := &APIError{Backend: backend, StatusCode: status}
	collectHeaderIDs(err, headers)
	var envelope map[string]json.RawMessage
	if json.Unmarshal(body, &envelope) != nil {
		return err
	}
	if backend == Graph {
		collectErrorObject(err, envelope["error"], 0)
	} else {
		collectErrorObject(err, body, 0)
		// ConfigAPI's TeamMiddletierErrorResponse.Error property serializes as "errors".
		var objects []json.RawMessage
		if json.Unmarshal(envelope["errors"], &objects) == nil {
			for _, object := range objects {
				collectErrorObject(err, object, 0)
			}
		} else {
			collectErrorObject(err, envelope["error"], 0)
		}
	}
	return err
}

func collectHeaderIDs(err *APIError, headers http.Header) {
	for _, value := range headers.Values("request-id") {
		err.RequestIDs = appendUnique(err.RequestIDs, value)
	}
	for _, value := range headers.Values("x-ms-request-id") {
		err.RequestIDs = appendUnique(err.RequestIDs, value)
	}
	for _, value := range headers.Values("client-request-id") {
		err.ClientRequestIDs = appendUnique(err.ClientRequestIDs, value)
	}
	for _, value := range headers.Values("x-ms-correlation-id") {
		err.CorrelationIDs = appendUnique(err.CorrelationIDs, value)
	}
}

func collectErrorObject(err *APIError, raw json.RawMessage, depth int) {
	if depth >= 16 {
		return
	}
	var object struct {
		Code             string          `json:"code"`
		ServiceErrorCode json.RawMessage `json:"errorCode"`
		RequestID        string          `json:"request-id"`
		ClientRequestID  string          `json:"client-request-id"`
		OperationID      string          `json:"operationId"`
		Diagnostics      json.RawMessage `json:"diagnostics"`
		Message          json.RawMessage `json:"message"`
		InnerError       json.RawMessage `json:"innerError"`
	}
	if json.Unmarshal(raw, &object) != nil {
		return
	}
	err.Codes = appendUnique(err.Codes, object.Code)
	var serviceCode int
	if len(object.ServiceErrorCode) != 0 && string(object.ServiceErrorCode) != "null" && json.Unmarshal(object.ServiceErrorCode, &serviceCode) == nil {
		err.ServiceErrorCodes = appendUniqueInt(err.ServiceErrorCodes, serviceCode)
	}
	err.RequestIDs = appendUnique(err.RequestIDs, object.RequestID)
	err.ClientRequestIDs = appendUnique(err.ClientRequestIDs, object.ClientRequestID)
	err.OperationIDs = appendUnique(err.OperationIDs, object.OperationID)
	collectDiagnostics(err, object.Diagnostics)
	var message struct {
		Diagnostics json.RawMessage `json:"diagnostics"`
	}
	if json.Unmarshal(object.Message, &message) == nil {
		collectDiagnostics(err, message.Diagnostics)
	}
	collectErrorObject(err, object.InnerError, depth+1)
}

func collectDiagnostics(err *APIError, raw json.RawMessage) {
	var diagnostics struct {
		Code          *int   `json:"code"`
		SubCode       *int   `json:"subCode"`
		CorrelationID string `json:"correlationID"`
	}
	if json.Unmarshal(raw, &diagnostics) != nil {
		return
	}
	err.CorrelationIDs = appendUnique(err.CorrelationIDs, diagnostics.CorrelationID)
	if diagnostics.Code != nil {
		err.DiagnosticCodes = append(err.DiagnosticCodes, *diagnostics.Code)
	}
	if diagnostics.SubCode != nil {
		err.DiagnosticSubCodes = append(err.DiagnosticSubCodes, *diagnostics.SubCode)
	}
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
