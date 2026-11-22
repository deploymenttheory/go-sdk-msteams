package externalaccesspolicyacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type externalPolicyCleanup struct{ service *external.Service }

func (h externalPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	var found string
	for _, v := range values {
		if v != nil && v.Identity != nil && labkit.SamePolicyName(*v.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *v.Identity
		}
	}
	if found == "" {
		return "", lab.ErrUnresolved
	}
	return found, nil
}

func (h externalPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
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
			if value == nil || value.Identity == nil {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (externalPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
