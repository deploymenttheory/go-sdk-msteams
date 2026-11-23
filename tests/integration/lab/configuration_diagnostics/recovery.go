package configurationdiagnosticsacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// probeType is the configuration type the lifecycle creates owned instances
// of. Feedback policy instances carry no licensing or user impact until
// granted, and these are never granted.
const probeType = "TeamsFeedbackPolicy"

// RecoveryHandlers returns journal recovery handlers for the owned
// configuration_diagnostics resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"generic-configuration": configurationCleanup{sdk.TeamsAdminAPI.ConfigurationDiagnostics},
	}
}

type configurationCleanup struct{ service *cd.Service }

// Resolve reconciles a planned create by its unique fixture identity.
func (h configurationCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.ListConfigurations(ctx, probeType, nil)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = entry.Fixture
		}
	}
	if found == "" {
		if _, _, err := h.service.GetConfiguration(ctx, probeType, entry.Fixture, nil); labkit.IsHTTP(err, 404) {
			return "", nil
		} else if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}

// Delete removes the owned instance and observes its absence from both the
// item and collection reads.
func (h configurationCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.RemoveConfiguration(ctx, probeType, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry.ResourceID)
}

func (h configurationCleanup) absent(ctx context.Context, name string) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetConfiguration(ctx, probeType, name, nil)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.ListConfigurations(ctx, probeType, nil)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if value == nil || value.Identity == nil {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(*value.Identity, name) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (configurationCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
