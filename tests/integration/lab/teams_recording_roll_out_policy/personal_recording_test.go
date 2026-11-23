package teamsrecordingrolloutpolicyacc

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

func recordingRollOutRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"MeetingRecordingOwnership": v.MeetingRecordingOwnership}, Fields: v.RawFields}, r, nil
	}
}

func TestAccRecordingRollOutPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "recording-rollout", "teams_recording_roll_out_policy", recordingRollOutRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"MeetingRecordingOwnership": v.MeetingRecordingOwnership}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
