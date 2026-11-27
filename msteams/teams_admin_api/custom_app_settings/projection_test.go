package customappsettings_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	cas "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/custom_app_settings"
)

func TestUnitSourceSetProjection(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name   string `json:"case"`
			Input  string `json:"input_json"`
			Output string `json:"request_serialized"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readFile(t, "testdata/pinned-set-projection.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 12 {
		t.Fatal("source observation cases missing")
	}
	for _, item := range fixture.Cases {
		t.Run(item.Name, func(t *testing.T) {
			got, err := cas.ProjectSideloadedAppsInteraction(json.RawMessage(item.Input), true)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			if json.Unmarshal(got.Body, &actual) != nil || json.Unmarshal([]byte(item.Output), &expected) != nil || !reflect.DeepEqual(actual, expected) {
				t.Fatal("source request projection differs")
			}
			if strings.Contains(fmt.Sprintf("%v %#v", got, got), "known-message") {
				t.Fatal("projection formatting exposes data")
			}
		})
	}
	for _, raw := range []string{"null", "[]", "{"} {
		if _, err := cas.ProjectSideloadedAppsInteraction(json.RawMessage(raw), false); !errors.Is(err, client.ErrDecode) {
			t.Fatal("invalid read accepted")
		}
	}
	projection, err := cas.ProjectSideloadedAppsInteraction(json.RawMessage(`{"future":"private","appAccessRequestConfig":{"future":42},"isSideloadedAppsInteractionEnabled":false}`), true)
	if err != nil || !reflect.DeepEqual(projection.OmittedPaths, []string{"appAccessRequestConfig", "appAccessRequestConfig.future", "future"}) || !reflect.DeepEqual(projection.ChangedPaths, []string{"appSettingsList", "isSideloadedAppsInteractionEnabled"}) {
		t.Fatal("projection diagnostics differ")
	}
}

func TestUnitPreparedSourceSet(t *testing.T) {
	for _, status := range []int{204, 200, 403} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			before := []byte(`{"LobLogo":"private-logo","isAppsEnabled":false,"appAccessRequestConfig":{"future":"drop"}}`)
			body := json.RawMessage(`{"LobLogo":"private-logo","isAppsEnabled":false,"isSideloadedAppsInteractionEnabled":true,"appSettingsList":[]}`)
			run(t, contracttest.Case{Name: fmt.Sprintf("prepared-source-set-%d", status), Mappings: mapping(setCommand), Want: contracttest.Expected{Status: status}, Steps: []contracttest.Exchange{discovery(t), wire("GET", "Get-CsTeamsSettingsCustomApp_Get", nil, 200, before), wire("PUT", setCommand, body, status, nil)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.CustomAppSettings
				prepared, read, err := service.PrepareSideloadedAppsInteraction(ctx, true)
				if err != nil || prepared == nil || read == nil {
					t.Fatal("prepare failed")
				}
				projection := prepared.Projection()
				projection.Body[0] = '!'
				projection.OmittedPaths[0] = "changed"
				snapshot := prepared.Before()
				snapshot[0] = '!'
				if string(prepared.Before()) != string(before) || prepared.Projection().OmittedPaths[0] != "appAccessRequestConfig" {
					t.Fatal("inspection changed frozen request")
				}
				if strings.Contains(fmt.Sprintf("%v %#v", prepared, prepared), "private-logo") {
					t.Fatal("prepared formatting exposes data")
				}
				if _, _, err := cas.NewService(nil).UpdateSideloadedAppsInteraction(ctx, prepared); !errors.Is(err, cas.ErrInput) {
					t.Fatal("foreign service accepted request")
				}
				canceled, cancel := context.WithCancel(ctx)
				cancel()
				canceledResult, response, err := service.UpdateSideloadedAppsInteraction(canceled, prepared)
				if !errors.Is(err, context.Canceled) || response != nil || canceledResult == nil || canceledResult.ReadResponse != read || canceledResult.UpdateInvoked {
					t.Fatal("canceled update lost preparation receipt")
				}
				result, response, err := service.UpdateSideloadedAppsInteraction(ctx, prepared)
				if result == nil || result.ReadResponse != read || result.UpdateResponse != response || !result.UpdateInvoked || result.HTTPAccepted != (status < 300) {
					t.Fatal("action receipts lost")
				}
				if status == 204 && err != nil || status == 200 && !errors.Is(err, cas.ErrResponse) || status == 403 && err == nil {
					t.Fatal("wrong action disposition")
				}
				if strings.Contains(fmt.Sprintf("%v %#v", result, result), "private-logo") {
					t.Fatal("result formatting exposes data")
				}
				return contracttest.Observation{Response: response}
			}})
		})
	}
}

func TestUnitSourceSetPrepareFailure(t *testing.T) {
	for _, item := range []struct {
		name   string
		status int
		body   string
	}{{"denied", 403, `{}`}, {"null", 200, `null`}, {"array", 200, `[]`}} {
		t.Run(item.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "prepare-failure-" + item.name, Mappings: mapping(setCommand), Want: contracttest.Expected{Status: item.status}, Steps: []contracttest.Exchange{discovery(t), wire("GET", "Get-CsTeamsSettingsCustomApp_Get", nil, item.status, []byte(item.body))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				prepared, response, err := sdk.TeamsAdminAPI.CustomAppSettings.PrepareSideloadedAppsInteraction(ctx, true)
				if prepared != nil || response == nil || err == nil {
					t.Fatal("failed prepare produced request or lost receipt")
				}
				return contracttest.Observation{Response: response}
			}})
		})
	}
}
