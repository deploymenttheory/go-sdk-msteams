package teammanagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

type creationTransport func(*http.Request) (*http.Response, error)

func (f creationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type creationProvider struct{}

func (creationProvider) AcquireToken(ctx context.Context, request auth.TokenRequest) (auth.Token, error) {
	if err := ctx.Err(); err != nil {
		return auth.Token{}, err
	}
	if request.Audience != auth.GraphAudience {
		return auth.Token{}, client.ErrDestination
	}
	return auth.Token{AccessToken: "synthetic-creation-token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

// This transport has no sockets or fallback; random aliases are asserted against
// the exact query actually observed, then matched to the subsequent POST body.
func creationClient(t *testing.T, transport creationTransport) *msteams.Client {
	t.Helper()
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ExternalProvider, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, Provider: creationProvider{}}, msteams.WithHTTPClient(&http.Client{Transport: creationTransport(func(request *http.Request) (*http.Response, error) {
		if request.Body != nil {
			defer func() { _ = request.Body.Close() }()
		}
		if request.URL.Scheme+"://"+request.URL.Host != auth.GraphAudience || request.Header.Get("Authorization") != "Bearer synthetic-creation-token" {
			t.Error("creation destination or audience differs")
			return nil, client.ErrRequest
		}
		return transport(request)
	})}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error(err)
		}
	})
	return sdk
}

func creationReply(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func TestGeneratedAliasRequiresCompleteSuccessfulAvailabilityRead(t *testing.T) {
	t.Parallel()
	pattern := regexp.MustCompile(`^groupTypes/any\(c:c eq 'Unified'\) and mailNickname eq '(msteams_[0-9a-f]{6})'$`)
	for _, name := range []string{"available", "collision", "later-page-collision", "limit", "denied", "foreign-continuation"} {
		t.Run(name, func(t *testing.T) {
			checks, calls, posts := 0, 0, 0
			selected := ""
			later := false
			sdk := creationClient(t, func(request *http.Request) (*http.Response, error) {
				calls++
				path := request.URL.Path
				if request.Method == "GET" && path == "/v1.0/groups" {
					query := request.URL.Query()
					if query.Get("$skiptoken") == "alias-page" {
						if !later {
							t.Error("unexpected alias continuation")
						}
						later = false
						return creationReply(request, 200, `{"value":[{"id":"`+groupID+`"}]}`), nil
					}
					parts := pattern.FindStringSubmatch(query.Get("$filter"))
					if len(parts) != 2 || len(query) != 2 || query.Get("$select") != "id" {
						t.Error("generated alias query differs")
						return nil, client.ErrRequest
					}
					checks++
					selected = parts[1]
					switch {
					case name == "denied":
						return creationReply(request, 403, `{"error":{"code":"Forbidden"}}`), nil
					case name == "foreign-continuation":
						return creationReply(request, 200, `{"value":[],"@odata.nextLink":"https://foreign.invalid/private"}`), nil
					case name == "later-page-collision" && checks == 1:
						later = true
						return creationReply(request, 200, `{"value":[],"@odata.nextLink":"`+auth.GraphAudience+`/v1.0/groups?$skiptoken=alias-page"}`), nil
					case name == "limit" || name == "collision" && checks == 1:
						return creationReply(request, 200, `{"value":[{"id":"`+groupID+`"}]}`), nil
					default:
						return creationReply(request, 200, `{"value":[]}`), nil
					}
				}
				if request.Method == "POST" && path == "/v1.0/groups" {
					posts++
					var body struct {
						Alias string `json:"mailNickname"`
					}
					if request.Body == nil || json.NewDecoder(request.Body).Decode(&body) != nil || body.Alias != selected {
						t.Error("generated alias changed between complete lookup and create")
					}
					return creationReply(request, 201, `{"id":"`+groupID+`"}`), nil
				}
				if request.Method == "GET" && path == "/v1.0/groups/"+groupID {
					return creationReply(request, 200, `{"id":"`+groupID+`","displayName":"Fixture"}`), nil
				}
				if request.Method == "PUT" && path == "/v1.0/groups/"+groupID+"/team" {
					return creationReply(request, 201, `{"id":"`+groupID+`"}`), nil
				}
				if request.Method == "GET" && path == "/v1.0/teams/"+groupID {
					return creationReply(request, 200, `{"id":"`+groupID+`"}`), nil
				}
				t.Error("unexpected generated-alias HTTP operation")
				return nil, client.ErrRequest
			})
			value := creationRequest()
			value.MailNickname = nil
			result, err := sdk.GraphAPI.TeamManagement.Create(t.Context(), value)
			if !result.GeneratedMailNickname || result.MailNickname != selected {
				t.Fatal("selected generated alias was lost")
			}
			switch name {
			case "limit":
				if !errors.Is(err, management.ErrMailNicknameLimit) || checks != 10 || posts != 0 || calls != 10 {
					t.Fatal("alias limit produced a write or was unbounded")
				}
			case "denied":
				if !errors.Is(err, auth.ErrForbidden) || posts != 0 || calls != 1 {
					t.Fatal("denied availability triggered a write")
				}
			case "foreign-continuation":
				if !errors.Is(err, client.ErrDestination) || posts != 0 || calls != 1 {
					t.Fatal("incomplete availability triggered a write")
				}
			default:
				if err != nil || posts != 1 || result.Value == nil {
					t.Fatal("available generated alias did not produce one complete creation")
				}
			}
		})
	}
}

func TestLostCreationResponsesNeverReplayOrBlindlyCompensate(t *testing.T) {
	t.Parallel()
	for _, phase := range []string{"group", "team"} {
		t.Run(phase, func(t *testing.T) {
			lost := errors.New("synthetic response loss")
			posts, puts, reconciles := 0, 0, 0
			sdk := creationClient(t, func(request *http.Request) (*http.Response, error) {
				path := request.URL.Path
				if request.Method == "POST" && path == "/v1.0/groups" {
					posts++
					if phase == "group" {
						return nil, lost
					}
					return creationReply(request, 201, `{"id":"`+groupID+`"}`), nil
				}
				if request.Method == "GET" && path == "/v1.0/groups/"+groupID {
					return creationReply(request, 200, `{"id":"`+groupID+`"}`), nil
				}
				if request.Method == "PUT" && path == "/v1.0/groups/"+groupID+"/team" {
					puts++
					return nil, lost
				}
				if request.Method == "GET" && path == "/v1.0/teams/"+groupID {
					reconciles++
					return creationReply(request, 404, `{"error":{"code":"Missing"}}`), nil
				}
				t.Error("ambiguous creation triggered an unexpected operation/delete")
				return nil, client.ErrRequest
			})
			result, err := sdk.GraphAPI.TeamManagement.Create(t.Context(), creationRequest())
			if !errors.Is(err, client.ErrTransport) || !errors.Is(result.PrimaryError, client.ErrTransport) || !errors.Is(err, management.ErrCreationReconciliation) || result.Cleanup.Disposition != management.CleanupUnresolved || posts != 1 {
				t.Fatal("ambiguous result lost primary error or unresolved cleanup state")
			}
			if phase == "group" {
				if result.ID != "" || result.CreatedGroupOwned || puts != 0 || reconciles != 0 {
					t.Fatal("lost group response fabricated identity or ownership")
				}
			} else {
				if result.ID != groupID || !result.CreatedGroupOwned || puts != 1 || reconciles != 1 {
					t.Fatal("lost team response changed known identity or blindly replayed")
				}
			}
		})
	}
}
