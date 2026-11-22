package applicationaccesspolicyacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type policyCleanup struct{ service *policy.Service }

func (h policyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	items, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	count := 0
	for _, item := range items {
		if labkit.SamePolicyName(item.Identity, entry.Fixture) {
			count++
		}
	}
	if count != 1 {
		return "", lab.ErrUnresolved
	}
	return entry.Fixture, nil
}

func (h policyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if entry.Fixture == "" || !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
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
		items, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, item := range items {
			if labkit.SamePolicyName(item.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (policyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
