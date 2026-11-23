package configapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

type doFunc func(context.Context, client.Request, any) (*resty.Response, error)

func (f doFunc) Do(ctx context.Context, request client.Request, result any) (*resty.Response, error) {
	return f(ctx, request, result)
}

func metadata(status int) *resty.Response {
	return &resty.Response{RawResponse: &http.Response{StatusCode: status, Header: make(http.Header)}}
}

func TestUnitDiscoveryNormalizationCachingAndHeaderOwnership(t *testing.T) {
	t.Parallel()
	var discoveries, management int
	base := doFunc(func(_ context.Context, request client.Request, result any) (*resty.Response, error) {
		if request.Path[0] == "Teams.Tenant" {
			discoveries++
			if request.Method != http.MethodGet || request.Replay != client.ReplaySafe || request.Header.Get("X-MS-Target-Uri") != "" {
				t.Error("discovery request contract changed")
			}
			return metadata(200), json.Unmarshal([]byte(`{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid:8443/path?q=ignored#ignored"}}`), result)
		}
		management++
		if request.Header.Get("X-MS-Target-Uri") != "https://admin.example.invalid/" {
			t.Error("target root normalization differs")
		}
		if request.Header.Get("X-Fixture") != "preserved" || request.Replay != client.ReplayNever {
			t.Error("headers/replay changed")
		}
		return metadata(204), nil
	})
	adapter := configapi.New(base)
	if configapi.New(adapter) != adapter {
		t.Fatal("existing discovery adapter was wrapped again")
	}
	headers := http.Header{"X-Fixture": {"preserved"}}
	for range 2 {
		if _, err := adapter.Do(t.Context(), client.Request{Method: http.MethodPut, Path: []string{"Skype.Policy", "fixture"}, Header: headers}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if discoveries != 1 || management != 2 {
		t.Fatal("unexpected discovery/cache count")
	}
	if headers.Get("X-MS-Target-Uri") != "" {
		t.Fatal("caller header map mutated")
	}
}

func TestUnitInvalidDiscoveryNeverFallsBackOrCaches(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, body string }{
		{name: "missing", body: `{}`},
		{name: "empty", body: `{"Endpoints":{"AdminServiceEndpoint":""}}`},
		{name: "absolute instead of bare", body: `{"Endpoints":{"AdminServiceEndpoint":"https://admin.example.invalid"}}`},
		{name: "credentials", body: `{"Endpoints":{"AdminServiceEndpoint":"user:password@admin.example.invalid"}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var count int
			adapter := configapi.New(doFunc(func(_ context.Context, request client.Request, result any) (*resty.Response, error) {
				count++
				if request.Path[0] != "Teams.Tenant" {
					t.Error("management sent after invalid discovery")
				}
				return metadata(200), json.Unmarshal([]byte(test.body), result)
			}))
			for range 2 {
				response, err := adapter.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "fixture"}}, nil)
				if !errors.Is(err, configapi.ErrDiscovery) || response == nil || response.StatusCode() != 200 {
					t.Fatal("invalid discovery or its metadata lost")
				}
			}
			if count != 2 {
				t.Fatal("discovery failure was cached")
			}
		})
	}
}

func TestUnitDiscoveryAuthorizationMetadataPropagates(t *testing.T) {
	t.Parallel()
	expected := metadata(403)
	adapter := configapi.New(doFunc(func(context.Context, client.Request, any) (*resty.Response, error) {
		return expected, &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}
	}))
	actual, err := adapter.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "fixture"}}, nil)
	if actual != expected {
		t.Fatal("discovery response metadata discarded")
	}
	var apiError *client.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != 403 {
		t.Fatal("discovery authorization failure changed")
	}
}

func TestUnitQueuedDiscoveryRespectsCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})
		var discoveries atomic.Int32
		adapter := configapi.New(doFunc(func(ctx context.Context, request client.Request, result any) (*resty.Response, error) {
			if request.Path[0] == "Teams.Tenant" {
				if discoveries.Add(1) == 1 {
					close(started)
				}
				select {
				case <-release:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				return metadata(200), json.Unmarshal([]byte(`{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid"}}`), result)
			}
			return metadata(200), nil
		}))
		first := make(chan error, 1)
		go func() {
			_, err := adapter.Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "fixture"}}, nil)
			first <- err
		}()
		<-started
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		second := make(chan error, 1)
		go func() {
			_, err := adapter.Do(ctx, client.Request{Method: http.MethodGet, Path: []string{"Skype.Policy", "fixture"}}, nil)
			second <- err
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		select {
		case err := <-second:
			if !errors.Is(err, context.Canceled) {
				t.Error("queued cancellation changed")
			}
		default:
			t.Error("queued discovery ignored cancellation")
		}
		close(release)
		if err := <-first; err != nil {
			t.Fatal(err)
		}
		if discoveries.Load() != 1 {
			t.Fatal("canceled caller started another discovery")
		}
	})
}

func TestUnitTargetSegmentExtendsForestAndRejectsPathText(t *testing.T) {
	t.Parallel()
	var targets []string
	base := doFunc(func(_ context.Context, request client.Request, result any) (*resty.Response, error) {
		if request.Path[0] == "Teams.Tenant" {
			return metadata(200), json.Unmarshal([]byte(`{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid"}}`), result)
		}
		targets = append(targets, request.Header.Get("X-MS-Target-Uri"))
		return metadata(200), nil
	})
	adapter := configapi.New(base)
	for _, segment := range []string{"HostedMigration", ""} {
		if _, err := adapter.DoTarget(t.Context(), segment, client.Request{Method: http.MethodPost, Path: []string{"Teams.HostedMigrationService", "ValidateUser"}}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(targets) != 2 || targets[0] != "https://admin.example.invalid/HostedMigration" || targets[1] != "https://admin.example.invalid/" {
		t.Fatalf("target segment differs: %q", targets)
	}
	for _, segment := range []string{"a/b", `a\b`, "a?b", "a#b", "%2f", "a b", "..", ".", "a\x7f"} {
		if _, err := adapter.DoTarget(t.Context(), segment, client.Request{Method: http.MethodPost, Path: []string{"x"}}, nil); !errors.Is(err, client.ErrRequest) {
			t.Fatalf("segment %q was accepted", segment)
		}
	}
	if len(targets) != 2 {
		t.Fatal("rejected segment reached the backend")
	}
}
