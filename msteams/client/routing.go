package client

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// ErrDestination means a route is not approved for this backend and cloud.
var ErrDestination = errors.New("destination is not allowed for this backend and cloud")

// ErrRequest means local request construction failed. Its text never includes input.
var ErrRequest = errors.New("invalid backend request")

// ValidateRequest checks a request's method, destination and headers without
// acquiring credentials or sending it. Batch adapters use the same route boundary
// for their nested operations. It does not serialize or validate resource bodies.
func (s *Scope) ValidateRequest(request Request) error {
	if s == nil || s.session == nil {
		return ErrClosed
	}
	profile, err := s.profile()
	if err != nil {
		return err
	}
	if _, err = profile.destination(request); err != nil {
		return err
	}
	switch request.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
	default:
		return ErrRequest
	}
	_, err = requestHeaders(request.Header)
	return err
}

type routeProfile struct {
	host     string
	audience string
	graph    bool
}

func (s *Scope) profile() (routeProfile, error) {
	s.session.mu.RLock()
	defer s.session.mu.RUnlock()
	if s.session.closed || s.session.auth == nil {
		return routeProfile{}, ErrClosed
	}
	environment := s.session.environment
	switch s.backend {
	case Graph:
		return routeProfile{host: strings.TrimPrefix(environment.Endpoints.Graph, "https://"), audience: environment.Endpoints.Graph, graph: true}, nil
	case TeamsAdmin:
		return routeProfile{host: strings.TrimPrefix(environment.Endpoints.TeamsAdmin, "https://"), audience: environment.Endpoints.TeamsAdminResource}, nil
	default:
		return routeProfile{}, ErrDestination
	}
}

func (p routeProfile) destination(request Request) (*url.URL, error) {
	var destination *url.URL
	if request.URL != "" {
		if len(request.Path) != 0 || len(request.Query) != 0 {
			return nil, ErrRequest
		}
		parsed, err := url.Parse(request.URL)
		if err != nil {
			return nil, ErrDestination
		}
		destination = parsed
	} else {
		if len(request.Path) == 0 {
			return nil, ErrRequest
		}
		segments := make([]string, len(request.Path))
		for i, segment := range request.Path {
			if segment == "" || segment == "." || segment == ".." || invalidPath(segment) {
				return nil, ErrRequest
			}
			segments[i] = url.PathEscape(segment)
		}
		parsed, err := url.Parse("https://" + p.host + "/" + strings.Join(segments, "/"))
		if err != nil {
			return nil, ErrRequest
		}
		parsed.RawQuery = request.Query.Encode()
		destination = parsed
	}
	if !p.allows(destination) {
		return nil, ErrDestination
	}
	return destination, nil
}

func (p routeProfile) allows(u *url.URL) bool {
	if u == nil || u.Scheme != "https" || u.Host != p.host || u.User != nil ||
		u.Opaque != "" || u.Fragment != "" || u.RawFragment != "" || invalidPath(u.Path) {
		return false
	}
	if !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return false
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	if p.graph {
		first := strings.Split(strings.TrimPrefix(u.EscapedPath(), "/"), "/")[0]
		if first != "v1.0" && first != "beta" {
			return false
		}
	}
	_, err := url.ParseQuery(u.RawQuery)
	return err == nil
}

func invalidPath(value string) bool {
	return strings.ContainsRune(value, '\\') || strings.ContainsFunc(value, unicode.IsControl)
}

func requestHeaders(headers http.Header) (http.Header, error) {
	result := make(http.Header, len(headers)+2)
	for name, values := range headers {
		if name == "" || strings.ContainsFunc(name, func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && !strings.ContainsRune("!#$%&'*+-.^_`|~", r)
		}) {
			return nil, ErrRequest
		}
		name = http.CanonicalHeaderKey(name)
		switch name {
		case "Authorization", "Proxy-Authorization", "Host", "Cookie", "Connection", "Proxy-Connection",
			"Content-Length", "Transfer-Encoding", "Trailer", "Upgrade", "Keep-Alive", "Te", "Accept-Encoding":
			return nil, ErrRequest
		}
		for _, value := range values {
			if strings.ContainsFunc(value, func(r rune) bool { return r == 127 || r < 32 && r != '\t' }) {
				return nil, ErrRequest
			}
			result.Add(name, value)
		}
	}
	return result, nil
}
