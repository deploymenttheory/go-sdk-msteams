package teamtemplate

import "encoding/json"

// Identity models the two selectors used by the PowerShell InputObject sets.
// Extract ODataID for Get/Update/Delete and Locale for Create. Neither field is
// treated as a destination URL; ODataID remains a single escaped path segment.
type Identity struct {
	ODataID string `json:"odataId,omitempty"`
	Locale  string `json:"locale,omitempty"`
}

// Template is a detail view. Definition projects the flattened known model;
// RawFields retains the original object, including nested/unknown fields and
// explicit nulls. JSON encoding preserves that original response snapshot.
// Use Definition as the explicit write input, not the read snapshot.
type Template struct {
	Definition
	ID        *string                    `json:"id,omitempty"`
	ODataID   *string                    `json:"@odata.id,omitempty"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (t *Template) UnmarshalJSON(raw []byte) error {
	fields, ok := object(raw)
	if t == nil || !ok {
		return ErrInput
	}
	definition, err := ParseDefinition(raw)
	if err != nil {
		return err
	}
	flat := flatten(fields)
	*t = Template{Definition: *definition, ID: stringField(flat, "id"), ODataID: stringField(flat, "@odata.id"), RawFields: fields}
	return nil
}

func (t Template) MarshalJSON() ([]byte, error) {
	if t.RawFields != nil {
		return json.Marshal(t.RawFields)
	}
	raw, err := json.Marshal(t.Definition)
	if err != nil {
		return nil, err
	}
	fields, _ := object(raw)
	for key, value := range map[string]*string{"id": t.ID, "@odata.id": t.ODataID} {
		if value != nil {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			fields[key] = encoded
		}
	}
	return json.Marshal(fields)
}

// TemplateSummary is the typed list/create/update response. Optional values and
// the raw snapshot preserve server omission, zero, empty and future fields.
type TemplateSummary struct {
	ID               *string                    `json:"id,omitempty"`
	Name             *string                    `json:"name,omitempty"`
	Scope            *string                    `json:"scope,omitempty"`
	Description      *string                    `json:"description,omitempty"`
	ShortDescription *string                    `json:"shortDescription,omitempty"`
	IconURI          *string                    `json:"iconUri,omitempty"`
	ChannelCount     *int64                     `json:"channelCount,omitempty"`
	AppCount         *int64                     `json:"appCount,omitempty"`
	ModifiedOn       *string                    `json:"modifiedOn,omitempty"`
	ModifiedBy       *string                    `json:"modifiedBy,omitempty"`
	Locale           *string                    `json:"locale,omitempty"`
	ODataID          *string                    `json:"@odata.id,omitempty"`
	RawFields        map[string]json.RawMessage `json:"-"`
}

func (s *TemplateSummary) UnmarshalJSON(raw []byte) error {
	fields, ok := object(raw)
	if s == nil || !ok {
		return ErrInput
	}
	*s = TemplateSummary{ID: stringField(fields, "id"), Name: stringField(fields, "name"),
		Scope: stringField(fields, "scope"), Description: stringField(fields, "description"),
		ShortDescription: stringField(fields, "shortDescription"), IconURI: stringField(fields, "iconUri"),
		ChannelCount: scalar[int64](fields["channelCount"]), AppCount: scalar[int64](fields["appCount"]),
		ModifiedOn: stringField(fields, "modifiedOn"), ModifiedBy: stringField(fields, "modifiedBy"),
		Locale: stringField(fields, "locale"), ODataID: stringField(fields, "@odata.id"), RawFields: fields}
	return nil
}

func (s TemplateSummary) MarshalJSON() ([]byte, error) {
	if s.RawFields != nil {
		return json.Marshal(s.RawFields)
	}
	type plain TemplateSummary
	return json.Marshal(plain(s))
}

// ListOptions preserves omitted locale versus an explicitly empty value.
// Nil options/locale select the wrapper's evidenced en-US default.
type ListOptions struct {
	PublicTemplateLocale *string
}

// ListResult contains the immediate collection, without an invented paginator.
// IsIncomplete exposes the module's TemplatesRetrievalFailure warning condition.
// Diagnostics and raw response headers are untrusted service data for deliberate
// inspection, never automatic logging. JSON encoding retains the wire envelope.
type ListResult struct {
	Items        []TemplateSummary          `json:"value"`
	IsIncomplete bool                       `json:"-"`
	Diagnostics  []string                   `json:"-"`
	RawFields    map[string]json.RawMessage `json:"-"`
}

func (r ListResult) MarshalJSON() ([]byte, error) {
	if r.RawFields != nil {
		return json.Marshal(r.RawFields)
	}
	type plain ListResult
	return json.Marshal(plain(r))
}

// ErrorDetail preserves the template service's errorCode/message/action model and
// unknown fields. These values can contain tenant data and are never formatted
// by Error.Error, String or GoString.
type ErrorDetail struct {
	Code      *string                    `json:"code,omitempty"`
	Message   *string                    `json:"message,omitempty"`
	Action    *string                    `json:"action,omitempty"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (e *ErrorDetail) UnmarshalJSON(raw []byte) error {
	fields, ok := object(raw)
	if e == nil || !ok {
		return ErrInput
	}
	*e = *errorDetail(fields)
	return nil
}

func errorDetail(fields map[string]json.RawMessage) *ErrorDetail {
	key := "code"
	if _, present := fields["errorCode"]; present {
		key = "errorCode"
	}
	return &ErrorDetail{Code: stringField(fields, key), Message: stringField(fields, "message"), Action: stringField(fields, "action"), RawFields: fields}
}

func (e ErrorDetail) MarshalJSON() ([]byte, error) {
	if e.RawFields != nil {
		return json.Marshal(e.RawFields)
	}
	type plain ErrorDetail
	return json.Marshal(plain(e))
}

type ErrorEnvelope struct {
	OperationID *string                    `json:"operationId,omitempty"`
	Errors      []*ErrorDetail             `json:"errors,omitempty"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (e *ErrorEnvelope) UnmarshalJSON(raw []byte) error {
	fields, ok := object(raw)
	if e == nil || !ok {
		return ErrInput
	}
	*e = ErrorEnvelope{OperationID: stringField(fields, "operationId"), RawFields: fields,
		Errors: objectItems(fields["errors"], errorDetail)}
	return nil
}

func (e ErrorEnvelope) MarshalJSON() ([]byte, error) {
	if e.RawFields != nil {
		return json.Marshal(e.RawFields)
	}
	type plain ErrorEnvelope
	return json.Marshal(plain(e))
}
