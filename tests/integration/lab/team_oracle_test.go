package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

var errTeamOracle = errors.New("pinned PowerShell team read could not be verified")

const maxTeamOracleBytes = 64 << 10

type teamOracleRequest struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Fixture string `json:"fixture"`
	Owner   string `json:"owner,omitempty"`
}

type teamOracleProjection struct {
	ID             string  `json:"id"`
	DisplayName    *string `json:"display_name"`
	Description    *string `json:"description"`
	Visibility     *string `json:"visibility"`
	Alias          *string `json:"alias"`
	Classification *string `json:"classification"`
	Archived       *bool   `json:"archived"`
	AllowGiphy     *bool   `json:"allow_giphy"`
}

type teamOracleUser struct {
	ID   string  `json:"id"`
	User *string `json:"user"`
	Name *string `json:"name"`
	Role string  `json:"role"`
}

type teamOracleResponse struct {
	Success bool                  `json:"success"`
	Stage   string                `json:"stage"`
	Team    *teamOracleProjection `json:"team"`
	Tenant  *teamOracleProjection `json:"tenant"`
	User    *teamOracleProjection `json:"user"`
	Users   []teamOracleUser      `json:"users"`
}

// Child stdout is private transient data. No command output, errors, identifiers
// or server text are published; only the fixed caller verdict is logged.
type teamOracleBuffer struct {
	data     []byte
	exceeded bool
}

func (b *teamOracleBuffer) Write(p []byte) (int, error) {
	if b.exceeded || len(p) > maxTeamOracleBytes-len(b.data) {
		b.exceeded = true
		return 0, errTeamOracle
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

func teamOracleEnabled(t *testing.T) bool {
	t.Helper()
	switch os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") {
	case "", "0":
		return false
	case "1":
		if !filepath.IsAbs(os.Getenv("MSTEAMS_LAB_MODULE_PATH")) {
			t.Fatal("explicit absolute pinned module path is required")
		}
		if _, err := exec.LookPath("pwsh"); err != nil {
			t.Fatal("PowerShell is required for the explicitly enabled read oracle")
		}
		return true
	default:
		t.Fatal("invalid PowerShell oracle flag")
	}
	return false
}

func decodeTeamOracle(raw []byte, kind string) (teamOracleResponse, error) {
	var value teamOracleResponse
	if len(raw) > maxTeamOracleBytes || (kind != "team" && kind != "users") {
		return value, errTeamOracle
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || !value.Success || value.Stage != "complete" {
		return teamOracleResponse{}, errTeamOracle
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return teamOracleResponse{}, errTeamOracle
	}
	if kind == "team" {
		if value.Team == nil || value.Tenant == nil || value.User == nil || value.Users != nil {
			return teamOracleResponse{}, errTeamOracle
		}
	} else if value.Team != nil || value.Tenant != nil || value.User != nil || value.Users == nil {
		return teamOracleResponse{}, errTeamOracle
	}
	return value, nil
}

func runTeamOracle(t *testing.T, ctx context.Context, request teamOracleRequest) teamOracleResponse {
	t.Helper()
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal("invalid oracle selectors")
	}
	defer clear(input)
	command := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-team.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1",
		"MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	command.Stdin = bytes.NewReader(input)
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.exceeded {
		t.Fatal("pinned PowerShell read oracle failed; no output retained")
	}
	observed, err := decodeTeamOracle(output.data, request.Kind)
	if err != nil {
		t.Fatal("pinned PowerShell did not produce a complete read observation")
	}
	return observed
}

func comparePowerShellTeam(t *testing.T, ctx context.Context, value *management.ManagedTeam, fixture, owner string) {
	t.Helper()
	if !teamOracleEnabled(t) {
		return
	}
	if value == nil || value.TeamSnapshot == nil {
		t.Fatal("complete Go snapshot required before PowerShell comparison")
	}
	want := teamOracleProjection{ID: value.ID, DisplayName: value.DisplayName, Description: value.Description, Alias: value.MailNickname,
		Classification: value.Classification, Archived: value.IsArchived}
	if value.Visibility != nil {
		text := strings.ToLower(string(*value.Visibility))
		want.Visibility = &text
	}
	if value.TeamSnapshot.FunSettings != nil {
		want.AllowGiphy = value.TeamSnapshot.FunSettings.AllowGiphy
	}
	observed := runTeamOracle(t, ctx, teamOracleRequest{Kind: "team", ID: value.ID, Fixture: fixture, Owner: owner})
	for _, actual := range []*teamOracleProjection{observed.Team, observed.Tenant, observed.User} {
		actual.ID = strings.ToLower(actual.ID)
		want.ID = strings.ToLower(want.ID)
		if actual.Visibility != nil {
			text := strings.ToLower(*actual.Visibility)
			actual.Visibility = &text
		}
		if !reflect.DeepEqual(*actual, want) {
			t.Fatal("PowerShell and Go selected team identity/list fields differ")
		}
	}
	t.Log("PowerShell 8.1.0 / Go selected team identity and list observations: passed")
}
