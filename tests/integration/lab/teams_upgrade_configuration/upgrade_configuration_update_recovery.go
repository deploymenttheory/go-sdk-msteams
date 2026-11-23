package teamsupgradeconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type upgradeConfigurationCleanup struct{ service *configuration.Service }

func (upgradeConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (upgradeConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h upgradeConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot struct {
		Value *bool `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Value == nil {
		return lab.ErrUnresolved
	}
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if before != nil && before.DownloadTeams != nil && *before.DownloadTeams == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{DownloadTeams: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.DownloadTeams != nil && *v.DownloadTeams == *snapshot.Value, e
	})
}
