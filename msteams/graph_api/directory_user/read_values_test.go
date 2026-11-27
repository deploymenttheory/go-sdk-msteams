package directoryuser_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	teamuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

func TestUnitDirectoryReadReplaysOwnAadUserStringProperties(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/implementation/core/evidence/directory-user-read-source.json")
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

func sourceReadCase(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "directory_user_resolution"
	c.Scope = "service"
	c.Source = "msteams/graph_api/directory_user/read_values_test.go"
	result, err := contracttest.Run(t.Context(), c)
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
			steps := []contracttest.Exchange{wire(auth.GraphAudience, path, 200, item.body)}
			if item.mail {
				steps = []contracttest.Exchange{wire(auth.GraphAudience, path, 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath(selector), 200, item.body)}
			}
			sourceReadCase(t, contracttest.Case{Name: item.name, Mode: mode, Steps: steps, Want: contracttest.Expected{Status: 200, Body: []byte(item.body), JSON: json.RawMessage(fmt.Sprintf("%q", item.id))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.GraphAPI.UserResolver.Resolve(ctx, selector, directoryuser.ResolveOptions{AllowCurrent: item.current})
				if err == nil && (value.User == nil || value.User.ID != item.id || value.User.RawFields["id"] == nil) {
					t.Fatal("resolved snapshot missing")
				}
				return contracttest.Observation{Value: value.ID, Response: response, Err: err}
			}})
		})
	}
}

func TestUnitDirectoryUnusableSnapshotsNeverResolveOrFallback(t *testing.T) {
	for _, body := range []string{`{}`, `{"id":null}`, `{"id":""}`} {
		sourceReadCase(t, contracttest.Case{Name: "unusable-id", Steps: []contracttest.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 200, body)}, Want: contracttest.Expected{Status: 200, Body: []byte(body), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
			if v.ID != "" || v.User == nil || v.Source != directoryuser.Direct {
				t.Fatal("failed resolution lost snapshot or invented usable ID")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitDirectoryOpaqueResolutionRetainsReceiptWhenReferenceGuardRejects(t *testing.T) {
	body := `{"id":"opaque-user"}`
	members := `{"value":[]}`
	sourceReadCase(t, contracttest.Case{Name: "opaque-reference-guard", Steps: []contracttest.Exchange{
		wire(auth.GraphAudience, directPath("fixture@example.invalid"), 200, body),
		wire(auth.GraphAudience, "/v1.0/groups/"+userID+"/members?$select=id", 200, members),
	}, Want: contracttest.Expected{Status: 200, Body: []byte(members), Error: client.ErrRequest}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.TeamUserManagement.Add(ctx, userID, "fixture@example.invalid", teamuser.ChangeOptions{})
		if string(v.UserID) != "opaque-user" || len(v.Stages) != 3 || v.Stages[0].Response == nil || string(v.Stages[0].Response.Body()) != body || v.Stages[2].Response != nil {
			t.Fatal("resolution or local rejection receipts changed")
		}
		return contracttest.Observation{Response: v.Response, Err: e}
	}})
}

func TestUnitDirectoryPrincipalNameMatchingRejectsUnrequestedOpaqueID(t *testing.T) {
	body := `{"value":[{"id":"` + userID + `","userPrincipalName":37},{"id":"opaque-user"}]}`
	sourceReadCase(t, contracttest.Case{Name: "opaque-principal-mismatch", Steps: []contracttest.Exchange{wire(auth.GraphAudience, namePath([]string{userID}), 200, body)}, Want: contracttest.Expected{Status: 200, Body: []byte(body), Error: client.ErrDecode, JSON: json.RawMessage(`{}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, []string{userID}, directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}})
}

func TestUnitDirectoryOpaqueMailIDsCannotEstablishFalseUniqueness(t *testing.T) {
	body := `{"value":[{"id":"opaque-A"},{"id":"opaque-a"}]}`
	sourceReadCase(t, contracttest.Case{Name: "opaque-mail-ambiguity", Steps: []contracttest.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath("fixture@example.invalid"), 200, body)}, Want: contracttest.Expected{Status: 200, Body: []byte(body), Error: directoryuser.ErrAmbiguous, JSON: json.RawMessage(`""`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
		return contracttest.Observation{Value: v.ID, Response: r, Err: e}
	}})
}

func TestUnitDirectoryPrincipalNameProjectionUsesOwnStringReader(t *testing.T) {
	body := `{"value":[{"id":"` + userID + `","userPrincipalName":9007199254740993}]}`
	sourceReadCase(t, contracttest.Case{Name: "scalar-principal-name", Steps: []contracttest.Exchange{wire(auth.GraphAudience, namePath([]string{userID}), 200, body)}, Want: contracttest.Expected{Status: 200, Body: []byte(body), JSON: json.RawMessage(`{"` + userID + `":"9007199254740993"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, []string{userID}, directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}})
}

func TestUnitDirectoryGUIDMailSpellingsRetainCaseInsensitiveDeduplication(t *testing.T) {
	const id = "abcdefab-0123-4567-89ab-0123456789ab"
	body := `{"value":[{"id":"` + id + `"},{"id":"ABCDEFAB-0123-4567-89AB-0123456789AB"}]}`
	sourceReadCase(t, contracttest.Case{Name: "guid-mail-same-identity", Steps: []contracttest.Exchange{wire(auth.GraphAudience, directPath("fixture@example.invalid"), 404, `{"error":{"code":"NotFound"}}`), wire(auth.GraphAudience, mailPath("fixture@example.invalid"), 200, body)}, Want: contracttest.Expected{Status: 200, Body: []byte(body), JSON: json.RawMessage(`"` + id + `"`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.Resolve(ctx, "fixture@example.invalid", directoryuser.ResolveOptions{})
		return contracttest.Observation{Value: v.ID, Response: r, Err: e}
	}})
}
