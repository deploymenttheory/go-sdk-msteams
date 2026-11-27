package targetinghierarchy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	hierarchy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_hierarchy"
)

func TestUnitHierarchyDeletionFailuresAreNotReplayed(t *testing.T) {
	t.Parallel()

	for _, version := range []hierarchy.Version{hierarchy.V1, hierarchy.V2} {
		for _, status := range []int{403, 429, 503} {
			path, backend := "/api/rest/v1/nodes/hierarchy", client.Targeting
			if version == hierarchy.V2 {
				path, backend = hierarchyPath(), client.Hierarchy
			}
			run(t, testutil.Case{Name: "delete-v" + string(version) + "-" + strconv.Itoa(status), Steps: []testutil.Exchange{wire(version, "DELETE", path, status, `{"error":{"code":"Denied","message":"private"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, err := sdk.TeamsAdminAPI.TargetingHierarchy.Delete(ctx, version)
				return testutil.Observation{Response: r, Err: err}
			}, Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: backend, StatusCode: status, Codes: []string{"Denied"}}}})
		}
	}
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
		run(t, testutil.Case{Name: tc.name, Cloud: tc.cloud, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s, input := source()
			v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, tc.version)
			if s.closes != 1 {
				t.Error("rejected profile/version lost source ownership")
			}
			return testutil.Observation{Value: v.ImportAttempted, Response: v.Response, Err: err}
		}, Want: testutil.Expected{Error: tc.err, JSON: json.RawMessage(`false`)}})
	}
	for _, name := range []string{"empty-filename", "dot-filename", "zero-limit", "negative-limit", "nil-source", "typed-nil-source"} {
		run(t, testutil.Case{Name: name, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
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
			return testutil.Observation{Value: v.CreateAttempted, Response: v.Response, Err: err}
		}, Want: testutil.Expected{Error: client.ErrRequest, JSON: json.RawMessage(`false`)}})
	}
}

func TestUnitHierarchyRequiresBoundCanonicalTenantBeforeV2Calls(t *testing.T) {
	t.Parallel()
	// Preserve the runner's default non-GUID tenant. A caller must configure a
	// canonical tenant in the session; this API accepts no substitutable tenant ID.
	c := testutil.Case{Name: "bound-tenant-required", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s, input := source()
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V2)
		if s.closes != 1 || v.CreateAttempted || v.ImportAttempted {
			t.Error("invalid bound tenant reached preparation or lost stream")
		}
		return testutil.Observation{Response: v.Response, Err: err}
	}, Want: testutil.Expected{Error: hierarchy.ErrTenantIdentity}}
	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatal("tenant binding contract differs")
	}
}

func TestUnitHierarchyCancellationAndCloseFailuresRetainClassification(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "cancelled-close-failure", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		s, input := source()
		s.closeErr = errors.New("private-close-canary")
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V2)
		if !errors.Is(err, client.ErrStreamClose) || s.closes != 1 || strings.Contains(err.Error(), "private-close-canary") {
			t.Error("close classification/ownership/redaction differs")
		}
		return testutil.Observation{Response: v.Response, Err: err}
	}, Want: testutil.Expected{Error: context.Canceled}})
	for _, name := range []string{"closed", "wrong-backend", "nil-native", "nil-composition"} {
		want := client.ErrClosed
		if name == "wrong-backend" {
			want = client.ErrDestination
		}
		run(t, testutil.Case{Name: name, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2()
			switch name {
			case "closed":
				if err := sdk.Close(); err != nil {
					return testutil.Observation{Err: err}
				}
			case "wrong-backend":
				s = hierarchy.NewV2Service(sdk.GraphAPI.Scope())
			case "nil-native":
				s = nil
			case "nil-composition":
				var empty *hierarchy.Service
				r, err := empty.Delete(ctx, hierarchy.V1)
				return testutil.Observation{Response: r, Err: err}
			}
			r, err := s.Delete(ctx)
			return testutil.Observation{Response: r, Err: err}
		}, Want: testutil.Expected{Error: want}})
	}
}

func TestUnitMalformedNativeHierarchyReadsAndImportHandles(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`null`, `[]`, `{}`, `{"id":"operation","status":false}`} {
		run(t, testutil.Case{Name: "malformed-status", Steps: []testutil.Exchange{wire(hierarchy.V2, "GET", hierarchyPath()+"/operations/latest", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().GetOperation(ctx, "")
			return testutil.Observation{Value: v, Response: r, Err: err}
		}, Want: testutil.Expected{Status: 200, Error: client.ErrDecode}})
	}
	for _, body := range []string{`null`, `{}`, `{"id":null}`, `{"id":""}`, `{"id":7}`} {
		run(t, testutil.Case{Name: "invalid-import-handle", Steps: []testutil.Exchange{uploadWire(hierarchy.V2, 202, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s, input := source()
			_, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.NativeV2().ImportCSV(ctx, input)
			if s.closes != 1 {
				t.Error("failed decode lost stream cleanup")
			}
			return testutil.Observation{Response: r, Err: err}
		}, Want: testutil.Expected{Status: 202, Error: client.ErrDecode}})
	}
}

func TestUnitFailedImportClosesSourceWithoutAcknowledgement(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "denied-close-failure", Steps: []testutil.Exchange{uploadWire(hierarchy.V1, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s, input := source()
		s.closeErr = errors.New("private-close-canary")
		v, err := sdk.TeamsAdminAPI.TargetingHierarchy.UploadCSV(ctx, input, hierarchy.V1)
		if s.closes != 1 || !v.ImportAttempted || v.ImportHTTPAccepted || !errors.Is(err, client.ErrStreamClose) || strings.Contains(err.Error(), "private-close-canary") {
			t.Error("failed import lost ownership or invented acknowledgement")
		}
		return testutil.Observation{Response: v.Response, Err: err}
	}, Want: testutil.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestUnitNilHierarchyContextIsLocal(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "nil-context", Invoke: func(_ context.Context, sdk *msteams.Client) testutil.Observation {
		var missing context.Context
		//nolint:contextcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		_, r, err := sdk.TeamsAdminAPI.TargetingHierarchy.GetStatus(missing, "", hierarchy.V1)
		return testutil.Observation{Response: r, Err: err}
	}, Want: testutil.Expected{Error: client.ErrNilContext}})
}
