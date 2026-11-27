package conversationrole_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
)

type comparisonProvider func(context.Context, auth.TokenRequest) (auth.Token, error)

func (p comparisonProvider) AcquireToken(ctx context.Context, r auth.TokenRequest) (auth.Token, error) {
	return p(ctx, r)
}

type comparisonTransport func(*http.Request) (*http.Response, error)

func (f comparisonTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type comparisonBody struct {
	io.Reader
	cancel context.CancelFunc
}

func (b comparisonBody) Close() error { b.cancel(); return nil }

func TestUnitConversationCancellationAfterComparisonPreventsCreate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	httpClient := &http.Client{Transport: comparisonTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.String() != route+"/configuration/Global" || r.Header.Get("X-MS-CmdletName") != "New-CsConversationRole" {
			t.Error("comparison sent unexpected method, target or cmdlet metadata")
			return nil, client.ErrRequest
		}
		// Close occurs after the complete comparison body has been read. This models
		// caller cancellation at the read/action boundary without timing sleeps.
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Request: r, Body: comparisonBody{Reader: strings.NewReader(strings.ReplaceAll(reply, identity, "Global")), cancel: cancel}}, nil
	})}
	provider := comparisonProvider(func(ctx context.Context, r auth.TokenRequest) (auth.Token, error) {
		if e := ctx.Err(); e != nil {
			return auth.Token{}, e
		}
		if r.Audience != auth.TeamsAdminAudience {
			t.Error("comparison requested another audience")
			return auth.Token{}, auth.ErrTokenMissing
		}
		return auth.Token{AccessToken: "offline-comparison-token", ExpiresAt: time.Now().Add(time.Hour)}, nil
	})
	session, e := client.NewSession(&auth.Config{Method: auth.ExternalProvider, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, Provider: provider}, httpClient, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(session.Close)
	scope, e := client.NewScope(client.TeamsAdmin, session)
	if e != nil {
		t.Fatal(e)
	}
	service := roles.NewService(scope)
	request, e := roles.ProjectCreateByIdentity(identity, parameters(t))
	if e != nil {
		t.Fatal(e)
	}
	result, response, e := service.CreateWithGlobalComparison(ctx, *request)
	if !errors.Is(e, context.Canceled) || result == nil || !result.ComparisonInvoked || result.CreateInvoked || result.CreateResponse != nil || result.HTTPAccepted || calls != 1 || response == nil || response.StatusCode() != 200 || result.ComparisonResponse != response {
		t.Fatal("cancellation lost comparison receipt or allowed creation")
	}
}
