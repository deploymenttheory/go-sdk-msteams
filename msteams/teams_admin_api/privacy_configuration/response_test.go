package privacyconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func readResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directory.Close(); err != nil {
			t.Error("testdata directory close failed")
		}
	})
	raw, err := directory.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatal("invalid permanent response fixture")
	}
	return raw
}

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []testutil.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []testutil.Exchange
	return append(steps, testutil.Exchange{
		Request: testutil.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-CsPrivacyConfiguration"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitGetGlobalResponseFixtureError(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global_403.json")
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	var codes []string
	if envelope.Code != "" {
		codes = []string{envelope.Code}
	}
	result, err := testutil.Run(t.Context(), testutil.Case{
		Name:  "get-global-403",
		Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/PrivacyConfiguration/configuration/Global", 403, raw),
		Want:  testutil.Expected{Status: 403, Body: raw, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{envelope.ErrorCode}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
			var apiErr *client.APIError
			if value != nil || !errors.As(err, &apiErr) || apiErr.StatusCode != 403 {
				t.Fatal("actual error response did not remain a typed API failure")
			}
			return testutil.Observation{Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}
