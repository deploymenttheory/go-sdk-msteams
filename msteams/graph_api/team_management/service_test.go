package teammanagement_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const (
	groupID = "00000000-0000-4000-8000-000000000002"
	source  = "contracts/endpoints/8.1.0/fixtures/core-team-compositions.json"
)

func runCase(t *testing.T, command string, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "team_management"
	if c.Source == "" {
		c.Source = source
	}
	if c.Mode == "" {
		c.Mode = auth.Delegated
	}
	if c.Name == "" {
		c.Name = "composition-case"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	set := "__AllParameterSets"
	if command == "Get-Team" {
		set = "Identity"
	}
	if len(c.Mappings) == 0 {
		c.Mappings = []contracttest.Mapping{{Command: command, ParameterSet: set}}
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("team composition differs: %+v", result.Differences())
	}
	return result
}

func wire(method, path, body string, status int, response string) contracttest.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: payload}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func fixture(t *testing.T, name string) []contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version  string `json:"baseline_version"`
		Captured bool   `json:"captured"`
		Cases    []struct {
			ID    string `json:"id"`
			Steps []struct {
				Method, Path string
				Query        [][2]string
				Headers      map[string]string
				Body         struct {
					Kind        string
					Value       json.RawMessage
					Base64      string
					ContentType string `json:"content_type"`
				}
				Response struct {
					Status  int
					JSON    json.RawMessage
					Headers map[string]string
				}
			}
		}
	}
	if json.Unmarshal(raw, &document) != nil || document.Version != "8.1.0" || document.Captured {
		t.Fatal("source-derived composition fixture unavailable")
	}
	for _, c := range document.Cases {
		if c.ID != name {
			continue
		}
		var steps []contracttest.Exchange
		for _, step := range c.Steps {
			body := ""
			if step.Body.Kind == "json" {
				body = string(step.Body.Value)
			}
			path := step.Path
			if len(step.Query) != 0 {
				query := make(url.Values)
				for _, pair := range step.Query {
					query.Add(pair[0], pair[1])
				}
				path += "?" + query.Encode()
			}
			exchange := wire(step.Method, path, body, step.Response.Status, string(step.Response.JSON))
			if step.Body.Kind == "bytes" {
				payload, err := base64.StdEncoding.Strict().DecodeString(step.Body.Base64)
				if err != nil {
					t.Fatal("source-derived binary fixture invalid")
				}
				exchange.Request.Binary = payload
				exchange.Request.Headers.Set("Content-Type", step.Body.ContentType)
			}
			for key, value := range step.Headers {
				exchange.Request.Headers.Set(key, value)
			}
			for key, value := range step.Response.Headers {
				exchange.Reply.Headers.Set(key, value)
			}
			steps = append(steps, exchange)
		}
		return steps
	}
	t.Fatal("source-derived composition case missing")
	return nil
}

func observed(result management.Result, err error) contracttest.Observation {
	stages := []string{}
	for _, s := range result.Stages {
		stages = append(stages, string(s.Action)+":"+string(s.State))
	}
	var value any
	if result.Value != nil {
		v := result.Value
		value = map[string]any{"id": v.ID, "name": v.DisplayName, "description": v.Description, "visibility": v.Visibility, "alias": v.MailNickname, "archived": v.IsArchived, "native_name": v.TeamSnapshot.DisplayName, "group_raw": v.GroupSnapshot.RawFields, "team_raw": v.TeamSnapshot.RawFields}
	}
	return contracttest.Observation{Value: map[string]any{"id": result.ID, "value": value, "stages": stages}, Response: result.Response, Err: err}
}

func evidence(t *testing.T, name string, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal("composition evidence directory unavailable")
	}
	defer func() { _ = root.Close() }()
	file, err := root.Create(name)
	if err != nil {
		t.Fatal("composition evidence file unavailable")
	}
	writeErr := contracttest.WriteEvidence(file, revision, results)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("composition evidence write failed")
	}
}

func enriched(stages string) json.RawMessage {
	return json.RawMessage(`{"id":"` + groupID + `","value":{"id":"` + groupID + `","name":"Group name","description":"Group description","visibility":"Private","alias":"fixture","archived":false,"native_name":"Team name","group_raw":{"id":"` + groupID + `","displayName":"Group name","description":"Group description","visibility":"Private","mailNickname":"fixture"},"team_raw":{"id":"` + groupID + `","displayName":"Team name","isArchived":false,"memberSettings":{}}},"stages":` + stages + `}`)
}

