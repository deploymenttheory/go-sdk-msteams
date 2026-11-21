package applicationmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/ApplicationMeetingConfiguration"
const getCommand = "Get-CsApplicationMeetingConfiguration"
const setCommand = "Set-CsApplicationMeetingConfiguration"
const source = "contracts/endpoints/8.1.0/fixtures/application-meeting-dispatch.json"

func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored"}}`)}}
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://forest.fixture.invalid/"}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	if len(c.Steps) > 0 {
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	} else {
		c.Scope = "local"
	}
	c.Family = "application_meeting_configuration"
	c.Source = source
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("configuration differs: %v %+v", err, r.Differences())
	}
	return r
}
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("application-meeting-configuration-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("configuration evidence write failed")
	}
}

type fixture struct {
	Cases []struct {
		Name     string          `json:"name"`
		Method   string          `json:"method"`
		URL      string          `json:"generated_url"`
		Body     json.RawMessage `json:"body"`
		Status   int             `json:"response_status"`
		Response string          `json:"response_body"`
	} `json:"cases"`
	Models []struct {
		Name     string          `json:"name"`
		Rejected bool            `json:"rejected"`
		Wire     json.RawMessage `json:"wire"`
	} `json:"models"`
}

func pinned(t *testing.T) fixture {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var f fixture
	if json.Unmarshal(raw, &f) != nil || len(f.Cases) != 3 || len(f.Models) != 11 {
		t.Fatal("invalid pinned observations")
	}
	return f
}

