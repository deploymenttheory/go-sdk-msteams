package teamstenantabuseconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type tenantAbuseConfigurationCleanup struct{ service *resource.Service }

func (tenantAbuseConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (tenantAbuseConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h tenantAbuseConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var original struct {
		Value *uint64 `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&original) != nil || d.Decode(new(any)) != io.EOF || original.Value == nil {
		return lab.ErrUnresolved
	}
	current, _, e := h.service.GetByIdentity(ctx, "Global")
	if e != nil {
		return e
	}
	if current != nil && current.LowSeatLimit != nil && *current.LowSeatLimit == *original.Value {
		return nil
	}
	if _, _, e := h.service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(*original.Value)}}); e != nil {
		return e
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.LowSeatLimit != nil && *v.LowSeatLimit == *original.Value, e
	})
}
