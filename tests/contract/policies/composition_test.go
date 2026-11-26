package policies_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	branding "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func policyCompositionFixture(t *testing.T, family, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../../msteams/teams_admin_api/" + family + "/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func policyCompositionWire(method, family, id, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	path := "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/" + family
	if id != "" {
		path += "/configuration/" + id
	}
	headers := http.Header{"X-MS-CmdletName": {command}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}
	if body != nil {
		headers.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: reply}}
}

// The read/acceptance bytes are retained primitive fixtures, composed offline;
// this sequence was not captured from a tenant. Only the blank 403 is authored.
// Equal caller names deliberately select different resource namespaces and
// different read identities (fixture-policy versus Tag:fixture-policy).
func TestUnitPolicyCompositionPreservesReceiptsAndIndependentRecovery(t *testing.T) {
	t.Parallel()
	external := policyCompositionFixture(t, "external_access_policy", "create_readback.json")
	before := policyCompositionFixture(t, "meeting_branding", "create_readback.json")
	created := policyCompositionFixture(t, "meeting_branding", "create_201.body")
	deleted := policyCompositionFixture(t, "meeting_branding", "delete_204.body")
	missing := policyCompositionFixture(t, "meeting_branding", "get_deleted_404.json")
	var externalRaw, brandingRaw map[string]json.RawMessage
	if json.Unmarshal(external, &externalRaw) != nil || json.Unmarshal(before, &brandingRaw) != nil {
		t.Fatal("invalid retained response")
	}
	steps := []contracttest.Exchange{
		policyCompositionWire("GET", "ExternalAccessPolicy", "fixture-policy", "Get-CsExternalAccessPolicy", nil, 200, external),
		policyCompositionWire("POST", "TeamsMeetingBrandingPolicy", "", "New-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"fixture-policy","EnableMeetingOptionsThemeOverride":false,"RequireBackgroundEffect":false}`), 201, created),
		policyCompositionWire("GET", "TeamsMeetingBrandingPolicy", "fixture-policy", "Get-CsTeamsMeetingBrandingPolicy", nil, 200, before),
		policyCompositionWire("GET", "TeamsMeetingBrandingPolicy", "Tag%3Afixture-policy", "Set-CsTeamsMeetingBrandingPolicy", nil, 200, before),
		policyCompositionWire("PATCH", "TeamsMeetingBrandingPolicy", "Tag%3Afixture-policy", "Set-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","EnableMeetingOptionsThemeOverride":true}`), 403, nil),
		policyCompositionWire("DELETE", "TeamsMeetingBrandingPolicy", "Tag%3Afixture-policy", "Remove-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy"}`), 204, deleted),
		policyCompositionWire("GET", "TeamsMeetingBrandingPolicy", "Tag%3Afixture-policy", "Get-CsTeamsMeetingBrandingPolicy", nil, 404, missing),
		policyCompositionWire("GET", "ExternalAccessPolicy", "fixture-policy", "Get-CsExternalAccessPolicy", nil, 200, external),
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "policy-composition-recovery", Family: "meeting_branding", Scope: "service", Source: "msteams/teams_admin_api/meeting_branding/testdata/create_readback.json", Steps: steps,
		Want: contracttest.Expected{Status: 200, JSON: external, Body: external},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			first, response, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || first == nil || first.Identity == nil || *first.Identity != "fixture-policy" || first.EnableFederationAccess == nil || *first.EnableFederationAccess || !reflect.DeepEqual(first.RawFields, externalRaw) {
				t.Fatal("first family lost its typed false or raw snapshot")
			}
			s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
			value, response, err := s.Create(ctx, branding.CreateRequest{Identity: "fixture-policy", Settings: branding.Settings{Fields: branding.Fields{EnableMeetingOptionsThemeOverride: policy.Set(false), RequireBackgroundEffect: policy.Set(false)}}})
			if err != nil || response == nil || response.StatusCode() != 201 || value != nil {
				t.Fatal("empty create acceptance was mistaken for a readback")
			}
			read, _, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || read == nil || read.Identity == nil || *read.Identity != "Tag:fixture-policy" || read.DefaultTheme != nil || string(read.RawFields["DefaultTheme"]) != "null" || !reflect.DeepEqual(read.RawFields, brandingRaw) {
				t.Fatal("second family lost its tagged identity or explicit null")
			}
			failed, response, denied := s.UpdateByIdentity(ctx, *read.Identity, branding.UpdateRequest{Fields: branding.Fields{EnableMeetingOptionsThemeOverride: policy.Set(true)}})
			if !errors.Is(denied, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 || failed == nil || !failed.PatchInvoked || failed.HTTPAccepted || failed.After != nil || failed.Before == nil || failed.Before.EnableMeetingOptionsThemeOverride == nil || *failed.Before.EnableMeetingOptionsThemeOverride || !reflect.DeepEqual(failed.Before.RawFields, brandingRaw) || !reflect.DeepEqual(first.RawFields, externalRaw) {
				t.Fatal("later denial lost preparation, prior evidence, or acceptance distinction")
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			if response, err := s.DeleteByIdentity(canceled, *read.Identity, policy.Controls{}); response != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("canceled cleanup unexpectedly made an attempt")
			}
			// Recovery is an explicit caller decision with an independent bound;
			// neither denial nor cancellation automatically rolls back resources.
			recovery, stop := context.WithTimeout(context.WithoutCancel(canceled), time.Second)
			defer stop()
			response, err = s.DeleteByIdentity(recovery, *read.Identity, policy.Controls{})
			if err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("independent cleanup was blocked by the canceled caller")
			}
			absent, response, err := s.GetByIdentity(recovery, *read.Identity)
			var api *client.APIError
			if absent != nil || response == nil || !errors.As(err, &api) || api.StatusCode != 404 {
				t.Fatal("delete acceptance was not followed by confirmed absence")
			}
			last, response, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(recovery, *first.Identity)
			if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, externalRaw)) {
				t.Fatal("recovering one namespace changed the other")
			}
			return contracttest.Observation{Value: last, Response: response, Err: err}
		}})
	if err != nil || !result.Passed() {
		t.Fatalf("composed policy protocol differs: %v %+v", err, result.Differences())
	}
}
