package directoryuser_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	teamuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

func TestUnitDirectoryReadReplaysOwnAadUserStringProperties(t *testing.T) {
	raw, err := os.ReadFile("testdata/directory_user_read_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	// Explicit tags keep the immutable source schema independent of Go field names.
	var source struct {
		Observations []struct {
			Field     string  `json:"field"`
			Input     string  `json:"input_json"`
			Value     *string `json:"value"`
			Exception *string `json:"exception"`
		} `json:"observations"`
	}
	if json.Unmarshal(raw, &source) != nil || len(source.Observations) != 70 {
		t.Fatal("source observations unavailable")
	}
	for index, row := range source.Observations {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			var user directoryuser.User
			err := json.Unmarshal([]byte(row.Input), &user)
			if row.Exception != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source rejection not retained")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got *string
			switch row.Field {
			case "id":
				got = &user.ID
			case "displayName":
				got = user.DisplayName
			case "userPrincipalName":
				got = user.UserPrincipalName
			case "userType":
				got = user.UserType
			case "mail":
				got = user.Mail
			}
			if row.Field == "id" && row.Value == nil {
				if user.ID != "" {
					t.Fatal("nullable ID representation changed")
				}
			} else if (got == nil) != (row.Value == nil) || got != nil && *got != *row.Value {
				t.Fatal("actual AadUser string projection differs")
			}
			var fields map[string]json.RawMessage
			if json.Unmarshal([]byte(row.Input), &fields) != nil || string(user.RawFields[row.Field]) != string(fields[row.Field]) {
				t.Fatal("raw token lost")
			}
		})
	}
}

func sourceReadCase(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("authored source protocol: setup=%v differences=%+v", err, result.Differences())
	}
}

func TestUnitDirectorySourceReadsResolveOpaqueAndScalarIDs(t *testing.T) {
	for _, item := range []struct {
		name, body, id string
		current, mail  bool
	}{
		{"direct-opaque", `{"id":"opaque-user","future":{"x":1}}`, "opaque-user", false, false},
		{"direct-scalar", `{"id":9007199254740993}`, "9007199254740993", false, false},
		{"current-scalar", `{"id":false}`, "false", true, false},
		{"mail-opaque", `{"value":[{"id":"opaque-user"}]}`, "opaque-user", false, true},
	} {
		t.Run(item.name, func(t *testing.T) {
			path := directPath("fixture@example.invalid")
			mode := auth.Application
			selector := "fixture@example.invalid"
			if item.current {
				path = "/v1.0/me?$select=id"
				mode = auth.Delegated
				selector = ""
			}
			steps := []testutil.Exchange{wire(auth.GraphAudience, path, 200, item.body)}
			if item.mail {
				steps = []testutil.Exchange{wire(auth.GraphAudience, path, 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath(selector), 200, item.body)}
			}
			sourceReadCase(t, testutil.Case{Name: item.name, Mode: mode, Steps: steps, Want: testutil.Expected{Status: 200, Body: []byte(item.body), JSON: json.RawMessage(fmt.Sprintf("%q", item.id))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := sdk.GraphAPI.UserResolver.Resolve(ctx, selector, directoryuser.ResolveOptions{AllowCurrent: item.current})
				if err == nil && (value.User == nil || value.User.ID != item.id || value.User.RawFields["id"] == nil) {
					t.Fatal("resolved snapshot missing")
				}
				return testutil.Observation{Value: value.ID, Response: response, Err: err}
			}})
		})
	}
}

