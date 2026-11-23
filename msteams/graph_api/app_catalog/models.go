package appcatalog

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	// ErrInput identifies a request the SDK rejects before sending it.
	ErrInput = errors.New("invalid app catalog request")
	// ErrResponse identifies an undeclared successful HTTP status.
	ErrResponse = errors.New("unexpected app catalog response")
)

// OrganizationDistribution is the only distribution method New-TeamsApp
// accepts (compared case-insensitively).
const OrganizationDistribution = "organization"

// TeamsApp is a Graph teamsApp from the tenant app catalog. RawFields keeps
// the complete object, including expanded appDefinitions.
type TeamsApp struct {
	ID                 string                     `json:"id"`
	ExternalID         *string                    `json:"externalId"`
	DisplayName        *string                    `json:"displayName"`
	DistributionMethod *string                    `json:"distributionMethod"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *TeamsApp) UnmarshalJSON(raw []byte) error {
	type plain TeamsApp
	var value plain
	if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(raw, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = TeamsApp(value)
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
