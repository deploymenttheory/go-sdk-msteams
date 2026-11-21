package meetingbranding_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "msteams/teams_admin_api/meeting_branding/testdata/get_global.json"
const assetSource = "msteams/teams_admin_api/meeting_branding/testdata/get_global.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"
const collection = "/Skype.Policy/configurations/TeamsMeetingBrandingPolicy"
const item = collection + "/configuration/Global"

type inputVectors struct {
	Name, Operation, Response string
	Status                    int
	Nested, Force             bool
	Request                   struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	value := struct {
		Cases []inputVectors `json:"cases"`
	}{
		Cases: []inputVectors{
			{
				Name:      "read",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "create",
				Operation: "create",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"NdiAssuranceSlateImages\": [],\n  \"MeetingBackgroundImages\": [],\n  \"MeetingBrandingThemes\": [],\n  \"DefaultTheme\": \"\",\n  \"EnableMeetingOptionsThemeOverride\": false,\n  \"EnableMeetingOptionsBrandedReactionsOverride\": false,\n  \"EnableNdiAssuranceSlate\": false,\n  \"EnableMeetingBackgroundImages\": false,\n  \"RequireBackgroundEffect\": false\n}",

					CmdletName: "New-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "patch",
				Operation: "patch",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"NdiAssuranceSlateImages\": [],\n  \"MeetingBackgroundImages\": [],\n  \"MeetingBrandingThemes\": [],\n  \"DefaultTheme\": \"\",\n  \"EnableMeetingOptionsThemeOverride\": false,\n  \"EnableMeetingOptionsBrandedReactionsOverride\": false,\n  \"EnableNdiAssuranceSlate\": false,\n  \"EnableMeetingBackgroundImages\": false,\n  \"RequireBackgroundEffect\": false\n}",

					CmdletName: "Set-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "user",
				Operation: "user",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/user%2Ffixture/policies/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"additionalParameters\": null\n}",

					CmdletName: "Grant-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "tenant",
				Operation: "tenant",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"PolicyName\": \"\",\n  \"additionalParameters\": null,\n  \"forceSwitchPresent\": false\n}",

					CmdletName: "Grant-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "group",
				Operation: "group",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "create-nested",
				Operation: "create",
				Status:    400,
				Nested:    true,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"NdiAssuranceSlateImages\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-slate\",\n      \"Name\": \"Fixture\",\n      \"NdiImageUri\": \"https://example.invalid/slate.png\",\n      \"IsDefault\": false\n    }\n  ],\n  \"MeetingBackgroundImages\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-image\",\n      \"Order\": 0,\n      \"Name\": \"Fixture\",\n      \"IsRequired\": false,\n      \"IsHidden\": null,\n      \"ImageUri\": \"https://example.invalid/background.png\"\n    }\n  ],\n  \"MeetingBrandingThemes\": [\n    {\n      \"MeetingReactions\": [],\n      \"DisplayName\": null,\n      \"LogoImageLightUri\": null,\n      \"LogoImageDarkUri\": null,\n      \"BackgroundImageLightUri\": null,\n      \"BackgroundImageDarkUri\": null,\n      \"EnableChatBackground\": null,\n      \"ChatBackgroundImageLightUri\": null,\n      \"ChatBackgroundImageDarkUri\": null,\n      \"ChatBackgroundImageAlignment\": null,\n      \"LogoImageLightPreAuthUri\": null,\n      \"LogoImageDarkPreAuthUri\": null,\n      \"MeetingInviteLogoImageLightPreAuthUri\": null,\n      \"MeetingInviteLogoImageDarkPreAuthUri\": null,\n      \"BackgroundImageLightPreAuthUri\": null,\n      \"BackgroundImageDarkPreAuthUri\": null,\n      \"BrandAccentColor\": null,\n      \"DisableOriginalReactions\": null,\n      \"Enabled\": false,\n      \"Identity\": \"fixture-theme\"\n    }\n  ],\n  \"DefaultTheme\": \"\",\n  \"EnableMeetingOptionsThemeOverride\": false,\n  \"EnableMeetingOptionsBrandedReactionsOverride\": false,\n  \"EnableNdiAssuranceSlate\": false,\n  \"EnableMeetingBackgroundImages\": false,\n  \"RequireBackgroundEffect\": false\n}",

					CmdletName: "New-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "patch-nested",
				Operation: "patch",
				Status:    400,
				Nested:    true,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"NdiAssuranceSlateImages\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-slate\",\n      \"Name\": \"Fixture\",\n      \"NdiImageUri\": \"https://example.invalid/slate.png\",\n      \"IsDefault\": false\n    }\n  ],\n  \"MeetingBackgroundImages\": [\n    {\n      \"Identity\": null,\n      \"Id\": \"fixture-image\",\n      \"Order\": 0,\n      \"Name\": \"Fixture\",\n      \"IsRequired\": false,\n      \"IsHidden\": null,\n      \"ImageUri\": \"https://example.invalid/background.png\"\n    }\n  ],\n  \"MeetingBrandingThemes\": [\n    {\n      \"MeetingReactions\": [],\n      \"DisplayName\": null,\n      \"LogoImageLightUri\": null,\n      \"LogoImageDarkUri\": null,\n      \"BackgroundImageLightUri\": null,\n      \"BackgroundImageDarkUri\": null,\n      \"EnableChatBackground\": null,\n      \"ChatBackgroundImageLightUri\": null,\n      \"ChatBackgroundImageDarkUri\": null,\n      \"ChatBackgroundImageAlignment\": null,\n      \"LogoImageLightPreAuthUri\": null,\n      \"LogoImageDarkPreAuthUri\": null,\n      \"MeetingInviteLogoImageLightPreAuthUri\": null,\n      \"MeetingInviteLogoImageDarkPreAuthUri\": null,\n      \"BackgroundImageLightPreAuthUri\": null,\n      \"BackgroundImageDarkPreAuthUri\": null,\n      \"BrandAccentColor\": null,\n      \"DisableOriginalReactions\": null,\n      \"Enabled\": false,\n      \"Identity\": \"fixture-theme\"\n    }\n  ],\n  \"DefaultTheme\": \"\",\n  \"EnableMeetingOptionsThemeOverride\": false,\n  \"EnableMeetingOptionsBrandedReactionsOverride\": false,\n  \"EnableNdiAssuranceSlate\": false,\n  \"EnableMeetingBackgroundImages\": false,\n  \"RequireBackgroundEffect\": false\n}",

					CmdletName: "Set-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "delete-force-false",
				Operation: "delete",
				Status:    400,
				Nested:    false,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-null-arrays",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-invalid-asset",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-null-asset-id",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-null-asset-order",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-lowercase-invalid-uri",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-lowercase-null-theme",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "delete-denied",
				Operation: "delete",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsMeetingBrandingPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsMeetingBrandingPolicy",
				},
			},
			{
				Name:      "group-opaque",
				Operation: "group",
				Status:    400,
				Nested:    false,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBrandingPolicy",
					Body:   "{\n  \"policyName\": null,\n  \"priority\": -1\n}",

					CmdletName: "Grant-CsTeamsMeetingBrandingPolicy",
				},
			},
		},
	}

	return value.Cases
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "meeting_branding"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("branding contract differs: %v %+v", e, r.Differences())
	}
	return r
}
