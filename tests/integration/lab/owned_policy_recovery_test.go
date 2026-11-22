package lab_test

import (
	"context"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

// TestAccRecoverOwnedPolicies uses an explicitly selected, identity-bound journal.
// The handlers verify SDK fixture ownership and independent deletion/absence.
func TestAccRecoverOwnedPolicies(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit owned-policy recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for recovery")
	}
	sdk, binding, _ := labSDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	_, err = journal.Recover(ctx, map[string]lab.Handler{
		"application-access-policy":  policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies},
		"external-access-policy":     externalPolicyCleanup{sdk.TeamsAdminAPI.ExternalAccessPolicies},
		"call-hold-policy":           callHoldPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallHoldPolicies},
		"call-park-policy":           callParkPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallParkPolicies},
		"channels-policy":            channelsPolicyCleanup{sdk.TeamsAdminAPI.TeamsChannelsPolicies},
		"cortana-policy":             cortanaPolicyCleanup{sdk.TeamsAdminAPI.TeamsCortanaPolicies},
		"events-policy":              eventsPolicyCleanup{sdk.TeamsAdminAPI.TeamsEventsPolicies},
		"enhanced-encryption-policy": enhancedEncryptionPolicyCleanup{sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies},
		"calling-policy":             callingPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallingPolicies},
	})
	if err != nil {
		t.Fatal("owned-policy recovery incomplete; private journal retained")
	}
	t.Log("owned-policy recovery verified")
}
