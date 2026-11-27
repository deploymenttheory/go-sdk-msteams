package meetingbranding_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func settings(t *testing.T, nested bool) resource.Settings {
	t.Helper()
	fields := resource.Fields{DefaultTheme: policy.Set(""), EnableMeetingOptionsThemeOverride: policy.Set(false), EnableMeetingOptionsBrandedReactionsOverride: policy.Set(false), EnableNdiAssuranceSlate: policy.Set(false), EnableMeetingBackgroundImages: policy.Set(false), RequireBackgroundEffect: policy.Set(false)}
	images, e := resource.Assets[resource.BackgroundImage]()
	if e != nil {
		t.Fatal(e)
	}
	themes, e := resource.Assets[resource.BrandingTheme]()
	if e != nil {
		t.Fatal(e)
	}
	slates, e := resource.Assets[resource.AssuranceSlate]()
	if e != nil {
		t.Fatal(e)
	}
	if nested {
		no := false
		image, e := resource.NewBackgroundImage("fixture-image", 0, "Fixture", "https://example.invalid/background.png", resource.BackgroundImageOptions{IsRequired: &no})
		if e != nil {
			t.Fatal(e)
		}
		theme, e := resource.NewBrandingTheme("fixture-theme", resource.ThemeOptions{ThemeFields: resource.ThemeFields{Enabled: &no}})
		if e != nil {
			t.Fatal(e)
		}
		slate, e := resource.NewAssuranceSlate("fixture-slate", "Fixture", "https://example.invalid/slate.png", resource.AssuranceSlateOptions{IsDefault: &no})
		if e != nil {
			t.Fatal(e)
		}
		images, e = resource.Assets(image)
		if e != nil {
			t.Fatal(e)
		}
		themes, e = resource.Assets(theme)
		if e != nil {
			t.Fatal(e)
		}
		slates, e = resource.Assets(slate)
		if e != nil {
			t.Fatal(e)
		}
	}
	return resource.Settings{Fields: fields, MeetingBackgroundImages: images, MeetingBrandingThemes: themes, NDIAssuranceSlateImages: slates}
}

func TestUnitBrandingTypedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {

		switch f.Operation {
		case "user":

		case "tenant":

		case "group":

		}
		var input json.RawMessage
		if f.Request.Body != "" {
			input = json.RawMessage(f.Request.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(f.Request.Method, strings.TrimPrefix(f.Request.Url, origin), f.Request.CmdletName, input, f.Status, f.Response)}}
			// These retained vectors prove rejected request bytes, not decoding.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
				switch f.Operation {
				case "list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "create":
					v, r, e := s.Create(ctx, resource.CreateRequest{Identity: scoped, Settings: settings(t, f.Nested)})
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "patch":
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings(t, f.Nested), BodyIdentity: policy.Set("read-model-identity")})
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "delete":
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return testutil.Observation{Response: r, Err: e}
				case "user":
					v, r, e := s.AssignUser(ctx, "user/fixture", nil)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					empty := ""
					v, r, e := s.AssignTenantDefault(ctx, &empty, false)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "group":
					v, r, e := s.AssignGroup(ctx, "group/fixture", nil, -1)
					return testutil.Observation{Value: v, Response: r, Err: e}
				default:
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
			}
			run(t, c)
		}
	}
}
