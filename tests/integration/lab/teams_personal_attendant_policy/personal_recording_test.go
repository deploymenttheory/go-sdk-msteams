package teamspersonalattendantpolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func personalRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"PersonalAttendant": v.PersonalAttendant, "CallScreening": v.CallScreening, "CalendarBookings": v.CalendarBookings, "InboundInternalCalls": v.InboundInternalCalls, "InboundFederatedCalls": v.InboundFederatedCalls, "InboundPSTNCalls": v.InboundPSTNCalls, "AutomaticTranscription": v.AutomaticTranscription, "AutomaticRecording": v.AutomaticRecording}, Fields: v.RawFields}, r, nil
	}
}

func TestAccPersonalAttendantPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "personal", "teams_personal_attendant_policy", personalRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"PersonalAttendant": v.PersonalAttendant, "CallScreening": v.CallScreening, "CalendarBookings": v.CalendarBookings, "InboundInternalCalls": v.InboundInternalCalls, "InboundFederatedCalls": v.InboundFederatedCalls, "InboundPSTNCalls": v.InboundPSTNCalls, "AutomaticTranscription": v.AutomaticTranscription, "AutomaticRecording": v.AutomaticRecording}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
