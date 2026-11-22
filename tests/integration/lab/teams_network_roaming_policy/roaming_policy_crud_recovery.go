package teamsnetworkroamingpolicyacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	roaming "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_network_roaming_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type roamingPolicyCleanup struct{ service *roaming.Service }

func (h roamingPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value.Identity == "" {
			return "", lab.ErrUnresolved
		}
		if labkit.SamePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
		// An empty list match alone cannot establish absence after a lost POST.
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if labkit.IsHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}

func (h roamingPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
	if err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if value.Identity == "" {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (roamingPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
