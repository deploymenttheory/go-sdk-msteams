package teamsappsetuppolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

func TestUnitAppSetupUpdateKeepsPreparationFalseAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","AllowSideLoading":true,"AllowUserPinning":true,"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsAppSetupPolicy", nil, c.read, c.readBody)}
		if c.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsAppSetupPolicy", json.RawMessage(`{"Identity":"read-model-identity","AllowSideLoading":false}`), c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + c.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAppSetupPolicy", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{AllowSideLoading: policy.Set(false)}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("setup preparation/acceptance lost")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.AllowSideLoading == nil || !*v.Before.AllowSideLoading || v.Before.AllowUserPinning == nil || !*v.Before.AllowUserPinning || v.After != nil {
					return contracttest.Observation{Err: errors.New("setup read model changed")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || c.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("setup partial failure lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitAppSetupComposesEveryListWithItsOwnKey(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, field := range []string{"AppPresetList", "PinnedAppBarApps", "PinnedMessageBarApps", "AppPresetMeetingList", "AdditionalCustomizationApps", "PinnedCallingBarApps"} {
		initial, expected := `[{"Identity":"first","Id":"a"}]`, `[{"Identity":"second","Id":"a"},{"Identity":null,"Id":"A"},{"Identity":null,"Id":"b"}]`
		more := `[{"Identity":"second","Id":"a","Future":false},{"Id":"A"}]`
		if field == "PinnedAppBarApps" || field == "PinnedMessageBarApps" {
			initial = `[{"Identity":"first","Id":"a","Order":1}]`
			more = `[{"Identity":"second","Id":"a","Order":1,"Future":false},{"Id":"a","Order":2}]`
			expected = `[{"Identity":"second","Id":"a","Order":1},{"Identity":null,"Id":"a","Order":2},{"Identity":null,"Id":"b","Order":9007199254740993}]`
		}
		if field == "AdditionalCustomizationApps" {
			initial = `[{"Identity":"first","Id":"a","AdditionalCustomizationId":"one"}]`
			more = `[{"Identity":"second","Id":"a","AdditionalCustomizationId":"one","Future":false},{"Id":"a","AdditionalCustomizationId":"two"}]`
			expected = `[{"Identity":"second","Id":"a","AdditionalCustomizationId":"one"},{"Identity":null,"Id":"a","AdditionalCustomizationId":"two"},{"Identity":null,"Id":"b","AdditionalCustomizationId":"three"}]`
		}
		read := `{"Identity":"read-model-identity","Future":false,"` + field + `":` + initial + `,"` + field + `":` + more + `}`
		body := `{"Identity":"read-model-identity","` + field + `":` + expected + `}`
		results = append(results, run(t, contracttest.Case{Name: "edit-" + strings.ToLower(field), Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAppSetupPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("GET", item, "Set-CsTeamsAppSetupPolicy", nil, 200, read), wire("PATCH", item, "Set-CsTeamsAppSetupPolicy", json.RawMessage(body), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			input := resource.UpdateRequest{}
			app := resource.ChangeApps([]*resource.App{resource.NewApp("b")}, []*resource.App{resource.NewApp("a")})
			pin := resource.ChangePins([]*resource.PinnedApp{resource.NewPinnedApp("b", largeOrder)}, []*resource.PinnedApp{resource.NewPinnedApp("a", 1)})
			switch field {
			case "AppPresetList":
				input.AppPresetList = app
			case "AppPresetMeetingList":
				input.AppPresetMeetingList = app
			case "PinnedCallingBarApps":
				input.PinnedCallingBarApps = app
			case "PinnedAppBarApps":
				input.PinnedAppBarApps = pin
			case "PinnedMessageBarApps":
				input.PinnedMessageBarApps = pin
			case "AdditionalCustomizationApps":
				input.AdditionalCustomizationApps = resource.ChangeCustomizations([]*resource.CustomizationApp{resource.NewCustomizationApp("b", "three")}, []*resource.CustomizationApp{resource.NewCustomizationApp("a", "one")})
			}
			v, r, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.UpdateByIdentity(ctx, "Global", input)
			if e == nil && (v == nil || v.Before == nil || v.Before.RawFields["Future"] == nil || !v.HTTPAccepted) {
				e = errors.New("setup read snapshot/acceptance lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
func TestUnitAppSetupBuildersNeverAcquireCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-setup-builders", Source: sourceModels, Invoke: func(_ context.Context, _ *msteams.Client) contracttest.Observation {
		apps, e := resource.ReplaceApps(resource.NewApp("")).Apply(nil)
		if e != nil || len(apps) != 1 || *apps[0].ID != "" {
			return contracttest.Observation{Err: errors.New("local app builder changed")}
		}
		pins, e := resource.ReplacePins(resource.NewPinnedApp("future-app", largeOrder)).Apply(nil)
		if e != nil || len(pins) != 1 || pins[0].Order != largeOrder {
			return contracttest.Observation{Err: errors.New("local pin builder changed")}
		}
		customs, e := resource.ReplaceCustomizations(resource.NewCustomizationApp("future-app", "")).Apply(nil)
		if e != nil || len(customs) != 1 || *customs[0].AdditionalCustomizationID != "" {
			return contracttest.Observation{Err: errors.New("local customization builder changed")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{result})
}
