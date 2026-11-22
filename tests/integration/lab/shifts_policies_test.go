package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func shiftsRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsShiftsPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"ShiftNoticeFrequency": v.ShiftNoticeFrequency, "ShiftNoticeMessageType": v.ShiftNoticeMessageType, "ShiftNoticeMessageCustom": v.ShiftNoticeMessageCustom, "AccessType": v.AccessType, "AccessGracePeriodMinutes": v.AccessGracePeriodMinutes, "EnableScheduleOwnerPermissions": v.EnableScheduleOwnerPermissions}, v.RawFields}, r, nil
	}
}
func TestAccShiftsPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "shifts", "teams_shifts_policy", shiftsRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsShiftsPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"ShiftNoticeFrequency": v.ShiftNoticeFrequency, "ShiftNoticeMessageType": v.ShiftNoticeMessageType, "ShiftNoticeMessageCustom": v.ShiftNoticeMessageCustom, "AccessType": v.AccessType, "AccessGracePeriodMinutes": v.AccessGracePeriodMinutes, "EnableScheduleOwnerPermissions": v.EnableScheduleOwnerPermissions}, v.RawFields})
			}
		}
		return out, r, e
	})
}
func shiftsAppRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, v.RawFields}, r, nil
	}
}
func TestAccShiftsAppPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "shifts-app", "teams_shifts_app_policy", shiftsAppRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			out = append(out, &compiledScalarView{v.Identity, map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, v.RawFields})
		}
		return out, r, e
	})
}
