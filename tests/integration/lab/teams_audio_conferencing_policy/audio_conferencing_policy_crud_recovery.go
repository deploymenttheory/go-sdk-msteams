package teamsaudioconferencingpolicyacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type audioPolicyCleanup struct{ service *audio.Service }

func (h audioPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if labkit.SamePolicyName(value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = value.Identity
		}
	}
	if found == "" {
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

func (h audioPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
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
			if labkit.SamePolicyName(value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (audioPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
