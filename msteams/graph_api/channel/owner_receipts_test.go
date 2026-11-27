package channel_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
)

// These authored outcomes exercise the composition directly. They do not pass
// through requestBoundary's intentional replacement of unverified successes.
func TestUnitOwnerResolutionReceiptsSurviveLaterCreationFailures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, ownerID, body string
		status              int
		want                error
	}{
		{"local-owner-guard", "opaque-owner", "", 0, client.ErrRequest},
		{"unusable-owner-snapshot", "", "", 0, client.ErrDecode},
		{"denied-post", userID, `{"error":{"code":"Denied"}}`, 403, auth.ErrForbidden},
		{"accepted-malformed", userID, `{`, 201, client.ErrDecode},
		{"created", userID, `{"id":"created"}`, 201, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ownerBody := `{"id":"` + tc.ownerID + `","future":{"kept":true}}`
			steps := []contracttest.Exchange{wire("GET", "https://graph.microsoft.com/v1.0/users/owner@fixture.invalid?$select=id", "", 200, ownerBody)}
			if tc.status != 0 {
				payload := `{"displayName":"Fixture","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('` + userID + `')","roles":["owner"]}]}`
				steps = append(steps, wire("POST", base+"/channels", payload, tc.status, tc.body))
			}
			status, body := tc.status, tc.body
			if status == 0 {
				status, body = 200, ownerBody
			}
			result, err := contracttest.Run(t.Context(), contracttest.Case{
				Name: tc.name, Family: "channels", Scope: "service", Source: "msteams/graph_api/channel/owner_receipts_test.go",
				Steps: steps, Want: contracttest.Expected{Status: status, Body: []byte(body), Error: tc.want},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					value, err := sdk.GraphAPI.Channels.CreateWithOwner(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Private}, "owner@fixture.invalid", directoryuser.ResolveOptions{})
					if value.OwnerResponse == nil || value.OwnerResponse.StatusCode() != 200 || string(value.OwnerResponse.Body()) != ownerBody {
						t.Fatal("owner resolution receipt lost")
					}
					if value.OwnerResolution == nil || value.OwnerResolution.ID != tc.ownerID || value.OwnerResolution.Source != directoryuser.Direct || value.OwnerResolution.User == nil || string(value.OwnerResolution.User.RawFields["future"]) != `{"kept":true}` {
						t.Fatal("typed/raw resolution result lost")
					}
					if tc.status == 0 && value.Response != value.OwnerResponse || tc.status != 0 && value.Response == value.OwnerResponse {
						t.Fatal("creation and preparation responses conflated")
					}
					if tc.want != nil && (value.Channel != nil || value.Operation != nil) {
						t.Fatal("failed creation returned a created resource")
					}
					if tc.want == nil && (value.Channel == nil || value.Channel.ID != "created") {
						t.Fatal("created channel lost")
					}
					return contracttest.Observation{Response: value.Response, Err: err}
				},
			})
			if err != nil || !result.Passed() {
				t.Fatalf("owner receipt contract differs: setup=%v differences=%+v", err, result.Differences())
			}
		})
	}
}

func TestUnitPrimitiveCreateHasNoOwnerResolutionReceipt(t *testing.T) {
	t.Parallel()
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "primitive-create", Family: "channels", Scope: "service", Source: "msteams/graph_api/channel/owner_receipts_test.go",
		Steps: []contracttest.Exchange{wire("POST", base+"/channels", `{"displayName":"Fixture","membershipType":"standard"}`, 201, `{"id":"created"}`)},
		Want:  contracttest.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, err := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture"})
			if value.OwnerResolution != nil || value.OwnerResponse != nil {
				t.Fatal("primitive creation invented resolution metadata")
			}
			return contracttest.Observation{Value: value.Channel, Response: value.Response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("primitive channel contract differs: setup=%v differences=%+v", err, result.Differences())
	}
}

func TestUnitIdentifierOwnerHasLocalResolutionAndCreationReceipt(t *testing.T) {
	t.Parallel()
	payload := `{"displayName":"Fixture","membershipType":"private","members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","user@odata.bind":"https://graph.microsoft.com/beta/users('` + userID + `')","roles":["owner"]}]}`
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "identifier-owner", Family: "channels", Scope: "service", Source: "msteams/graph_api/channel/owner_receipts_test.go",
		Steps: []contracttest.Exchange{wire("POST", base+"/channels", payload, 201, `{"id":"created"}`)},
		Want:  contracttest.Expected{Status: 201, JSON: json.RawMessage(`{"id":"created"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, err := sdk.GraphAPI.Channels.CreateWithOwner(ctx, teamID, channel.CreateRequest{DisplayName: "Fixture", MembershipType: channel.Private}, userID, directoryuser.ResolveOptions{})
			if value.OwnerResponse != nil || value.OwnerResolution == nil || value.OwnerResolution.ID != userID || value.OwnerResolution.Source != directoryuser.Identifier || value.OwnerResolution.User != nil {
				t.Fatal("local identifier resolution invented a directory lookup")
			}
			return contracttest.Observation{Value: value.Channel, Response: value.Response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("identifier-owner contract differs: setup=%v differences=%+v", err, result.Differences())
	}
}
