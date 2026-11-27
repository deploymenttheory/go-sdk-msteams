package compliancerecording_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	cr "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/compliance_recording"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// This body is derived from the pinned application serializer, not a successful
// tenant application. It constrains the existing model and sparse request path;
// successful tenant application CRUD remains unqualified.
func TestUnitApplicationSchemaPreservesScopedIdentityAndPresence(t *testing.T) {
	t.Parallel()
	const identity = "Tag:fixture-policy/" + appID
	body := `{"Identity":"` + identity + `","Id":"` + appID + `","ComplianceRecordingPairedApplications":[],"RequiredBeforeMeetingJoin":false,"RequiredDuringMeeting":null,"ConcurrentInvitationCount":4294967295,"FutureField":{"Enabled":false}}`
	run(t, contracttest.Case{Name: "application-model-boundaries", Source: "msteams/teams_admin_api/compliance_recording/schema_boundaries_test.go", Steps: []contracttest.Exchange{
		compiledWire("GET", appPath, "Get-CsTeamsComplianceRecordingApplication", nil, 200, body),
		compiledWire("PATCH", appPath, "Set-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"`+identity+`","RequiredDuringCall":null,"ConcurrentInvitationCount":0,"ComplianceRecordingPairedApplications":[],"Force":"False"}`), 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, _, err := sdk.TeamsAdminAPI.ComplianceRecording.GetApplicationByID(ctx, "Tag:fixture-policy", appID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if value == nil || value.Identity == nil || *value.Identity != identity || value.ID == nil || *value.ID != appID || value.ConcurrentInvitationCount == nil || *value.ConcurrentInvitationCount != ^uint32(0) || value.RequiredBeforeMeetingJoin == nil || *value.RequiredBeforeMeetingJoin || value.ComplianceRecordingPairedApplications == nil || len(value.ComplianceRecordingPairedApplications) != 0 {
			t.Fatal("scoped application model lost identity, explicit false, empty list or unsigned value")
		}
		if !bytes.Equal(value.RawFields["RequiredDuringMeeting"], []byte("null")) || value.RawFields["RequiredDuringCall"] != nil || !bytes.Equal(value.RawFields["FutureField"], []byte(`{"Enabled":false}`)) || !bytes.Equal(value.RawJSON, []byte(body)) {
			t.Fatal("application raw model lost null/omission or unknown data")
		}
		_, response, err := sdk.TeamsAdminAPI.ComplianceRecording.PatchApplicationByIdentity(ctx, identity, cr.ApplicationSettings{RequiredDuringCall: policy.Null[bool](), ConcurrentInvitationCount: policy.Set[uint32](0), ComplianceRecordingPairedApplications: policy.Set([]*cr.PairedApplication{})}, policy.Controls{Force: new(false)})
		return contracttest.Observation{Response: response, Err: err}
	}})
}
