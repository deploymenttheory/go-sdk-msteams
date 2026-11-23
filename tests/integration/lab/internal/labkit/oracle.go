package labkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Return fixed stage labels only; PowerShell exception text stays private.
func ReadOracleFailureStage(raw []byte) string {
	var report struct {
		Success bool   `json:"success"`
		Stage   string `json:"failure_stage"`
	}
	if json.Unmarshal(raw, &report) != nil {
		return "unknown"
	}
	if report.Success {
		return ""
	}
	switch report.Stage {
	case "module", "authentication", "read", "projection":
		return report.Stage
	default:
		return "unknown"
	}
}

var ErrOracle = errors.New("pinned PowerShell team read could not be verified")

const MaxOracleBytes = 64 << 10

type TeamOracleRequest struct {
	Kind            string `json:"kind"`
	ID              string `json:"id"`
	Fixture         string `json:"fixture"`
	Owner           string `json:"owner,omitempty"`
	Channel         string `json:"channel,omitempty"`
	OperationID     string `json:"operation_id,omitempty"`
	SharedTeamID    string `json:"shared_team_id,omitempty"`
	APIVersion      string `json:"api_version,omitempty"`
	RequestID       string `json:"request_id,omitempty"`
	OrchestrationID string `json:"orchestration_id,omitempty"`
}

type TeamOracleProjection struct {
	ID             string  `json:"id"`
	DisplayName    *string `json:"display_name"`
	Description    *string `json:"description"`
	Visibility     *string `json:"visibility"`
	Alias          *string `json:"alias"`
	Classification *string `json:"classification"`
	Archived       *bool   `json:"archived"`
	AllowGiphy     *bool   `json:"allow_giphy"`
}

type TeamOracleUser struct {
	ID       string  `json:"id"`
	User     *string `json:"user"`
	Name     *string `json:"name"`
	Role     string  `json:"role"`
	TenantID *string `json:"tenant_id,omitempty"`
}

type TeamOracleResponse struct {
	Success          bool                         `json:"success"`
	Stage            string                       `json:"stage"`
	Team             *TeamOracleProjection        `json:"team"`
	Tenant           *TeamOracleProjection        `json:"tenant"`
	User             *TeamOracleProjection        `json:"user"`
	Users            []TeamOracleUser             `json:"users"`
	Operation        *TeamOracleOperation         `json:"operation"`
	SharedTeams      []SharedTeamObservation      `json:"shared_teams"`
	AssociatedTeams  []SharedTeamObservation      `json:"associated_teams"`
	TargetingPolicy  *TargetingPolicyObservation  `json:"targeting_policy"`
	HierarchyStatus  *HierarchyStatusObservation  `json:"hierarchy_status"`
	BatchStatus      *string                      `json:"batch_status"`
	PrivateMigration *PrivateMigrationObservation `json:"private_migration"`
}

// Child stdout is private transient data. No command output, errors, identifiers
// or server text are published; only the fixed caller verdict is logged.
type OracleBuffer struct {
	// Data is the captured output; callers clear it after decoding.
	Data []byte
	// Exceeded reports that the output limit was reached and data was dropped.
	Exceeded bool
}

func (b *OracleBuffer) Write(p []byte) (int, error) {
	if b.Exceeded || len(p) > MaxOracleBytes-len(b.Data) {
		b.Exceeded = true
		return 0, ErrOracle
	}
	b.Data = append(b.Data, p...)
	return len(p), nil
}

