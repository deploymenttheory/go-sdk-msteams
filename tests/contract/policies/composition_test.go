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
	recording "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
	banner "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func policyCompositionFixture(t *testing.T, family, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("../../../msteams/teams_admin_api")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := directory.Close(); err != nil {
			t.Error(err)
		}
	}()
	raw, err := directory.ReadFile(family + "/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// Primary-authored offline values from the pinned recording/banner DTOs and
// local builder/setter contracts, not captured nonempty tenant responses.
// Core.dll 68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49:
// source types/builders/relative-ID binding are traced in the published POL04
// primary audit at dc2bb9a. Unknown RawFields below test SDK compatibility only.
func TestUnitPolicyCompositionKeepsRecordingAndBannerIdentitySeparate(t *testing.T) {
	t.Parallel()
	const id = "00000000-0000-4000-8000-000000000031"
	const identity = "Global/" + id
	const message = `{"Identity":"` + identity + `","Id":"` + id + `","Description":null,"RecordingAndTranscriptionLocalizationCustomMessage":[{"Language":"en-US","ParticipantImplicit":"","InitiatorExplicit":null,"Future":false}],"Future":false}`
	const text = `{"Identity":"` + identity + `","Id":"` + id + `","Text":"","Description":null,"Future":false}`
	steps := []contracttest.Exchange{
		policyCompositionWire("POST", "TeamsRecordingAndTranscriptionCustomMessage", "", "New-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{"Identity":"`+identity+`","Id":null,"Description":null,"RecordingAndTranscriptionLocalizationCustomMessage":[{"Identity":null,"Language":"en-US","InitiatorImplicit":null,"ParticipantImplicit":"","InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}]}`), 200, []byte(message)),
		policyCompositionWire("GET", "CustomBannerText", identity, "Get-CsTeamsCustomBannerText", nil, 200, []byte(text)),
		policyCompositionWire("GET", "TeamsRecordingAndTranscriptionCustomMessage", identity, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, []byte(message)),
		policyCompositionWire("PATCH", "TeamsRecordingAndTranscriptionCustomMessage", identity, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{"Identity":"`+identity+`","Id":"`+id+`","RecordingAndTranscriptionLocalizationCustomMessage":[]}`), 204, nil),
		policyCompositionWire("PATCH", "CustomBannerText", identity, "Set-CsTeamsCustomBannerText", json.RawMessage(`{"Id":"`+id+`","Text":"changed"}`), 403, nil),
		policyCompositionWire("DELETE", "TeamsRecordingAndTranscriptionCustomMessage", identity, "Remove-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{"Id":"`+id+`"}`), 204, nil),
		policyCompositionWire("GET", "TeamsRecordingAndTranscriptionCustomMessage", identity, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 404, nil),
		policyCompositionWire("GET", "CustomBannerText", identity, "Get-CsTeamsCustomBannerText", nil, 200, []byte(text)),
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "recording-banner-composition", Family: "recording_messages", Scope: "service", Source: "tests/contract/policies/composition_test.go", Mappings: []contracttest.Mapping{
		{Command: "New-CsTeamsRecordingAndTranscriptionLocalizationCustomMessage", ParameterSet: "Identity"},
		{Command: "New-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "ParentAndRelativeKey"},
		{Command: "Set-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "ParentAndRelativeKey"},
		{Command: "Remove-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "ParentAndRelativeKey"},
		{Command: "Get-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "Identity"},
		{Command: "Get-CsTeamsCustomBannerText", ParameterSet: "ParentAndRelativeKey"},
		{Command: "Get-CsTeamsCustomBannerText", ParameterSet: "Identity"},
		{Command: "Set-CsTeamsCustomBannerText", ParameterSet: "ParentAndRelativeKey"},
	}, Steps: steps, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(text), Body: []byte(text)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		localization, err := recording.NewLocalization(recording.LocalizationInput{Language: "en-US", ParticipantImplicit: new("")})
		if err != nil {
			t.Fatal(err)
		}
		s := sdk.TeamsAdminAPI.RecordingMessages
		created, response, err := s.CreateByID(ctx, id, recording.CreateRequest{Description: policy.Null[string](), Localizations: recording.ReplaceLocalizations(localization)})
		if err != nil || response == nil || response.StatusCode() != 200 || created == nil || created.ID == nil || *created.ID != id || created.Identity == nil || *created.Identity != identity || created.Description != nil || string(created.RawFields["Description"]) != "null" || len(created.Localizations) != 1 || created.Localizations[0] == nil || created.Localizations[0].ParticipantImplicit == nil || *created.Localizations[0].ParticipantImplicit != "" || created.Localizations[0].InitiatorExplicit != nil || string(created.Localizations[0].RawFields["Future"]) != "false" {
			t.Fatal("created localization lost typed identity, empty/null or raw extension")
		}
		other, _, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.GetByID(ctx, id)
		if err != nil || other == nil || other.ID == nil || *other.ID != id || other.Text == nil || *other.Text != "" || other.Description != nil || string(other.RawFields["Description"]) != "null" {
			t.Fatal("equal-ID banner lost its distinct model or explicit empty text")
		}
		changed, response, err := s.UpdateByID(ctx, id, recording.UpdateRequest{Localizations: recording.ClearLocalizations()})
		if err != nil || response == nil || response.StatusCode() != 204 || changed == nil || !changed.HTTPAccepted || !changed.PatchInvoked || changed.After != nil || changed.Before == nil || len(changed.Before.Localizations) != 1 || len(created.Localizations) != 1 {
			t.Fatal("successful clear mutated preparation or fabricated a readback")
		}
		failed, response, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.PatchByID(ctx, id, banner.PatchRequest{Settings: banner.Settings{Text: policy.Set("changed")}, BodyID: policy.Set(*other.ID)})
		if failed != nil || response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) || len(changed.Before.Localizations) != 1 || *other.Text != "" {
			t.Fatal("later banner denial erased successful recording preparation or mutated prior read")
		}
		deleted, response, err := s.DeleteByID(ctx, id, policy.Controls{})
		if err != nil || response == nil || response.StatusCode() != 204 || deleted != nil {
			t.Fatal("explicit recording cleanup lost empty acceptance")
		}
		absent, response, err := s.GetByIdentity(ctx, identity)
		var api *client.APIError
		if absent != nil || response == nil || !errors.As(err, &api) || api.StatusCode != 404 {
			t.Fatal("recording cleanup was not independently confirmed")
		}
		last, response, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.GetByIdentity(ctx, identity)
		if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, other.RawFields)) {
			t.Fatal("same literal identity caused cross-family cleanup or mutation")
		}
		return contracttest.Observation{Value: last, Response: response, Err: err}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("recording/banner composition differs: %v %+v", err, result.Differences())
	}
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
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "policy-composition-recovery", Family: "meeting_branding", Scope: "service", Source: "msteams/teams_admin_api/meeting_branding/testdata/create_readback.json", Mappings: []contracttest.Mapping{
		{Command: "Get-CsExternalAccessPolicy", ParameterSet: "Identity"},
		{Command: "New-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"},
		{Command: "Get-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"},
		{Command: "Set-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"},
		{Command: "Remove-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"},
	}, Steps: steps,
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
