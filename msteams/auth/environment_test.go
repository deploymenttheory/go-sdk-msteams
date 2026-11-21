package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestUnitEnvironmentOverridesValidateAndIsolateBindings(t *testing.T) {
	t.Parallel()
	base := Endpoints{"https://identity.example.test", "https://graph.example.test", "https://teams.example.test", "api://teams.example.test"}
	first, err := ResolveEnvironment("custom", &base)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{first.BindingCloud: true}
	for field := range 4 {
		changed := base
		fields := []*string{&changed.Authority, &changed.Graph, &changed.TeamsAdmin, &changed.TeamsAdminResource}
		*fields[field] += "x"
		next, err := ResolveEnvironment("custom", &changed)
		if err != nil || seen[next.BindingCloud] {
			t.Fatalf("endpoint tuple did not isolate binding: %v", err)
		}
		seen[next.BindingCloud] = true
	}
	canonical := base
	canonical.Authority = "https://IDENTITY.EXAMPLE.TEST/"
	second, err := ResolveEnvironment("CUSTOM", &canonical)
	if err != nil || second != first {
		t.Fatal("equivalent origins produced different bindings")
	}
	for _, bad := range []string{"", "http://identity.test", "https://user:pass@identity.test", "https://identity.test/path", "https://identity.test?", "https://identity.test?q=x", "https://identity.test#fragment", "https://identity.test/%2f"} {
		changed := base
		changed.Authority = bad
		if _, err := ResolveEnvironment("custom", &changed); !errors.Is(err, ErrInvalidConfig) {
			t.Errorf("invalid origin accepted: %q", bad)
		}
	}
	for _, bad := range []string{"", "arbitrary", "https://resource.test/.default", "api://user:pass@resource.test", "api://resource.test?q=x", "api://resource.test#fragment", "resource with spaces"} {
		changed := base
		changed.TeamsAdminResource = bad
		if _, err := ResolveEnvironment("custom", &changed); !errors.Is(err, ErrInvalidConfig) {
			t.Errorf("invalid audience accepted: %q", bad)
		}
	}
	for _, name := range []string{"Noam", "Emea", "Apac", "INT", "government", "custom"} {
		if _, err := ResolveEnvironment(name, nil); !errors.Is(err, ErrUnsupportedCloud) {
			t.Errorf("unknown/internal profile accepted: %q", name)
		}
	}
	for _, value := range []any{base, &base, first, &first} {
		if strings.Contains(fmt.Sprintf("%v %+v %#v", value, value, value), "example.test") {
			t.Fatal("ordinary diagnostic exposed custom endpoints")
		}
	}
}

// These expected tuples come from the pinned Connect/HostContract assemblies;
// the China authority uses the documented current Entra canonical origin.
var environmentCases = []struct{ name, authority, graph, teams string }{
	{"TeamsCloud", "https://login.microsoftonline.com", "https://graph.microsoft.com", "https://api.interfaces.records.teams.microsoft.com"},
	{"gcc", "https://login.microsoftonline.com", "https://graph.microsoft.com", "https://api.interfaces.records.teams.microsoft.com"},
	{"TeamsGCCH", "https://login.microsoftonline.us", "https://graph.microsoft.us", "https://api.interfaces.records.gov.teams.microsoft.us"},
	{"TeamsDOD", "https://login.microsoftonline.us", "https://dod-graph.microsoft.us", "https://api.interfaces.records.dod.teams.microsoft.us"},
	{"TeamsChina", "https://login.chinacloudapi.cn", "https://microsoftgraph.chinacloudapi.cn", "https://api.interfaces.records.teams.microsoftonline.cn"},
	{"TeamsFrance", "https://login.sovcloud-identity.fr", "https://graph.svc.sovcloud.fr", "https://api.interfaces.records.communications.svc.sovcloud.fr"},
	{"TeamsGermany", "https://login.sovcloud-identity.de", "https://graph.svc.sovcloud.de", "https://api.interfaces.records.communications.svc.sovcloud.de"},
	{"custom", "https://identity.example.test:8443", "https://graph.example.test:8443", "https://teams.example.test:8443"},
}

