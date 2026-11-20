package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (a *delegatedHTTP) trustFederation(endpoints []string) error {
	a.federation = make(map[string]*url.URL, len(endpoints))
	for _, value := range endpoints {
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || u.RawPath != "" || u.Hostname() == "login.microsoftonline.com" || !strings.HasPrefix(u.Path, "/") || strings.ContainsAny(u.Path, "\\\r\n") {
			return &ConfigError{Field: "federation_endpoints", Reason: "requires exact HTTPS endpoint URLs without credentials, query or fragment"}
		}
		for _, segment := range strings.Split(u.Path, "/") {
			if segment == "." || segment == ".." {
				return &ConfigError{Field: "federation_endpoints", Reason: "must not contain dot segments"}
			}
		}
		a.federation[value] = u
	}
	return nil
}

// The endpoint was explicitly trusted by the caller, independently of user-realm
// and MEX metadata. MSAL may follow metadata only to these exact configured URLs.
func (a *delegatedHTTP) sendFederation(request *http.Request, target *url.URL) (*http.Response, error) {
	if request.Method != http.MethodGet && request.Method != http.MethodPost || request.Host != "" && request.Host != target.Host {
		return rejectAuthorityRequest(request)
	}
	trusted := request.Clone(request.Context())
	destination := *target
	trusted.URL = &destination
	trusted.Host = target.Host
	httpClient := *a.base.client
	httpClient.Jar = nil
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrAuthorityDestination }
	response, err := httpClient.Do(trusted)
	if err != nil {
		return nil, err
	}
	const limit = 1 << 20
	content, readErr := io.ReadAll(io.LimitReader(response.Body, limit+1))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || len(content) > limit {
		return nil, ErrDelegatedAuthentication
	}
	response.Body = io.NopCloser(bytes.NewReader(content))
	response.ContentLength = int64(len(content))
	return response, nil
}
