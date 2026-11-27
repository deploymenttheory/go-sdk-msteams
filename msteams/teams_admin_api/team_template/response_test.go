package teamtemplate_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

func templateResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	raw, err := root.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func templateResponseWire(method, path, command string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: reply}}
}

func responseDefinition(t *testing.T, body bool, short, description string, channels bool) (template.Definition, json.RawMessage) {
	t.Helper()
	definition, wire := inputDefinition(t, body)
	definition.DisplayName = new("fixture-template")
	definition.ShortDescription = &short
	definition.Description = &description
	var fields map[string]json.RawMessage
	if json.Unmarshal(wire, &fields) != nil {
		t.Fatal("invalid pinned input fixture")
	}
	for key, value := range map[string]string{"displayName": "fixture-template", "shortDescription": short, "description": description} {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		fields[key] = raw
	}
	if channels {
		fields["channels"] = json.RawMessage(`[{"id":"General","displayName":"General","isFavoriteByDefault":true,"tabs":[]}]`)
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return definition, raw
}

func verifyActualTemplate(t *testing.T, value *template.Template, raw []byte, short, description string) {
	t.Helper()
	var envelope struct {
		Value map[string]json.RawMessage `json:"value"`
	}
	if json.Unmarshal(raw, &envelope) != nil || value == nil || !reflect.DeepEqual(value.RawFields, envelope.Value) {
		t.Fatal("complete actual template snapshot changed")
	}
	if value.DisplayName == nil || *value.DisplayName != "fixture-template" || value.ShortDescription == nil || *value.ShortDescription != short || value.Description == nil || *value.Description != description || len(value.Channels) != 1 || value.Channels[0] == nil || value.Channels[0].ID == nil || *value.Channels[0].ID != "General" || value.Channels[0].DisplayName == nil || *value.Channels[0].DisplayName != "General" || value.Channels[0].IsFavoriteByDefault == nil || !*value.Channels[0].IsFavoriteByDefault {
		t.Fatal("actual template fields or General channel were lost")
	}
}

func TestUnitTemplateCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, body := range []bool{false, true} {
		prefix := "expanded"
		if body {
			prefix = "json_body"
		}
		t.Run(prefix, func(t *testing.T) {
			read := func(label string) []byte { t.Helper(); return templateResponseFixture(t, prefix+"_"+label) }
			created, before, updated, after := read("create_201.json"), read("create_readback.json"), read("update.json"), read("update_readback.json")
			listed, deleted, missing := read("owned_list.json"), read("delete_204.body"), read("get_deleted_404.json")
			var failureReceipt struct {
				OperationID string `json:"operationId"`
				Errors      []struct {
					Code string `json:"errorCode"`
				} `json:"errors"`
			}
			if json.Unmarshal(missing, &failureReceipt) != nil || failureReceipt.OperationID == "" || len(failureReceipt.Errors) != 1 || failureReceipt.Errors[0].Code == "" {
				t.Fatal("actual error response lacks its operation identifier")
			}
			var receipt struct {
				Reference string `json:"@odata.id"`
				Name      string `json:"name"`
			}
			if json.Unmarshal(created, &receipt) != nil || receipt.Reference == "" || receipt.Name != "fixture-template" {
				t.Fatal("actual template receipt lacks its selector")
			}
			const initialShort = "SDK lab non-default initial summary"
			const initialDescription = "SDK lab non-default initial description"
			const updatedShort = "SDK lab non-default updated summary"
			const updatedDescription = "SDK lab non-default updated description"
			definition, createWire := responseDefinition(t, body, initialShort, initialDescription, false)
			update, updateWire := responseDefinition(t, body, updatedShort, updatedDescription, true)
			path := "/Teams.Templates/" + url.PathEscape(receipt.Reference)
			steps := []testutil.Exchange{
				{Request: testutil.WireRequest{Method: "GET", URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: serviceDiscoveryFixture(t)}},
				templateResponseWire("POST", "/Teams.Templates/api/TeamTemplates/v1.0/Tenant/en-US", "New-CsTeamTemplate", expectedCreateBody(t, createWire), 201, created),
				templateResponseWire("GET", path, "Get-CsTeamTemplate", nil, 200, before),
				templateResponseWire("PUT", path, "Update-CsTeamTemplate", updateWire, 200, updated),
				templateResponseWire("GET", path, "Get-CsTeamTemplate", nil, 200, after),
				templateResponseWire("GET", "/Teams.Templates/api/TeamTemplates/v1.0/en-US", "Get-CsTeamTemplateList", nil, 200, listed),
				templateResponseWire("DELETE", path, "Remove-CsTeamTemplate", nil, 204, deleted),
				templateResponseWire("GET", path, "Get-CsTeamTemplate", nil, 404, missing),
			}
			passCase(t, testutil.Case{Name: prefix + "-actual-crud", Steps: steps, Want: testutil.Expected{Status: 404, Body: missing, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, OperationIDs: []string{failureReceipt.OperationID}, Codes: []string{failureReceipt.Errors[0].Code}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamTemplates
				value, response, err := s.Create(ctx, "en-US", definition)
				if err != nil || response == nil || response.StatusCode() != 201 || value == nil || value.ODataID == nil || *value.ODataID != receipt.Reference || !reflect.DeepEqual(value.RawFields, mustTemplateObject(t, created)) {
					t.Fatal("actual creation receipt was changed")
				}
				selected, response, err := s.GetByID(ctx, receipt.Reference)
				if err != nil || response == nil || response.StatusCode() != 200 {
					t.Fatal("actual template detail was not decoded")
				}
				verifyActualTemplate(t, selected, before, initialShort, initialDescription)
				update.Channels = selected.Channels
				value, response, err = s.UpdateByID(ctx, receipt.Reference, update)
				if err != nil || response == nil || response.StatusCode() != 200 || value == nil || !reflect.DeepEqual(value.RawFields, mustTemplateObject(t, updated)) {
					t.Fatal("actual update receipt was changed")
				}
				selected, response, err = s.GetByID(ctx, receipt.Reference)
				if err != nil || response == nil || response.StatusCode() != 200 {
					t.Fatal("updated template detail was not decoded")
				}
				verifyActualTemplate(t, selected, after, updatedShort, updatedDescription)
				values, response, err := s.List(ctx, nil)
				if err != nil || response == nil || response.StatusCode() != 200 || values == nil || values.IsIncomplete || !reflect.DeepEqual(values.RawFields, mustTemplateObject(t, listed)) {
					t.Fatal("complete actual template list was changed")
				}
				found := 0
				for _, item := range values.Items {
					if item.ODataID != nil && *item.ODataID == receipt.Reference && item.Name != nil && *item.Name == receipt.Name {
						found++
					}
				}
				if found != 1 {
					t.Fatal("owned template absent or ambiguous in actual list")
				}
				response, err = s.DeleteByID(ctx, receipt.Reference)
				if err != nil || response == nil || response.StatusCode() != 204 {
					t.Fatal("actual empty delete reply changed")
				}
				_, response, err = s.GetByID(ctx, receipt.Reference)
				var failure *template.Error
				if !errors.As(err, &failure) || failure.Envelope == nil || len(failure.Envelope.Errors) != 1 || failure.Envelope.Errors[0].Code == nil || *failure.Envelope.Errors[0].Code != failureReceipt.Errors[0].Code {
					t.Fatal("actual template absence lost its error envelope")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
		})
	}
}

func mustTemplateObject(t *testing.T, raw []byte) map[string]json.RawMessage {
	t.Helper()
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil || value == nil {
		t.Fatal("invalid actual template object fixture")
	}
	return value
}
