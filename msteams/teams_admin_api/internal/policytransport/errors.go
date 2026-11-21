package policytransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// PolicyRP identifies flight-disabled types with both a numeric error code and
// a specific message. Generic bad requests, forbidden responses and arbitrary
// mentions of unavailable features do not establish this condition.
func classifyError(err error, status int, body []byte) error {
	if err == nil || status != http.StatusBadRequest {
		return err
	}
	var api *client.APIError
	if !errors.As(err, &api) || api == nil || api.Backend != client.TeamsAdmin || api.StatusCode != status {
		return err
	}
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"message"`
	}
	if json.Unmarshal(body, &envelope) != nil || envelope.Code != "ClientError" || envelope.ErrorCode != 40003 ||
		!strings.HasPrefix(envelope.Message, "Type ") || !strings.HasSuffix(envelope.Message, " is not currently enabled in flighting") {
		return err
	}
	return errors.Join(err, auth.ErrTenantCapabilityUnavailable)
}
