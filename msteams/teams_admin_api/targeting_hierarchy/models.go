// Package targetinghierarchy manages the singleton operational hierarchy through
// the separately scoped targeting v1 and hierarchy v2 protocols.
package targetinghierarchy

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrVersion        = errors.New("hierarchy API version must be 1 or 2")
	ErrTenantIdentity = errors.New("hierarchy v2 requires a GUID tenant identity bound to the SDK session")
)

// Version selects a service protocol, independently of PowerShell releases.
type Version string

const (
	V1 Version = "1"
	V2 Version = "2"
)

// CSV transfers ownership of Body immediately when an import/upload is called.
// Close must unblock Read. MaxBytes bounds the complete multipart wire payload,
// including framing; this caller-selected budget is not a claimed service limit.
// Filename is normalized to its portable basename. Original bytes are streamed.
type CSV struct {
	Filename string
	Body     io.ReadCloser
	MaxBytes int64
}

// ImportResult preserves the complete response map and optional operation ID.
// V1 does not require requestId. V2 requires a nonblank id before it can expose
// a usable handle; decoding failure retains the response/partial result.
type ImportResult struct {
	RequestID *string
	RawFields map[string]json.RawMessage
}

func (ImportResult) String() string     { return "hierarchy import result (redacted)" }
func (v ImportResult) GoString() string { return v.String() }

type RetailErrorDetails struct {
	ErrorMessage    *string `json:"errorMessage"`
	LastKnownStatus *string `json:"lastKnownStatus"`
}

// RetailOperation retains raw status/timestamps and nested JSON error text.
// Timestamps stay strings because the pinned wire model does not require RFC3339.
type RetailOperation struct {
	ID             string                     `json:"id"`
	Status         *string                    `json:"status"`
	ErrorDetails   *RetailErrorDetails        `json:"errorDetails"`
	LastUpdatedAt  *string                    `json:"lastUpdatedAt"`
	LastModifiedBy *string                    `json:"lastModifiedBy"`
	Filename       *string                    `json:"filename"`
	RawFields      map[string]json.RawMessage `json:"-"`
}

func (v *RetailOperation) UnmarshalJSON(data []byte) error {
	type snapshot RetailOperation
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = RetailOperation(value)
	return nil
}
func (RetailOperation) String() string     { return "retail hierarchy operation (redacted)" }
func (v RetailOperation) GoString() string { return v.String() }

type TargetingError struct {
	Code        *string                    `json:"errorCode"`
	Description *string                    `json:"description"`
	Args        []map[string]string        `json:"args"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *TargetingError) UnmarshalJSON(data []byte) error {
	type snapshot TargetingError
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = TargetingError(value)
	return nil
}

// DecodeErrors explicitly interprets the nested JSON text without changing the
// native read result. Malformed text remains available in ErrorDetails/RawFields.
func (v RetailOperation) DecodeErrors() ([]TargetingError, error) {
	if v.ErrorDetails == nil || v.ErrorDetails.ErrorMessage == nil || strings.TrimSpace(*v.ErrorDetails.ErrorMessage) == "" {
		return nil, nil
	}
	var values []TargetingError
	if json.Unmarshal([]byte(*v.ErrorDetails.ErrorMessage), &values) != nil {
		return nil, client.ErrDecode
	}
	return values, nil
}

type HierarchyError struct {
	Message *string                `json:"message"`
	Details []HierarchyErrorDetail `json:"details"`
}
type HierarchyErrorDetail struct {
	Code      *string                    `json:"code"`
	Message   *string                    `json:"message"`
	Arguments map[string]string          `json:"arguments"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *HierarchyErrorDetail) UnmarshalJSON(data []byte) error {
	type snapshot HierarchyErrorDetail
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = HierarchyErrorDetail(value)
	return nil
}

type HierarchyOperation struct {
	ID                 string                     `json:"id"`
	Status             *string                    `json:"status"`
	Error              *HierarchyError            `json:"error"`
	LastActionDateTime *string                    `json:"lastActionDateTime"`
	LastModifiedBy     *string                    `json:"lastModifiedBy"`
	Filename           *string                    `json:"fileName"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *HierarchyOperation) UnmarshalJSON(data []byte) error {
	type snapshot HierarchyOperation
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = HierarchyOperation(value)
	return nil
}
func (HierarchyOperation) String() string     { return "hierarchy v2 operation (redacted)" }
func (v HierarchyOperation) GoString() string { return v.String() }

// Status is the pinned command projection plus retained native data. ErrorMessage
// is presentation text, never an automatically logged error. RawStatus remains
// distinct from the v2 normalized presentation status, including future values.
type Status struct {
	Version         Version
	ID              string
	Status          *string
	RawStatus       *string
	LastKnownStatus *string
	ErrorMessage    *string
	LastUpdatedAt   *string
	LastModifiedBy  *string
	Filename        *string
	Retail          *RetailOperation
	Hierarchy       *HierarchyOperation
}

func (Status) String() string     { return "hierarchy status projection (redacted)" }
func (v Status) GoString() string { return v.String() }

// UploadResult retains preparation and mutation outcomes even if a later stage
// fails. A successful import HTTP response is not proof of completed ingestion.
// Failed/ambiguous create/import writes are never rolled back or replayed.
type UploadResult struct {
	Version              Version
	LookupHTTPAccepted   bool
	CreateAttempted      bool
	CreationHTTPAccepted bool
	ImportAttempted      bool
	ImportHTTPAccepted   bool
	Import               *ImportResult
	LookupResponse       *resty.Response
	CreationResponse     *resty.Response
	Response             *resty.Response
}

func (UploadResult) String() string     { return "hierarchy upload result (redacted)" }
func (v UploadResult) GoString() string { return v.String() }