func TestUnitEnvironmentAuthenticationFlows(t *testing.T) {
	t.Parallel()
	pfx, err := os.ReadFile("testdata/client-credentials-sha1.pfx")
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range environmentCases {
		for _, method := range []Method{ClientSecret, Certificate, WorkloadIdentity, Password, DeviceCode, Browser} {
			t.Run(profile.name+"/"+string(method), func(t *testing.T) {
				cfg := &Config{Cloud: profile.name, Method: method, TenantID: appTenant, ClientID: appID}
				resource := TeamsAdminAudience
				if profile.name == "custom" {
					resource = "api://teams.example.test"
					cfg.Endpoints = &Endpoints{profile.authority, profile.graph, profile.teams, resource}
				}
				environment, err := ResolveEnvironment(cfg.Cloud, cfg.Endpoints)
				if err != nil {
					t.Fatal(err)
				}
				if environment.Endpoints != (Endpoints{profile.authority, profile.graph, profile.teams, resource}) {
					t.Fatal("pinned endpoint tuple changed")
				}
				fixture := &environmentIdentity{t: t, authority: profile.authority, resource: resource, graph: profile.graph, method: method}
				switch method {
				case ClientSecret:
					cfg.ClientSecret = "fixture-secret"
				case Certificate:
					cfg.PFX, cfg.PFXPassword = pfx, []byte("fixture-only")
				case WorkloadIdentity:
					cfg.Assertion = func(context.Context) (string, error) { return "fixture-assertion", nil }
				case Password:
					cfg.Username, cfg.Password = "fixture@example.test", "fixture-password"
				case DeviceCode:
					cfg.DevicePrompt = func(_ context.Context, message DeviceCodeMessage) error {
						if message.VerificationURL != profile.authority+"/common/oauth2/deviceauth" {
							t.Error("device instructions escaped authority")
						}
						return nil
					}
				case Browser:
					cfg.OpenBrowser = func(ctx context.Context, target string) error {
						if !strings.HasPrefix(target, profile.authority+"/") {
							t.Error("browser escaped authority")
						}
						return postBrowserResult(ctx, target, false, &delegatedService{})
					}
				}
				httpClient := &http.Client{Transport: fixture}
				var provider Provider
				if method == Password || method == DeviceCode || method == Browser {
					p, err := NewDelegatedProvider(cfg, httpClient)
					if err != nil {
						t.Fatal(err)
					}
					p.authority.wait = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }
					t.Cleanup(func() { p.Close(); p.Wait() })
					provider = p
				} else {
					p, err := NewApplicationProvider(cfg, httpClient)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { p.Close(); p.Wait() })
					provider = p
				}
				for _, audience := range []string{profile.graph, resource} {
					if _, err := provider.AcquireToken(t.Context(), cfg.TokenBinding().Request(audience)); err != nil {
						t.Fatalf("profile exchange failed: %v (%v)", err, errors.Unwrap(err))
					}
				}
				if fixture.posts != 2 {
					t.Fatalf("expected independent audiences, got %d exchanges", fixture.posts)
				}
				if _, err := provider.AcquireToken(t.Context(), cfg.TokenBinding().Request("https://foreign.example.test")); !errors.Is(err, ErrUnsupportedAudience) {
					t.Fatalf("foreign audience accepted: %v", err)
				}
				if fixture.posts != 2 {
					t.Fatal("foreign audience reached identity transport")
				}
			})
		}
	}
}

type environmentIdentity struct {
	t                          *testing.T
	authority, resource, graph string
	method                     Method
	posts                      int
	foreignMetadata            string
}