func OracleEnabled(t *testing.T) bool {
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

func DecodeTeamOracle(raw []byte, kind string) (TeamOracleResponse, error) {
	var value TeamOracleResponse
	if len(raw) > MaxOracleBytes || (kind != "team" && kind != "users" && kind != "channel-users" && kind != "operation" && kind != "shared-teams" && kind != "associated-teams" && kind != "shared-users" && kind != "targeting-policy" && kind != "hierarchy-status" && kind != "batch-status" && kind != "private-migration") {
		return value, ErrOracle
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || !value.Success || value.Stage != "complete" {
		return TeamOracleResponse{}, ErrOracle
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return TeamOracleResponse{}, ErrOracle
	}
	if kind == "batch-status" || kind == "private-migration" {
		if value.Team != nil || value.Tenant != nil || value.User != nil || value.Users != nil || value.Operation != nil || value.SharedTeams != nil || value.AssociatedTeams != nil || value.TargetingPolicy != nil || value.HierarchyStatus != nil {
			return TeamOracleResponse{}, ErrOracle
		}
		if kind == "batch-status" && (value.BatchStatus == nil || value.PrivateMigration != nil) || kind == "private-migration" && (value.PrivateMigration == nil || value.BatchStatus != nil) {
			return TeamOracleResponse{}, ErrOracle
		}
		return value, nil
	}
	if value.BatchStatus != nil || value.PrivateMigration != nil {
		return TeamOracleResponse{}, ErrOracle
	}
	if kind == "targeting-policy" || kind == "hierarchy-status" {
		if value.Team != nil || value.Tenant != nil || value.User != nil || value.Users != nil || value.Operation != nil || value.SharedTeams != nil || value.AssociatedTeams != nil {
			return TeamOracleResponse{}, ErrOracle
		}
		if kind == "targeting-policy" && (value.TargetingPolicy == nil || value.TargetingPolicy.Identity == "" || value.HierarchyStatus != nil) || kind == "hierarchy-status" && (value.HierarchyStatus == nil || value.HierarchyStatus.ID == "" || value.TargetingPolicy != nil) {
			return TeamOracleResponse{}, ErrOracle
		}
		return value, nil
	}
	if value.TargetingPolicy != nil || value.HierarchyStatus != nil {
		return TeamOracleResponse{}, ErrOracle
	}
	if kind == "shared-teams" || kind == "associated-teams" {
		if value.Team != nil || value.Tenant != nil || value.User != nil || value.Users != nil || value.Operation != nil {
			return TeamOracleResponse{}, ErrOracle
		}
		if kind == "shared-teams" && (value.SharedTeams == nil || value.AssociatedTeams != nil) || kind == "associated-teams" && (value.AssociatedTeams == nil || value.SharedTeams != nil) {
			return TeamOracleResponse{}, ErrOracle
		}
		return value, nil
	}
	if value.SharedTeams != nil || value.AssociatedTeams != nil {
		return TeamOracleResponse{}, ErrOracle
	}
	if kind == "team" {
		if value.Team == nil || value.Tenant == nil || value.User == nil || value.Users != nil || value.Operation != nil {
			return TeamOracleResponse{}, ErrOracle
		}
	} else if kind == "operation" {
		if value.Operation == nil || value.Operation.ID == "" || value.Team != nil || value.Tenant != nil || value.User != nil || value.Users != nil {
			return TeamOracleResponse{}, ErrOracle
		}
	} else if value.Team != nil || value.Tenant != nil || value.User != nil || value.Users == nil || value.Operation != nil {
		return TeamOracleResponse{}, ErrOracle
	}
	return value, nil
}

func RunTeamOracle(t *testing.T, ctx context.Context, request TeamOracleRequest) TeamOracleResponse {
	t.Helper()
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal("invalid oracle selectors")
	}
	defer clear(input)
	command := OracleCommand(ctx, "read-team.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1",
		"MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdin = bytes.NewReader(input)
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell read oracle failed; no output retained")
	}
	observed, err := DecodeTeamOracle(output.Data, request.Kind)
	if err != nil {
		t.Fatal("pinned PowerShell did not produce a complete read observation")
	}
	return observed
}

type TeamOracleOperation struct {
	ID             string     `json:"id"`
	Type           *string    `json:"type"`
	Status         *string    `json:"status"`
	Created        *time.Time `json:"created"`
	Updated        *time.Time `json:"updated"`
	Attempts       *int64     `json:"attempts"`
	TargetID       *string    `json:"target_id"`
	TargetLocation *string    `json:"target_location"`
}

type SharedTeamObservation struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"display_name"`
	TenantID    *string `json:"tenant_id"`
	IsHost      *bool   `json:"is_host"`
}

type TargetingPolicyObservation struct {
	Identity                           string  `json:"identity"`
	CustomTagsMode                     *string `json:"custom_tags_mode"`
	Description                        *string `json:"description"`
	ManageTagsPermissionMode           *string `json:"manage_tags_permission_mode"`
	ShiftBackedTagsMode                *string `json:"shift_backed_tags_mode"`
	SuggestedPresetTags                *string `json:"suggested_preset_tags"`
	TeamOwnersEditWhoCanManageTagsMode *string `json:"team_owners_edit_who_can_manage_tags_mode"`
}

type HierarchyStatusObservation struct {
	ID              string  `json:"id"`
	Status          *string `json:"status"`
	LastKnownStatus *string `json:"last_known_status"`
	ErrorMessage    *string `json:"error_message"`
	LastUpdatedAt   *string `json:"last_updated_at"`
	LastModifiedBy  *string `json:"last_modified_by"`
	Filename        *string `json:"filename"`
}

type PrivateMigrationObservation struct {
	TenantID *string `json:"tenant_id"`
	Status   *string `json:"status"`
	Details  *string `json:"details"`
}
