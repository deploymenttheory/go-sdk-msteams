package teamsremotelogcollectionconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_configuration"
)

const (
	origin     = "https://api.interfaces.records.teams.microsoft.com"
	collection = "/Skype.Policy/configurations/TeamsRemoteLogCollectionConfiguration"
	getCmd     = "Get-CsTeamsRemoteLogCollectionConfiguration"
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

func wire(path string, status int, reply []byte) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", getCmd)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("remote log collection configuration contract %s differed: %+v", c.Name, result.Differences())
	}
}

func TestUnitConfigurationReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	list, global := readFile(t, "list.json"), readFile(t, "get_global.json")
	run(t, testutil.Case{Name: "actual-list",
		Steps: []testutil.Exchange{wire(collection, 200, list)}, Want: testutil.Expected{Status: 200, Body: list},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.List(ctx)
			if err != nil || len(values) != 1 || *values[0].Identity != "Global" || values[0].Devices == nil || values[0].RawFields["ConfigMetadata"] == nil {
				t.Fatal("global configuration was lost")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "actual-get-global",
		Steps: []testutil.Exchange{wire(collection+"/configuration/Global", 200, global)}, Want: testutil.Expected{Status: 200, Body: global},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.GetByIdentity(ctx, "Global")
			if err != nil || value == nil || len(value.Devices) != 0 || value.RawFields["SchemaVersion"] == nil {
				t.Fatal("global configuration was lost")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationEmbeddedDevicesAndFilter(t *testing.T) {
	t.Parallel()
	reply := []byte(`[{"Identity":"Tag:Other","Devices":[]},{"Identity":"Global","Devices":[{"Identity":"Global/d1","Id":"d1","DeviceId":"x","UserId":"u","ExpireAfter":"2026-11-26T00:00:00Z","Future":1}]}]`)
	for index, c := range []struct {
		pattern string
		count   int
	}{{"Glo*", 1}, {"Other", 1}, {"missing", 0}} {
		run(t, testutil.Case{Name: fmt.Sprintf("filter-%d", index),
			Steps: []testutil.Exchange{wire(collection, 200, reply)}, Want: testutil.Expected{Status: 200, Body: reply},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
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
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitConfigurationFailures(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "malformed",
		Steps: []testutil.Exchange{wire(collection+"/configuration/Global", 200, []byte(`{"Devices":{}}`))},
		Want:  testutil.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.GetByIdentity(ctx, "Global")
			if value != nil {
				t.Fatal("malformed reply returned a value")
			}
			return testutil.Observation{Response: response, Err: err}
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
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: configuration.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: invoke(ctx, sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations)}
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
