package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

type migrationProjection struct {
	ID    string `json:"id"`
	State string `json:"state"`
}
type migrationObservation struct {
	Success bool                  `json:"success"`
	Tenant  string                `json:"tenant_id"`
	Jobs    []migrationProjection `json:"jobs"`
}

func decodeMigrationOracle(raw []byte, tenant string) (migrationObservation, error) {
	var value migrationObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) {
		return value, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&value) != nil || !value.Success || !strings.EqualFold(value.Tenant, tenant) || value.Jobs == nil {
		return migrationObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return migrationObservation{}, labkit.ErrOracle
	}
	seen := map[string]bool{}
	for _, job := range value.Jobs {
		id := strings.ToLower(job.ID)
		if !labkit.DirectoryID.MatchString(id) || job.State == "" || seen[id] {
			return migrationObservation{}, labkit.ErrOracle
		}
		seen[id] = true
	}
	return value, nil
}

func migrationOracle(t *testing.T, ctx context.Context, tenant string) migrationObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid migration oracle tenant")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-assignment-migration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned migration read failed; no output retained")
	}
	v, err := decodeMigrationOracle(output.Data, tenant)
	if err != nil {
		t.Fatal("pinned migration read projection unavailable")
	}
	return v
}

// TestAccAssignmentMigrationReadOnly lists status and reads returned job IDs.
// It never submits migrations/cleanup or changes effective assignments.
func TestAccAssignmentMigrationReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	jobs, response, err := sdk.TeamsAdminAPI.AssignmentMigrations.List(ctx)
	labkit.WriteResponseFixture(t, "assignment_migration", "list", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("migration status list failed (HTTP %d); no response payload retained", status)
	}
	projected := make([]migrationProjection, 0, len(jobs))
	for _, job := range jobs {
		if job == nil || job.MigrationEventID == nil || !labkit.DirectoryID.MatchString(*job.MigrationEventID) {
			t.Fatal("migration status has no stable job identity")
		}
		state, ok := job.State.Text()
		if !ok {
			number, numeric := job.State.Number()
			if !numeric {
				t.Fatal("migration status omitted its state")
			}
			names := []resource.StateName{resource.NotStarted, resource.PopulationInProgress, resource.GPACreationInProgress, resource.WaitingForGPAProcessed, resource.CleanUpInProgress, resource.Completed, resource.Failed, resource.ConnectionLost}
			state = resource.StateName(strconv.FormatInt(int64(number), 10))
			if number >= 0 && int64(number) < int64(len(names)) {
				state = names[int(number)]
			}
		}
		projected = append(projected, migrationProjection{ID: strings.ToLower(*job.MigrationEventID), State: string(state)})
		value, fixtureResponse, err := sdk.TeamsAdminAPI.AssignmentMigrations.GetByID(ctx, *job.MigrationEventID)
		labkit.WriteResponseFixture(t, "assignment_migration", "get_by_id", fixtureResponse, nil)
		if err != nil || value == nil || value.MigrationEventID == nil || !strings.EqualFold(*value.MigrationEventID, *job.MigrationEventID) {
			t.Fatal("migration read did not return the selected job")
		}
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("migration tenant binding required")
		}
		actual := migrationOracle(t, ctx, tenant).Jobs
		for i := range actual {
			actual[i].ID = strings.ToLower(actual[i].ID)
		}
		sort.Slice(projected, func(i, j int) bool { return projected[i].ID < projected[j].ID })
		sort.Slice(actual, func(i, j int) bool { return actual[i].ID < actual[j].ID })
		if !reflect.DeepEqual(projected, actual) {
			t.Fatal("Go and pinned migration status projections differ")
		}
	}
	t.Log("status read observations verified; submission/cleanup and effective assignments remain separately qualified")
}

func TestUnitMigrationOracleRejectsWrongTenantAndMalformedProjection(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	for _, raw := range []string{`{"success":false}`, `{"success":true,"tenant_id":"00000000-0000-4000-8000-000000000004","jobs":[]}`, `{"success":true,"tenant_id":"` + tenant + `"}`, `{"success":true,"tenant_id":"` + tenant + `","jobs":null}`, `{"success":true,"tenant_id":"` + tenant + `","jobs":[],"unknown":false}`, `{"success":true,"tenant_id":"` + tenant + `","jobs":[]} {}`} {
		if _, err := decodeMigrationOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid migration oracle projection accepted")
		}
	}
	if _, err := decodeMigrationOracle([]byte(`{"success":true,"tenant_id":"`+tenant+`","jobs":[]}`), tenant); err != nil {
		t.Fatal("empty status projection rejected")
	}
}
