// Package tenantfederationconfiguration models the modern tenant federation
// configuration and its local domain choices. Local constructors perform no I/O.
package tenantfederationconfiguration

import (
	"encoding/json"
	"errors"
)

var ErrInput = errors.New("invalid tenant federation configuration input")

// DomainPattern is the local domain value built by New-CsEdgeDomainPattern.
// It does not register a domain or prove DNS ownership/reachability.
type DomainPattern struct {
	Domain string `json:"Domain"`
}

// NewDomainPattern requires the bound nonempty string. It does not invent DNS
// syntax, wildcard, ownership or reachability checks absent from the helper.
func NewDomainPattern(domain string) (DomainPattern, error) {
	if domain == "" {
		return DomainPattern{}, ErrInput
	}
	return DomainPattern{Domain: domain}, nil
}

// AllowedDomains is the request-side allow-all/allow-list choice. Its JSON is
// the flattened native request value, not PowerShell's formatting wrapper.
// Response choice/discriminator semantics remain a separate raw-data contract.
type AllowedDomains struct {
	domains []string
	valid   bool
}

// NewAllowList copies string domain values. Omitted/nil/empty emits [] on the
// wire. Values remain ordered and are not normalized or deduplicated. Existing
// DomainPattern values can be passed through their Domain field.
func NewAllowList(domains ...string) *AllowedDomains {
	return &AllowedDomains{domains: append([]string{}, domains...), valid: true}
}

// NewAllowAllKnownDomains builds the pinned request sentinel array. It does
// not assert that a nullable response means the same service-side choice.
func NewAllowAllKnownDomains() *AllowedDomains {
	return NewAllowList("AllowAllKnownDomains")
}

func (v AllowedDomains) MarshalJSON() ([]byte, error) {
	if !v.valid {
		return nil, ErrInput
	}
	return json.Marshal(v.domains)
}

func (DomainPattern) String() string      { return "federation domain pattern (redacted)" }
func (v DomainPattern) GoString() string  { return v.String() }
func (AllowedDomains) String() string     { return "federation allowed domain choice (redacted)" }
func (v AllowedDomains) GoString() string { return v.String() }