func TestUnitGetIdentityEnrichesFromSourceDerivedPairedRead(t *testing.T) {
	t.Parallel()
	r := runCase(t, "Get-Team", contracttest.Case{Name: "identity-enriched", Steps: fixture(t, "team-get-enriched"), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		return observed(sdk.GraphAPI.TeamManagement.GetByID(ctx, groupID))
	}, Want: contracttest.Expected{Status: 200, JSON: enriched(`["read-group:succeeded","read-team:succeeded"]`)}})
	evidence(t, "team-get-service.json", []contracttest.Result{r})
}

type observedPictureSource struct {
	io.Reader
	closes int
}

func (s *observedPictureSource) Close() error { s.closes++; return nil }

func TestUnitDeleteAndPictureUseSourceDerivedPrimitiveMappings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		command, sourceCase, report string
	}{
		{command: "Remove-Team", sourceCase: "team-delete-backing-group", report: "team-delete-service.json"},
		{command: "Set-TeamPicture", sourceCase: "team-picture-raw-bytes", report: "team-picture-service.json"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			t.Parallel()
			var results []contracttest.Result
			for _, denied := range []bool{false, true} {
				steps := fixture(t, tc.sourceCase)
				want := contracttest.Expected{Status: http.StatusNoContent}
				name := tc.sourceCase
				if denied {
					name += "-denied"
					steps[0].Reply = contracttest.Reply{Status: http.StatusForbidden,
						Headers: http.Header{"Content-Type": {"application/json"}, "Request-Id": {"fixture-denial"}},
						Body:    []byte(`{"error":{"code":"Authorization_RequestDenied"}}`)}
					want = contracttest.Expected{Status: http.StatusForbidden, Error: auth.ErrForbidden,
						Headers: http.Header{"Request-Id": {"fixture-denial"}}}
				}
				results = append(results, runCase(t, tc.command, contracttest.Case{Name: name, Steps: steps,
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						if tc.command == "Remove-Team" {
							response, err := sdk.GraphAPI.Teams.DeleteByID(ctx, groupID)
							return contracttest.Observation{Response: response, Err: err}
						}
						// The immutable source fixture deliberately contains synthetic,
						// non-image bytes. This proves transfer, not server image validity.
						stream := &observedPictureSource{Reader: bytes.NewReader(steps[0].Request.Binary)}
						response, err := sdk.GraphAPI.Teams.UploadPicture(ctx, groupID, stream)
						if stream.closes != 1 {
							t.Error("picture source was not closed exactly once")
						}
						return contracttest.Observation{Response: response, Err: err}
					}, Want: want}))
			}
			evidence(t, tc.report, results)
		})
	}
}

func TestUnitUpdateOrdersWritesAndPreservesPartialOutcomes(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, name := range []string{"team-update-partial", "team-update-readback"} {
		t.Run(name, func(t *testing.T) {
			empty, disabled := "", false
			value := management.UpdateRequest{Properties: group.UpdateRequest{Description: &empty}, Settings: team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateChannels: &disabled}}}
			want := contracttest.Expected{Status: 200, JSON: enriched(`["update-group:succeeded","update-team:succeeded","read-group:succeeded","read-team:succeeded"]`)}
			if name == "team-update-partial" {
				rating := team.GiphyModerate
				value.Settings.FunSettings = &team.FunSettings{GiphyContentRating: &rating}
				want = contracttest.Expected{Status: 403, Error: auth.ErrForbidden, JSON: json.RawMessage(`{"id":"` + groupID + `","value":null,"stages":["update-group:succeeded","update-team:unconfirmed"]}`)}
			}
			results = append(results, runCase(t, "Set-Team", contracttest.Case{Name: name, Steps: fixture(t, name), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return observed(sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, value))
			}, Want: want}))
			if value.Settings.GuestSettings != nil || value.Settings.MessagingSettings != nil || value.Settings.DiscoverySettings != nil {
				t.Fatal("composition modified caller settings")
			}
		})
	}
	evidence(t, "team-update-service.json", results)
}

