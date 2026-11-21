package channelusermanagement_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestUnitCallerDeadlineBoundsWholeCompositionBeforeMutation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "synthetic-offline-token", ExpiresAt: time.Now().Add(time.Hour)}}},
			msteams.WithHTTPTimeout(5*time.Minute), msteams.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				calls++
				if request.Method != "GET" || calls > 2 {
					t.Error("expired composition attempted an extra request or write")
					return nil, context.Canceled
				}
				select {
				case <-request.Context().Done():
					return nil, request.Context().Err()
				case <-time.After(30 * time.Second):
				}
				body := `{"value":[{"id":"` + channelID + `","displayName":"fixture","membershipType":"private"}]}`
				if calls == 2 {
					body = `{"id":"` + userID + `"}`
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			})}))
		if err != nil {
			t.Fatal("offline client construction failed")
		}
		defer func() {
			if err := sdk.Close(); err != nil {
				t.Error("offline client close failed")
			}
		}()
		started := time.Now()
		ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
		defer cancel()
		result, err := sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", "member@fixture.invalid", management.ChangeOptions{})
		if !errors.Is(err, context.DeadlineExceeded) || calls != 2 || time.Since(started) != 45*time.Second || len(result.Stages) != 2 || result.Stages[1].Action != management.ResolveUser || result.Stages[1].State != management.Unconfirmed {
			t.Fatal("whole-operation deadline or retained failure stages differed")
		}
	})
}
