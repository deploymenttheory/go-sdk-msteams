package teamsremotelogcollectiondevice_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	device "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_device"
	"github.com/go-resty/resty/v2"
)

const (
	origin     = "https://api.interfaces.records.teams.microsoft.com"
	source     = "msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go"
	collection = "/Skype.Policy/configurations/TeamsRemoteLogCollectionDevice"
	probe      = "00000000-0000-4000-8000-000000000001"
	probeUser  = "00000000-0000-4000-8000-000000000002"
	selected   = collection + "/configuration/Global/" + probe
	expiry     = "2026-11-26T00:00:00Z"
	getCmd     = "Get-CsTeamsRemoteLogCollectionDevice"
	newCmd     = "New-CsTeamsRemoteLogCollectionDevice"
	setCmd     = "Set-CsTeamsRemoteLogCollectionDevice"
	removeCmd  = "Remove-CsTeamsRemoteLogCollectionDevice"
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

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	contentType := "application/json; charset=utf-8"
	if status >= 400 {
		contentType = "text/plain; charset=utf-8"
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "teams_remote_log_collection_device"
	if c.Source == "" {
		c.Source = source
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
		t.Fatalf("remote log collection device contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(command, set string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}

func apiError(status int, body []byte) *client.APIError {
	return client.NewAPIError(client.TeamsAdmin, status, nil, body)
}

func TestUnitDeviceReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	list, missing := readFile(t, "list.json"), readFile(t, "get_missing_404.json")
	run(t, contracttest.Case{Name: "actual-list", Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/list.json", Mappings: mapping(getCmd, "Identity"),
		Steps: []contracttest.Exchange{wire("GET", collection, getCmd, nil, 200, list)}, Want: contracttest.Expected{Status: 200, Body: list},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.List(ctx)
			if err != nil || values == nil || len(values) != 0 {
				t.Fatal("empty collection was not an empty list")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	for name, invoke := range map[string]func(context.Context, *device.Service) (*device.Device, *resty.Response, error){
		"actual-get-by-id-missing": func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
			return s.GetByID(ctx, probe)
		},
		"actual-get-by-identity-missing": func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
			return s.GetByIdentity(ctx, "Global/"+probe)
		},
	} {
		run(t, contracttest.Case{Name: name, Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/get_missing_404.json", Mappings: mapping(getCmd, "ParentAndRelativeKey"),
			Steps: []contracttest.Exchange{wire("GET", selected, getCmd, nil, 404, missing)},
			Want:  contracttest.Expected{Status: 404, Body: missing, APIError: apiError(404, missing)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := invoke(ctx, sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
}

func TestUnitDeviceListFilterMatchesIdentityLocally(t *testing.T) {
	t.Parallel()
	reply := []byte(`[{"Identity":"Global/` + probe + `","Id":"` + probe + `"},{"Identity":"Global/` + probeUser + `","Id":"` + probeUser + `"}]`)
	run(t, contracttest.Case{Name: "filter", Mappings: mapping(getCmd, "Filter"),
		Steps: []contracttest.Exchange{wire("GET", collection, getCmd, nil, 200, reply)}, Want: contracttest.Expected{Status: 200, Body: reply},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.ListFiltered(ctx, "*-000000000001")
			if err != nil || len(values) != 1 || *values[0].ID != probe {
				t.Fatal("local filter result differs")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitDeviceRejectedWritesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	invalid := readFile(t, "create_invalid_device_400.json")
	run(t, contracttest.Case{Name: "actual-create-invalid-device", Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/create_invalid_device_400.json", Mappings: mapping(newCmd, "ParentAndRelativeKey"),
		Steps: []contracttest.Exchange{wire("POST", collection, newCmd, json.RawMessage(`{"Identity":"Global/`+probe+`","Id":null,"DeviceId":"msteams-sdk-not-a-guid","UserId":"`+probeUser+`","ExpireAfter":"`+expiry+`"}`), 400, invalid)},
		Want:  contracttest.Expected{Status: 400, Body: invalid, APIError: apiError(400, invalid)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.CreateByID(ctx, probe, device.CreateRequest{DeviceID: "msteams-sdk-not-a-guid", UserID: probeUser, ExpireAfter: expiry})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})

	missing := readFile(t, "update_missing_404.json")
	run(t, contracttest.Case{Name: "actual-update-missing", Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/update_missing_404.json", Mappings: mapping(setCmd, "ParentAndRelativeKey"),
		Steps: []contracttest.Exchange{wire("GET", selected, setCmd, nil, 404, missing)},
		Want:  contracttest.Expected{Status: 404, Body: missing, APIError: apiError(404, missing)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.UpdateByID(ctx, probe, device.UpdateRequest{Settings: device.Settings{UserID: policy.Set(probeUser)}})
			if result == nil || result.PatchInvoked || result.ReadResponse == nil {
				t.Fatal("failed preparation read was not reported")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})

	patch := readFile(t, "patch_missing_404.json")
	run(t, contracttest.Case{Name: "actual-patch-missing", Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/patch_missing_404.json", Mappings: mapping(setCmd, "ParentAndRelativeKey"),
		Steps: []contracttest.Exchange{wire("PATCH", selected, setCmd, json.RawMessage(`{"UserId":"`+probeUser+`","Id":"`+probe+`"}`), 404, patch)},
		Want:  contracttest.Expected{Status: 404, Body: patch, APIError: apiError(404, patch)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.PatchByID(ctx, probe, device.PatchRequest{Settings: device.Settings{UserID: policy.Set(probeUser)}, BodyID: policy.Set(probe)})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})

	deleted := readFile(t, "delete_missing_404.json")
	run(t, contracttest.Case{Name: "actual-delete-missing", Source: "msteams/teams_admin_api/teams_remote_log_collection_device/testdata/delete_missing_404.json", Mappings: mapping(removeCmd, "ParentAndRelativeKey"),
		Steps: []contracttest.Exchange{wire("DELETE", selected, removeCmd, json.RawMessage(`{"Id":"`+probe+`"}`), 404, deleted)},
		Want:  contracttest.Expected{Status: 404, Body: deleted, APIError: apiError(404, deleted)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.DeleteByID(ctx, probe, policy.Controls{})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitDeviceWriteShapes(t *testing.T) {
	t.Parallel()
	stored := []byte(`{"Identity":"Global/` + probe + `","Id":"` + probe + `","DeviceId":"` + probe + `","UserId":"` + probeUser + `","ExpireAfter":"` + expiry + `","Future":1}`)
	settings := `"DeviceId":"` + probe + `","UserId":"` + probeUser + `","ExpireAfter":"` + expiry + `"`
	input := device.CreateRequest{DeviceID: probe, UserID: probeUser, ExpireAfter: expiry}
	forced := input
	forced.Force = new(true)
	for _, c := range []struct {
		name, method, path, command, set, body string
		status                                 int
		invoke                                 func(context.Context, *device.Service) (*device.Device, *resty.Response, error)
	}{
		{"create-by-identity", "POST", collection, newCmd, "Identity", `{` + settings + `,"Identity":"Global/` + probe + `","Id":"` + probe + `","Force":"True"}`, 200,
			func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
				return s.CreateByIdentity(ctx, "Global/"+probe, forced)
			}},
		{"create-by-id", "POST", collection, newCmd, "ParentAndRelativeKey", `{` + settings + `,"Identity":"Global/` + probe + `","Id":null}`, 200,
			func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
				return s.CreateByID(ctx, probe, input)
			}},
		{"patch-by-identity", "PATCH", selected, setCmd, "Identity", `{"ExpireAfter":null,"Identity":"Global/` + probe + `","Force":"False"}`, 200,
			func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
				return s.PatchByIdentity(ctx, "Global/"+probe, device.PatchRequest{Settings: device.Settings{ExpireAfter: policy.Null[string]()}, Controls: policy.Controls{Force: new(false)}})
			}},
		{"delete-by-identity", "DELETE", selected, removeCmd, "Identity", `{"Identity":"Global/` + probe + `"}`, 204,
			func(ctx context.Context, s *device.Service) (*device.Device, *resty.Response, error) {
				return s.DeleteByIdentity(ctx, "Global/"+probe, policy.Controls{})
			}},
	} {
		reply := stored
		if c.status == 204 {
			reply = nil
		}
		run(t, contracttest.Case{Name: c.name, Mappings: mapping(c.command, c.set),
			Steps: []contracttest.Exchange{wire(c.method, c.path, c.command, json.RawMessage(c.body), c.status, reply)},
			Want:  contracttest.Expected{Status: c.status, Body: reply},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices)
				if err != nil {
					t.Fatal(err)
				}
				if c.status == 204 && value != nil || c.status == 200 && (value == nil || *value.UserID != probeUser || value.RawFields["Future"] == nil) {
					t.Fatal("device reply was fabricated or lost")
				}
				if value != nil && fmt.Sprintf("%v %#v", *value, *value) != "remote log collection device (redacted) remote log collection device (redacted)" {
					t.Fatal("device formatting exposed identifiers")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitDeviceCreateGeneratesAndKeepsID(t *testing.T) {
	t.Parallel()
	invalid := readFile(t, "create_invalid_device_400.json")
	var result *device.CreateResult
	// The body carries a random Id, so the only permitted difference is that
	// unasserted request body; Create-by-ID/Identity cases assert the shape.
	outcome, err := contracttest.Run(t.Context(), contracttest.Case{Name: "generated-id", Family: "teams_remote_log_collection_device", Source: source, Scope: "service", Mappings: mapping(newCmd, "ParentAndRelativeKey"),
		Steps: []contracttest.Exchange{wire("POST", collection, newCmd, nil, 400, invalid)},
		Want:  contracttest.Expected{Status: 400, Body: invalid, APIError: apiError(400, invalid)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var response *resty.Response
			var err error
			result, response, err = sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.Create(ctx, device.CreateRequest{DeviceID: "x", UserID: probeUser, ExpireAfter: expiry})
			return contracttest.Observation{Response: response, Err: err}
		}})
	if err != nil {
		t.Fatal(err)
	}
	if differences := outcome.Differences(); len(differences) != 1 || differences[0].Kind != "unexpected-body" {
		t.Fatalf("generated create differed beyond its body: %+v", differences)
	}
	if result == nil || result.HTTPAccepted || result.Identity != "Global/"+result.ID ||
		!regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(result.ID) {
		t.Fatal("generated identity was not retained on failure")
	}
}

func TestUnitDeviceInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	complete := device.CreateRequest{DeviceID: probe, UserID: probeUser, ExpireAfter: expiry}
	for name, invoke := range map[string]func(context.Context, *device.Service) error{
		"blank-id":       func(ctx context.Context, s *device.Service) error { _, _, err := s.GetByID(ctx, ""); return err },
		"slash-id":       func(ctx context.Context, s *device.Service) error { _, _, err := s.GetByID(ctx, "a/b"); return err },
		"blank-identity": func(ctx context.Context, s *device.Service) error { _, _, err := s.GetByIdentity(ctx, " "); return err },
		"empty-filter":   func(ctx context.Context, s *device.Service) error { _, _, err := s.ListFiltered(ctx, ""); return err },
		"missing-device": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.CreateByID(ctx, probe, device.CreateRequest{UserID: probeUser, ExpireAfter: expiry})
			return err
		},
		"missing-user": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.Create(ctx, device.CreateRequest{DeviceID: probe, ExpireAfter: expiry})
			return err
		},
		"missing-expiry": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.CreateByIdentity(ctx, "Global/"+probe, device.CreateRequest{DeviceID: probe, UserID: probeUser})
			return err
		},
		"bad-identity": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.CreateByIdentity(ctx, "Global\\x", complete)
			return err
		},
		"empty-update": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.UpdateByID(ctx, probe, device.UpdateRequest{})
			return err
		},
		"empty-patch": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.PatchByIdentity(ctx, "Global/"+probe, device.PatchRequest{})
			return err
		},
		"patch-no-bodyid": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.PatchByID(ctx, probe, device.PatchRequest{Settings: device.Settings{UserID: policy.Set(probeUser)}})
			return err
		},
		"patch-mixed": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.PatchByIdentity(ctx, "Global/"+probe, device.PatchRequest{Settings: device.Settings{UserID: policy.Set(probeUser)}, BodyID: policy.Set(probe)})
			return err
		},
		"delete-blank-id": func(ctx context.Context, s *device.Service) error {
			_, _, err := s.DeleteByID(ctx, "", policy.Controls{})
			return err
		},
	} {
		run(t, contracttest.Case{Name: name, Mappings: mapping(newCmd, "ParentAndRelativeKey"), Want: contracttest.Expected{Error: device.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: invoke(ctx, sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices)}
			}})
	}
}

func TestUnitDeviceMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{"array": `[]`, "string-id": `{"Id":1}`, "scalar": `7`} {
		run(t, contracttest.Case{Name: "malformed-" + name, Mappings: mapping(getCmd, "ParentAndRelativeKey"),
			Steps: []contracttest.Exchange{wire("GET", selected, getCmd, nil, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices.GetByID(ctx, probe)
				if value != nil {
					t.Fatal("malformed reply returned a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitDeviceNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *device.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.GetByID(t.Context(), probe); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := device.NewService(nil).List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
