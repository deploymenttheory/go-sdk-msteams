package lab_test

import (
	"encoding/json"
	"testing"
)

// Return fixed stage labels only; PowerShell exception text stays private.
func readOracleFailureStage(raw []byte) string {
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

func TestUnitReadOracleFailureStageRejectsDynamicDiagnostics(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ raw, want string }{
		{`{"success":true}`, ""}, {`{"success":false}`, "unknown"},
		{`{"success":false,"failure_stage":"module"}`, "module"},
		{`{"success":false,"failure_stage":"authentication"}`, "authentication"},
		{`{"success":false,"failure_stage":"read"}`, "read"},
		{`{"success":false,"failure_stage":"projection"}`, "projection"},
		{`{"success":false,"failure_stage":"private-token"}`, "unknown"},
		{`private-exception`, "unknown"},
	} {
		if readOracleFailureStage([]byte(tc.raw)) != tc.want {
			t.Fatal("oracle failure exposed dynamic text or lost its fixed stage")
		}
	}
}
