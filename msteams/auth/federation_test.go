package auth

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Synthetic WS-Trust 1.3 shapes follow MSAL v1.10.1's XML model contracts.
const fixtureMEX = `<definitions><Policy Id="fixture"><ExactlyOne><All>
<TransportBinding sp="fixture"/><SignedSupportingTokens><Policy><UsernameToken><Policy><WssUsernameToken10/></Policy></UsernameToken></Policy></SignedSupportingTokens>
</All></ExactlyOne></Policy><binding name="fixture"><PolicyReference URI="#fixture"/>
<binding transport="http://schemas.xmlsoap.org/soap/http"/><operation><operation soapAction="http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue"/></operation></binding>
<service><port binding="fixture"><EndpointReference><Address>https://federation.example.test/password</Address></EndpointReference></port></service></definitions>`
const fixtureSAML = `<Envelope><Body><RequestSecurityTokenResponseCollection><RequestSecurityTokenResponse><RequestedSecurityToken>
<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" AssertionID="fixture-assertion"/>
</RequestedSecurityToken></RequestSecurityTokenResponse></RequestSecurityTokenResponseCollection></Body></Envelope>`

func TestTrustedFederationPasswordUsesMicrosoftWSExchange(t *testing.T) {
	t.Parallel()
	for _, allowPassword := range []bool{false, true} {
		t.Run(map[bool]string{false: "unapproved-password", true: "trusted-password"}[allowPassword], func(t *testing.T) {
			service := &delegatedService{t: t, federated: true, federationMetadata: "https://federation.example.test/mex"}
			var passwordPosts int
			transport := tokenRoundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host == "login.microsoftonline.com" {
					return service.RoundTrip(r)
				}
				payload := fixtureMEX
				switch r.URL.String() {
				case "https://federation.example.test/mex":
					if r.Method != http.MethodGet {
						t.Error("MEX request method changed")
					}
				case "https://federation.example.test/password":
					passwordPosts++
					if !allowPassword {
						t.Error("unapproved federation endpoint received credentials")
					}
					body, err := io.ReadAll(r.Body)
					_ = r.Body.Close()
					if err != nil {
						return nil, err
					}
					if !strings.Contains(string(body), "fixture-password") || !strings.Contains(string(body), "fixture@example.test") {
						t.Error("MSAL SOAP request lost password credentials")
					}
					payload = fixtureSAML
				default:
					t.Error("request escaped federation allowlist")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/soap+xml"}}, Body: io.NopCloser(strings.NewReader(payload)), Request: r}, nil
			})
			cfg := delegatedConfig(Password)
			cfg.FederationEndpoints = []string{"https://federation.example.test/mex"}
			if allowPassword {
				cfg.FederationEndpoints = append(cfg.FederationEndpoints, "https://federation.example.test/password")
			}
			provider, err := NewDelegatedProvider(cfg, &http.Client{Transport: transport})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { provider.Close(); provider.Wait() })
			cfg.FederationEndpoints[0] = "https://mutated.invalid/mex"
			_, err = provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience))
			if allowPassword {
				if err != nil || passwordPosts != 1 || len(service.grants) != 1 {
					t.Fatalf("federated acquisition failed: %v (%v)", err, errors.Unwrap(err))
				}
			} else if err == nil || passwordPosts != 0 || len(service.grants) != 0 {
				t.Fatal("MEX metadata bypassed explicit endpoint trust")
			}
		})
	}
}

func TestFederationEndpointValidation(t *testing.T) {
	t.Parallel()
	for _, endpoint := range []string{"http://federation.test/mex", "https://user:secret@federation.test/mex", "https://federation.test/mex#fragment", "https://federation.test/mex?x=y", "https://federation.test/../mex", "https://login.microsoftonline.com/other/token"} {
		cfg := delegatedConfig(Password)
		cfg.FederationEndpoints = []string{endpoint}
		if _, err := NewDelegatedProvider(cfg, &http.Client{}); !errors.Is(err, ErrInvalidConfig) {
			t.Fatal("ambiguous federation destination accepted")
		}
	}
}
