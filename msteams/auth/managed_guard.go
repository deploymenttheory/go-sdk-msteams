package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/managedidentity"
)

// managedGuard confines Azure's pipeline to its selected local platform endpoint.
// A pipeline policy preserves Service Fabric's library-owned certificate pinning.
// It is instantiated per acquisition and called synchronously by that credential.
type managedGuard struct {
	endpoint           *url.URL
	source             managedidentity.Source
	environment        map[string]string
	clientID, audience string
	status             int
	failure            error
}

func newManagedGuard(source managedidentity.Source, env map[string]string, clientID, audience string) (*managedGuard, error) {
	var endpoint string
	switch source {
	case managedidentity.DefaultToIMDS:
		endpoint = "http://169.254.169.254/metadata/identity/oauth2/token"
	case managedidentity.AppService, managedidentity.ServiceFabric:
		endpoint = env["IDENTITY_ENDPOINT"]
	case managedidentity.CloudShell, managedidentity.AzureML:
		endpoint = env["MSI_ENDPOINT"]
	case managedidentity.AzureArc:
		endpoint = env["IDENTITY_ENDPOINT"]
		if endpoint == "" {
			endpoint = "http://127.0.0.1:40342/metadata/identity/oauth2/token"
		}
	default:
		return nil, ErrManagedIdentityUnavailable
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" || u.RawFragment != "" || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.RawPath != "" || strings.ContainsAny(u.Path, "\\\r\n") {
		return nil, ErrAuthorityDestination
	}
	return &managedGuard{endpoint: u, source: source, environment: env, clientID: clientID, audience: audience}, nil
}
func (g *managedGuard) Do(request *policy.Request) (*http.Response, error) {
	raw := request.Raw()
	if !sameManagedEnvironment(g.environment) {
		g.failure = ErrIdentityEnvironmentChanged
		closeManagedBody(raw)
		return nil, g.failure
	}
	if !g.validRequest(raw) {
		g.failure = ErrAuthorityDestination
		closeManagedBody(raw)
		return nil, g.failure
	}
	response, err := request.Next()
	if err != nil {
		return nil, err
	}
	g.status = response.StatusCode
	const limit = 1 << 20
	content, readErr := io.ReadAll(io.LimitReader(response.Body, limit+1))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || len(content) > limit {
		g.failure = ErrManagedIdentityUnavailable
		return nil, g.failure
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		var echo struct {
			Resource string `json:"resource"`
			ClientID string `json:"client_id"`
		}
		if json.Unmarshal(content, &echo) != nil {
			g.failure = ErrInvalidToken
			return nil, g.failure
		}
		if echo.Resource != "" && echo.Resource != g.audience || echo.ClientID != "" && g.clientID != "" && !strings.EqualFold(echo.ClientID, g.clientID) {
			g.failure = ErrBindingMismatch
			return nil, g.failure
		}
	}
	response.Body = io.NopCloser(bytes.NewReader(content))
	response.ContentLength = int64(len(content))
	return response, nil
}
func (g *managedGuard) validRequest(r *http.Request) bool {
	u := r.URL
	if u == nil || u.Scheme != g.endpoint.Scheme || u.Host != g.endpoint.Host || u.EscapedPath() != g.endpoint.EscapedPath() || u.User != nil || u.Opaque != "" || u.Fragment != "" || r.Host != "" && r.Host != u.Host {
		return false
	}
	if g.source == managedidentity.CloudShell {
		if r.Method != http.MethodPost || r.Body == nil {
			return false
		}
		content, err := io.ReadAll(io.LimitReader(r.Body, 8193))
		closeManagedBody(r)
		if err != nil || len(content) > 8192 {
			return false
		}
		r.Body = io.NopCloser(bytes.NewReader(content))
		values, err := url.ParseQuery(string(content))
		return err == nil && len(values["resource"]) == 1 && values.Get("resource") == g.audience
	}
	if r.Method != http.MethodGet {
		return false
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(values["resource"]) != 1 || values.Get("resource") != g.audience {
		return false
	}
	identity := g.clientID
	key := "client_id"
	if g.source == managedidentity.AzureML {
		key = "clientid"
		if identity == "" {
			identity = g.environment["DEFAULT_IDENTITY_CLIENT_ID"]
		}
	}
	return values.Get(key) == identity
}
func closeManagedBody(r *http.Request) {
	if r != nil && r.Body != nil {
		_ = r.Body.Close()
	}
}
