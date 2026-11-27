package applicationmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/ApplicationMeetingConfiguration"
const getCommand = "Get-CsApplicationMeetingConfiguration"
const setCommand = "Set-CsApplicationMeetingConfiguration"

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()
	if len(c.Steps) > 0 {
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}

	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("configuration differs: %v %+v", err, r.Differences())
	}
	return r
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

	f := fixture{
		Cases: []struct {
			Name     string          `json:"name"`
			Method   string          `json:"method"`
			URL      string          `json:"generated_url"`
			Body     json.RawMessage `json:"body"`
			Status   int             `json:"response_status"`
			Response string          `json:"response_body"`
		}{

			{
				Name:   "application-meeting-list",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationMeetingConfiguration",
				Status: 400,
			},

			{
				Name:   "application-meeting-read",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationMeetingConfiguration/configuration/Global",
				Status: 400,
			},

			{
				Name:   "application-meeting-update",
				Method: "PUT",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationMeetingConfiguration/configuration/Global",
				Body:   json.RawMessage("{\"AllowRemoveParticipantAppIds\":[\"fixture-a\",\"fixture-b\"],\"Identity\":\"Global\"}"),
				Status: 400,
			},
		},
		Models: []struct {
			Name     string          `json:"name"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		}{

			{
				Name:     "omitted",
				Rejected: false,
				Wire:     json.RawMessage("{\"Identity\":\"Global\"}"),
			},

			{
				Name:     "scalar",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":[\"fixture-app\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "replacement",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":[\"fixture-a\",\"fixture-b\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "empty-replacement",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":[],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "null-replacement",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":[],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "operators",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":{\"Add\":[\"fixture-a\"],\"Remove\":[\"fixture-b\"]},\"Identity\":\"Global\"}"),
			},

			{
				Name:     "empty-operators",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":{},\"Identity\":\"Global\"}"),
			},

			{
				Name:     "legacy-object",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowRemoveParticipantAppIds\":[{\"AppId\":\"fixture-app\",\"Enabled\":false}],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "replace-operator-rejected",
				Rejected: true,
			},

			{
				Name:     "numeric-item-rejected",
				Rejected: true,
			},

			{
				Name:     "ps-list-modifier-rejected",
				Rejected: true,
			},
		},
	}

	return f
}

func TestUnitTypedDispatchFromRootClient(t *testing.T) {
	t.Parallel()
	f := pinned(t)

	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, observed := range f.Cases {
			command := getCommand
			if observed.Method == "PUT" {
				command = setCommand
			}
			body := observed.Body
			if string(body) == "null" {
				body = nil
			}
			run(t, testutil.Case{Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: observed.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{discovery(t), wire(observed.Method, strings.TrimPrefix(observed.URL, origin), command, body, observed.Status, observed.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
				switch observed.Name {
				case "application-meeting-list":
					v, r, e := s.List(ctx)
					if e == nil && (v == nil || len(v) != 0) {
						e = errors.New("empty array changed")
					}
					return testutil.Observation{Response: r, Err: e}
				case "application-meeting-read":
					v, r, e := s.GetByIdentity(ctx, "Global")
					if e == nil && (v.Identity != "Global" || v.AppIDs == nil || len(v.AppIDs) != 0) {
						e = errors.New("read projection changed")
					}
					return testutil.Observation{Response: r, Err: e}
				case "application-meeting-update":
					r, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{AppIDs: configuration.ReplaceAppIDs([]string{"fixture-a", "fixture-b"})})
					return testutil.Observation{Response: r, Err: e}
				default:
					return testutil.Observation{Err: errors.New("unexpected fixture")}
				}
			}})
		}
	}
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

func TestUnitUpdateMatchesTypedListNormalizer(t *testing.T) {
	t.Parallel()

	for _, observed := range pinned(t).Models {
		if observed.Rejected {
			continue
		}
		input := configuration.UpdateRequest{AppIDs: change(t, observed.Name)}
		run(t, testutil.Case{Name: observed.Name, Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("PUT", collection+"/configuration/Global", setCommand, observed.Wire, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.UpdateGlobal(ctx, input)
			return testutil.Observation{Response: r, Err: e}
		}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin,
			StatusCode: 400}},
		})
	}
}

func TestUnitFiltersUseRealResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil {
		t.Fatal("invalid real list fixture")
	}
	identities := func(row json.RawMessage) string {
		var value struct {
			Identity string `json:"Identity"`
		}
		if json.Unmarshal(row, &value) != nil {
			t.Fatal("invalid real list identity")
		}
		return value.Identity
	}
	slices.SortStableFunc(rows, func(a, b json.RawMessage) int {
		left, right := strings.EqualFold(identities(a), "Global"), strings.EqualFold(identities(b), "Global")
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	cases := []struct {
		Name, Pattern, OnlyID string
		Empty                 bool
	}{
		{Name: "list"}, {Name: "all", Pattern: "*"}, {Name: "global", Pattern: "Global", OnlyID: "Global"},
		{Name: "missing", Pattern: "__sdk_missing_fixture__", Empty: true},
		{Name: "whitespace", Pattern: " ", Empty: false},
	}
	for _, row := range rows {
		id := identities(row)
		if strings.HasPrefix(id, "Tag:") && !strings.ContainsAny(id, "*?[]`") {
			cases = append(cases, struct {
				Name, Pattern, OnlyID string
				Empty                 bool
			}{Name: "tag-prefix", Pattern: strings.TrimPrefix(id, "Tag:"), OnlyID: id})
			break
		}
	}
	for _, tc := range cases {
		want := make([]json.RawMessage, 0, len(rows))
		for _, row := range rows {
			if !tc.Empty && (tc.OnlyID == "" || strings.EqualFold(identities(row), tc.OnlyID)) {
				want = append(want, row)
			}
		}
		expected, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "real-filter-" + tc.Name,
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/ApplicationMeetingConfiguration", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return testutil.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}

func TestUnitIdentityEscapingAndMutationDenialPreserveMetadata(t *testing.T) {
	t.Parallel()

	var snapshot configuration.Configuration
	if json.Unmarshal([]byte(`{"Identity":"tag:Fixture / Config","AllowRemoveParticipantAppIds":[{"Future":false}]}`), &snapshot) != nil || snapshot.AppIDs != nil || string(snapshot.RawFields["AllowRemoveParticipantAppIds"]) != `[{"Future":false}]` {
		t.Fatal("local compatibility model input lost raw fields")
	}
	run(t, testutil.Case{Name: "escaped", Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("GET", collection+"/configuration/tag:Fixture%20%2F%20Config", getCommand, nil, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.GetByIdentity(ctx, "tag:Fixture / Config")
		if e == nil && (v.AppIDs != nil || string(v.RawFields["AllowRemoveParticipantAppIds"]) != `[{"Future":false}]`) {
			e = errors.New("raw compatibility shape lost")
		}
		return testutil.Observation{Response: r, Err: e}
	}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin,
		StatusCode: 400}},
	})
	for _, status := range []int{400, 403, 429, 503} {
		run(t, testutil.Case{Name: fmt.Sprint("denied-", status), Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("PUT", collection+"/configuration/Global", setCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"code":"FutureDenied","message":"fixture private text"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.UpdateGlobal(ctx, configuration.UpdateRequest{})
			var api *client.APIError
			if !errors.As(e, &api) || r == nil || r.StatusCode() != status || !strings.Contains(string(r.Body()), "fixture private text") {
				return testutil.Observation{Err: errors.New("denial metadata lost")}
			}
			if strings.Contains(fmt.Sprint(e), "fixture private text") {
				return testutil.Observation{Err: errors.New("error exposed payload")}
			}
			return testutil.Observation{Response: r}
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
		run(t, testutil.Case{Name: tc.name, Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire(tc.method, path, command, body, tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
			if tc.method == "PUT" {
				response, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{})
				if !errors.Is(e, configuration.ErrResponse) {
					return testutil.Observation{Err: errors.New("unexpected write status accepted")}
				}
				return testutil.Observation{Response: response}
			}
			if tc.list {
				_, response, e := s.List(ctx)
				if !errors.Is(e, client.ErrDecode) {
					return testutil.Observation{Err: errors.New("malformed list accepted")}
				}
				return testutil.Observation{Response: response}
			}
			_, response, e := s.GetByIdentity(ctx, "Global")
			want := client.ErrDecode
			if tc.status != 200 {
				want = configuration.ErrResponse
			}
			if !errors.Is(e, want) {
				return testutil.Observation{Err: errors.New("malformed item/status accepted")}
			}
			return testutil.Observation{Response: response}
		}})
	}
}

func TestUnitInvalidInputsStayLocal(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "invalid-input", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, configuration.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid read identity accepted")}
			}
			if _, e := s.UpdateByIdentity(ctx, id, configuration.UpdateRequest{}); !errors.Is(e, configuration.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid write identity accepted")}
			}
		}
		if _, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{AppIDs: &configuration.AppIDChange{}}); !errors.Is(e, configuration.ErrInput) {
			return testutil.Observation{Err: errors.New("zero change accepted")}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, configuration.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid wildcard accepted")}
		}

		//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		if _, _, e := s.List(nil); !errors.Is(e, client.ErrNilContext) {
			return testutil.Observation{Err: errors.New("nil context accepted")}
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.GetByIdentity(c, "Global"); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		var missing *configuration.Service
		if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
			return testutil.Observation{Err: errors.New("nil service accepted")}
		}
		return testutil.Observation{}
	}})
}
