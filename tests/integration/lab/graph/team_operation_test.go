package graphacc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	operation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type operationFixture struct {
	TenantID    string `json:"tenant_id"`
	TeamID      string `json:"team_id"`
	OperationID string `json:"operation_id"`
}

// TestAccTeamOperationReadOnly reads only an explicitly selected existing
// operation. It never initiates an operation or follows the target resource URL.
func TestAccTeamOperationReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("operation lab requires explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid operation lab enablement")
	}
	_ = labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_TEAM_OPERATION_FILE"))
	if err != nil {
		t.Fatal("explicit private operation fixture is required")
	}
	defer clear(raw)
	var fixture operationFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.TeamID) || strings.TrimSpace(fixture.OperationID) == "" {
		t.Fatal("invalid operation fixture binding")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, _, err := sdk.GraphAPI.TeamOperations.GetByID(ctx, fixture.TeamID, fixture.OperationID)
	if err != nil || value == nil || value.ID != fixture.OperationID || value.Status == nil {
		t.Fatal("selected beta team-operation read failed")
	}
	if labkit.OracleEnabled(t) {
		want := operationProjection(*value)
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "operation", ID: fixture.TeamID, OperationID: fixture.OperationID}).Operation
		if actual == nil || !sameOperationProjection(want, *actual) {
			t.Fatal("PowerShell and Go selected operation fields differ")
		}
		t.Log("PowerShell 8.1.0 / Go selected beta team-operation fields: passed")
	}
	t.Log("selected team-operation read observed; authorization, active polling and mutation scenarios remain separately qualified")
}

func operationProjection(value operation.Operation) labkit.TeamOracleOperation {
	result := labkit.TeamOracleOperation{ID: value.ID, Created: value.CreatedDateTime, Updated: value.LastActionDateTime, Attempts: value.AttemptsCount, TargetID: value.TargetResourceID, TargetLocation: value.TargetResourceLocation}
	if value.OperationType != nil {
		v := strings.ToLower(string(*value.OperationType))
		result.Type = &v
	}
	if value.Status != nil {
		v := strings.ToLower(string(*value.Status))
		result.Status = &v
	}
	return result
}

func sameOperationProjection(a, b labkit.TeamOracleOperation) bool {
	// JSON decoding discards monotonic clocks; compare offsets representing the
	// same instant without losing nil versus present timestamp fields.
	for _, pair := range [][2]*time.Time{{a.Created, b.Created}, {a.Updated, b.Updated}} {
		if pair[0] == nil || pair[1] == nil {
			if pair[0] != pair[1] {
				return false
			}
		} else if !pair[0].Equal(*pair[1]) {
			return false
		}
	}
	a.Created, a.Updated, b.Created, b.Updated = nil, nil, nil, nil
	return reflect.DeepEqual(a, b)
}

func TestUnitOperationOracleComparesPresenceAndTimestampInstants(t *testing.T) {
	t.Parallel()
	utc := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	offset := utc.In(time.FixedZone("fixture", 3600))
	zero := int64(0)
	want := labkit.TeamOracleOperation{ID: "fixture", Created: &utc, Attempts: &zero}
	if !sameOperationProjection(want, labkit.TeamOracleOperation{ID: "fixture", Created: &offset, Attempts: &zero}) {
		t.Fatal("equal timestamp instants differed")
	}
	if sameOperationProjection(want, labkit.TeamOracleOperation{ID: "fixture", Created: &utc}) {
		t.Fatal("unknown and zero attempts collapsed")
	}
	if sameOperationProjection(want, labkit.TeamOracleOperation{ID: "fixture", Attempts: &zero}) {
		t.Fatal("unknown and present timestamps collapsed")
	}
}
