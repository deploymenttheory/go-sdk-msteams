package teamsmultitenantorganizationconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_multi_tenant_organization_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type multiTenantConfigurationCleanup struct{ service *configuration.Service }

func (multiTenantConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (multiTenantConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h multiTenantConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot struct {
		Value *string `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Value == nil || (*snapshot.Value != "Enabled" && *snapshot.Value != "Disabled") {
		return lab.ErrUnresolved
	}
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if before != nil && before.CopilotFromHomeTenant != nil && *before.CopilotFromHomeTenant == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{CopilotFromHomeTenant: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.CopilotFromHomeTenant != nil && *v.CopilotFromHomeTenant == *snapshot.Value, e
	})
}
