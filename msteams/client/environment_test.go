package client_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	"github.com/go-resty/resty/v2"
)

func TestUnitEnvironmentRoutesAndContinuationsStayBound(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany", "custom"} {
		t.Run(cloud, func(t *testing.T) {
			provider := &tokenRecorder{}
			cfg := &auth.Config{Method: auth.ExternalProvider, TenantID: "fixture", Cloud: cloud, IdentityID: "fixture", TokenMode: auth.Application, Provider: provider}
			if cloud == "custom" {
				cfg.Endpoints = &auth.Endpoints{Authority: "https://authority.example.test", Graph: "https://graph.example.test", TeamsAdmin: "https://teams.example.test", TeamsAdminResource: "api://teams.example.test"}
			}
			environment, err := auth.ResolveEnvironment(cloud, cfg.Endpoints)
			if err != nil {
				t.Fatal(err)
			}
			var destinations []string
			httpClient := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				destinations = append(destinations, r.URL.String())
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{}`)), Request: r}, nil
			})}
			session, err := client.NewSession(cfg, httpClient, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(session.Close)
			// Caller mutation cannot redirect an existing session or change its cache identity.
			cfg.Cloud = "other"
			if cfg.Endpoints != nil {
				cfg.Endpoints.Graph = "https://foreign.example.test"
			}
			copied, err := session.Environment()
			if err != nil || copied != environment {
				t.Fatal("environment was not copied")
			}
			copied.Endpoints.Graph = "https://foreign.example.test"
			for _, backend := range []client.Backend{client.Graph, client.TeamsAdmin} {
				scope, err := client.NewScope(backend, session)
				if err != nil {
					t.Fatal(err)
				}
				origin, audience, path := environment.Endpoints.Graph, environment.Endpoints.Graph, []string{"v1.0", "teams"}
				if backend == client.TeamsAdmin {
					origin, audience, path = environment.Endpoints.TeamsAdmin, environment.Endpoints.TeamsAdminResource, []string{"fixture", "policy"}
				}
				if _, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, Path: path}, nil); err != nil {
					t.Fatal(err)
				}
				last := provider.requests[len(provider.requests)-1]
				if last.Cloud != environment.BindingCloud || last.Audience != audience || destinations[len(destinations)-1] != origin+"/"+strings.Join(path, "/") {
					t.Fatal("route or binding escaped selected environment")
				}
				before := len(provider.requests)
				for _, target := range []string{"https://foreign.example.test/v1.0/teams", origin + ".foreign.test/v1.0/teams"} {
					if _, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, URL: target}, nil); !errors.Is(err, client.ErrDestination) {
						t.Fatal("foreign continuation accepted")
					}
				}
				if backend == client.Graph {
					if _, err := scope.Do(t.Context(), client.Request{Method: http.MethodGet, URL: origin + "/unversioned"}, nil); !errors.Is(err, client.ErrDestination) {
						t.Fatal("custom Graph route lost version guard")
					}
					accepted := &resty.Response{RawResponse: &http.Response{StatusCode: http.StatusAccepted, Header: http.Header{"Location": {"/teams('fixture')/operations('fixture')"}}}}
					operation, err := graphapi.NewTeamOperation(scope, accepted)
					if err != nil || operation == nil {
						t.Fatalf("relative operation location failed in selected profile: %v", err)
					}
				}
				if len(provider.requests) != before {
					t.Fatal("rejected destination or operation construction acquired tokens")
				}
			}
			session.Close()
			if _, err := session.Environment(); !errors.Is(err, client.ErrClosed) {
				t.Fatal("closed session retained accessible environment")
			}
		})
	}
}
