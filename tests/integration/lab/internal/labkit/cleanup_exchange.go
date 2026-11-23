package labkit

import (
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func CleanupExchange(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{
		Request: contracttest.WireRequest{
			Method:   method,
			URL:      "https://api.interfaces.records.teams.microsoft.com" + path,
			Audience: auth.TeamsAdminAudience,
		},
		Reply: contracttest.Reply{
			Status:  status,
			Headers: http.Header{"Content-Type": {"application/json"}},
			Body:    []byte(body),
		},
	}
}
