package sipdevicesgateway

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies local validation failures raised before any I/O.
var ErrInput = errors.New("invalid SIP devices gateway input")

// ErrResponse identifies an undeclared successful HTTP status. The module
// accepts only 200 for every gateway operation.
var ErrResponse = errors.New("unexpected SIP devices gateway response")

// BulkSignInItem is one row of a bulk sign-in request (SdgBulkSignInRequestItem).
// Nil fields are omitted from the body, as the module omits null values.
type BulkSignInItem struct {
	UserName   *string `json:"userName,omitempty"`
	HardwareID *string `json:"hardwareId,omitempty"`
}

// String redacts the account name and hardware ID.
func (BulkSignInItem) String() string     { return "bulk sign-in item (redacted)" }
func (i BulkSignInItem) GoString() string { return i.String() }

// BulkSignInResponse is the created batch (SdgBulkSignInResponse).
type BulkSignInResponse struct {
	BatchID         *string                    `json:"batchId"`
	BatchItemsCount *int64                     `json:"batchItemsCount"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

// BulkSignInRequestSummary is one batch in the tenant's summary
// (SdgBulkSignInRequestsSummaryResponseItem).
type BulkSignInRequestSummary struct {
	BatchID         *string                    `json:"batchId"`
	BatchStatus     *string                    `json:"batchStatus"`
	BatchItemsCount *int64                     `json:"batchItemsCount"`
	UserID          *string                    `json:"userId"`
	RequestTime     *string                    `json:"requestTime"`
	Region          *string                    `json:"region"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

// BulkSignInRequestsSummary is the summary envelope; Requests holds the
// items that Get-CsSdgBulkSignInRequestsSummary writes to the pipeline.
type BulkSignInRequestsSummary struct {
	Requests  []*BulkSignInRequestSummary `json:"requests"`
	RawFields map[string]json.RawMessage  `json:"-"`
}

// BulkSignInItemStatus is one device's result within a batch.
type BulkSignInItemStatus struct {
	UserName   *string                    `json:"userName"`
	HardwareID *string                    `json:"hardwareId"`
	Status     *string                    `json:"status"`
	Error      *string                    `json:"error"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

// Failed reports whether the service recorded an error for this device.
func (s *BulkSignInItemStatus) Failed() bool { return s != nil && s.Error != nil && *s.Error != "" }

// BulkSignInRequestStatus is one batch's status (SdgBulkSignInRequestStatusResult).
// A batch can complete with some devices failed; inspect BatchItems.
type BulkSignInRequestStatus struct {
	BatchID     *string                    `json:"batchId"`
	BatchStatus *string                    `json:"batchStatus"`
	UserID      *string                    `json:"userId"`
	RequestTime *string                    `json:"requestTime"`
	Region      *string                    `json:"region"`
	BatchItems  []*BulkSignInItemStatus    `json:"batchItems"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

// DeviceTaggingRequest is the tagging body (SdgDeviceTaggingRequest). The
// module serializes these exact, mixed-case property names.
type DeviceTaggingRequest struct {
	OCEUserName *string `json:"OCEUserName,omitempty"`
	HardwareID  *string `json:"HardwareId,omitempty"`
	TenantID    *string `json:"TenantId,omitempty"`
	SdhRegion   *string `json:"SdhRegion,omitempty"`
	IcmID       *string `json:"IcmId,omitempty"`
}

// DeviceTaggingResponse is the tagging result (SdgDeviceTaggingResponse).
type DeviceTaggingResponse struct {
	CorrelationID *string                    `json:"correlationId"`
	Status        *string                    `json:"status"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

// DeviceTransferRequest selects the device and both regions. All are required.
type DeviceTransferRequest struct {
	SdhDeviceID, SourceRegion, TargetRegion string
}

// DeviceTransferResponse is the transfer result (SdgDeviceTransferResponse).
type DeviceTransferResponse struct {
	Message   *string                    `json:"Message"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func decodeFields[T any](raw []byte, value *T, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

func (v *BulkSignInResponse) UnmarshalJSON(raw []byte) error {
	type plain BulkSignInResponse
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BulkSignInResponse(value)
	return nil
}

func (v *BulkSignInRequestSummary) UnmarshalJSON(raw []byte) error {
	type plain BulkSignInRequestSummary
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BulkSignInRequestSummary(value)
	return nil
}

func (v *BulkSignInRequestsSummary) UnmarshalJSON(raw []byte) error {
	type plain BulkSignInRequestsSummary
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BulkSignInRequestsSummary(value)
	return nil
}

func (v *BulkSignInItemStatus) UnmarshalJSON(raw []byte) error {
	type plain BulkSignInItemStatus
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BulkSignInItemStatus(value)
	return nil
}

func (v *BulkSignInRequestStatus) UnmarshalJSON(raw []byte) error {
	type plain BulkSignInRequestStatus
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = BulkSignInRequestStatus(value)
	return nil
}

func (v *DeviceTaggingResponse) UnmarshalJSON(raw []byte) error {
	type plain DeviceTaggingResponse
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = DeviceTaggingResponse(value)
	return nil
}

func (v *DeviceTransferResponse) UnmarshalJSON(raw []byte) error {
	type plain DeviceTransferResponse
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = DeviceTransferResponse(value)
	return nil
}

// Status models carry user principal names and hardware IDs; keep them out of logs.
func (BulkSignInItemStatus) String() string        { return "bulk sign-in item status (redacted)" }
func (v BulkSignInItemStatus) GoString() string    { return v.String() }
func (BulkSignInRequestStatus) String() string     { return "bulk sign-in request status (redacted)" }
func (v BulkSignInRequestStatus) GoString() string { return v.String() }
func (DeviceTaggingRequest) String() string        { return "device tagging request (redacted)" }
func (v DeviceTaggingRequest) GoString() string    { return v.String() }
