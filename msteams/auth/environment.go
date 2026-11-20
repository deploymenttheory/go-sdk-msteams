package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"unicode"
)

// Endpoints selects explicit HTTPS service origins and a Teams token audience.
// Graph is both the Graph origin and resource audience. Nonempty fields override
// the selected built-in profile; a custom profile requires all four fields.
// These are trusted caller configuration, never learned from service responses.
type Endpoints struct {
	Authority          string
	Graph              string
	TeamsAdmin         string
	TeamsAdminResource string
}

// Environment is a resolved, immutable-by-copy endpoint tuple. BindingCloud is
// the cache-isolation identity: explicit overrides include a tuple fingerprint so
// otherwise-identical tenants/apps cannot share tokens across different authorities.
type Environment struct {
	Name         string
	BindingCloud string
	Endpoints    Endpoints
}

// ResolveEnvironment resolves pinned Teams environment aliases and optional
// explicit endpoint overrides without network discovery. Routing availability is
// not a claim of operation support in that cloud or live tenant qualification.
func ResolveEnvironment(cloud string, overrides *Endpoints) (Environment, error) {
	name := strings.ToLower(cloud)
	var endpoints Endpoints
	switch name {
	case "", "commercial", "teamscloud", "azurecloud", "public":
		name = "commercial"
		endpoints = Endpoints{"https://login.microsoftonline.com", GraphAudience, "https://api.interfaces.records.teams.microsoft.com", TeamsAdminAudience}
	case "gcc":
		endpoints = Endpoints{"https://login.microsoftonline.com", GraphAudience, "https://api.interfaces.records.teams.microsoft.com", TeamsAdminAudience}
	case "gcch", "teamsgcch":
		name = "gcch"
		endpoints = Endpoints{"https://login.microsoftonline.us", "https://graph.microsoft.us", "https://api.interfaces.records.gov.teams.microsoft.us", TeamsAdminAudience}
	case "dod", "teamsdod":
		name = "dod"
		endpoints = Endpoints{"https://login.microsoftonline.us", "https://dod-graph.microsoft.us", "https://api.interfaces.records.dod.teams.microsoft.us", TeamsAdminAudience}
	case "china", "teamschina", "gallatin":
		name = "china"
		endpoints = Endpoints{"https://login.chinacloudapi.cn", "https://microsoftgraph.chinacloudapi.cn", "https://api.interfaces.records.teams.microsoftonline.cn", TeamsAdminAudience}
	case "france", "teamsfrance", "bleu":
		name = "france"
		endpoints = Endpoints{"https://login.sovcloud-identity.fr", "https://graph.svc.sovcloud.fr", "https://api.interfaces.records.communications.svc.sovcloud.fr", TeamsAdminAudience}
	case "germany", "teamsgermany", "delos":
		name = "germany"
		endpoints = Endpoints{"https://login.sovcloud-identity.de", "https://graph.svc.sovcloud.de", "https://api.interfaces.records.communications.svc.sovcloud.de", TeamsAdminAudience}
	case "custom":
		if overrides == nil {
			return Environment{}, ErrUnsupportedCloud
		}
	default:
		return Environment{}, ErrUnsupportedCloud
	}
	if overrides != nil {
		if overrides.Authority != "" {
			endpoints.Authority = overrides.Authority
		}
		if overrides.Graph != "" {
			endpoints.Graph = overrides.Graph
		}
		if overrides.TeamsAdmin != "" {
			endpoints.TeamsAdmin = overrides.TeamsAdmin
		}
		if overrides.TeamsAdminResource != "" {
			endpoints.TeamsAdminResource = overrides.TeamsAdminResource
		}
	}
	for _, field := range []*string{&endpoints.Authority, &endpoints.Graph, &endpoints.TeamsAdmin} {
		origin, ok := httpsOrigin(*field)
		if !ok {
			return Environment{}, &ConfigError{Field: "endpoints", Reason: "requires explicit HTTPS origins without paths, queries or credentials"}
		}
		*field = origin
	}
	if !validResourceAudience(endpoints.TeamsAdminResource) {
		return Environment{}, &ConfigError{Field: "teams_admin_resource", Reason: "requires a resource identifier without scope suffix or whitespace"}
	}
	binding := name
	if overrides != nil {
		digest := sha256.Sum256([]byte(endpoints.Authority + "\x00" + endpoints.Graph + "\x00" + endpoints.TeamsAdmin + "\x00" + endpoints.TeamsAdminResource))
		binding += "@" + hex.EncodeToString(digest[:])
	}
	return Environment{Name: name, BindingCloud: binding, Endpoints: endpoints}, nil
}

func httpsOrigin(value string) (string, bool) {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawFragment != "" || u.Opaque != "" || u.RawPath != "" || (u.Path != "" && u.Path != "/") {
		return "", false
	}
	return "https://" + strings.ToLower(u.Host), true
}

func validResourceAudience(value string) bool {
	if value == "" || strings.ContainsFunc(value, unicode.IsSpace) || strings.ContainsFunc(value, unicode.IsControl) || strings.HasSuffix(value, "/.default") {
		return false
	}
	if tenantGUID.MatchString(value) {
		return true
	}
	u, err := url.Parse(value)
	return err == nil && (u.Scheme == "https" || u.Scheme == "api") && u.Host != "" && u.User == nil && u.Fragment == "" && u.RawQuery == "" && !u.ForceQuery && u.Opaque == ""
}

func (e Environment) supportsAudience(audience string) bool {
	return audience == e.Endpoints.Graph || audience == e.Endpoints.TeamsAdminResource
}

// String avoids printing custom origins from ordinary diagnostics.
func (Endpoints) String() string       { return "msteams endpoints (redacted)" }
func (e Endpoints) GoString() string   { return e.String() }
func (Environment) String() string     { return "msteams environment (redacted)" }
func (e Environment) GoString() string { return e.String() }
