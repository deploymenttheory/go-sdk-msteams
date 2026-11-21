package client_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestHierarchyProfilesMatchPinnedOriginAudienceAndCloudMaps(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../contracts/endpoints/8.1.0/special-service-profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Profiles []struct {
			ID          string `json:"id"`
			Audience    string `json:"audience"`
			CloudRoutes map[string]struct {
				Origin string `json:"origin"`
			} `json:"cloud_routes"`
		} `json:"profiles"`
	}
	if json.Unmarshal(raw, &source) != nil {
		t.Fatal("invalid pinned profile contract")
	}
	for _, tc := range []struct {
		id      string
		backend client.Backend
		path    []string
	}{
		{"targeting-v1", client.Targeting, []string{"api", "rest", "v1", "nodes", "upload", "status"}},
		{"hierarchy-v2", client.Hierarchy, []string{"api", "v1", "hierarchies", "fixture-tenant", "operations", "latest"}},
	} {
		for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
			t.Run(tc.id+"-"+cloud, func(t *testing.T) {
				origin, audience := "", ""
				for _, profile := range source.Profiles {
					if profile.ID == tc.id {
						origin, audience = profile.CloudRoutes[cloud].Origin, profile.Audience
					}
				}
				calls := 0
				httpClient := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if r.URL.String() != origin+"/"+strings.Join(tc.path, "/") || r.Header.Get("Authorization") != "Bearer fixture-token-"+audience {
						t.Error("profile substituted origin, namespace or audience")
					}
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{}`)), Request: r}, nil
				})}
				scope, _, provider := transportScope(t, tc.backend, httpClient, cloud)
				if calls != 0 || len(provider.requests) != 0 {
					t.Fatal("profile construction acquired credentials")
				}
				_, err := scope.Do(t.Context(), client.Request{Method: "GET", Path: tc.path}, nil)
				if origin == "" {
					if !errors.Is(err, client.ErrDestination) || calls != 0 || len(provider.requests) != 0 {
						t.Fatal("unqualified cloud fell back or acquired credentials")
					}
					return
				}
				if err != nil || calls != 1 || len(provider.requests) != 1 || provider.requests[0].Audience != audience {
					t.Fatal("pinned profile route failed")
				}
				for _, destination := range []string{origin + "/outside", origin + "/api/v1/hierarchies-foreign/fixture", origin + ".foreign.invalid/" + strings.Join(tc.path, "/"), "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/fixture"} {
					if _, err := scope.Do(t.Context(), client.Request{Method: "GET", URL: destination}, nil); !errors.Is(err, client.ErrDestination) {
						t.Fatal("foreign namespace or origin accepted")
					}
				}
				if calls != 1 || len(provider.requests) != 1 {
					t.Fatal("rejected route acquired another token or reached HTTP")
				}
			})
		}
	}
}

func TestHierarchyProfileErrorsKeepTypedBackendAndRedactDiagnostics(t *testing.T) {
	t.Parallel()
	for _, backend := range []client.Backend{client.Targeting, client.Hierarchy} {
		value := client.NewAPIError(backend, 403, http.Header{"request-id": {"private-canary"}}, []byte(`{"error":{"code":"PrivateCode","message":"private-canary"}}`))
		if value.Backend != backend || len(value.Codes) != 1 || value.Codes[0] != "PrivateCode" {
			t.Fatal("native backend/error details were not retained")
		}
		for _, format := range []string{"%v", "%+v", "%#v"} {
			text := fmt.Sprintf(format, value)
			if strings.Contains(text, "private-canary") || strings.Contains(text, "PrivateCode") || !strings.Contains(text, string(backend)) {
				t.Fatal("backend error formatting leaked diagnostics or lost its family")
			}
		}
	}
}
