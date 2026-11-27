package appcatalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	// ErrInput identifies a request the SDK rejects before sending it.
	ErrInput = errors.New("invalid app catalog request")
	// ErrResponse is retained for earlier callers. Current operations accept any
	// 2xx receipt and report invalid response content as client.ErrDecode.
	ErrResponse = errors.New("unexpected app catalog response")
)

// OrganizationDistribution is the only distribution method New-TeamsApp
// accepts (compared case-insensitively).
const OrganizationDistribution = "organization"

// TeamsApp is a Graph teamsApp from the tenant app catalog. RawFields keeps
// the complete object, including expanded appDefinitions. ID and the other
// strings preserve source nullability; string-valued numeric/bool projections
// retain their original number lexemes and lowercase boolean text.
type TeamsApp struct {
	ID                 *string                    `json:"id"`
	ExternalID         *string                    `json:"externalId"`
	DisplayName        *string                    `json:"displayName"`
	DistributionMethod *string                    `json:"distributionMethod"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

// coercedString follows the pinned JsonReader string contract without passing
// opaque identifiers through float64. Missing/null values remain nil.
type coercedString struct{ value *string }

func (v *coercedString) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		v.value = nil
		return nil
	}
	if len(raw) == 0 {
		return client.ErrDecode
	}
	var text string
	switch raw[0] {
	case '"':
		if json.Unmarshal(raw, &text) != nil {
			return client.ErrDecode
		}
	case 't', 'f':
		var flag bool
		if json.Unmarshal(raw, &flag) != nil {
			return client.ErrDecode
		}
		text = strconv.FormatBool(flag)
	default:
		var number json.Number
		if json.Unmarshal(raw, &number) != nil {
			return client.ErrDecode
		}
		text = number.String()
	}
	v.value = &text
	return nil
}

func (v *TeamsApp) UnmarshalJSON(raw []byte) error {
	var value struct {
		ID                 coercedString `json:"id"`
		ExternalID         coercedString `json:"externalId"`
		DisplayName        coercedString `json:"displayName"`
		DistributionMethod coercedString `json:"distributionMethod"`
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil || json.Unmarshal(raw, &value) != nil {
		return client.ErrDecode
	}
	*v = TeamsApp{ID: value.ID.value, ExternalID: value.ExternalID.value, DisplayName: value.DisplayName.value, DistributionMethod: value.DistributionMethod.value, RawFields: fields}
	return nil
}

// ListOptions are Get-TeamsApp's filters. Each non-blank value adds one
// "<property> eq '<value>'" clause, joined with " and " in the module's order
// (id, externalId, displayName, distributionMethod). Single quotes are doubled
// for OData; the module sends them unescaped, which Graph rejects with 400.
type ListOptions struct {
	ID                 string
	ExternalID         string
	DisplayName        string
	DistributionMethod string
	Paging             client.PagingOptions
}

// CreateRequest is New-TeamsApp. Package is the app package (.zip) bytes,
// sent unchanged as application/zip.
type CreateRequest struct {
	DistributionMethod string
	Package            []byte
}
