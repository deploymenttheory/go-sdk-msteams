package teamsremotelogcollectionconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_configuration"
)

const (
	origin     = "https://api.interfaces.records.teams.microsoft.com"
	collection = "/Skype.Policy/configurations/TeamsRemoteLogCollectionConfiguration"
	getCmd     = "Get-CsTeamsRemoteLogCollectionConfiguration"
	fixtures   = "msteams/teams_admin_api/teams_remote_log_collection_configuration/testdata/"
)

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = directory.Close() }()
	body, err := directory.ReadFile(name)
	if err != nil || !json.Valid(body) {
		t.Fatal("invalid acceptance fixture")
	}
	return body
}

func wire(path string, status int, reply []byte) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", getCmd)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "teams_remote_log_collection_configuration"
	if c.Source == "" {
		c.Source = "msteams/teams_admin_api/teams_remote_log_collection_configuration/service_test.go"
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("remote log collection configuration contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(set string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: getCmd, ParameterSet: set}}
}

func TestUnitConfigurationReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	list, global := readFile(t, "list.json"), readFile(t, "get_global.json")
	run(t, contracttest.Case{Name: "actual-list", Source: fixtures + "list.json", Mappings: mapping("Identity"),
		Steps: []contracttest.Exchange{wire(collection, 200, list)}, Want: contracttest.Expected{Status: 200, Body: list},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.List(ctx)
			if err != nil || len(values) != 1 || *values[0].Identity != "Global" || values[0].Devices == nil || values[0].RawFields["ConfigMetadata"] == nil {
				t.Fatal("global configuration was lost")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "actual-get-global", Source: fixtures + "get_global.json", Mappings: mapping("Identity"),
		Steps: []contracttest.Exchange{wire(collection+"/configuration/Global", 200, global)}, Want: contracttest.Expected{Status: 200, Body: global},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.GetByIdentity(ctx, "Global")
			if err != nil || value == nil || len(value.Devices) != 0 || value.RawFields["SchemaVersion"] == nil {
				t.Fatal("global configuration was lost")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationEmbeddedDevicesAndFilter(t *testing.T) {
	t.Parallel()
	reply := []byte(`[{"Identity":"Tag:Other","Devices":[]},{"Identity":"Global","Devices":[{"Identity":"Global/d1","Id":"d1","DeviceId":"x","UserId":"u","ExpireAfter":"2026-11-26T00:00:00Z","Future":1}]}]`)
	for index, c := range []struct {
		pattern string
		count   int
	}{{"Glo*", 1}, {"Other", 1}, {"missing", 0}} {
		run(t, contracttest.Case{Name: fmt.Sprintf("filter-%d", index), Mappings: mapping("Filter"),
			Steps: []contracttest.Exchange{wire(collection, 200, reply)}, Want: contracttest.Expected{Status: 200, Body: reply},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.ListFiltered(ctx, c.pattern)
				if err != nil || len(values) != c.count {
					t.Fatal("local filter result differs")
				}
				if c.pattern == "Glo*" {
					d := values[0].Devices[0]
					if *d.ID != "d1" || d.RawFields["Future"] == nil || fmt.Sprintf("%v %v", *values[0], *d) != "remote log collection configuration (redacted) remote log collection device (redacted)" {
						t.Fatal("embedded device lost or exposed")
					}
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitConfigurationFailures(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "malformed", Mappings: mapping("Identity"),
		Steps: []contracttest.Exchange{wire(collection+"/configuration/Global", 200, []byte(`{"Devices":{}}`))},
		Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.GetByIdentity(ctx, "Global")
			if value != nil {
				t.Fatal("malformed reply returned a value")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	for name, invoke := range map[string]func(context.Context, *configuration.Service) error{
		"empty-filter": func(ctx context.Context, s *configuration.Service) error {
			_, _, err := s.ListFiltered(ctx, "")
			return err
		},
		"blank-identity": func(ctx context.Context, s *configuration.Service) error {
			_, _, err := s.GetByIdentity(ctx, "")
			return err
		},
	} {
		run(t, contracttest.Case{Name: name, Mappings: mapping("Filter"), Want: contracttest.Expected{Error: configuration.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: invoke(ctx, sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations)}
			}})
	}
	var s *configuration.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.List(t.Context()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
}