func TestUnitEmptyUpdateUsesActualPinnedSettingsContainers(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-team-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Cases []struct {
			Name string
			Wire json.RawMessage
		}
	}
	if json.Unmarshal(raw, &file) != nil {
		t.Fatal("serializer capture unavailable")
	}
	var body json.RawMessage
	for _, c := range file.Cases {
		if c.Name == "team-settings-omitted" {
			body = c.Wire
		}
	}
	if len(body) == 0 {
		t.Fatal("pinned empty settings missing")
	}
	steps := []contracttest.Exchange{wire("PATCH", "/v1.0/groups/"+groupID, `{}`, 204, ""), wire("PATCH", "/v1.0/teams/"+groupID, string(body), 204, "")}
	steps = append(steps, fixture(t, "team-get-enriched")...)
	runCase(t, "Set-Team", contracttest.Case{Name: "pinned-empty-containers", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		return observed(sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, management.UpdateRequest{}))
	}, Want: contracttest.Expected{Status: 200, JSON: enriched(`["update-group:succeeded","update-team:succeeded","read-group:succeeded","read-team:succeeded"]`)}})
}

func TestUnitFailuresStopLaterStagesWithoutReplayingOrRollingBack(t *testing.T) {
	t.Parallel()
	for _, phase := range []string{"group-write", "team-write", "group-read", "team-read", "wrong-group-id", "wrong-team-id"} {
		t.Run(phase, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("PATCH", "/v1.0/groups/"+groupID, `{}`, 204, ""), wire("PATCH", "/v1.0/teams/"+groupID, `{"memberSettings":{},"guestSettings":{},"messagingSettings":{},"funSettings":{},"discoverySettings":{}}`, 204, "")}
			steps = append(steps, fixture(t, "team-get-enriched")...)
			index := map[string]int{"group-write": 0, "team-write": 1, "group-read": 2, "team-read": 3, "wrong-group-id": 2, "wrong-team-id": 3}[phase]
			want := error(auth.ErrForbidden)
			status := 403
			if phase == "wrong-group-id" || phase == "wrong-team-id" {
				steps[index].Reply.Body = []byte(`{"id":"11111111-1111-4111-8111-111111111111"}`)
				want, status = client.ErrDecode, 200
			} else {
				steps[index].Reply.Status = 403
				steps[index].Reply.Body = []byte(`{"error":{"code":"Forbidden"}}`)
			}
			steps = steps[:index+1]
			stages := []string{"update-group:succeeded", "update-team:succeeded", "read-group:succeeded", "read-team:succeeded"}[:index+1]
			stages[index] = []string{"update-group:unconfirmed", "update-team:unconfirmed", "read-group:unconfirmed", "read-team:unconfirmed"}[index]
			encoded, _ := json.Marshal(stages)
			runCase(t, "Set-Team", contracttest.Case{Name: phase, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return observed(sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, management.UpdateRequest{}))
			}, Want: contracttest.Expected{Status: status, Error: want, JSON: json.RawMessage(`{"id":"` + groupID + `","value":null,"stages":` + string(encoded) + `}`)}})
		})
	}
}

func TestUnitInvalidInputAndCancellationPerformNoIO(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"get", "update"} {
		for _, condition := range []string{"invalid-id", "nil-context", "cancelled", "closed", "nil-service", "wrong-backend"} {
			want := client.ErrRequest
			switch condition {
			case "nil-context":
				want = client.ErrNilContext
			case "cancelled":
				want = context.Canceled
			case "closed", "nil-service":
				want = client.ErrClosed
			case "wrong-backend":
				want = client.ErrDestination
			}
			runCase(t, map[string]string{"get": "Get-Team", "update": "Set-Team"}[method], contracttest.Case{Name: method + "-" + condition, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service, id := sdk.GraphAPI.TeamManagement, groupID
				switch condition {
				case "invalid-id":
					id = "../private"
				case "nil-context":
					if method == "get" {
						//nolint:contextcheck,staticcheck // Deliberately exercise rejection of nil contexts before authentication or I/O.
						_, err := service.GetByID(nil, id)
						return contracttest.Observation{Err: err}
					}
					//nolint:contextcheck,staticcheck // Deliberately exercise rejection of nil contexts before authentication or I/O.
					_, err := service.UpdateByID(nil, id, management.UpdateRequest{})
					return contracttest.Observation{Err: err}
				case "cancelled":
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					ctx = cancelled
				case "closed":
					if err := sdk.Close(); err != nil {
						return contracttest.Observation{Err: err}
					}
				case "nil-service":
					service = nil
				case "wrong-backend":
					service = management.NewService(sdk.TeamsAdminAPI.Scope())
				}
				if method == "get" {
					_, err := service.GetByID(ctx, id)
					return contracttest.Observation{Err: err}
				}
				_, err := service.UpdateByID(ctx, id, management.UpdateRequest{})
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		}
	}
}
