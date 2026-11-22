package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func personalRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"PersonalAttendant": v.PersonalAttendant, "CallScreening": v.CallScreening, "CalendarBookings": v.CalendarBookings, "InboundInternalCalls": v.InboundInternalCalls, "InboundFederatedCalls": v.InboundFederatedCalls, "InboundPSTNCalls": v.InboundPSTNCalls, "AutomaticTranscription": v.AutomaticTranscription, "AutomaticRecording": v.AutomaticRecording}, v.RawFields}, r, nil
	}
}
func TestAccPersonalAttendantPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "personal", "teams_personal_attendant_policy", personalRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"PersonalAttendant": v.PersonalAttendant, "CallScreening": v.CallScreening, "CalendarBookings": v.CalendarBookings, "InboundInternalCalls": v.InboundInternalCalls, "InboundFederatedCalls": v.InboundFederatedCalls, "InboundPSTNCalls": v.InboundPSTNCalls, "AutomaticTranscription": v.AutomaticTranscription, "AutomaticRecording": v.AutomaticRecording}, v.RawFields})
			}
		}
		return out, r, e
	})
}
func recordingRollOutRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"MeetingRecordingOwnership": v.MeetingRecordingOwnership}, v.RawFields}, r, nil
	}
}
func TestAccRecordingRollOutPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "recording-rollout", "teams_recording_roll_out_policy", recordingRollOutRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"MeetingRecordingOwnership": v.MeetingRecordingOwnership}, v.RawFields})
			}
		}
		return out, r, e
	})
}