func TestUnitDirectoryUnusableSnapshotsNeverResolveOrFallback(t *testing.T) {
	for _, body := range []string{`{}`, `{"id":null}`, `{"id":""}`} {
		sourceReadCase(t, testutil.Case{Name: "unusable-id", Steps: []testutil.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
			if v.ID != "" || v.User == nil || v.Source != directoryuser.Direct {
				t.Fatal("failed resolution lost snapshot or invented usable ID")
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitDirectoryOpaqueResolutionRetainsReceiptWhenReferenceGuardRejects(t *testing.T) {
	body := `{"id":"opaque-user"}`
	members := `{"value":[]}`
	sourceReadCase(t, testutil.Case{Name: "opaque-reference-guard", Steps: []testutil.Exchange{
		wire(auth.GraphAudience, directPath("fixture@example.invalid"), 200, body),
		wire(auth.GraphAudience, "/v1.0/groups/"+userID+"/members?$select=id", 200, members),
	}, Want: testutil.Expected{Status: 200, Body: []byte(members), Error: client.ErrRequest}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, e := sdk.GraphAPI.TeamUserManagement.Add(ctx, userID, "fixture@example.invalid", teamuser.ChangeOptions{})
		if string(v.UserID) != "opaque-user" || len(v.Stages) != 3 || v.Stages[0].Response == nil || string(v.Stages[0].Response.Body()) != body || v.Stages[2].Response != nil {
			t.Fatal("resolution or local rejection receipts changed")
		}
		return testutil.Observation{Response: v.Response, Err: e}
	}})
}

func TestUnitDirectoryPrincipalNameMatchingRejectsUnrequestedOpaqueID(t *testing.T) {
	body := `{"value":[{"id":"` + userID + `","userPrincipalName":37},{"id":"opaque-user"}]}`
	sourceReadCase(t, testutil.Case{Name: "opaque-principal-mismatch", Steps: []testutil.Exchange{wire(auth.GraphAudience, namePath([]string{userID}), 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body), Error: client.ErrDecode, JSON: json.RawMessage(`{}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, []string{userID}, directoryuser.PrincipalNamesOptions{})
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}

func TestUnitDirectoryOpaqueMailIDsCannotEstablishFalseUniqueness(t *testing.T) {
	body := `{"value":[{"id":"opaque-A"},{"id":"opaque-a"}]}`
	sourceReadCase(t, testutil.Case{Name: "opaque-mail-ambiguity", Steps: []testutil.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath("fixture@example.invalid"), 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body), Error: directoryuser.ErrAmbiguous, JSON: json.RawMessage(`""`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
		return testutil.Observation{Value: v.ID, Response: r, Err: e}
	}})
}

func TestUnitDirectoryPrincipalNameProjectionUsesOwnStringReader(t *testing.T) {
	body := `{"value":[{"id":"` + userID + `","userPrincipalName":9007199254740993}]}`
	sourceReadCase(t, testutil.Case{Name: "scalar-principal-name", Steps: []testutil.Exchange{wire(auth.GraphAudience, namePath([]string{userID}), 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body), JSON: json.RawMessage(`{"` + userID + `":"9007199254740993"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, []string{userID}, directoryuser.PrincipalNamesOptions{})
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}

func TestUnitDirectoryGUIDMailSpellingsRetainCaseInsensitiveDeduplication(t *testing.T) {
	const id = "abcdefab-0123-4567-89ab-0123456789ab"
	body := `{"value":[{"id":"` + id + `"},{"id":"ABCDEFAB-0123-4567-89AB-0123456789AB"}]}`
	sourceReadCase(t, testutil.Case{Name: "guid-mail-same-identity", Steps: []testutil.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath("fixture@example.invalid"), 200, body)}, Want: testutil.Expected{Status: 200, Body: []byte(body), JSON: json.RawMessage(`"` + id + `"`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
		return testutil.Observation{Value: v.ID, Response: r, Err: e}
	}})
}

func TestUnitDirectoryReadPreservesSourcePropertyMatchingAndAssignmentOrder(t *testing.T) {
	raw, err := os.ReadFile("testdata/directory_user_read_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Controls []struct {
			Input      string `json:"input_json"`
			Projection struct {
				ID            *string `json:"id"`
				DisplayName   *string `json:"displayName"`
				PrincipalName *string `json:"userPrincipalName"`
				UserType      *string `json:"userType"`
				Mail          *string `json:"mail"`
			} `json:"projection"`
			Exception *string `json:"exception"`
		} `json:"property_controls"`
	}
	if json.Unmarshal(raw, &source) != nil || len(source.Controls) != 11 {
		t.Fatal("source property controls unavailable")
	}
	for index, row := range source.Controls {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			var user directoryuser.User
			err := json.Unmarshal([]byte(row.Input), &user)
			if row.Exception != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("earlier rejected assignment was hidden")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantID := ""
			if row.Projection.ID != nil {
				wantID = *row.Projection.ID
			}
			if user.ID != wantID {
				t.Fatal("source ID alias order/null reset changed")
			}
			for _, pair := range [][2]*string{{user.DisplayName, row.Projection.DisplayName}, {user.UserPrincipalName, row.Projection.PrincipalName}, {user.UserType, row.Projection.UserType}, {user.Mail, row.Projection.Mail}} {
				if (pair[0] == nil) != (pair[1] == nil) || pair[0] != nil && *pair[0] != *pair[1] {
					t.Fatal("source property match/assignment changed")
				}
			}
			var fields map[string]json.RawMessage
			if json.Unmarshal([]byte(row.Input), &fields) != nil || len(fields) != len(user.RawFields) {
				t.Fatal("raw unknown/alias properties lost")
			}
			for name, value := range fields {
				if string(user.RawFields[name]) != string(value) {
					t.Fatal("raw token changed")
				}
			}
		})
	}
}
