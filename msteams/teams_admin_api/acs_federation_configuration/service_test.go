package acsfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitPinnedProviderDispatch(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	seen := 0
	for _, f := range captured(t) {
		if !strings.HasPrefix(f.Name, "acs-") {
			continue
		}
		seen++
		command := "Get-CsTeamsAcsFederationConfiguration"
		if f.Capture.Method == "PATCH" {
			command = "Set-CsTeamsAcsFederationConfiguration"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{exchange(f)}}
			if f.Capture.FailureType != nil {
				c.Want.Error = client.ErrDecode
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.ACSFederationConfigurations
				if strings.HasSuffix(f.Name, "-list") && f.Name != "acs-null-list" {
					v, r, e := s.List(ctx)
					if e == nil && (v == nil || len(v) != 0) {
						e = errors.New("empty list changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Capture.Method == "PATCH" {
					v, r, e := s.PatchByIdentity(ctx, scoped, config.PatchRequest{Settings: config.Settings{Fields: config.Fields{EnableAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.Strings()}, BodyIdentity: policy.Set("read-model-identity")})
					if e == nil && v != nil {
						e = errors.New("empty PATCH invented resource")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
				v, r, e := s.GetByIdentity(ctx, scoped)
				if e == nil && v == nil {
					e = errors.New("object read lost")
				}
				if e == nil && f.Name == "acs-null-list" && string(v.RawFields["AllowedAcsResources"]) != "null" {
					e = errors.New("raw null list changed")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	if seen != 4 {
		t.Fatal("provider fixture count differs")
	}
	evidence(t, results)
}
func TestUnitComposedUpdatePreservesReadIdentityAndSparseFields(t *testing.T) {
	t.Parallel()
	path := "/Skype.Policy/configurations/TeamsAcsFederationConfiguration/configuration/scope%20/Fixture%20%2F%20Child"
	r := run(t, contracttest.Case{Name: "composed-update", Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAcsFederationConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		wire("GET", path, "Set-CsTeamsAcsFederationConfiguration", nil, 200, `{"Identity":"read-model-identity","AllowedAcsResources":["a","a","A","b"],"LabelForAllowedAcsUsers":"preserved","Future":false}`), wire("PATCH", path, "Set-CsTeamsAcsFederationConfiguration", json.RawMessage(`{"Identity":"read-model-identity","AllowedAcsResources":["a","A","b","a"],"HideBannerForAllowedAcsUsers":false}`), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.UpdateByIdentity(ctx, scoped, config.UpdateRequest{Fields: config.Fields{HideBannerForAllowedAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.ChangeStrings([]string{"b", "a"}, []string{"a", "b"})})
		if e == nil && (v == nil || !v.PatchInvoked || !v.HTTPAccepted || v.After != nil || v.Before == nil || v.ReadResponse == nil) {
			e = errors.New("update preparation/HTTP acceptance lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{r})
}
