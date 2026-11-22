package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func roomVideoRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{&v.Identity, map[string]any{"Description": v.Description, "Enabled": v.Enabled, "AreaCode": v.AreaCode, "ReceiveExternalCalls": v.ReceiveExternalCalls, "ReceiveInternalCalls": v.ReceiveInternalCalls, "PlaceExternalCalls": v.PlaceExternalCalls, "PlaceInternalCalls": v.PlaceInternalCalls}, v.RawFields}, r, nil
	}
}
func TestAccRoomVideoTeleConferencingPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "room-video", "teams_room_video_tele_conferencing_policy", roomVideoRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			out = append(out, &compiledScalarView{new(v.Identity), map[string]any{"Description": v.Description, "Enabled": v.Enabled, "AreaCode": v.AreaCode, "ReceiveExternalCalls": v.ReceiveExternalCalls, "ReceiveInternalCalls": v.ReceiveInternalCalls, "PlaceExternalCalls": v.PlaceExternalCalls, "PlaceInternalCalls": v.PlaceInternalCalls}, v.RawFields})
		}
		return out, r, e
	})
}
func sharedCallingRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "ResourceAccount": v.ResourceAccount, "EmergencyNumbers": v.EmergencyNumbers}, v.RawFields}, r, nil
	}
}
func TestAccSharedCallingRoutingPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "shared-calling", "teams_shared_calling_routing_policy", sharedCallingRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "ResourceAccount": v.ResourceAccount, "EmergencyNumbers": v.EmergencyNumbers}, v.RawFields})
			}
		}
		return out, r, e
	})
}
