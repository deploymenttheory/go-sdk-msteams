package callintegration

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
)

// Both association types must cross the evidenced server draft boundary. The
// public module wrapper uses the normalized reply, including provider Priority;
// locally constructing the Holiday member bypasses that normalization.
func draftAssociations(ctx context.Context, service *aa.Service, weeklyID, fixedID, flowID string) ([]aa.CallHandlingAssociation, error) {
	values := []aa.CallHandlingAssociation{}
	for _, request := range []aa.CallHandlingAssociation{
		{Type: ptr(aa.AssociationAfterHours), ScheduleID: &weeklyID, CallFlowID: &flowID},
		{Type: ptr(aa.AssociationHoliday), ScheduleID: &fixedID, CallFlowID: &flowID},
	} {
		value, _, err := service.DraftCallHandlingAssociation(ctx, request)
		if err != nil {
			return values, err
		}
		values = append(values, *value)
	}
	return values, nil
}

func TestUnitBothAssociationsPreserveServerDraftPriority(t *testing.T) {
	t.Parallel()
	const fixedID = "00000000-0000-4000-8000-000000000004"
	after := `{"Type":"AfterHours","ScheduleId":"` + scheduleID + `","CallFlowId":"` + flowID + `","Enabled":true}`
	holiday := `{"Type":"Holiday","ScheduleId":"` + fixedID + `","CallFlowId":"` + flowID + `","Enabled":true}`
	// Priorities here are synthetic normalized DTO fields, not a claim about the
	// provider's holiday ordering rules; the caller must preserve either value.
	run(t, "server-normalized-associations", []contracttest.Exchange{
		exchange("POST", "/Teams.VoiceApps/auto-attendants/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation", after, 200, `{"CallHandlingAssociation":{"Type":"AfterHours","ScheduleId":"`+scheduleID+`","CallFlowId":"`+flowID+`","Enabled":true,"Priority":1}}`),
		exchange("POST", "/Teams.VoiceApps/auto-attendants/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation", holiday, 200, `{"CallHandlingAssociation":{"Type":"Holiday","ScheduleId":"`+fixedID+`","CallFlowId":"`+flowID+`","Enabled":true,"Priority":0}}`),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		values, err := draftAssociations(ctx, sdk.TeamsAdminAPI.AutoAttendants, scheduleID, fixedID, flowID)
		if err == nil && (len(values) != 2 || values[0].Priority == nil || *values[0].Priority != 1 || values[1].Priority == nil || *values[1].Priority != 0) {
			t.Fatal("composition dropped provider-normalized association priorities")
		}
		return contracttest.Observation{Err: err}
	})
}
