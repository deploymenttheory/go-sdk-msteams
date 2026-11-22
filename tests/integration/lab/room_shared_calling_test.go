package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func roomVideoRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: &v.Identity, Settings: map[string]any{"Description": v.Description, "Enabled": v.Enabled, "AreaCode": v.AreaCode, "ReceiveExternalCalls": v.ReceiveExternalCalls, "ReceiveInternalCalls": v.ReceiveInternalCalls, "PlaceExternalCalls": v.PlaceExternalCalls, "PlaceInternalCalls": v.PlaceInternalCalls}, Fields: v.RawFields}, r, nil
	}
}
func TestAccRoomVideoTeleConferencingPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "room-video", "teams_room_video_tele_conferencing_policy", roomVideoRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			out = append(out, &labkit.CompiledScalarView{Identity: new(v.Identity), Settings: map[string]any{"Description": v.Description, "Enabled": v.Enabled, "AreaCode": v.AreaCode, "ReceiveExternalCalls": v.ReceiveExternalCalls, "ReceiveInternalCalls": v.ReceiveInternalCalls, "PlaceExternalCalls": v.PlaceExternalCalls, "PlaceInternalCalls": v.PlaceInternalCalls}, Fields: v.RawFields})
		}
		return out, r, e
	})
}
func sharedCallingRead(sdk *msteams.Client) labkit.CompiledScalarRead {
	return func(ctx context.Context) (*labkit.CompiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "ResourceAccount": v.ResourceAccount, "EmergencyNumbers": v.EmergencyNumbers}, Fields: v.RawFields}, r, nil
	}
}
func TestAccSharedCallingRoutingPolicyReadOnly(t *testing.T) {
	labkit.RunCompiledScalarGlobal(t, "shared-calling", "teams_shared_calling_routing_policy", sharedCallingRead, func(ctx context.Context, sdk *msteams.Client) ([]*labkit.CompiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.ListFiltered(ctx, "Global")
		var out []*labkit.CompiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &labkit.CompiledScalarView{Identity: v.Identity, Settings: map[string]any{"Description": v.Description, "ResourceAccount": v.ResourceAccount, "EmergencyNumbers": v.EmergencyNumbers}, Fields: v.RawFields})
			}
		}
		return out, r, e
	})
}