func (f *environmentIdentity) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme+"://"+r.URL.Host != f.authority {
		f.t.Fatal("identity request crossed authority")
	}
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}
	issuer := f.authority + "/" + appTenant + "/v2.0"
	endpoint := f.authority + "/" + appTenant + "/oauth2/v2.0/token"
	var payload any
	switch {
	case strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration"):
		metadata := map[string]any{"issuer": issuer, "token_endpoint": endpoint, "authorization_endpoint": strings.TrimSuffix(endpoint, "token") + "authorize"}
		if f.foreignMetadata != "" {
			metadata[f.foreignMetadata] = "https://login.microsoftonline.com/" + appTenant + "/oauth2/v2.0/token"
		}
		payload = metadata
	case strings.HasPrefix(r.URL.Path, "/common/UserRealm/"):
		payload = map[string]any{"account_type": "Managed", "domain_name": "example.test", "cloud_instance_name": strings.TrimPrefix(f.authority, "https://"), "cloud_audience_urn": "urn:federation:MicrosoftOnline"}
	case strings.HasSuffix(r.URL.Path, "/devicecode"):
		payload = map[string]any{"device_code": "fixture-device", "user_code": "FIXTURE", "verification_uri": f.authority + "/common/oauth2/deviceauth", "expires_in": 900, "interval": 1}
	case r.Method == http.MethodPost && r.URL.String() == endpoint:
		f.posts++
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		var scope string
		for _, part := range strings.Fields(r.Form.Get("scope")) {
			if strings.HasSuffix(part, "/.default") {
				if scope != "" {
					f.t.Error("mixed resource scopes")
				}
				scope = part
			}
		}
		if scope != f.graph+"/.default" && scope != f.resource+"/.default" {
			f.t.Error("resource scope escaped environment")
		}
		if r.Form.Get("client_id") != appID {
			f.t.Error("client identity changed")
		}
		if f.method == Certificate {
			parts := strings.Split(r.Form.Get("client_assertion"), ".")
			if len(parts) != 3 {
				f.t.Fatal("missing certificate assertion")
			}
			claims, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				f.t.Fatal(err)
			}
			var assertion struct {
				Audience string `json:"aud"`
			}
			if json.Unmarshal(claims, &assertion) != nil || assertion.Audience != endpoint {
				f.t.Error("certificate assertion uses wrong authority")
			}
		}
		if f.method == WorkloadIdentity && r.Form.Get("client_assertion") != "fixture-assertion" {
			f.t.Error("workload assertion missing")
		}
		info, _ := json.Marshal(map[string]string{"uid": "fixture-user", "utid": appTenant})
		claims, _ := json.Marshal(map[string]any{"sub": "fixture-user", "oid": "fixture-user", "tid": appTenant, "iss": issuer, "aud": appID, "preferred_username": "fixture@example.test", "exp": time.Now().Add(time.Hour).Unix()})
		id := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + base64.RawURLEncoding.EncodeToString(claims) + ".fixture"
		payload = map[string]any{"access_token": "fixture-access", "expires_in": 3600, "token_type": "Bearer", "scope": r.Form.Get("scope"), "refresh_token": "fixture-refresh", "client_info": base64.RawURLEncoding.EncodeToString(info), "id_token": id}
	default:
		f.t.Errorf("unapproved identity route: %s", r.URL.Path)
		return nil, errors.New("unapproved identity route")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
}

func TestUnitEnvironmentRejectsForeignIdentityMetadataBeforeCredentialPOST(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"issuer", "token_endpoint", "authorization_endpoint"} {
		for _, method := range []Method{ClientSecret, Password, Browser} {
			t.Run(field+"/"+string(method), func(t *testing.T) {
				if field == "authorization_endpoint" && method != Browser {
					return // Only browser flows consume an authorization endpoint.
				}
				cfg := delegatedConfig(method)
				cfg.Cloud = "gcch"
				fixture := &environmentIdentity{t: t, authority: "https://login.microsoftonline.us", foreignMetadata: field}
				var provider Provider
				if method == ClientSecret {
					cfg.ClientSecret = "fixture-secret"
					p, err := NewApplicationProvider(cfg, &http.Client{Transport: fixture})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { p.Close(); p.Wait() })
					provider = p
				} else {
					p, err := NewDelegatedProvider(cfg, &http.Client{Transport: fixture})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { p.Close(); p.Wait() })
					provider = p
				}
				_, err := provider.AcquireToken(t.Context(), cfg.TokenBinding().Request("https://graph.microsoft.us"))
				if err == nil || fixture.posts != 0 {
					t.Fatalf("foreign metadata allowed credential exchange: %v", err)
				}
			})
		}
	}
}
