package callintegration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

// Emit fixed categories only: provider messages and arbitrary codes can contain
// tenant identifiers. HTTP status retains the operation's failure classification.
func logCreateFailure(t *testing.T, response *resty.Response, err error) {
	t.Helper()
	labkit.LogAccessResponse(t, response, err)
	if response == nil {
		return
	}
	t.Log("integrated attendant validation category: " + createFailureCategory(response.Body()))
}

func createFailureCategory(raw []byte) string {
	var body struct{ ErrorDetails []struct{ ErrorCode string } }
	if json.Unmarshal(raw, &body) != nil {
		return "unavailable"
	}
	for _, detail := range body.ErrorDetails {
		code := detail.ErrorCode
		switch {
		case strings.Contains(code, "PriorityMismatch"):
			return "association-priority"
		case strings.Contains(code, "Duplicate"):
			return "duplicate-reference"
		case strings.Contains(code, "InvalidScheduleType"):
			return "schedule-type"
		case strings.Contains(code, "CallFlow") && strings.Contains(code, "NotFound"):
			return "missing-flow"
		case strings.Contains(code, "Schedule") && strings.Contains(code, "NotFound"):
			return "missing-schedule"
		}
	}
	return "other"
}

func TestUnitCreateFailureCategoriesDoNotPublishProviderValues(t *testing.T) {
	for _, tc := range []struct{ body, want string }{
		{`{"errorDetails":[{"ErrorCode":"AutoAttendant_CallHandlingAssociationsContainsDuplicates","Message":"private-tenant-value"}]}`, "duplicate-reference"},
		{`{"errorDetails":[{"ErrorCode":"InvalidScheduleTypeUsedWithAssociation"}]}`, "schedule-type"},
		{`{"errorDetails":[{"ErrorCode":"CallHandlingAssociationTypePriorityMismatch"}]}`, "association-priority"},
		{`{"errorDetails":[{"ErrorCode":"AssociatedCallFlowNotFound"}]}`, "missing-flow"},
		{`{"errorDetails":[{"ErrorCode":"AssociatedScheduleNotFound"}]}`, "missing-schedule"},
		{`{"errorDetails":[{"ErrorCode":"private-tenant-value"}]}`, "other"},
		{`invalid private-value`, "unavailable"},
	} {
		if got := createFailureCategory([]byte(tc.body)); got != tc.want {
			t.Fatal("bounded create diagnostic category changed")
		}
	}
}
