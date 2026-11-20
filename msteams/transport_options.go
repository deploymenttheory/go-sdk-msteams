package msteams

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// WithTransport borrows a RoundTripper. The SDK still applies its destination,
// authentication and cancellation guards around it. It must honor request context.
// Later WithHTTPClient/WithTransport options replace earlier transport selections.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(o *options) error {
		if nilValue(transport) {
			return fmt.Errorf("transport is required: %w", ErrInvalidOption)
		}
		if o.http == nil {
			o.http = &http.Client{}
		}
		o.http.Transport = transport
		return nil
	}
}

// WithProxy configures an explicit HTTP/HTTPS/SOCKS5 proxy on an owned clone of
// the selected *http.Transport. It cannot configure an opaque custom RoundTripper.
// Proxy credentials remain transport configuration and are never diagnostic fields.
func WithProxy(raw string) ClientOption {
	return func(o *options) error {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "socks5" && u.Scheme != "socks5h") || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
			return fmt.Errorf("invalid proxy configuration: %w", ErrInvalidOption)
		}
		o.proxy = u
		return nil
	}
}

// WithTLSClientConfig snapshots TLS settings on an owned clone of the selected
// *http.Transport. TLS 1.2 or newer and server verification are required. Certificate
// callbacks, private-key signers and session caches remain borrowed and must be safe
// for concurrent use. TLS key logging is not enabled through this option.
func WithTLSClientConfig(configuration *tls.Config) ClientOption {
	copy := cloneTLSConfig(configuration)
	return func(o *options) error {
		if copy == nil {
			return fmt.Errorf("TLS configuration is required: %w", ErrInvalidOption)
		}
		minimum := uint16(tls.VersionTLS12)
		if len(copy.EncryptedClientHelloConfigList) > 0 {
			minimum = tls.VersionTLS13
		}
		if copy.InsecureSkipVerify || copy.KeyLogWriter != nil || copy.MinVersion != 0 && copy.MinVersion < minimum || copy.MaxVersion != 0 && (copy.MaxVersion < minimum || copy.MinVersion > copy.MaxVersion) {
			return fmt.Errorf("TLS requires verified servers, supported versions and no key logging: %w", ErrInvalidOption)
		}
		o.tlsConfig = cloneTLSConfig(copy)
		if o.tlsConfig.MinVersion == 0 {
			o.tlsConfig.MinVersion = minimum
		}
		return nil
	}
}

// WithGlobalHeaders copies management API defaults. Per-request headers take
// precedence. Credentials, cookies, host/framing headers are rejected by the session.
// Defaults do not apply to identity-provider requests.
func WithGlobalHeaders(headers map[string]string) ClientOption {
	copy := make(map[string]string, len(headers))
	for key, value := range headers {
		copy[key] = value
	}
	return func(o *options) error {
		if o.headers == nil {
			o.headers = make(http.Header)
		}
		seen := make(map[string]bool, len(copy))
		for key, value := range copy {
			canonical := http.CanonicalHeaderKey(key)
			if seen[canonical] {
				return fmt.Errorf("duplicate default header: %w", ErrInvalidOption)
			}
			seen[canonical] = true
			o.headers.Set(key, value)
		}
		return nil
	}
}

// WithGlobalHeader sets one management API default, under the protected-header guard.
func WithGlobalHeader(key, value string) ClientOption {
	return WithGlobalHeaders(map[string]string{key: value})
}

// WithUserAgent sets a nonempty management API User-Agent.
func WithUserAgent(value string) ClientOption {
	return func(o *options) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("user agent is required: %w", ErrInvalidOption)
		}
		return WithGlobalHeader("User-Agent", value)(o)
	}
}

// WithDebug enables sanitized attempt logs on the caller's logger. The logger must
// allow Debug level. Bodies, URLs, headers, credentials and service text are omitted.
func WithDebug() ClientOption { return func(o *options) error { o.debug = true; return nil } }

// WithMaxConcurrentRequests sets the shared Graph/Teams in-flight attempt limit.
// A later WithRequestPolicy replaces this value along with the other policy fields.
func WithMaxConcurrentRequests(limit int) ClientOption {
	return func(o *options) error {
		if limit < 1 || limit > 1024 {
			return fmt.Errorf("concurrency limit must be between 1 and 1024: %w", ErrInvalidOption)
		}
		o.requestPolicy.MaxConcurrent = limit
		return nil
	}
}

func cloneTLSConfig(source *tls.Config) *tls.Config {
	if source == nil {
		return nil
	}
	copy := source.Clone()
	if source.RootCAs != nil {
		copy.RootCAs = source.RootCAs.Clone()
	}
	if source.ClientCAs != nil {
		copy.ClientCAs = source.ClientCAs.Clone()
	}
	copy.NextProtos = slices.Clone(source.NextProtos)
	copy.CipherSuites = slices.Clone(source.CipherSuites)
	copy.CurvePreferences = slices.Clone(source.CurvePreferences)
	copy.EncryptedClientHelloConfigList = slices.Clone(source.EncryptedClientHelloConfigList)
	copy.EncryptedClientHelloKeys = slices.Clone(source.EncryptedClientHelloKeys)
	for i := range copy.EncryptedClientHelloKeys {
		copy.EncryptedClientHelloKeys[i].Config = slices.Clone(source.EncryptedClientHelloKeys[i].Config)
		copy.EncryptedClientHelloKeys[i].PrivateKey = slices.Clone(source.EncryptedClientHelloKeys[i].PrivateKey)
	}
	copy.Certificates = slices.Clone(source.Certificates)
	for i := range copy.Certificates {
		cert := &copy.Certificates[i]
		cert.Certificate = slices.Clone(cert.Certificate)
		for j := range cert.Certificate {
			cert.Certificate[j] = slices.Clone(cert.Certificate[j])
		}
		cert.OCSPStaple = slices.Clone(cert.OCSPStaple)
		cert.SignedCertificateTimestamps = slices.Clone(cert.SignedCertificateTimestamps)
		for j := range cert.SignedCertificateTimestamps {
			cert.SignedCertificateTimestamps[j] = slices.Clone(cert.SignedCertificateTimestamps[j])
		}
		cert.SupportedSignatureAlgorithms = slices.Clone(cert.SupportedSignatureAlgorithms)
		if cert.Leaf != nil {
			// Reparse owned DER rather than sharing the caller's mutable parsed leaf.
			leaf, err := x509.ParseCertificate(slices.Clone(cert.Leaf.Raw))
			if err == nil {
				cert.Leaf = leaf
			} else {
				cert.Leaf = nil
			}
		}
	}
	// Deprecated certificate-name lookup can alias caller certificates; clients use
	// Certificates/GetClientCertificate and do not need a server-side name lookup.
	//nolint:staticcheck // SA1019: deliberately clear the deprecated lookup to prevent retaining caller-owned certificate pointers.
	copy.NameToCertificate = nil
	return copy
}
