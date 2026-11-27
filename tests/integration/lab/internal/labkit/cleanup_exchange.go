package labkit

import (
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func CleanupExchange(method, path string, status int, body string) testutil.Exchange {
	return testutil.Exchange{
		Request: testutil.WireRequest{
			Method:   method,
			URL:      "https://api.interfaces.records.teams.microsoft.com" + path,
			Audience: auth.TeamsAdminAudience,
		},
		Reply: testutil.Reply{
			Status:  status,
			Headers: http.Header{"Content-Type": {"application/json"}},
			Body:    []byte(body),
		},
	}
}