func TestUnitPinnedDispatchFromRootClient(t *testing.T) {
	t.Parallel()
	f := pinned(t)
	var results []contracttest.Result
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, observed := range f.Cases {
			command, set := getCommand, "Identity"
			if observed.Method == "PUT" {
				command = setCommand
			}
			body := observed.Body
			if string(body) == "null" {
				body = nil
			}
			results = append(results, run(t, contracttest.Case{Name: observed.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(), wire(observed.Method, strings.TrimPrefix(observed.URL, origin), command, body, observed.Status, observed.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
				switch observed.Name {
				case "application-meeting-list":
					v, r, e := s.List(ctx)
					if e == nil && (v == nil || len(v) != 0) {
						e = errors.New("empty array changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				case "application-meeting-read":
					v, r, e := s.GetByIdentity(ctx, "Global")
					if e == nil && (v.Identity != "Global" || v.AppIDs == nil || len(v.AppIDs) != 0) {
						e = errors.New("read projection changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				case "application-meeting-update":
					r, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{AppIDs: configuration.ReplaceAppIDs([]string{"fixture-a", "fixture-b"})})
					return contracttest.Observation{Response: r, Err: e}
				default:
					return contracttest.Observation{Err: errors.New("unexpected fixture")}
				}
			}}))
		}
	}
	evidence(t, results)
}

func change(t *testing.T, name string) *configuration.AppIDChange {
	t.Helper()
	switch name {
	case "omitted":
		return nil
	case "scalar":
		return configuration.ReplaceAppIDs([]string{"fixture-app"})
	case "replacement":
		return configuration.ReplaceAppIDs([]string{"fixture-a", "fixture-b"})
	case "empty-replacement":
		return configuration.ReplaceAppIDs([]string{})
	case "null-replacement":
		return configuration.ReplaceAppIDs(nil)
	case "operators":
		return configuration.ChangeAppIDs([]string{"fixture-a"}, []string{"fixture-b"})
	case "empty-operators":
		return configuration.ChangeAppIDs(nil, nil)
	case "legacy-object":
		v, err := configuration.LegacyAppIDs([]json.RawMessage{json.RawMessage(`{"AppId":"fixture-app","Enabled":false,"DataSource":"fixture-source","Key":"fixture-key"}`)})
		if err != nil {
			t.Fatal(err)
		}
		return v
	default:
		t.Fatal("unrecognized accepted model")
		return nil
	}
}

func TestUnitUpdateMatchesPinnedListNormalizer(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, observed := range pinned(t).Models {
		if observed.Rejected {
			continue
		}
		input := configuration.UpdateRequest{AppIDs: change(t, observed.Name)}
		results = append(results, run(t, contracttest.Case{Name: observed.Name, Mode: auth.Application, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire("PUT", collection+"/configuration/Global", setCommand, observed.Wire, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.UpdateGlobal(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitListFilterIsLocalAndReadSnapshotsPreserveData(t *testing.T) {
	t.Parallel()
	reply := `[{"Identity":"tag:Fixture","AllowRemoveParticipantAppIds":null,"Future":{"enabled":false}},{"Identity":"Global","AllowRemoveParticipantAppIds":[]},{"Identity":"Other"},{"Identity":"tag:fixture-2","AllowRemoveParticipantAppIds":["a",""]}]`
	var results []contracttest.Result
	for index, pattern := range []string{"", "Fixture*"} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprint("filtered-", index), Mode: auth.Delegated, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{discovery(), wire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.ListFiltered(ctx, pattern)
			if e == nil {
				want := []string{"Global", "tag:Fixture", "Other", "tag:fixture-2"}
				if pattern != "" {
					want = []string{"tag:Fixture", "tag:fixture-2"}
				}
				var ids []string
				for _, item := range v {
					ids = append(ids, item.Identity)
					raw, err := json.Marshal(item)
					var fields map[string]json.RawMessage
					if err != nil || json.Unmarshal(raw, &fields) != nil || !reflect.DeepEqual(fields, item.RawFields) {
						e = errors.New("snapshot changed")
					}
				}
				if !reflect.DeepEqual(ids, want) {
					e = errors.New("filter/order changed")
				}
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitIdentityEscapingAndMutationDenialPreserveMetadata(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "escaped", Mode: auth.Application, Steps: []contracttest.Exchange{discovery(), wire("GET", collection+"/configuration/tag:Fixture%20%2F%20Config", getCommand, nil, 200, `{"Identity":"tag:Fixture / Config","AllowRemoveParticipantAppIds":[{"Future":false}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.GetByIdentity(ctx, "tag:Fixture / Config")
		if e == nil && (v.AppIDs != nil || string(v.RawFields["AllowRemoveParticipantAppIds"]) != `[{"Future":false}]`) {
			e = errors.New("raw compatibility shape lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	for _, status := range []int{400, 403, 429, 503} {
		run(t, contracttest.Case{Name: fmt.Sprint("denied-", status), Mode: auth.Application, Steps: []contracttest.Exchange{discovery(), wire("PUT", collection+"/configuration/Global", setCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"code":"FutureDenied","message":"fixture private text"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.UpdateGlobal(ctx, configuration.UpdateRequest{})
			var api *client.APIError
			if !errors.As(e, &api) || r == nil || r.StatusCode() != status || !strings.Contains(string(r.Body()), "fixture private text") {
				return contracttest.Observation{Err: errors.New("denial metadata lost")}
			}
			if strings.Contains(fmt.Sprint(e), "fixture private text") {
				return contracttest.Observation{Err: errors.New("error exposed payload")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestUnitMalformedAndUnexpectedSuccess(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, method, body string
		status             int
		list               bool
	}{
		{"item-null", "GET", "null", 200, false}, {"item-empty", "GET", "{}", 200, false}, {"item-array", "GET", "[]", 200, false}, {"list-null", "GET", "null", 200, true}, {"list-envelope", "GET", `{"value":[]}`, 200, true}, {"list-invalid-item", "GET", `[{}]`, 200, true}, {"unexpected-item-status", "GET", `{"Identity":"Global"}`, 201, false}, {"unexpected-write-status", "PUT", "", 200, false},
	} {
		path, command := collection+"/configuration/Global", getCommand
		var body json.RawMessage
		if tc.list {
			path = collection
		}
		if tc.method == "PUT" {
			command = setCommand
			body = json.RawMessage(`{"Identity":"Global"}`)
		}
		run(t, contracttest.Case{Name: tc.name, Mode: auth.Application, Steps: []contracttest.Exchange{discovery(), wire(tc.method, path, command, body, tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
			if tc.method == "PUT" {
				response, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{})
				if !errors.Is(e, configuration.ErrResponse) {
					return contracttest.Observation{Err: errors.New("unexpected write status accepted")}
				}
				return contracttest.Observation{Response: response}
			}
			if tc.list {
				_, response, e := s.List(ctx)
				if !errors.Is(e, client.ErrDecode) {
					return contracttest.Observation{Err: errors.New("malformed list accepted")}
				}
				return contracttest.Observation{Response: response}
			}
			_, response, e := s.GetByIdentity(ctx, "Global")
			want := client.ErrDecode
			if tc.status != 200 {
				want = configuration.ErrResponse
			}
			if !errors.Is(e, want) {
				return contracttest.Observation{Err: errors.New("malformed item/status accepted")}
			}
			return contracttest.Observation{Response: response}
		}})
	}
}

func TestUnitInvalidInputsStayLocal(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "invalid-input", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, configuration.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid read identity accepted")}
			}
			if _, e := s.UpdateByIdentity(ctx, id, configuration.UpdateRequest{}); !errors.Is(e, configuration.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid write identity accepted")}
			}
		}
		if _, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{AppIDs: &configuration.AppIDChange{}}); !errors.Is(e, configuration.ErrInput) {
			return contracttest.Observation{Err: errors.New("zero change accepted")}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, configuration.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid wildcard accepted")}
		}
		//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		if _, _, e := s.List(nil); !errors.Is(e, client.ErrNilContext) {
			return contracttest.Observation{Err: errors.New("nil context accepted")}
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.GetByIdentity(c, "Global"); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		var missing *configuration.Service
		if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
			return contracttest.Observation{Err: errors.New("nil service accepted")}
		}
		return contracttest.Observation{}
	}})
}
