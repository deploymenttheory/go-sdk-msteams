package compliancerecording_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	cr "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/compliance_recording"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_CreatePolicyByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.ComplianceRecording.CreatePolicyByIdentity(context.Background(), "Fixture", cr.PolicySettings{Enabled: policy.Set(false), Description: policy.Set("Recorded lines")}, policy.Controls{})
}

func ExampleService_CreateApplicationByID() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// The Id is the recorder's resource account object ID; the service rejects
	// any other GUID. Paired applications and ConcurrentInvitationCount above
	// one are mutually exclusive.
	paired, err := cr.NewPairedApplication("00000000-0000-4000-8000-000000000002")
	if err != nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.ComplianceRecording.CreateApplicationByID(context.Background(), "Tag:Fixture", "00000000-0000-4000-8000-000000000001",
		cr.ApplicationSettings{RequiredDuringCall: policy.Set(true), ComplianceRecordingPairedApplications: policy.Set([]*cr.PairedApplication{paired})}, policy.Controls{})
}

func ExampleService_UpdatePolicyByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// The list edit applies to the policy read before the PATCH, like
	// -ComplianceRecordingApplications @{Remove=...}.
	_, _, _ = sdk.TeamsAdminAPI.ComplianceRecording.UpdatePolicyByIdentity(context.Background(), "Tag:Fixture", cr.PolicyUpdate{
		Applications: &cr.ListEdit[cr.Application]{Remove: []string{"00000000-0000-4000-8000-000000000001"}},
	})
}

func ExampleService_Assign() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.ComplianceRecording.Assign(context.Background(), cr.Assignment{Group: &cr.GroupGrant{ID: "fixture-group", PolicyName: "Fixture", Rank: new(int32(1))}})
}
