// Package serviceprofile owns the closed origin/audience pairs shared by native
// authentication and transport for non-ConfigAPI Teams services. It does not
// select management operations, infer authorization, or perform discovery.
package serviceprofile

// Service selects a fixed resource protocol independently of a PowerShell release.
type Service uint8

const (
	Targeting Service = iota + 1
	Hierarchy
)

// Endpoint separates the request origin, fixed namespace and token audience.
// The current two profiles both use audience + /.default; explicit delegated
// scopes for other services must introduce their own scope/cache contract.
type Endpoint struct {
	Origin     string
	Audience   string
	PathPrefix string
}

// Resolve accepts only the canonical cloud names from auth.ResolveEnvironment.
// These pairs derive from the pinned special-service profile contract. Missing
// maps are local routing gaps, not proof that a service rejects a cloud or mode.
func Resolve(service Service, cloud string) (Endpoint, bool) {
	switch service {
	case Targeting:
		switch cloud {
		case "commercial":
			return Endpoint{Origin: "https://retailservices.teams.microsoft.com", Audience: "https://retailservices.teams.microsoft.com", PathPrefix: "/api/rest/v1/nodes"}, true
		case "gcc":
			return Endpoint{Origin: "https://retailservices.gcc.teams.microsoft.com", Audience: "https://retailservices.teams.microsoft.com", PathPrefix: "/api/rest/v1/nodes"}, true
		}
	case Hierarchy:
		if cloud == "commercial" {
			return Endpoint{Origin: "https://hierarchy.teams.microsoft.com", Audience: "https://chatsvcagg.teams.microsoft.com", PathPrefix: "/api/v1/hierarchies"}, true
		}
	}
	return Endpoint{}, false
}
