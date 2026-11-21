package meetingbranding_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
		set := "Identity"
		switch f.Operation {
		case "user":
			set = "GrantToUser"
		case "tenant":
			set = "GrantToTenant"
		case "group":
			set = "GrantToGroup"
		}
		var input json.RawMessage
		if f.Request.Body != "" {
			input = json.RawMessage(f.Request.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: f.Request.CmdletName, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(f.Request.Method, strings.TrimPrefix(f.Request.Url, origin), f.Request.CmdletName, input, f.Status, f.Response)}}
			if f.Request.FailureType != nil {
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				} else {
					c.Want.Error = client.ErrDecode
				}
			} else if f.Response != "" && f.Operation != "delete" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			if len(c.Steps) != 0 {
				c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
				switch f.Operation {
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "create":
					v, r, e := s.Create(ctx, resource.CreateRequest{Identity: scoped, Settings: settings(t, f.Nested)})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "patch":
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings(t, f.Nested), BodyIdentity: policy.Set("read-model-identity")})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Response: r, Err: e}
				case "user":
					v, r, e := s.AssignUser(ctx, "user/fixture", nil)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					empty := ""
					v, r, e := s.AssignTenantDefault(ctx, &empty, false)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "group":
					v, r, e := s.AssignGroup(ctx, "group/fixture", nil, -1)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				default:
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
			}
			run(t, c)
		}
	}
}
