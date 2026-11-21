package targetinghierarchy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
)

func TestUnitHierarchyDeletionFailuresAreNotReplayed(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		for _, status := range []int{403, 429, 503} {
			path, backend := "/api/rest/v1/nodes/hierarchy", client.Targeting
			if version == hierarchy.V2 {
				path, backend = hierarchyPath(), client.Hierarchy
			}
			results = append(results, run(t, contracttest.Case{Name: "delete-v" + string(version) + "-" + strconv.Itoa(status), Mappings: mapping("Remove-TeamTargetingHierarchy"), Steps: []contracttest.Exchange{wire(version, "DELETE", path, status, `{"error":{"code":"Denied","message":"private"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, err := sdk.TeamsAdminAPI.TargetingHierarchy.Delete(ctx, version)
				return contracttest.Observation{Response: r, Err: err}
			}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: backend, StatusCode: status, Codes: []string{"Denied"}}}}))
		}
	}
	evidence(t, results)
}

func TestUnitHierarchyOptionsProfilesAndSourceValidationAreLocal(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, cloud string
		version     hierarchy.Version
		err         error
	}{
		{"invalid-version", "commercial", "3", hierarchy.ErrVersion},
		{"unqualified-v1", "gcch", hierarchy.V1, client.ErrDestination},
		{"unqualified-v2", "gcc", hierarchy.V2, client.ErrDestination},
	} {
		run(t, contracttest.Case{Name: tc.name, Scope: "local", Cloud: tc.cloud, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s, input := source()
			v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, tc.version)
			if s.closes != 1 {
				t.Error("rejected profile/version lost source ownership")
			}
			return contracttest.Observation{Value: v.ImportAttempted, Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Error: tc.err, JSON: json.RawMessage(`false`)}})
	}
	for _, name := range []string{"empty-filename", "dot-filename", "zero-limit", "negative-limit", "nil-source", "typed-nil-source"} {
		run(t, contracttest.Case{Name: name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s, input := source()
			hasBody := true
			switch name {
			case "empty-filename":
				input.Filename = ""
			case "dot-filename":
				input.Filename = ".."
			case "zero-limit":
				input.MaxBytes = 0
			case "negative-limit":
				input.MaxBytes = -1
			case "nil-source":
				input.Body = nil
				hasBody = false
			case "typed-nil-source":
				var missing *csvSource
				input.Body = missing
				hasBody = false
			}
			v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V2)
			if hasBody && s.closes != 1 || !hasBody && s.closes != 0 {
				t.Error("invalid CSV ownership differs")
			}
			return contracttest.Observation{Value: v.CreateAttempted, Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Error: client.ErrRequest, JSON: json.RawMessage(`false`)}})
	}
}

func TestUnitHierarchyRequiresBoundCanonicalTenantBeforeV2Calls(t *testing.T) {
	t.Parallel()
	// Preserve the runner's default non-GUID tenant. A caller must configure a
	// canonical tenant in the session; this API accepts no substitutable tenant ID.
	c := contracttest.Case{Name: "bound-tenant-required", Family: "targeting_hierarchy", Scope: "local", Source: "contracts/endpoints/8.1.0/special-service-profiles.json", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s, input := source()
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V2)
		if s.closes != 1 || v.CreateAttempted || v.ImportAttempted {
			t.Error("invalid bound tenant reached preparation or lost stream")
		}
		return contracttest.Observation{Response: v.Response, Err: err}
	}, Want: contracttest.Expected{Error: hierarchy.ErrTenantIdentity}}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatal("tenant binding contract differs")
	}
}

func TestUnitHierarchyCancellationAndCloseFailuresRetainClassification(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "cancelled-close-failure", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		s, input := source()
		s.closeErr = errors.New("private-close-canary")
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V2)
		if !errors.Is(err, client.ErrStreamClose) || s.closes != 1 || strings.Contains(err.Error(), "private-close-canary") {
			t.Error("close classification/ownership/redaction differs")
		}
		return contracttest.Observation{Response: v.Response, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
	for _, name := range []string{"closed", "wrong-backend", "nil-native", "nil-composition"} {
		want := client.ErrClosed
		if name == "wrong-backend" {
			want = client.ErrDestination
		}
		run(t, contracttest.Case{Name: name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2()
			switch name {
			case "closed":
				if err := sdk.Close(); err != nil {
					return contracttest.Observation{Err: err}
				}
			case "wrong-backend":
				s = hierarchy.NewV2Service(sdk.GraphAPI.Scope())
			case "nil-native":
				s = nil
			case "nil-composition":
				var empty *hierarchy.Service
				r, err := empty.Delete(ctx, hierarchy.V1)
				return contracttest.Observation{Response: r, Err: err}
			}
			r, err := s.Delete(ctx)
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestUnitMalformedNativeHierarchyReadsAndImportHandles(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`null`, `[]`, `{}`, `{"id":"operation","status":false}`} {
		run(t, contracttest.Case{Name: "malformed-status", Steps: []contracttest.Exchange{wire(hierarchy.V2, "GET", hierarchyPath()+"/operations/latest", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().GetOperation(ctx, "")
			return contracttest.Observation{Value: v, Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
	for _, body := range []string{`null`, `{}`, `{"id":null}`, `{"id":""}`, `{"id":7}`} {
		run(t, contracttest.Case{Name: "invalid-import-handle", Steps: []contracttest.Exchange{uploadWire(hierarchy.V2, 202, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s, input := source()
			_, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().ImportCSV(ctx, input)
			if s.closes != 1 {
				t.Error("failed decode lost stream cleanup")
			}
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: 202, Error: client.ErrDecode}})
	}
}

func TestUnitAcceptedImportCloseFailureKeepsHTTPAcknowledgement(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "accepted-close-failure", Steps: []contracttest.Exchange{uploadWire(hierarchy.V1, 202, `{"requestId":"operation"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s, input := source()
		s.closeErr = errors.New("private-close-canary")
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V1)
		if s.closes != 1 || !v.ImportAttempted || !v.ImportHTTPAccepted || strings.Contains(err.Error(), "private-close-canary") {
			t.Error("accepted import lost acknowledgement or leaked close error")
		}
		return contracttest.Observation{Response: v.Response, Err: err}
	}, Want: contracttest.Expected{Status: 202, Error: client.ErrStreamClose}})
}

func TestUnitNilHierarchyContextIsLocal(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "nil-context", Scope: "local", Invoke: func(_ context.Context, sdk *msteams.Client) contracttest.Observation {
		var missing context.Context
		//nolint:contextcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		_, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(missing, "", hierarchy.V1)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Error: client.ErrNilContext}})
}
