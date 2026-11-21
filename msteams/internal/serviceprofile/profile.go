// Package serviceprofile owns the closed origin/audience pairs shared by native
// authentication and transport for non-ConfigAPI Teams services. It does not
// select management operations, infer authorization, or perform discovery.
package serviceprofile

// Service selects a fixed resource protocol independently of a PowerShell release.
type Service uint8

const (
	Targeting Service = iota + 1
	Hierarchy
	PrivateChannelMigration
	Regions
)

// Endpoint separates the request origin, fixed namespace and token audience.
// Empty DelegatedScope uses audience + /.default. A nonempty value requires
// delegated mode and the exact scope in acquisition, cache and refresh keys.
type Endpoint struct {
	Origin         string
	Audience       string
	PathPrefix     string
	DelegatedScope string
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
	case PrivateChannelMigration:
		switch cloud {
		case "commercial":
			return Endpoint{Origin: "https://teams.microsoft.com", Audience: "https://templates.teams.microsoft.com", PathPrefix: "/fabric/templates/api"}, true
		case "gcc":
			return Endpoint{Origin: "https://teams.microsoft.com", Audience: "https://templates.teams.microsoft.com", PathPrefix: "/fabric/gcc/templates/api"}, true
		case "gcch":
			return Endpoint{Origin: "https://gov.teams.microsoft.us", Audience: "https://templates.teams.microsoft.us", PathPrefix: "/fabric/gov/templates/api"}, true
		case "dod":
			return Endpoint{Origin: "https://dod.teams.microsoft.us", Audience: "https://templates.teams.microsoft.us", PathPrefix: "/fabric/dod/templates/api"}, true
		case "china":
			return Endpoint{Origin: "https://teams.microsoftonline.cn", Audience: "https://templates.teams.microsoftonline.cn", PathPrefix: "/fabric/templates/api"}, true
		}
	case Regions:
		if cloud == "commercial" || cloud == "gcc" {
			return Endpoint{Origin: "https://teams.cloud.microsoft", Audience: "https://migrationsvc.teams.microsoft.com", PathPrefix: "/api/tms/prod", DelegatedScope: "https://migrationsvc.teams.microsoft.com/Region.Read.All"}, true
		}
	}
	return Endpoint{}, false
}
